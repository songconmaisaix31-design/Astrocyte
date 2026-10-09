package importers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SummarizeExtractor runs the existing pure Node website extractor, never the
// downloader, transcription or LLM path. Paths come from trusted startup config.
type SummarizeExtractor struct{ NodeExecutable, CLIPath string }

func NewSummarizeExtractor(nodeExecutable, cliPath string) (*SummarizeExtractor, error) {
	for _, p := range []string{nodeExecutable, cliPath} {
		if !filepath.IsAbs(p) {
			return nil, fmt.Errorf("extractor executable paths must be absolute")
		}
		info, err := os.Stat(p)
		if err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("summarize extractor unavailable")
		}
	}
	return &SummarizeExtractor{NodeExecutable: nodeExecutable, CLIPath: cliPath}, nil
}

func (e *SummarizeExtractor) Extract(ctx context.Context, locator string) (Export, error) {
	// This optional path is for fixed official arXiv HTML only. Other imported
	// sources still use already-exported data, not arbitrary CLI URL execution.
	id, err := NormalizeArxivID(locator)
	if err != nil || !revisionSuffix.MatchString(id) || locator != "https://arxiv.org/html/"+id {
		return Export{}, fmt.Errorf("extractor requires fixed official arXiv HTML")
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	home, err := os.MkdirTemp("", "astrocyte-summarize-")
	if err != nil {
		return Export{}, err
	}
	defer os.RemoveAll(home)
	clean := []string{"HOME=" + home, "USERPROFILE=" + home, "TEMP=" + home, "TMP=" + home, "PATH=" + filepath.Dir(e.NodeExecutable)}
	for _, key := range []string{"SystemRoot", "WINDIR", "COMSPEC"} {
		if v := os.Getenv(key); v != "" {
			clean = append(clean, key+"="+v)
		}
	}
	version := exec.CommandContext(ctx, e.NodeExecutable, e.CLIPath, "--version")
	version.Env = clean
	v, err := version.Output()
	if err != nil || !strings.Contains(string(v), "0.21.8") {
		return Export{}, fmt.Errorf("summarize extractor requires installed 0.21.8")
	}
	cmd := exec.CommandContext(ctx, e.NodeExecutable, e.CLIPath, locator, "--extract", "--json", "--format", "text", "--firecrawl", "off", "--youtube", "web", "--video-mode", "transcript", "--embedded-video", "off", "--timeout", "30s", "--retries", "0", "--metrics", "off")
	cmd.Env = clean
	cmd.Dir = home
	var stdout cappedOutput
	var stderr cappedOutput
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return Export{}, ctx.Err()
		}
		return Export{}, fmt.Errorf("summarize extractor unavailable: %v", err)
	}
	result, err := ParseSummarizeJSON(stdout.Bytes())
	if err != nil {
		return Export{}, err
	}
	if result.URL != locator || !result.HasOriginal {
		return Export{}, fmt.Errorf("summarize extractor lacks matching original HTML text")
	}
	return result, nil
}

// Bound subprocess output, including a misconfigured/unexpected executable.
type cappedOutput struct{ bytes.Buffer }

func (b *cappedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 16<<20 {
		return 0, fmt.Errorf("extractor output exceeds size limit")
	}
	return b.Buffer.Write(p)
}
