package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type liveCapture struct {
	Rows                []app.StoredEvent
	Clients             map[string]map[string]int `json:"client_counts"`
	Git                 map[string]int            `json:"git_counts"`
	Heads               map[string]string         `json:"captured_git_heads"`
	Errors              int                       `json:"source_errors"`
	StructureFailed     int                       `json:"structure_failed"`
	NormalizationFailed int                       `json:"normalization_failed"`
	Noncognitive        int                       `json:"noncognitive"`
	Duplicates          int                       `json:"capture_duplicates"`
	seen                map[string]app.StoredEvent
	mapper              *events.NativeSemanticMapper
	raw                 *json.Encoder
	selection           []map[string]any
}

func (c *liveCapture) add(row app.StoredEvent) {
	if old, ok := c.seen[row.OriginKey]; ok {
		a, _ := json.Marshal(old.Observation)
		b, _ := json.Marshal(row.Observation)
		if !bytes.Equal(old.RawBytes, row.RawBytes) || !bytes.Equal(a, b) {
			c.Errors++
		} else {
			c.Duplicates++
		}
		return
	}
	c.seen[row.OriginKey] = row
	c.Rows = append(c.Rows, row)
}

func captureLive(ctx context.Context, config liveConfig, out string) (liveCapture, error) {
	c := liveCapture{Clients: map[string]map[string]int{}, Git: map[string]int{}, Heads: map[string]string{}, seen: map[string]app.StoredEvent{}, mapper: events.NewNativeSemanticMapper()}
	f, err := os.OpenFile(filepath.Join(out, "raw-local.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return c, err
	}
	defer f.Close()
	c.raw = json.NewEncoder(f)
	for _, source := range []string{"claude_code", "codex"} {
		c.captureNative(ctx, source, config.Native[source], config.Start, config.End)
	}
	for _, topic := range []string{"Morphogenesis", "Astrocyte/Attention Engine", "McPet"} {
		c.captureGit(ctx, topic, config.Git[topic], config.Start, config.End)
	}
	if err := writeExclusive(filepath.Join(out, "native-selection-local.json"), c.selection); err != nil {
		c.Errors++
		return c, err
	}
	sort.Slice(c.Rows, func(i, j int) bool {
		if c.Rows[i].Observation.TS != c.Rows[j].Observation.TS {
			return c.Rows[i].Observation.TS < c.Rows[j].Observation.TS
		}
		return c.Rows[i].OriginKey < c.Rows[j].OriginKey
	})
	return c, f.Sync()
}

func (c *liveCapture) captureNative(ctx context.Context, source, root string, start, end time.Time) {
	counts := map[string]int{}
	c.Clients[source] = counts
	if events.ValidateLocalReplayPath(root) != nil {
		c.Errors++
		counts["root_errors"]++
		return
	}
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			counts["directory_errors"]++
			c.Errors++
			return nil
		}
		if events.ValidateLocalReplayPath(path) != nil {
			counts["redirect_errors"]++
			c.Errors++
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		c.Errors++
		counts["directory_errors"]++
	}
	sort.Strings(files)
	var relativeFiles []string
	for _, path := range files {
		relative, _ := filepath.Rel(root, path)
		relativeFiles = append(relativeFiles, filepath.ToSlash(relative))
	}
	c.selection = append(c.selection, map[string]any{"source": source, "root": root, "files": relativeFiles, "window_start": start.UnixMilli(), "window_end": end.UnixMilli(), "selection_complete": err == nil && counts["directory_errors"] == 0 && counts["redirect_errors"] == 0, "scope": "all authorized JSONL files; not restricted to tool hits or capture day"})
	counts["files"] = len(files)
	for _, path := range files {
		if ctx.Err() != nil {
			c.Errors++
			counts["cancelled"]++
			return
		}
		c.captureNativeFile(source, root, path, start, end, counts)
	}
}

