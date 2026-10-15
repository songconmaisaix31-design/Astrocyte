package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
)

const (
	protocolVersion = "2025-11-25"
	maxMessageBytes = 16 << 20
)

// Server is a minimal stdio MCP server: initialize, ping, tools/list and
// tools/call over newline-delimited JSON-RPC.
type Server struct {
	name    string
	version string
	tools   []Tool
	byName  map[string]Tool
}

func NewServer(name, version string) *Server {
	return &Server{name: name, version: version, byName: map[string]Tool{}}
}

func (s *Server) RegisterTool(t Tool) {
	s.tools = append(s.tools, t)
	s.byName[t.Name] = t
}

// Serve reads newline-delimited JSON-RPC requests from in and writes one
// newline-delimited response per request until EOF or a write error.
func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), maxMessageBytes)
	w := bufio.NewWriter(out)
	initialized := false
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		resp := s.handle(ctx, line, &initialized)
		if resp == nil {
			continue
		}
		b, err := json.Marshal(resp)
		if err != nil {
			return err
		}
		if _, err := w.Write(b); err != nil {
			return err
		}
		if err := w.WriteByte('\n'); err != nil {
			return err
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, line []byte, initialized *bool) *response {
	if !json.Valid(line) {
		return errorResponse(nil, codeParseError, "Parse error")
	}
	var req request
	if err := json.Unmarshal(line, &req); err != nil {
		return errorResponse(nil, codeInvalidRequest, "Invalid request")
	}
	if req.JSONRPC != "2.0" || req.Method == "" || !validID(req.ID) {
		return errorResponse(nil, codeInvalidRequest, "Invalid request")
	}
	if isNotification(req.ID) {
		return nil
	}
	switch req.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string                     `json:"protocolVersion"`
			Capabilities    map[string]json.RawMessage `json:"capabilities"`
			ClientInfo      struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"clientInfo"`
		}
		if *initialized || json.Unmarshal(req.Params, &params) != nil || params.ProtocolVersion == "" || params.Capabilities == nil || params.ClientInfo.Name == "" || params.ClientInfo.Version == "" {
			return errorResponse(req.ID, codeInvalidParams, "Invalid initialize params")
		}
		version := protocolVersion
		switch params.ProtocolVersion {
		case "2024-11-05", "2025-03-26", "2025-06-18", protocolVersion:
			version = params.ProtocolVersion
		}
		*initialized = true
		return okResponse(req.ID, s.initializeResult(version))
	case "ping":
		return okResponse(req.ID, map[string]any{})
	case "tools/list":
		if !*initialized {
			return errorResponse(req.ID, codeInvalidRequest, "Initialize first")
		}
		return okResponse(req.ID, map[string]any{"tools": s.toolList()})
	case "tools/call":
		if !*initialized {
			return errorResponse(req.ID, codeInvalidRequest, "Initialize first")
		}
		return s.call(req.ID, ctx, req.Params)
	default:
		return errorResponse(req.ID, codeMethodNotFound, "Method not found")
	}
}

func (s *Server) initializeResult(version string) map[string]any {
	return map[string]any{
		"protocolVersion": version,
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": s.name, "version": s.version},
	}
}

type toolListEntry struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

func (s *Server) toolList() []toolListEntry {
	out := make([]toolListEntry, 0, len(s.tools))
	for _, t := range s.tools {
		out = append(out, toolListEntry{Name: t.Name, Description: t.Description, InputSchema: t.InputSchema})
	}
	return out
}

type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) call(id json.RawMessage, ctx context.Context, params json.RawMessage) *response {
	var p callParams
	if err := json.Unmarshal(params, &p); err != nil {
		return errorResponse(id, codeInvalidParams, "Invalid params")
	}
	if len(p.Arguments) == 0 {
		p.Arguments = json.RawMessage(`{}`)
	}
	p.Arguments = bytes.TrimSpace(p.Arguments)
	if len(p.Arguments) == 0 || p.Arguments[0] != '{' {
		return errorResponse(id, codeInvalidParams, "Arguments must be an object")
	}
	t, ok := s.byName[p.Name]
	if !ok {
		return errorResponse(id, codeInvalidParams, "Unknown tool")
	}
	result, err := t.Handler(ctx, p.Arguments)
	if err != nil {
		if delivery, ok := result.(interface{ MCPDeliveryState() string }); ok && delivery.MCPDeliveryState() != "" {
			// Text is authoritative for clients that discard structuredContent
			// when isError is true. Never drop a committed receipt on this path.
			payload, marshalErr := json.Marshal(result)
			var receipt map[string]any
			if marshalErr == nil && json.Unmarshal(payload, &receipt) == nil && receipt != nil {
				receipt["error"] = "Tool operation failed"
				payload, marshalErr = json.Marshal(receipt)
				if marshalErr == nil {
					return okResponse(id, toolResult{Content: []contentItem{{Type: "text", Text: string(payload)}}, IsError: true, StructuredContent: receipt})
				}
			}
		}
		// A committed receipt takes precedence even when its cause wraps an
		// argument error. Classification must never erase delivery uncertainty.
		var invalid *InvalidParamsError
		if errors.As(err, &invalid) {
			return errorResponse(id, codeInvalidParams, "Invalid tool arguments")
		}
		return okResponse(id, toolResultError(err))
	}
	text, err := json.Marshal(result)
	if err != nil {
		return errorResponse(id, codeInternalError, "Failed to marshal tool result")
	}
	return okResponse(id, toolResultText(string(text)))
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolResult struct {
	Content           []contentItem  `json:"content"`
	IsError           bool           `json:"isError"`
	StructuredContent map[string]any `json:"structuredContent,omitempty"`
}

func toolResultText(text string) toolResult {
	return toolResult{Content: []contentItem{{Type: "text", Text: text}}, IsError: false}
}

func toolResultError(err error) toolResult {
	return toolResult{Content: []contentItem{{Type: "text", Text: "Tool operation failed"}}, IsError: true}
}
