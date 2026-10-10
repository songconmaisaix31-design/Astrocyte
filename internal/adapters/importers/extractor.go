package importers

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:embed summarize-media.mjs
var summarizeMediaBridge []byte

// All paths are trusted startup configuration. Empty tools permit captions,
// never a cloud fallback. The bridge uses only pinned upstream media code.
type SummarizeOptions struct {
	Version, YtDlpPath, FFmpegPath, WhisperBinary, WhisperModel string
	// UVXPath points at the uvx binary summarize uses to convert a local PDF to
	// markdown. It is only required for paper_pdf extraction; empty means PDF
	// extraction is unavailable, which is reported honestly rather than guessed.
	UVXPath string
	Timeout time.Duration
}
type SummarizeExtractor struct {
	NodeExecutable, CLIPath string
	Options                 SummarizeOptions
}

func NewSummarizeExtractor(nodeExecutable, cliPath string) (*SummarizeExtractor, error) {
	return NewSummarizeExtractorWithOptions(nodeExecutable, cliPath, SummarizeOptions{})
}
func NewSummarizeExtractorWithOptions(nodeExecutable, cliPath string, o SummarizeOptions) (*SummarizeExtractor, error) {
	if o.Version == "" {
		o.Version = "0.25.1"
	}
	if o.Version != "0.25.1" && o.Version != "0.21.8" {
		return nil, fmt.Errorf("unsupported summarize pin")
	}
	for _, p := range []string{nodeExecutable, cliPath} {
		if err := regularAbsolute(p); err != nil {
			return nil, err
		}
	}
	for _, p := range []string{o.YtDlpPath, o.FFmpegPath, o.WhisperBinary, o.WhisperModel} {
		if p != "" {
			if err := regularAbsolute(p); err != nil {
				return nil, err
			}
		}
	}
	if o.Timeout < 0 {
		return nil, fmt.Errorf("summarize timeout must be positive")
	}
	if o.Timeout == 0 {
		o.Timeout = 30 * time.Minute
	}
	if o.UVXPath != "" {
		if err := regularAbsolute(o.UVXPath); err != nil {
			return nil, err
		}
	} else if resolved, err := exec.LookPath("uvx"); err == nil {
		// Best-effort: use the ambient uvx when no explicit path is configured.
		// PDF extraction remains optional; its absence is reported as unavailable.
		if abs, err := filepath.Abs(resolved); err == nil {
			o.UVXPath = abs
		}
	}
	return &SummarizeExtractor{nodeExecutable, cliPath, o}, nil
}
func regularAbsolute(p string) error {
	if !filepath.IsAbs(p) {
		return fmt.Errorf("summarize dependency paths must be absolute")
	}
	info, err := os.Stat(p)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("summarize dependency unavailable: %s", p)
	}
	return nil
}
func (e *SummarizeExtractor) isolatedEnv(home string, media bool) []string {
	paths := []string{filepath.Dir(e.NodeExecutable)}
	env := []string{"HOME=" + home, "USERPROFILE=" + home, "TEMP=" + home, "TMP=" + home, "XDG_CONFIG_HOME=" + home, "XDG_CACHE_HOME=" + home, "APPDATA=" + home, "LOCALAPPDATA=" + home, "NO_COLOR=1", "SUMMARIZE_DISABLE_LOCAL_WHISPER_CPP=1"}
	if media {
		env[len(env)-1] = "SUMMARIZE_DISABLE_LOCAL_WHISPER_CPP=0"
		for _, p := range [][2]string{{"YT_DLP_PATH", e.Options.YtDlpPath}, {"FFMPEG_PATH", e.Options.FFmpegPath}, {"SUMMARIZE_WHISPER_CPP_BINARY", e.Options.WhisperBinary}, {"SUMMARIZE_WHISPER_CPP_MODEL_PATH", e.Options.WhisperModel}} {
			if p[1] != "" {
				env = append(env, p[0]+"="+p[1])
				paths = append(paths, filepath.Dir(p[1]))
			}
		}
	}
	for _, k := range []string{"SystemRoot", "WINDIR", "COMSPEC"} {
		if v := os.Getenv(k); v != "" {
			env = append(env, k+"="+v)
		}
	}
	// Upstream cleanup needs the OS taskkill; unrelated PATH tools stay excluded.
	if root := os.Getenv("SystemRoot"); root != "" {
		paths = append(paths, filepath.Join(root, "System32"))
	}
	return append(env, "PATH="+strings.Join(paths, string(os.PathListSeparator)))
}
func dependencyError(message string) error {
	return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: message, RequiredAction: extractionAction(message, "install_pinned_summarize_and_configure_local_media_tools")}
}

