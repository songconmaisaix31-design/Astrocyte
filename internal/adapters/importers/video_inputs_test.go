package importers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func videoActualInputs(t *testing.T, timed bool) (cli, media, stderr []byte) {
	t.Helper()
	key := "ASTROCYTE_TEST_VIDEO_BASELINE_DIR"
	if timed {
		key = "ASTROCYTE_TEST_VIDEO_TIMED_DIR"
	}
	root := os.Getenv(key)
	if root == "" || !filepath.IsAbs(root) {
		t.Fatal("actual video original directory must be explicitly configured", key)
	}
	read := func(name string) []byte {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal("read actual video original", name, err)
		}
		return data
	}
	cli = read("summarize-cli-original.json")
	media = read("summarize-upstream-media.json")
	if timed {
		stderr = read("summarize-cli.stderr.txt")
	} else {
		var material app.MaterialDetail
		if err := json.Unmarshal(read("video-material.json"), &material); err != nil {
			t.Fatal(err)
		}
		for _, revision := range material.Revisions {
			for _, attachment := range revision.Attachments {
				if attachment.Name != "summarize-cli.stderr.txt" {
					continue
				}
				digest, err := hex.DecodeString(attachment.ObjectRef)
				if err != nil || len(digest) != 32 {
					t.Fatal("invalid actual original object reference")
				}
				return cli, media, read(filepath.Join("objects", attachment.ObjectRef))
			}
		}
		t.Fatal("actual baseline CLI stderr attachment missing")
	}
	return cli, media, stderr
}

// 按生产桥接的字段映射重放已保存的实际原件，测试不启动替代 CLI。
func videoActualWire(t *testing.T, cli, media, stderr []byte) []byte {
	t.Helper()
	wire, err := json.Marshal(struct {
		CLIJSON   string          `json:"cliJSON"`
		CLIStderr string          `json:"cliStderr"`
		CLIExit   int             `json:"cliExit"`
		Media     json.RawMessage `json:"media"`
	}{CLIJSON: string(cli), CLIStderr: string(stderr), Media: media})
	if err != nil {
		t.Fatal(err)
	}
	return wire
}

func videoActualExport(t *testing.T, timed bool) Export {
	t.Helper()
	cli, media, stderr := videoActualInputs(t, timed)
	e := &SummarizeExtractor{Options: SummarizeOptions{Version: "0.25.1"}}
	export, err := e.decodeVideoResult("https://www.bilibili.com/video/BV1PReT6EEqR/", videoActualWire(t, cli, media, stderr), "", nil)
	if err != nil {
		t.Fatal("decode actual saved video originals", err)
	}
	return export
}

func TestVideoNativeOffsetsMatchSavedTranscriptAndSpans(t *testing.T) {
	export := videoActualExport(t, true)
	root := os.Getenv("ASTROCYTE_TEST_VIDEO_TIMED_DIR")
	raw, err := os.ReadFile(filepath.Join(root, "whisper-original.json"))
	if err != nil {
		t.Fatal(err)
	}
	var native struct {
		Transcription []struct {
			Offsets struct{ From, To int64 }
			Text    string
		}
	}
	if err := json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	spans := exportSpans(export)
	if len(native.Transcription) != 952 || len(native.Transcription) != len(export.Segments) || len(spans) != len(export.Segments) {
		t.Fatal("saved native time, parsed segment and source span counts differ")
	}
	var previousEnd int64
	for index, nativeSegment := range native.Transcription {
		segment := export.Segments[index]
		if nativeSegment.Offsets.From < previousEnd || nativeSegment.Offsets.To < nativeSegment.Offsets.From || float64(nativeSegment.Offsets.From) != segment.StartMS || segment.EndMS == nil || float64(nativeSegment.Offsets.To) != *segment.EndMS || strings.TrimSpace(nativeSegment.Text) != segment.Text || !strings.Contains(export.Text, segment.Text) {
			t.Fatal("actual native milliseconds or text ownership differs", index)
		}
		interval := fmt.Sprintf("%s [%.3fs–%.3fs]: ", export.URL, float64(nativeSegment.Offsets.From)/1000, float64(nativeSegment.Offsets.To)/1000)
		if spans[index] != interval+segment.Text {
			t.Fatal("source span differs from actual native interval/text", index)
		}
		previousEnd = nativeSegment.Offsets.To
	}
}
