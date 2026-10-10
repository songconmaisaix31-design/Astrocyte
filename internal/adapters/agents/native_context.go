package agents

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type nativeContextItem struct {
	Type    string          `json:"type"`
	Role    string          `json:"role"`
	Text    string          `json:"text"`
	Content json.RawMessage `json:"content"`
}

type contextProjection struct {
	events []domain.NativeEvent
	bytes  int
}

func (c *contextProjection) add(role, text string) error {
	if text == "" {
		return nil
	}
	if len(c.events) >= 256 || c.bytes+len(text) > 128*1024 {
		return nativeError(apierrors.BudgetExhausted, "native context exceeds 256 messages or 128 KiB; no partial history returned")
	}
	c.bytes += len(text)
	kind := "text"
	if role == "user" {
		kind = "user_text"
	}
	c.events = append(c.events, domain.NativeEvent{Sequence: len(c.events) + 1, Kind: kind, Text: text})
	return nil
}
func (c *contextProjection) item(item nativeContextItem) error {
	role := item.Role
	if item.Type == "agentMessage" {
		role = "assistant"
	}
	if item.Type == "userMessage" {
		role = "user"
	}
	if role != "user" && role != "assistant" {
		return nil
	}
	if item.Text != "" {
		return c.add(role, item.Text)
	}
	var plain string
	if json.Unmarshal(item.Content, &plain) == nil {
		return c.add(role, plain)
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(item.Content, &parts); err != nil {
		return nativeError(apierrors.EvidenceMissing, "native message content schema is unavailable")
	}
	for _, part := range parts {
		if part.Type == "text" {
			if err := c.add(role, part.Text); err != nil {
				return err
			}
		}
	}
	return nil
}

// Reads only the exact owned original native ID. No global thread/list, disk
// scan, inference turn, read-state mutation or implicit process start occurs.
func (n *Native) readOwnedContext(ctx context.Context, p *nativeProcess, s domain.NativeSession) ([]domain.NativeEvent, error) {
	if p.observe().StopConfirmed {
		return nil, nativeError(apierrors.EvidenceMissing, "native process is stopped; original history requires explicit resume or a registered history root")
	}
	projection := contextProjection{events: []domain.NativeEvent{}}
	if n.id == "codex" {
		metadata, err := p.call(ctx, "thread/read", map[string]any{"threadId": s.NativeID, "includeTurns": false}, false)
		if err != nil {
			return nil, err
		}
		var identity struct {
			Thread struct {
				ID  string `json:"id"`
				Cwd string `json:"cwd"`
			} `json:"thread"`
		}
		if json.Unmarshal(metadata["result"], &identity) != nil || identity.Thread.ID != s.NativeID || filepath.Clean(identity.Thread.Cwd) != filepath.Clean(p.cmd.Dir) {
			return nil, nativeError(apierrors.ScopeDenied, "native history identity or project differs")
		}
		cursor := ""
		seen := map[string]bool{}
		for page := 0; page < 8; page++ {
			params := map[string]any{"threadId": s.NativeID, "limit": 32, "sortDirection": "asc"}
			if cursor != "" {
				params["cursor"] = cursor
			}
			response, err := p.call(ctx, "thread/items/list", params, false)
			if err != nil {
				return nil, err
			}
			var result struct {
				Data []struct {
					Item nativeContextItem `json:"item"`
				} `json:"data"`
				NextCursor *string `json:"nextCursor"`
			}
			if json.Unmarshal(response["result"], &result) != nil || result.Data == nil {
				return nil, nativeError(apierrors.EvidenceMissing, "native history item schema is unavailable")
			}
			for _, entry := range result.Data {
				if err := projection.item(entry.Item); err != nil {
					return nil, err
				}
			}
			if result.NextCursor == nil || *result.NextCursor == "" {
				return projection.events, nil
			}
			cursor = *result.NextCursor
			if seen[cursor] {
				return nil, nativeError(apierrors.EvidenceMissing, "native history cursor repeated; completeness unknown")
			}
			seen[cursor] = true
		}
		return nil, nativeError(apierrors.BudgetExhausted, "native history exceeds eight pages; no partial history returned")
	}
	if n.id == "pi" {
		response, err := p.call(ctx, "get_messages", map[string]any{}, true)
		if err != nil {
			return nil, err
		}
		var result struct {
			Messages []nativeContextItem `json:"messages"`
		}
		if json.Unmarshal(response["data"], &result) != nil || result.Messages == nil {
			return nil, nativeError(apierrors.EvidenceMissing, "native conversation schema is unavailable")
		}
		if len(result.Messages) > 256 {
			return nil, nativeError(apierrors.BudgetExhausted, "native conversation exceeds message bound")
		}
		for _, message := range result.Messages {
			if err := projection.item(message); err != nil {
				return nil, err
			}
		}
		return projection.events, nil
	}
	return nil, nativeError(apierrors.UnsupportedCapability, "complete native history requires a verified protocol or explicit history root; observe supplies current attachment events only")
}