func extractionAction(message, fallback string) string {
	if strings.Contains(message, "blocked local network") || strings.Contains(message, "DNS pinning") || strings.Contains(message, "redirected too many") {
		return "configure_public_source_network"
	}
	return fallback
}

// paperPDFResult is the local markdown extraction of one already-downloaded PDF,
// plus the immutable original summarize JSON for provenance. Text is empty when
// the document yields no readable content; it is never guessed or OCR-filled
// without an explicit model key, which this path does not use.
type paperPDFResult struct {
	Text     string
	Original []byte
}

// ExtractPDF converts an already-downloaded public PDF to markdown using the
// pinned summarize local-PDF path (uvx + markitdown), never an LLM. The PDF
// bytes are written to a private temp file; no URL is fetched here.
func (e *SummarizeExtractor) ExtractPDF(ctx context.Context, pdfURL string, pdfBytes []byte) (paperPDFResult, error) {
	if e.Options.Version != "0.25.1" {
		return paperPDFResult{}, dependencyError("paper PDF extraction requires summarize 0.25.1")
	}
	if e.Options.UVXPath == "" {
		return paperPDFResult{}, dependencyError("paper PDF extraction requires uvx (markitdown) configured")
	}
	if len(pdfBytes) == 0 || len(pdfBytes) > 64<<20 {
		return paperPDFResult{}, invalid("paper PDF must contain from 1 byte to 64MiB")
	}
	ctx, cancel := context.WithTimeout(ctx, e.Options.Timeout)
	defer cancel()
	home, err := os.MkdirTemp("", "astrocyte-paper-pdf-")
	if err != nil {
		return paperPDFResult{}, err
	}
	defer os.RemoveAll(home)
	env := e.isolatedEnv(home, false)
	env = append(env, "UVX_PATH="+e.Options.UVXPath)
	// uv reuses its package cache across runs; pointing it at a stable location
	// avoids re-downloading markitdown[all] on every single-paper extraction.
	env = append(env, "UV_CACHE_DIR="+filepath.Join(home, "uv-cache"))
	if err = e.verifyVersion(ctx, env); err != nil {
		return paperPDFResult{}, err
	}
	pdfPath := filepath.Join(home, "paper.pdf")
	if err = os.WriteFile(pdfPath, pdfBytes, 0o600); err != nil {
		return paperPDFResult{}, err
	}
	cmd := exec.CommandContext(ctx, e.NodeExecutable, e.CLIPath, pdfPath, "--extract", "--json", "--format", "text", "--firecrawl", "off", "--youtube", "web", "--video-mode", "transcript", "--embedded-video", "off", "--timeout", "120s", "--retries", "0", "--metrics", "off")
	cmd.Env = env
	cmd.Dir = home
	cmd.WaitDelay = 5 * time.Second
	var out, stderr cappedOutput
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return paperPDFResult{}, ctx.Err()
		}
		return paperPDFResult{}, dependencyError(fmt.Sprintf("paper PDF extraction failed (%v): %s", err, stderr.String()))
	}
	var parsed struct {
		Extracted struct {
			Content string `json:"content"`
		} `json:"extracted"`
	}
	if err = json.Unmarshal(out.Bytes(), &parsed); err != nil {
		return paperPDFResult{}, fmt.Errorf("paper PDF output: %w", err)
	}
	if strings.TrimSpace(parsed.Extracted.Content) == "" {
		return paperPDFResult{}, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "paper PDF yielded no readable full text", RequiredAction: "choose_accessible_public_paper_or_provide_existing_export"}
	}
	return paperPDFResult{Text: parsed.Extracted.Content, Original: append([]byte(nil), out.Bytes()...)}, nil
}
func (e *SummarizeExtractor) verifyVersion(ctx context.Context, env []string) error {
	cmd := exec.CommandContext(ctx, e.NodeExecutable, e.CLIPath, "--version")
	cmd.Env = env
	cmd.WaitDelay = 3 * time.Second
	var out, stderr cappedOutput
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return dependencyError("summarize version check failed: " + stderr.String())
	}
	if strings.TrimSpace(out.String()) != e.Options.Version {
		return dependencyError("summarize installed version differs from configured pin " + e.Options.Version)
	}
	return nil
}

