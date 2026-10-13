package events

import (
	"bufio"
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
			continue
		}
		approved, e := unredirectedReplayPath(root)
		if e != nil {
			continue
		}
		info, e := os.Stat(approved)
		if e != nil || !info.IsDir() {
			continue
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

// ReadRelaySession reads only a selected file's frozen byte extent. ReadBytes
// has no Scanner token ceiling, and no text/record-count limit is substituted
// for session coverage. A non-newline tail never advances the checkpoint.
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
	if info.Size() < c.Offset {
		return fail("source_truncated")
	}
	if c.Offset > 0 {
		var boundary [1]byte
		if _, e := f.ReadAt(boundary[:], c.Offset-1); e != nil || boundary[0] != '\n' {
			return fail("checkpoint_not_line_boundary")
		}
	}
	// Revalidate the first complete identity record even after a serialized restart.
	headReader := bufio.NewReader(sessionContextReader{ctx, io.NewSectionReader(f, 0, info.Size())})
	header, headErr := headReader.ReadBytes('\n')
	if headErr == nil {
		if b.Format == "codex" || b.Format == "claude_code" {
			if !sessionHeaderMatches(b, header) {
				return fail("session_header_unverifiable")
			}
			batch.HeaderVerified = true
		}
	} else if headErr != io.EOF {
		return fail("header_read_failed")
	}
	reader := bufio.NewReader(sessionContextReader{ctx, io.NewSectionReader(f, c.Offset, info.Size()-c.Offset)})
	for {
		if ctx.Err() != nil {
			return fail("capture_cancelled")
		}
		line, e := reader.ReadBytes('\n')
		if e == io.EOF {
			batch.TailBytes = int64(len(line))
			break
		}
		if e != nil {
			return fail("source_read_failed")
		}
		start := batch.Next.Offset
		batch.Next.Offset += int64(len(line))
		batch.Next.Counts.Lines++
		if !json.Valid(bytes.TrimSpace(line)) {
			markSessionError(&batch.Next, "malformed_json", false)
			continue
		}
		if !batch.HeaderVerified {
			batch.Next.Counts.Unknown++
			continue
		}
		observation, known, scopeOK := sessionLine(b, line)
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
		if old, exists := batch.Next.Seen[id]; exists {
			left, _ := json.Marshal(old)
			right, _ := json.Marshal(p)
			if !bytes.Equal(left, right) {
				return fail("duplicate_event_conflict")
			}
			batch.Next.Counts.Duplicates++
			continue
		}
		batch.Next.Seen[id] = p
		batch.Next.Counts.Projected++
		batch.Items = append(batch.Items, RelaySessionItem{Origin: origin, Projection: p})
	}
	batch.Remaining = info.Size() - batch.Next.Offset
	// Detect shrink/replacement during the read. Appends are left for the next call.
	after, err := os.Stat(b.Path)
	if err != nil || !os.SameFile(info, after) || after.Size() < info.Size() {
		return fail("source_changed_during_read")
	}
	batch.SnapshotComplete = batch.HeaderVerified && batch.TailBytes == 0 && after.Size() == info.Size() && batch.Next.Offset == info.Size()
	if b.ArchiveFinalized && batch.TailBytes != 0 {
		markSessionError(&batch.Next, "closed_incomplete_tail", true)
	}
	batch.SessionComplete = batch.SnapshotComplete && b.ExitConfirmed && b.ArchiveFinalized && b.PrefixComplete && batch.Next.Counts.Unknown == 0 && batch.Next.Counts.Errors == 0 && batch.Next.Blocked == ""
	return batch, nil
}

func sessionHeaderMatches(b RelaySessionBinding, raw []byte) bool {
	if b.Format != "codex" && b.Format != "claude_code" {
		return false
	}
	o, err := nativeRecord(b.Format, raw)
	v := object(raw)
	if b.Format == "codex" && str(v, "type") != "session_meta" {
		return false
	}
	return err == nil && sessionMetadataTyped(v) && o.session == b.NativeID && o.cwd != "" && sameReplayProject(o.cwd, b.WorkspaceRoot)
}

// Uses only the existing fixture-backed vocabulary, without changing any mapper.
// Unknown envelopes are counted and never treated as complete format support.
func sessionLine(b RelaySessionBinding, raw []byte) (nativeObservation, bool, bool) {
	o, err := nativeRecord(b.Format, raw)
	if err != nil {
		return nativeObservation{}, false, false
	}
	if (o.session != "" && o.session != b.NativeID) || (o.cwd != "" && !sameReplayProject(o.cwd, b.WorkspaceRoot)) {
		return o, false, false
	}
	v := object(raw)
	kind := str(v, "type")
	if !sessionMetadataTyped(v) {
		return o, false, false
	}
	known := false
	if b.Format == "codex" {
		p := object(v["payload"])
		if cwd := str(p, "cwd"); (kind == "session_meta" || kind == "turn_context") && cwd != "" && !sameReplayProject(cwd, b.WorkspaceRoot) {
			return o, false, false
		}
		known = sessionCodexEnvelope(kind, p, &o)
	} else if b.Format == "claude_code" {
		m := object(v["message"])
		known = kind == "assistant" || kind == "user" || kind == "system"
		if len(m["usage"]) != 0 {
			var valid bool
			o.tokens, valid = sessionTokenTotal(object(m["usage"]))
			known = known && valid
		}
	}
	if o.tokens != nil && *o.tokens < 0 {
		o.tokens = nil
		known = false
	}
	return o, known, true
}

func sessionMetadataTyped(v map[string]json.RawMessage) bool {
	for _, fields := range []map[string]json.RawMessage{v, object(v["payload"])} {
		for _, key := range []string{"sessionId", "cwd"} {
			if value, ok := fields[key]; ok {
				var s string
				if json.Unmarshal(value, &s) != nil || s == "" || string(value) == "null" {
					return false
				}
			}
		}
	}
	return true
}

// Validate every present component, even when a source supplies total_tokens.
// Missing components are unknown, never inferred from cumulative/cache totals.
func sessionTokenTotal(u map[string]json.RawMessage) (*int, bool) {
	values := map[string]int{}
	for _, key := range []string{"total_tokens", "input_tokens", "output_tokens"} {
		if raw, ok := u[key]; ok {
			var n int
			if string(raw) == "null" || json.Unmarshal(raw, &n) != nil || n < 0 {
				return nil, false
			}
			values[key] = n
		}
	}
	i, inputOK := values["input_tokens"]
	o, outputOK := values["output_tokens"]
	if inputOK && outputOK && i > math.MaxInt-o {
		return nil, false
	}
	if total, ok := values["total_tokens"]; ok {
		return &total, true
	}
	if !inputOK || !outputOK {
		return nil, false
	}
	total := i + o
	return &total, true
}
