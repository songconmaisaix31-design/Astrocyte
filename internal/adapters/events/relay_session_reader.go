package events

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/olivere/ndjson"
)

func validateSessionSelection(b RelaySessionBinding) error {
	if !b.Owned || !b.Authorized {
		return sessionFailure("ownership_or_permission_missing")
	}
	for _, id := range []string{b.TaskID, b.TerminalID, b.SessionID, b.WorktreeID} {
		u, err := uuid.Parse(id)
		if err != nil || u == uuid.Nil || u.String() != id {
			return sessionFailure("binding_invalid")
		}
	}
	if b.ProjectID == "" || b.NativeID == "" || !filepath.IsAbs(b.Path) || !filepath.IsAbs(b.WorkspaceRoot) {
		return sessionFailure("binding_unverifiable")
	}
	workspace, err := unredirectedReplayPath(b.WorkspaceRoot)
	if err != nil {
		return sessionFailure("workspace_unverifiable")
	}
	info, err := os.Stat(workspace)
	if err != nil || !info.IsDir() {
		return sessionFailure("workspace_unverifiable")
	}
	selected, err := unredirectedReplayPath(b.Path)
	if err != nil {
		return sessionFailure("source_unavailable_or_redirected")
	}
	for _, root := range b.ApprovedRoots {
		if !filepath.IsAbs(root) {
			return sessionFailure("approved_root_invalid")
		}
		approved, e := unredirectedReplayPath(root)
		if e != nil {
			return sessionFailure("approved_root_redirected")
		}
		info, e := os.Stat(approved)
		if e != nil || !info.IsDir() {
			return sessionFailure("approved_root_unavailable")
		}
		rel, e := filepath.Rel(approved, selected)
		if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return nil
		}
	}
	return sessionFailure("source_outside_approved_roots")
}

type sessionContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r sessionContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