// Extract remains restricted to fixed official arXiv HTML, with media disabled.
func (e *SummarizeExtractor) Extract(ctx context.Context, locator string) (Export, error) {
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
	env := e.isolatedEnv(home, false)
	if err = e.verifyVersion(ctx, env); err != nil {
		return Export{}, err
	}
	cmd := exec.CommandContext(ctx, e.NodeExecutable, e.CLIPath, locator, "--extract", "--json", "--format", "text", "--firecrawl", "off", "--youtube", "web", "--video-mode", "transcript", "--embedded-video", "off", "--timeout", "30s", "--retries", "0", "--metrics", "off")
	cmd.Env = env
	cmd.Dir = home
	cmd.WaitDelay = 3 * time.Second
	var out, stderr cappedOutput
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return Export{}, ctx.Err()
		}
		return Export{}, dependencyError(fmt.Sprintf("summarize extraction failed (%v): %s", err, stderr.String()))
	}
	r, err := ParseSummarizeJSON(out.Bytes())
	if err != nil {
		return Export{}, err
	}
	if r.URL != locator || !r.HasOriginal {
		return Export{}, fmt.Errorf("summarize extractor lacks matching original HTML text")
	}
	r.Version = e.Options.Version
	return r, nil
}

var bilibiliVideoPath = regexp.MustCompile(`^/video/BV[0-9A-Za-z]{10}/?$`)
var youtubeVideoID = regexp.MustCompile(`^[0-9A-Za-z_-]{11}$`)

// Do not send playlists/search, credentials, arbitrary hosts or LAN URLs to yt-dlp.
func validateVideoURL(locator string) error {
	u, err := url.Parse(locator)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
		return invalid("use a single public HTTPS video URL without credentials or fragment")
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return invalid("invalid video URL query")
	}
	for _, values := range q {
		if len(values) != 1 {
			return invalid("video URL must select a single value per query parameter")
		}
	}
	switch strings.ToLower(u.Hostname()) {
	case "bilibili.com", "www.bilibili.com", "m.bilibili.com":
		if !bilibiliVideoPath.MatchString(u.Path) {
			return invalid("use a single Bilibili BV video URL")
		}
		for k := range q {
			if k != "p" && k != "spm_id_from" && k != "spm" {
				return invalid("unsupported Bilibili video query")
			}
		}
		if p := q.Get("p"); p != "" && p != "1" {
			return invalid("multipart selection is unsupported by the upstream single-video route")
		}
	case "youtu.be":
		if !youtubeVideoID.MatchString(strings.Trim(u.Path, "/")) {
			return invalid("use a single YouTube video URL")
		}
		for k := range q {
			if k != "t" && k != "si" {
				return invalid("playlists/search are not supported")
			}
		}
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		id := q.Get("v")
		if u.Path != "/watch" {
			id = ""
			for _, p := range []string{"/shorts/", "/embed/"} {
				if strings.HasPrefix(u.Path, p) {
					id = strings.TrimPrefix(u.Path, p)
				}
			}
		}
		if !youtubeVideoID.MatchString(id) {
			return invalid("use a single YouTube video URL")
		}
		for k := range q {
			if k != "v" && k != "t" {
				return invalid("playlists/search are not supported")
			}
		}
	default:
		return invalid("public URL video import supports single Bilibili and YouTube videos")
	}
	return nil
}

// Original failure outputs can be saved by live probes without accepting page text.
type ExtractionFailure struct {
	Err                           error
	CLIJSON, CLIStderr, MediaJSON []byte
}

func (e *ExtractionFailure) Error() string { return e.Err.Error() }
func (e *ExtractionFailure) Unwrap() error { return e.Err }
func (e *SummarizeExtractor) ExtractVideo(ctx context.Context, locator string) (Export, error) {
	if err := validateVideoURL(locator); err != nil {
		return Export{}, err
	}
	if e.Options.Version != "0.25.1" {
		return Export{}, dependencyError("media bridge requires summarize 0.25.1")
	}
	ctx, cancel := context.WithTimeout(ctx, e.Options.Timeout)
	defer cancel()
	home, err := os.MkdirTemp("", "astrocyte-summarize-media-")
	if err != nil {
		return Export{}, err
	}
	defer os.RemoveAll(home)
	env := e.isolatedEnv(home, true)
	if err = e.verifyVersion(ctx, env); err != nil {
		return Export{}, err
	}
	bridge := filepath.Join(home, "summarize-media.mjs")
	if err = os.WriteFile(bridge, summarizeMediaBridge, 0600); err != nil {
		return Export{}, err
	}
	cmd := exec.CommandContext(ctx, e.NodeExecutable, bridge, e.CLIPath, locator)
	cmd.Env = env
	cmd.Dir = home
	cmd.WaitDelay = 5 * time.Second
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return Export{}, err
	}
	defer stdin.Close()
	// EOF uses upstream tracked-child cleanup before WaitDelay's kill fallback.
	cmd.Cancel = func() error { return stdin.Close() }
	var out, stderr cappedOutput
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	if ctx.Err() != nil {
		return Export{}, ctx.Err()
	}
	return e.decodeVideoResult(locator, out.Bytes(), stderr.String(), runErr)
}

