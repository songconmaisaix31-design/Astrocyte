package distillers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestCodexProcessFailureChild(t *testing.T) {
	if os.Getenv("ASTROCYTE_PROCESS_FAILURE_CHILD") == "1" {
		os.Exit(7)
	}
	if os.Getenv("ASTROCYTE_PROCESS_FAILURE_CHILD") == "sleep" {
		time.Sleep(time.Hour)
		os.Exit(0)
	}
}

// Real native subprocess failure/cancellation, without any model submission.
func TestStartedProcessLossIsUnknownAndUnstartedIsUnavailable(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"1", "sleep"} {
		ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
		cmd := exec.CommandContext(ctx, executable, "-test.run=^TestCodexProcessFailureChild$")
		cmd.Env = append(os.Environ(), "ASTROCYTE_PROCESS_FAILURE_CHILD="+mode)
		var stderr boundedOutput
		cmd.Stderr = &stderr
		err := runProcessor(ctx, cmd, &stderr)
		cancel()
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != apierrors.DeliveryUnknown || service.Retryable {
			t.Fatalf("started process loss invited retry: %v", err)
		}
	}
	cmd := exec.Command(filepath.Join(t.TempDir(), "unavailable.exe"))
	var stderr boundedOutput
	cmd.Stderr = &stderr
	err = runProcessor(context.Background(), cmd, &stderr)
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != apierrors.ProviderUnavailable {
		t.Fatalf("unstarted process incorrectly unknown: %v", err)
	}
}

func TestRejectUnselectedSnapshotsBeforeCLI(t *testing.T) {
	c := &Codex{options: CodexOptions{AllowedSourceKeys: []string{"arxiv:2504.16054"}}}
	for _, source := range []app.SourceSnapshot{{SourceKey: "private:unselected", Text: "synthetic"}, {SourceKey: "arxiv:2504.16054", Text: "synthetic", ContentDigest: "forged"}} {
		_, err := c.Distill(context.Background(), app.DistillationInput{Inputs: []app.SourceSnapshot{source}})
		var e *apierrors.ServiceError
		if !errors.As(err, &e) || e.Code != apierrors.ScopeDenied {
			t.Fatalf("unauthorized input reached CLI: %v", err)
		}
	}
}

func TestStructuredOutputDoesNotTrustModelRefsOrProvenance(t *testing.T) {
	ref := app.SourceRef{MaterialID: "selected", Revision: 1, Locator: "https://arxiv.org/html/2504.16054v1"}
	out := app.DistillationOutput{OutputText: "evidence-bound output", RelatedRefs: []app.SourceRef{ref}, RelatedIdeas: []string{}, Conflicts: []string{}, PendingQuestions: []string{}, GoalRefs: []string{}, ExistingAssets: []string{}, MissingEvidence: []string{}}
	raw, _ := json.Marshal(out)
	var fields map[string]any
	_ = json.Unmarshal(raw, &fields)
	delete(fields, "provenance")
	raw, _ = json.Marshal(fields)
	if _, err := decodeOutput(raw, []app.SourceSnapshot{{Ref: ref}}); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { v["provenance"] = map[string]string{"processor": "forged"} },
		func(v map[string]any) { v["goal_refs"] = nil },
		func(v map[string]any) { delete(v, "missing_evidence") },
		func(v map[string]any) {
			v["related_refs"] = []app.SourceRef{{MaterialID: "unselected", Revision: 1, Locator: "https://private.invalid"}}
		},
	} {
		var v map[string]any
		_ = json.Unmarshal(raw, &v)
		mutate(v)
		bad, _ := json.Marshal(v)
		if _, err := decodeOutput(bad, []app.SourceSnapshot{{Ref: ref}}); err == nil {
			t.Fatal("untrusted output accepted")
		}
	}
}

// Explicit opt-in uses only the supplied public full HTML extraction, never
// reads a repository/database/object directory or chooses a private source.
func TestCodexLiveSelectedPublicPaper(t *testing.T) {
	executable := os.Getenv("ASTROCYTE_TEST_CODEX_EXE")
	textPath := os.Getenv("ASTROCYTE_TEST_PUBLIC_ARXIV_TEXT")
	if executable == "" || textPath == "" {
		t.Skip("set native Codex path and approved public arXiv extracted-text path")
	}
	text, err := os.ReadFile(textPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(text) < 5000 {
		t.Fatal("full original public text required")
	}
	digest := sha256.Sum256(text)
	c, err := NewCodex(CodexOptions{Executable: executable, WorkRoot: t.TempDir(), Model: "gpt-6.1-sol", AllowedSourceKeys: []string{"arxiv:2504.16054"}})
	if err != nil {
		t.Fatal(err)
	}
	config, err := c.ConfigurationID(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	input := app.DistillationInput{JobID: "public-paper-live", OperationID: "public-paper-live-operation", Stage: "content", ProcessingConfig: config, Question: "根据真实论文原文，解释π0.5的关键方法、实际证据及局限；区分实验结果与可迁移的项目想法，未知目标保持未知。", Inputs: []app.SourceSnapshot{{Ref: app.SourceRef{MaterialID: "selected-public-paper", Revision: 1, Locator: "https://arxiv.org/html/2504.16054v1"}, SourceKey: "arxiv:2504.16054", Title: "π0.5: a Vision-Language-Action Model with Open-World Generalization", Text: string(text), ContentDigest: hex.EncodeToString(digest[:]), Provenance: app.Provenance{Processor: "arxiv+summarize", Version: "2504.16054v1; summarize 0.21.8", Mode: "official_atom_pdf_and_html_text", Source: "https://arxiv.org/html/2504.16054v1"}}}}
	output, err := c.Distill(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if output.Provenance.Processor != "codex-cli" || output.Provenance.Model != "gpt-6.1-sol" || output.OutputText == "" {
		t.Fatal("missing actual processor output/provenance")
	}
	t.Logf("actual processor=%s version=%s model=%s mode=%s output_chars=%d", output.Provenance.Processor, output.Provenance.Version, output.Provenance.Model, output.Provenance.Mode, len(output.OutputText))
	if dir := os.Getenv("ASTROCYTE_TEST_PROCESSOR_EXPORT_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.MarshalIndent(output, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "codex-public-paper-output.json"), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}