func liveTimestamp(raw json.RawMessage) int64 {
	var v struct {
		Timestamp json.RawMessage `json:"timestamp"`
		Type      string          `json:"type"`
		Payload   struct {
			Timestamp json.RawMessage `json:"timestamp"`
		} `json:"payload"`
	}
	if json.Unmarshal(raw, &v) != nil {
		return 0
	}
	parse := func(value json.RawMessage) int64 {
		var number int64
		if json.Unmarshal(value, &number) == nil {
			return number
		}
		var text string
		if json.Unmarshal(value, &text) != nil {
			return 0
		}
		t, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return 0
		}
		return t.UnixMilli()
	}
	ts := parse(v.Timestamp)
	if ts == 0 && v.Type == "session_meta" {
		ts = parse(v.Payload.Timestamp)
	}
	return ts
}

func liveAssociationScope(raw json.RawMessage, previous events.NativeAssociation) string {
	if previous.Scope != "" {
		return previous.Scope
	}
	var v struct {
		CWD     string `json:"cwd"`
		Type    string `json:"type"`
		Payload struct {
			CWD string `json:"cwd"`
		} `json:"payload"`
	}
	_ = json.Unmarshal(raw, &v)
	if v.Type == "session_meta" {
		return v.Payload.CWD
	}
	return v.CWD
}

func liveToolOrigin(source, relative string, line, block int, raw json.RawMessage, association events.NativeAssociation) string {
	var v struct {
		Payload struct {
			CallID string `json:"call_id"`
		} `json:"payload"`
		Message struct {
			Content []struct {
				ID string `json:"id"`
			} `json:"content"`
		} `json:"message"`
	}
	_ = json.Unmarshal(raw, &v)
	id := v.Payload.CallID
	if source == "claude_code" && block < len(v.Message.Content) {
		id = v.Message.Content[block].ID
	}
	if association.Session != "" && id != "" {
		return fmt.Sprintf("live-native:%s:%q:%q", source, association.Session, id)
	}
	return fmt.Sprintf("live-native:%s:file:%q:line:%d:block:%d", source, filepath.ToSlash(relative), line, block)
}

