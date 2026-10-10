package importers

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed paper-dom.mjs
var paperDOMModule []byte

//go:embed paper-dom-bridge.mjs
var paperDOMBridge []byte

// PaperSnapshot is adapter-local observation, not authority to import or process.
// Text is empty unless body evidence establishes readable full-paper structure.
type PaperSnapshot struct {
	SchemaVersion   int      `json:"schema_version"`
	SourceURL       string   `json:"source_url"`
	HostFamily      string   `json:"host_family"`
	SourceKey       string   `json:"source_key"`
	Title           string   `json:"title"`
	Abstract        string   `json:"abstract"`
	Authors         []string `json:"authors"`
	DOI             string   `json:"doi"`
	ArxivID         string   `json:"arxiv_id"`
	ObservedVersion string   `json:"observed_version"`
	PDFURLs         []string `json:"pdf_urls"`
	ContentState    string   `json:"content_state"`
	Text            string   `json:"text"`
	Warning         string   `json:"warning"`
	Truncated       bool     `json:"truncated"`
	Provenance      struct {
		Processor, Version, Mode, Source string
	} `json:"provenance"`
}

// InspectPaperHTML parses only supplied HTML bytes with existing pinned
// summarize DOM dependencies, without fetching URLs or invoking a model.
// It supports rendered current-page observations and explicit public snapshots
// without selecting either pending product path or freezing a transport.
func (e *SummarizeExtractor) InspectPaperHTML(ctx context.Context, locator string, html []byte) (PaperSnapshot, error) {
	var snapshot PaperSnapshot
	u, err := url.Parse(locator)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Port() != "" {
		return snapshot, invalid("paper observation requires a public HTTPS URL without credentials")
	}
	if len(html) == 0 || len(html) > 16<<20 {
		return snapshot, invalid("paper HTML must contain from 1 byte to 16MiB; no truncation")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	home, err := os.MkdirTemp("", "astrocyte-paper-dom-")
	if err != nil {
		return snapshot, err
	}
	defer os.RemoveAll(home)
	for name, data := range map[string][]byte{"paper-dom.mjs": paperDOMModule, "paper-dom-bridge.mjs": paperDOMBridge} {
		if err = os.WriteFile(filepath.Join(home, name), data, 0o600); err != nil {
			return snapshot, err
		}
	}
	cmd := exec.CommandContext(ctx, e.NodeExecutable, filepath.Join(home, "paper-dom-bridge.mjs"), e.CLIPath, locator)
	cmd.Env = e.isolatedEnv(home, false)
	cmd.Dir = home
	cmd.Stdin = bytes.NewReader(html)
	cmd.WaitDelay = 3 * time.Second
	var out, stderr cappedOutput
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return snapshot, ctx.Err()
		}
		return snapshot, dependencyError(fmt.Sprintf("paper DOM extraction failed (%v): %s", err, stderr.String()))
	}
	if err = json.Unmarshal(out.Bytes(), &snapshot); err != nil {
		return snapshot, fmt.Errorf("paper DOM output: %w", err)
	}
	if snapshot.SourceURL != locator || snapshot.Truncated || (snapshot.ContentState != "readable_fulltext" && strings.TrimSpace(snapshot.Text) != "") {
		return PaperSnapshot{}, invalid("paper DOM result lacks matching untruncated provenance")
	}
	return snapshot, nil
}