// ReadRelaySession 使用 ndjson 库读取选定文件的固定范围，未完成尾行不推进游标。
func ReadRelaySession(ctx context.Context, b RelaySessionBinding, prior RelaySessionCursor) (RelaySessionBatch, error) {
	c := copySessionCursor(prior)
	batch := RelaySessionBatch{Next: c, Items: []RelaySessionItem{}}
	fail := func(code string) (RelaySessionBatch, error) {
		markSessionError(&batch.Next, code, true)
		if batch.SnapshotBytes >= batch.Next.Offset {
			batch.Remaining = batch.SnapshotBytes - batch.Next.Offset
		}
		batch.SessionComplete = false
		return batch, sessionFailure(code)
	}
	if c.Blocked != "" {
		return batch, sessionFailure("capture_blocked_requires_reconciliation")
	}
	if err := validateSessionSelection(b); err != nil {
		markSessionError(&batch.Next, "selection_denied", true)
		return batch, err
	}
	key := sessionBindingID(b)
	if c.Version == 0 && (c.Offset != 0 || c.FileIdentity != "" || len(c.Seen) != 0 || c.Counts != (RelaySessionCounts{})) {
		return fail("checkpoint_invalid")
	}
	if c.Version != 0 && (c.Version != 1 || c.BindingID != key || c.Offset < 0 || c.FileIdentity == "") {
		return fail("checkpoint_binding_mismatch")
	}
	batch.Next.Version, batch.Next.BindingID = 1, key
	f, err := os.Open(b.Path)
	if err != nil {
		return fail("source_open_failed")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return fail("source_not_regular")
	}
	identity, err := relaySessionFileIdentity(f, info)
	if err != nil {
		return fail("source_identity_unverifiable_or_aliased")
	}
	if c.FileIdentity != "" && c.FileIdentity != identity {
		return fail("source_replaced")
	}
	batch.Next.FileIdentity = identity
	batch.SnapshotBytes = info.Size()
	if info.Size() >= int64(math.MaxInt) {
		return fail("source_extent_unsupported")
	}
	if info.Size() < c.Offset {
		return fail("source_truncated")
	}
	if c.Offset > 0 {
		var boundary [1]byte
		if _, e := f.ReadAt(boundary[:], c.Offset-1); e != nil || boundary[0] != '\n' {
			return fail("checkpoint_not_line_boundary")
		}
	}
	// 库负责分行，缓冲上限使用文件范围，避免固定行长限制。
	headReader := ndjson.NewReaderSize(sessionContextReader{ctx, io.NewSectionReader(f, 0, info.Size())}, int(info.Size())+1)
	if headReader.Next() {
		header := headReader.Bytes()
		_, complete, boundaryErr := sessionLineEnd(f, 0, int64(len(header)), info.Size())
		if boundaryErr != nil {
			return fail("header_read_failed")
		}
		if !complete {
			batch.TailBytes, batch.Remaining = info.Size(), info.Size()
			if b.ArchiveFinalized {
				return fail("closed_incomplete_tail")
			}
			return batch, nil
		}
		if b.Format == "codex" || b.Format == "claude_code" {
			matches, decodeErr := sessionHeaderMatches(b, header)
			if decodeErr != nil {
				return fail("malformed_json")
			}
			if !matches {
				return fail("session_header_unverifiable")
			}
			batch.HeaderVerified = true
		}
	} else if headReader.Err() != nil {
		return fail("header_read_failed")
	}
	reader := ndjson.NewReaderSize(sessionContextReader{ctx, io.NewSectionReader(f, c.Offset, info.Size()-c.Offset)}, int(info.Size())+1)
	for reader.Next() {
		if ctx.Err() != nil {
			return fail("capture_cancelled")
		}
		line := reader.Bytes()
		start := batch.Next.Offset
		end, complete, e := sessionLineEnd(f, start, int64(len(line)), info.Size())
		if e != nil {
			return fail("source_read_failed")
		}
		if !complete {
			batch.TailBytes = info.Size() - start
			break
		}
		batch.Next.Counts.Lines++
		var decoded json.RawMessage
		if reader.Decode(&decoded) != nil {
			return fail("malformed_json")
		}
		batch.Next.Offset = end
		if !batch.HeaderVerified {
			batch.Next.Counts.Unknown++
			continue
		}
		observation, known, scopeOK, decodeErr := sessionLine(b, decoded)
		if decodeErr != nil {
			return fail("invalid_native_record")
		}
		if !scopeOK {
			return fail("session_or_cwd_changed")
		}
		if !known {
			batch.Next.Counts.Unknown++
		}
		if observation.ts <= 0 {
			if known {
				batch.Next.Counts.Unknown++
			}
			continue
		}
		kind := observation.kind
		if kind == "unknown" {
			kind = "native_activity"
		}
		origin, id := sessionOrigin(key, identity, start)
		p := RelaySessionProjection{ID: id, Kind: kind, Topic: "relay:" + b.TaskID, Tokens: observation.tokens, Scope: b.ProjectID, TS: time.UnixMilli(observation.ts).UTC().Format(time.RFC3339Nano)}
		if p.Tokens == nil {
			batch.Next.Counts.UnknownTokens++
		}
		if batch.Next.Seen[id] {
			batch.Next.Counts.Duplicates++
			continue
		}
		batch.Next.Seen[id] = true
		batch.Next.Counts.Projected++
		batch.Items = append(batch.Items, RelaySessionItem{Origin: origin, Projection: p})
	}
	if reader.Err() != nil {
		return fail("source_read_failed")
	}
	batch.Remaining = info.Size() - batch.Next.Offset
	// Detect shrink/replacement during the read. Appends are left for the next call.
	after, err := os.Stat(b.Path)
	if err != nil || !os.SameFile(info, after) || after.Size() < info.Size() {
		return fail("source_changed_during_read")
	}
	batch.SnapshotComplete = batch.HeaderVerified && batch.TailBytes == 0 && after.Size() == info.Size() && batch.Next.Offset == info.Size()
	if b.ArchiveFinalized && batch.TailBytes != 0 {
		return fail("closed_incomplete_tail")
	}
	batch.SessionComplete = batch.SnapshotComplete && b.ExitConfirmed && b.ArchiveFinalized && b.PrefixComplete && batch.Next.Counts.Unknown == 0 && batch.Next.Counts.Errors == 0 && batch.Next.Blocked == ""
	return batch, nil
}

func sessionLineEnd(f *os.File, start, length, extent int64) (int64, bool, error) {
	end := start + length
	if end == extent {
		return start, false, nil
	}
	var boundary [1]byte
	if _, err := f.ReadAt(boundary[:], end); err != nil {
		return start, false, err
	}
	if boundary[0] == '\r' {
		end++
		if end == extent {
			return start, false, nil
		}
		if _, err := f.ReadAt(boundary[:], end); err != nil {
			return start, false, err
		}
	}
	if boundary[0] != '\n' {
		return start, false, sessionFailure("source_boundary_invalid")
	}
	return end + 1, true, nil
}