func (c *liveCapture) captureNativeFile(source, root, path string, start, end time.Time, counts map[string]int) {
	f, err := os.Open(path)
	if err != nil {
		c.Errors++
		counts["read_errors"]++
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		c.Errors++
		counts["read_errors"]++
		return
	}
	// Snapshot each file's observed byte length; concurrent appends belong to
	// the next capture. A torn last record or shrink is an incomplete source.
	reader := bufio.NewReader(io.LimitReader(f, info.Size()))
	association := events.NativeAssociation{}
	badAssociation := false
	relative, _ := filepath.Rel(root, path)
	if source == "claude_code" && strings.Contains(filepath.ToSlash(relative), "/subagents/") {
		counts["subagent_files"]++
	}
	if source == "codex" {
		parts := strings.Split(filepath.ToSlash(relative), "/")
		if len(parts) >= 4 {
			if _, err := strconv.Atoi(parts[0]); err == nil {
				counts["dated_files"]++
			} else {
				counts["legacy_files"]++
			}
		} else {
			counts["legacy_files"]++
		}
	}
	line := 0
	var readErr error
	for {
		lineBytes, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			readErr = err
			break
		}
		if len(lineBytes) == 0 {
			break
		}
		line++
		counts["scanned_lines"]++
		raw := json.RawMessage(bytes.TrimRight(lineBytes, "\r\n"))
		trimmed := bytes.TrimSpace(raw)
		if !json.Valid(raw) || len(trimmed) == 0 || trimmed[0] != '{' {
			c.Errors++
			counts["parse_errors"]++
			continue
		}
		project := liveAssociationScope(raw, association)
		next, err := events.ObserveNativeAssociation(source, project, association, raw)
		if err != nil {
			badAssociation = true
		} else {
			association = next
		}
		ts := liveTimestamp(raw)
		if ts == 0 {
			counts["timestamp_unknown"]++
			if len(events.NativeToolBlocks(source, raw)) > 0 {
				c.Errors++
				counts["tool_timestamp_unknown"]++
			}
			continue
		}
		if ts < start.UnixMilli() || ts >= end.UnixMilli() {
			continue
		}
		counts["window_records"]++
		if c.raw.Encode(map[string]any{"source": source, "relative_file": relative, "line": line, "association": association, "record": raw}) != nil {
			c.Errors++
			counts["evidence_errors"]++
			return
		}
		blocks := events.NativeToolBlocks(source, raw)
		if len(blocks) == 0 {
			counts["non_tool_records"]++
			continue
		}
		if badAssociation || association.Scope == "" || association.Session == "" {
			c.Errors++
			counts["association_errors"]++
			continue
		}
		for _, block := range blocks {
			counts["tool_parents"]++
			origin := liveToolOrigin(source, relative, line, block, raw, association)
			rows, semantics, err := c.mapper.MapTool(source, origin, association.Scope, raw, ts, association, block)
			if err != nil {
				c.Errors++
				counts["mapping_errors"]++
				continue
			}
			for i, row := range rows {
				counts["semantic_units"]++
				if row.Observation.TS != ts {
					c.Errors++
					counts["timestamp_errors"]++
					continue
				}
				if semantics[i].Bucket == "unresolved_structure" {
					c.StructureFailed++
					counts["structure_failed"]++
					c.add(row)
				} else if strings.HasPrefix(semantics[i].Bucket, "cognitive_") {
					if row.Event == nil {
						c.Errors++
						counts["mapping_errors"]++
						continue
					}
					c.add(row)
					signal, err := events.NormalizeValidationSignal(row)
					if err != nil {
						c.Errors++
						counts["normalization_errors"]++
					} else if signal == nil {
						c.NormalizationFailed++
						counts["normalization_failed"]++
					} else {
						counts["cognitive_success"]++
					}
				} else {
					c.Noncognitive++
					counts["noncognitive"]++
				}
			}
		}
	}
	current, statErr := os.Stat(path)
	if readErr != nil || statErr != nil || current.Size() < info.Size() || !os.SameFile(info, current) {
		c.Errors++
		counts["read_errors"]++
	}
}

func (c *liveCapture) captureGit(ctx context.Context, topic, root string, start, end time.Time) {
	c.Git[topic] = 0
	if events.ValidateLocalReplayPath(root) != nil || c.mapper.Repository(root).Topic != topic {
		c.Errors++
		return
	}
	run := func(args ...string) ([]byte, error) {
		command := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
		return command.Output()
	}
	head, err := run("rev-parse", "HEAD")
	if err != nil {
		c.Errors++
		return
	}
	sha := strings.TrimSpace(string(head))
	c.Heads[topic] = sha
	data, err := run("log", sha, "--since-as-filter="+start.Format(time.RFC3339), "--until="+end.Format(time.RFC3339), "--format=%H%x00%ct%x00%s")
	if err != nil {
		c.Errors++
		return
	}
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		parts := bytes.SplitN(bytes.TrimSuffix(line, []byte("\r")), []byte{0}, 3)
		if len(parts) != 3 {
			c.Errors++
			continue
		}
		seconds, err := strconv.ParseInt(string(parts[1]), 10, 64)
		if err != nil {
			c.Errors++
			continue
		}
		ts := seconds * 1000
		if ts < start.UnixMilli() || ts >= end.UnixMilli() {
			continue
		}
		record := events.GitLogRecord{Commit: string(parts[0]), TS: ts, Subject: string(parts[2])}
		raw, _ := json.Marshal(record)
		if c.raw.Encode(map[string]any{"source": "git", "topic": topic, "project": root, "record": record, "git_line": line}) != nil {
			c.Errors++
			return
		}
		row, err := c.mapper.MapGit(root, raw, ts)
		if err != nil || row.Event == nil || row.Event.Topic != topic {
			c.Errors++
			continue
		}
		c.Git[topic]++
		c.add(row)
	}
}