func (e *SummarizeExtractor) decodeVideoResult(locator string, output []byte, stderr string, runErr error) (Export, error) {
	var result struct {
		CLIJSON   string          `json:"cliJSON"`
		CLIStderr string          `json:"cliStderr"`
		CLIExit   int             `json:"cliExit"`
		Media     json.RawMessage `json:"media"`
		Error     string          `json:"error"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return Export{}, dependencyError(fmt.Sprintf("summarize media bridge failed (%v): %s", runErr, stderr))
	}
	fail := func(message string, code apierrors.Code) (Export, error) {
		return Export{}, &ExtractionFailure{Err: &apierrors.ServiceError{Code: code, Message: message, RequiredAction: extractionAction(message, "check_public_video_access_and_local_yt_dlp_ffmpeg_whisper_or_provide_existing_export")}, CLIJSON: []byte(result.CLIJSON), CLIStderr: []byte(result.CLIStderr), MediaJSON: result.Media}
	}
	if result.Error != "" {
		return fail(result.Error, apierrors.ProviderUnavailable)
	}
	if runErr != nil {
		return fail("summarize media bridge exited unsuccessfully", apierrors.ProviderUnavailable)
	}
	parsed, parseErr := ParseSummarizeJSON([]byte(result.CLIJSON))
	if parseErr == nil && result.CLIExit == 0 && parsed.HasOriginal && parsed.TranscriptSource != "" {
		if canonicalWebKey(parsed.URL) != canonicalWebKey(locator) {
			return fail("summarize transcript source mismatch", apierrors.EvidenceMissing)
		}
		parsed.Version = e.Options.Version
		parsed.Mode = "upstream_cli_transcript"
		return parsed, nil
	}
	var media struct {
		Text     *string         `json:"text"`
		Provider string          `json:"provider"`
		Error    *string         `json:"error"`
		Segments json.RawMessage `json:"segments"`
	}
	if err := json.Unmarshal(result.Media, &media); err != nil {
		return fail("selected video has no real transcript", apierrors.EvidenceMissing)
	}
	if media.Error != nil && *media.Error != "" {
		return fail(*media.Error, apierrors.ProviderUnavailable)
	}
	if media.Text == nil || strings.TrimSpace(*media.Text) == "" {
		return fail("selected video has no real transcript", apierrors.EvidenceMissing)
	}
	if media.Provider != "whisper.cpp" {
		return fail("unexpected nonlocal transcriber", apierrors.EvidenceMissing)
	}
	var cliMetadata struct {
		Extracted struct {
			Title string `json:"title"`
		} `json:"extracted"`
	}
	_ = json.Unmarshal([]byte(result.CLIJSON), &cliMetadata)
	normalized, err := json.Marshal(map[string]any{"input": map[string]string{"url": locator}, "extracted": map[string]any{"title": cliMetadata.Extracted.Title, "content": *media.Text, "transcriptSegments": media.Segments}})
	if err != nil {
		return Export{}, err
	}
	parsed, err = ParseSummarizeJSON(normalized)
	if err != nil {
		return Export{}, err
	}
	parsed.Original = append([]byte(nil), result.Media...)
	parsed.Version = e.Options.Version
	parsed.Mode = "upstream_media_transcript"
	parsed.TranscriptSource = media.Provider
	parsed.ExtraAttachments = []ExportAttachment{{Name: "summarize-cli-original.json", MediaType: "application/json", Data: []byte(result.CLIJSON)}, {Name: "summarize-cli.stderr.txt", MediaType: "text/plain", Data: []byte(result.CLIStderr)}}
	return parsed, nil
}

type cappedOutput struct{ bytes.Buffer }

func (b *cappedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 16<<20 {
		return 0, fmt.Errorf("extractor output exceeds size limit")
	}
	return b.Buffer.Write(p)
}