func sessionHeaderMatches(b RelaySessionBinding, raw []byte) (bool, error) {
	if b.Format != "codex" && b.Format != "claude_code" {
		return false, nil
	}
	var identity struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &identity); err != nil {
		return false, err
	}
	if b.Format == "codex" && identity.Type != "session_meta" {
		return false, nil
	}
	o, _, scopeOK, err := sessionLine(b, raw)
	if err != nil {
		return false, err
	}
	return scopeOK && o.session == b.NativeID && o.cwd != "" && sameReplayProject(o.cwd, b.WorkspaceRoot), nil
}

// 使用类型化解码，未知协议保持未知，类型错误立即返回。
func sessionLine(b RelaySessionBinding, raw []byte) (nativeObservation, bool, bool, error) {
	var record struct {
		Type      string          `json:"type"`
		Session   string          `json:"sessionId"`
		CWD       string          `json:"cwd"`
		Timestamp json.RawMessage `json:"timestamp"`
		Payload   json.RawMessage `json:"payload"`
		Message   json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(raw, &record); err != nil {
		return nativeObservation{}, false, false, err
	}
	ts, err := sessionTimestamp(record.Timestamp)
	if err != nil {
		return nativeObservation{}, false, false, err
	}
	o := nativeObservation{session: record.Session, cwd: record.CWD, kind: "unknown", outcome: "unknown", depth: "unknown", ts: ts}
	known := false
	if b.Format == "codex" {
		p, e := sessionObject(record.Payload)
		if e != nil {
			return o, false, false, e
		}
		if record.Type == "session_meta" {
			o.session, err = sessionString(p, "id")
			if err != nil {
				return o, false, false, err
			}
			o.cwd, err = sessionString(p, "cwd")
			if err != nil {
				return o, false, false, err
			}
			if o.ts == 0 {
				o.ts, err = sessionTimestamp(p["timestamp"])
				if err != nil {
					return o, false, false, err
				}
			}
		} else if record.Type == "turn_context" {
			o.cwd, err = sessionString(p, "cwd")
			if err != nil {
				return o, false, false, err
			}
		}
		known, err = sessionCodexEnvelope(record.Type, p, &o)
		if err != nil {
			return o, false, false, err
		}
	} else if b.Format == "claude_code" {
		m, e := sessionObject(record.Message)
		if e != nil {
			return o, false, false, e
		}
		known = record.Type == "assistant" || record.Type == "user" || record.Type == "system"
		if len(m["usage"]) != 0 {
			u, e := sessionObject(m["usage"])
			if e != nil {
				return o, false, false, e
			}
			tokens, valid, e := sessionTokenTotal(u)
			if e != nil {
				return o, false, false, e
			}
			o.tokens = tokens
			known = known && valid
		}
	}
	if (o.session != "" && o.session != b.NativeID) || (o.cwd != "" && !sameReplayProject(o.cwd, b.WorkspaceRoot)) {
		return o, false, false, nil
	}
	return o, known, true, nil
}

func sessionObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func sessionString(v map[string]json.RawMessage, key string) (string, error) {
	if len(v[key]) == 0 {
		return "", nil
	}
	var s string
	if err := json.Unmarshal(v[key], &s); err != nil {
		return "", err
	}
	return s, nil
}

func sessionTimestamp(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return 0, nil
	}
	var v any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&v); err != nil {
		return 0, err
	}
	switch value := v.(type) {
	case string:
		t, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return 0, err
		}
		return t.UnixMilli(), nil
	case json.Number:
		return value.Int64()
	}
	return 0, sessionFailure("timestamp_type_invalid")
}

// 缺少使用量字段保持未知，不使用累计值填补。
func sessionTokenTotal(u map[string]json.RawMessage) (*int, bool, error) {
	values := map[string]int{}
	for _, key := range []string{"total_tokens", "input_tokens", "output_tokens"} {
		if raw, ok := u[key]; ok {
			var n int
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, false, nil
			}
			if err := json.Unmarshal(raw, &n); err != nil {
				return nil, false, err
			}
			if n < 0 {
				return nil, false, sessionFailure("token_usage_negative")
			}
			values[key] = n
		}
	}
	i, inputOK := values["input_tokens"]
	o, outputOK := values["output_tokens"]
	if inputOK && outputOK && i > math.MaxInt-o {
		return nil, false, sessionFailure("token_usage_overflow")
	}
	if total, ok := values["total_tokens"]; ok {
		return &total, true, nil
	}
	if !inputOK || !outputOK {
		return nil, false, nil
	}
	total := i + o
	return &total, true, nil
}
