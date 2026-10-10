package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type historyRecord struct {
	root, path, projectID, nativeID string
}

var discoveredHistory = struct {
	sync.Mutex
	records map[string]historyRecord
}{records: map[string]historyRecord{}}

// Only explicit human-registered history roots are walked. Read at most a 64KiB
// header before checking cwd; content of unmatched projects is never decoded.
func discoverHistory(ctx context.Context, cli string, p domain.LocalProject) ([]domain.NativeSession, error) {
	if cli != "codex" && cli != "pi" {
		return nil, nativeError(apierrors.UnsupportedCapability, "this CLI history header scope protocol has not been verified")
	}
	chosen := p.Settings.HistoryRoots[cli]
	if chosen == "" {
		return nil, nativeError(apierrors.ApprovalRequired, "register this CLI history root explicitly before discovery")
	}
	root, err := CanonicalRoot(chosen)
	if err != nil {
		return nil, err
	}
	result := []domain.NativeSession{}
	count := 0
	err = filepath.WalkDir(root, func(path string, e fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			return scopeError("registered history root is inaccessible")
		}
		count++
		if count > 2000 {
			return nativeError(apierrors.BudgetExhausted, "history discovery exceeds 2000 entries")
		}
		if e.Type()&os.ModeSymlink != 0 {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if e.IsDir() {
			if strings.Count(rel, string(filepath.Separator)) > 4 {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".jsonl" {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		line, err := bufio.NewReaderSize(io.LimitReader(f, 64*1024), 64*1024).ReadBytes('\n')
		_ = f.Close()
		if err != nil && err != io.EOF {
			return nil
		}
		var record struct {
			Type    string `json:"type"`
			ID      string `json:"id"`
			Cwd     string `json:"cwd"`
			Payload struct {
				ID  string `json:"id"`
				Cwd string `json:"cwd"`
			} `json:"payload"`
		}
		if json.Unmarshal(line, &record) != nil {
			return nil
		}
		id, cwd := record.ID, record.Cwd
		if cli == "codex" {
			if record.Type != "session_meta" {
				return nil
			}
			id, cwd = record.Payload.ID, record.Payload.Cwd
		} else if record.Type != "session" {
			return nil
		}
		if id == "" || filepath.Clean(cwd) != filepath.Clean(p.Root) {
			return nil
		}
		if len(result) >= 100 {
			return nativeError(apierrors.BudgetExhausted, "history discovery exceeds 100 sessions")
		}
		key := uuid.NewString()
		discoveredHistory.Lock()
		discoveredHistory.records[key] = historyRecord{root: root, path: path, projectID: p.ID, nativeID: id}
		discoveredHistory.Unlock()
		result = append(result, domain.NativeSession{ID: key, ProjectID: p.ID, CLI: cli, NativeID: id, Version: "historical_version_not_verified", Ownership: "external_observed", Mode: "observed", Status: "unknown", StopConfirmed: false, Limitations: []string{"history is read-only evidence; occupancy and original process ownership are unknown"}})
		return nil
	})
	return result, err
}

func readHistory(ctx context.Context, s domain.NativeSession) ([]domain.NativeEvent, error) {
	discoveredHistory.Lock()
	r, ok := discoveredHistory.records[s.ID]
	discoveredHistory.Unlock()
	if !ok || r.projectID != s.ProjectID || r.nativeID != s.NativeID {
		return nil, nativeError(apierrors.ScopeDenied, "history must be rediscovered in the current approved project")
	}
	rel, err := filepath.Rel(r.root, r.path)
	if err != nil {
		return nil, err
	}
	// Generic project path validation excludes all native-state dirs; here the
	// history root is separately human approved, still reject symlinks/escapes.
	resolved, err := filepath.EvalSymlinks(r.path)
	if err != nil || resolved != r.path || !inside(r.root, resolved) || rel == ".." {
		return nil, scopeError("history path leaves registered scope")
	}
	info, err := os.Stat(resolved)
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1024*1024 {
		return nil, nativeError(apierrors.BudgetExhausted, "history read exceeds 1 MiB input bound")
	}
	f, err := os.Open(resolved)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scan := bufio.NewScanner(io.LimitReader(f, 1024*1024+1))
	scan.Buffer(make([]byte, 4096), 128*1024)
	events := []domain.NativeEvent{}
	output := 0
	for scan.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		var rec struct {
			Type    string `json:"type"`
			Message struct {
				Role    string          `json:"role"`
				Content json.RawMessage `json:"content"`
			} `json:"message"`
			Payload struct {
				Type    string `json:"type"`
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"payload"`
		}
		if json.Unmarshal(scan.Bytes(), &rec) != nil {
			return nil, nativeError(apierrors.EvidenceMissing, "native history JSONL is invalid")
		}
		texts := []string{}
		if s.CLI == "codex" && rec.Type == "response_item" && rec.Payload.Type == "message" && (rec.Payload.Role == "assistant" || rec.Payload.Role == "user") {
			for _, part := range rec.Payload.Content {
				if part.Type == "input_text" || part.Type == "output_text" {
					texts = append(texts, part.Text)
				}
			}
		}
		if s.CLI == "pi" && rec.Type == "message" && (rec.Message.Role == "assistant" || rec.Message.Role == "user") {
			var text string
			if json.Unmarshal(rec.Message.Content, &text) == nil {
				texts = append(texts, text)
			} else {
				var parts []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}
				_ = json.Unmarshal(rec.Message.Content, &parts)
				for _, part := range parts {
					if part.Type == "text" {
						texts = append(texts, part.Text)
					}
				}
			}
		}
		for _, text := range texts {
			output += len(text)
			if output > 128*1024 || len(events) >= 256 {
				return nil, nativeError(apierrors.BudgetExhausted, "history context exceeds output bound")
			}
			events = append(events, domain.NativeEvent{Sequence: len(events) + 1, Kind: "historical_text", Text: text})
		}
	}
	return events, scan.Err()
}
