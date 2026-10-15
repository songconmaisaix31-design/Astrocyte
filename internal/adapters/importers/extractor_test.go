package importers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func appVideoCommand(adapter string) app.ImportMaterialCommand {
	return app.ImportMaterialCommand{Adapter: adapter, Kind: "video", SourceLocator: "https://www.bilibili.com/video/BV1PReT6EEqR/"}
}

func installedExtractor(t *testing.T) *SummarizeExtractor {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("locate Node executable", err)
	}
	node, err = filepath.Abs(node)
	if err != nil {
		t.Fatal("resolve Node executable path", err)
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err = os.Stat(filepath.Join(root, "package.json")); err == nil {
			break
		}
		if !os.IsNotExist(err) || filepath.Dir(root) == root {
			t.Fatal("installed summarize repository unavailable", err)
		}
		root = filepath.Dir(root)
	}
	// Node 读取自身包管理器安装的链接，保留真实文件系统错误。
	resolveCLI := exec.Command(node, "-p", "require('node:fs').realpathSync(process.argv[1])", filepath.Join(root, "node_modules", "@steipete", "summarize", "dist", "cli.js"))
	resolveCLI.Env = (&SummarizeExtractor{NodeExecutable: node}).isolatedEnv(t.TempDir(), false)
	cliBytes, err := resolveCLI.Output()
	if err != nil {
		t.Fatal("resolve installed summarize CLI", root, err)
	}
	cli := strings.TrimSpace(string(cliBytes))
	e, err := NewSummarizeExtractor(node, cli)
	if err != nil {
		t.Fatal("initialize installed extractor", node, cli, err)
	}
	return e
}
func TestExtractorSubprocessIsolationAndFixedRevision(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "synthetic-not-real")
	t.Setenv("GROQ_API_KEY", "synthetic-not-real")
	t.Setenv("NODE_OPTIONS", "--require nonexisting-sentinel")
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:1")
	e := installedExtractor(t)
	for _, value := range e.isolatedEnv(t.TempDir(), false) {
		for _, key := range []string{"OPENAI_API_KEY=", "GROQ_API_KEY=", "NODE_OPTIONS=", "HTTP_PROXY="} {
			if strings.HasPrefix(value, key) {
				t.Fatal("child inherited unapproved configuration", key)
			}
		}
		if strings.HasPrefix(value, "PATH=") && strings.Contains(value, "ShadowBot") {
			t.Fatal("child inherited unrelated PATH tools")
		}
	}
	r, err := e.Extract(context.Background(), "https://arxiv.org/html/2501.12948v1")
	if err != nil || len(r.Text) < 1000 || r.URL != "https://arxiv.org/html/2501.12948v1" || r.Version != "0.25.1" || !r.HasOriginal || len(r.Original) == 0 {
		t.Fatal("actual fixed-revision extraction missing source, original or version", err)
	}
	for _, bad := range []string{"https://arxiv.org/html/2504.16054", "https://evil.test/html/2504.16054v1", "https://arxiv.org/abs/2504.16054v1"} {
		if _, err = e.Extract(context.Background(), bad); err == nil {
			t.Fatal("accepted nonfixed official HTML", bad)
		}
	}
}
func TestExtractorPhysicalCancellationAndOutputBound(t *testing.T) {
	e := installedExtractor(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := e.Extract(ctx, "https://arxiv.org/html/2504.16054v1")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 4*time.Second {
		t.Fatal("cancellation not bounded", err)
	}
	var cap cappedOutput
	if _, err = cap.Write(make([]byte, 16<<20)); err != nil || cap.Len() != 16<<20 {
		t.Fatal("valid output boundary rejected", err)
	}
	if _, err = cap.Write([]byte{0}); err == nil || cap.Len() != 16<<20 {
		t.Fatal("accepted accumulated oversized output")
	}
	cap.Reset()
	if _, err = cap.Write(make([]byte, (16<<20)+1)); err == nil || cap.Len() != 0 {
		t.Fatal("unbounded output writer")
	}
}
func TestExtractorVersionMismatchFailsClosed(t *testing.T) {
	e := installedExtractor(t)
	e.Options.Version = "0.21.8"
	_, err := e.Extract(context.Background(), "https://arxiv.org/html/2504.16054v1")
	if err == nil || !strings.Contains(err.Error(), "pin") {
		t.Fatal("version mismatch accepted", err)
	}
}
func TestVideoURLSinglePublicSourceAndLegacySemantics(t *testing.T) {
	for _, valid := range []string{"https://www.bilibili.com/video/BV1PReT6EEqR/", "https://youtu.be/dQw4w9WgXcQ", "https://www.youtube.com/watch?v=dQw4w9WgXcQ"} {
		if err := validateVideoURL(valid); err != nil {
			t.Fatal(err)
		}
	}
	for _, bad := range []string{"file:///etc/passwd", "https://127.0.0.1/video/BV1PReT6EEqR/", "https://user:pass@www.bilibili.com/video/BV1PReT6EEqR/", "https://www.bilibili.com/video/BV1PReT6EEqR/?p=2", "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=private", "https://www.bilibili.com/search?q=x", "https://evil.test/video.mp4"} {
		if err := validateVideoURL(bad); err == nil {
			t.Fatal("unsafe/multipart/playlist URL reached media", bad)
		}
	}
	r := NewReader(nil)
	_, err := r.ReadSource(context.Background(), appVideoCommand("summarize_url"))
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != apierrors.ProviderUnavailable {
		t.Fatal("disabled URL backend fake success", err)
	}
	_, err = r.ReadSource(context.Background(), appVideoCommand("summarize"))
	if !errors.As(err, &service) || service.Code != apierrors.EvidenceMissing {
		t.Fatal("legacy export request started URL media", err)
	}
}

func TestUpstreamMediaResultsRemainSeparateFromCLIAndSummary(t *testing.T) {
	locator := "https://www.bilibili.com/video/BV1PReT6EEqR/"
	e := &SummarizeExtractor{Options: SummarizeOptions{Version: "0.25.1"}}
	for _, timed := range []bool{false, true} {
		cli, media, stderr := videoActualInputs(t, timed)
		original, err := parseSummarizeJSON(cli, true)
		if err != nil {
			t.Fatal(err)
		}
		var actual struct {
			Text     string
			Segments []Segment
		}
		if err := json.Unmarshal(media, &actual); err != nil {
			t.Fatal(err)
		}
		wire := videoActualWire(t, cli, media, stderr)
		got, err := e.decodeVideoResult(locator, wire, "", nil)
		if err != nil || got.Title != original.Title || got.Text != actual.Text || got.Summary != "" || got.Mode != "upstream_media_transcript" || string(got.Original) != string(media) || len(got.ExtraAttachments) != 2 || string(got.ExtraAttachments[0].Data) != string(cli) || string(got.ExtraAttachments[1].Data) != string(stderr) {
			t.Fatal("actual source, separate originals or transcript changed", err)
		}
		if !timed && len(got.Segments) != 0 {
			t.Fatal("invented timeline")
		}
		if timed && len(got.Segments) == 0 {
			t.Fatal("lost actual native timing")
		}
		if len(got.Segments) != len(actual.Segments) {
			t.Fatal("actual segment count changed")
		}
		for index, segment := range got.Segments {
			if segment.StartMS != actual.Segments[index].StartMS || segment.Text != actual.Segments[index].Text || segment.EndMS == nil || *segment.EndMS != *actual.Segments[index].EndMS {
				t.Fatal("actual segment timing or text changed", index)
			}
		}
	}
	cli, media, stderr := videoActualInputs(t, true)
	for _, corrupt := range []func(map[string]any){
		func(value map[string]any) { value["provider"] = "openai" },
		func(value map[string]any) { delete(value["segments"].([]any)[0].(map[string]any), "startMs") },
		func(value map[string]any) { value["text"] = nil },
	} {
		var value map[string]any
		if err := json.Unmarshal(media, &value); err != nil {
			t.Fatal(err)
		}
		corrupt(value)
		invalidMedia, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		wire := videoActualWire(t, cli, invalidMedia, stderr)
		if _, err := e.decodeVideoResult(locator, wire, "", nil); err == nil {
			t.Fatal("accepted corrupted provider, timing or text")
		}
	}
	_, err := e.decodeVideoResult(locator, videoActualWire(t, cli, media, stderr), "", context.Canceled)
	var failed *ExtractionFailure
	if !errors.As(err, &failed) || string(failed.CLIJSON) != string(cli) || string(failed.CLIStderr) != string(stderr) || string(failed.MediaJSON) != string(media) {
		t.Fatal("original failure discarded", err)
	}
	if extractionAction("blocked local network address", "check_tools") != "configure_public_source_network" {
		t.Fatal("network failure action changed")
	}
}
