package distillers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	_ "modernc.org/sqlite"
)

// This captures the actual adapter's assembled request, then stops before any
// CLI/model call. It verifies source fidelity/capacity, not model acceptance.
type capacityCaptureProcessor struct{ request app.SelectedTextRequest }

func (*capacityCaptureProcessor) ConfigurationID(context.Context, app.Principal, string, string) (string, error) {
	return "capacity-only-no-model", nil
}
func (*capacityCaptureProcessor) ProjectSpaceID(context.Context, app.Principal, string, string) (string, error) {
	return "capacity-only-no-model", nil
}
func (p *capacityCaptureProcessor) ProcessSelectedText(_ context.Context, _ app.Principal, request app.SelectedTextRequest) (app.SelectedTextResult, error) {
	p.request = request
	return app.SelectedTextResult{}, context.Canceled
}

func verifyCombinedPrompt(t *testing.T, inputs []app.SourceSnapshot) {
	t.Helper()
	processor := &capacityCaptureProcessor{}
	distiller, err := NewSelectedTextFactory(processor, 30*time.Minute).Resolve(context.Background(), app.Principal{ID: "capacity-check", Kind: "human"}, "capacity-only", "capacity-only")
	if err != nil {
		t.Fatal(err)
	}
	_, err = distiller.Distill(context.Background(), app.DistillationInput{JobID: "capacity-only", OperationID: "not-started", Inputs: inputs, Stage: "association", Question: "Use both complete selected source versions; retain uncertainty."})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("request did not reach bounded capture before model call", err)
	}
	request := processor.request
	if len(request.Prompt) <= 128*1024 || request.OutputLimit != 128*1024 || selectedNativePromptPreflight(request.Prompt) != nil {
		t.Fatalf("expanded input/output boundary incorrect: prompt=%d output=%d", len(request.Prompt), request.OutputLimit)
	}
	_, data, found := strings.Cut(request.Prompt, "\nINPUT_DATA:\n")
	var decoded app.DistillationInput
	if !found || json.Unmarshal([]byte(data), &decoded) != nil || len(decoded.Inputs) != len(inputs) {
		t.Fatal("assembled input lost schema/reference payload")
	}
	for i, source := range inputs {
		if decoded.Inputs[i].Text != source.Text || decoded.Inputs[i].ContentDigest != source.ContentDigest || !sameRef(decoded.Inputs[i].Ref, source.Ref) {
			t.Fatal("source text/digest/ref truncated or changed", i)
		}
	}
	t.Logf("complete caller prompt=%d UTF-8 bytes; native envelope preflight PASS; output=%d; no model call", len(request.Prompt), request.OutputLimit)
}

func TestCombinedPromptPreservesCompleteSelectedText(t *testing.T) {
	inputs := []app.SourceSnapshot{{Ref: app.SourceRef{MaterialID: "paper", Revision: 1}, Text: strings.Repeat("Paper observation with evidence. ", 4000)}, {Ref: app.SourceRef{MaterialID: "video", Revision: 1}, Text: strings.Repeat("视频正文与明确未知。", 2500)}}
	verifyCombinedPrompt(t, inputs)
}

func TestActualKnownPaperVideoCapacityReadOnly(t *testing.T) {
	root := os.Getenv("ASTROCYTE_CAPACITY_SOURCE_ROOT")
	if root == "" {
		t.Skip("read-only existing source store not explicitly provided")
	}
	// Only this user-authorized old acceptance store is inspected. No schema
	// migration, transaction, service start, source refresh or model call.
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(root, "state.sqlite"))+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT data FROM attention_materials")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	inputs := []app.SourceSnapshot{}
	for rows.Next() {
		var raw string
		if err = rows.Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var material app.MaterialDetail
		if err = json.Unmarshal([]byte(raw), &material); err != nil {
			t.Fatal(err)
		}
		for _, revision := range material.Revisions {
			if revision.Revision != material.Material.CurrentRevision || !(revision.SourceKey == "arxiv:2504.16054" || strings.Contains(revision.SourceKey, "BV1PReT6EEqR")) {
				continue
			}
			if len(revision.ObjectRef) != 64 || strings.ContainsAny(revision.ObjectRef, "/\\") {
				t.Fatal("unsafe object ref")
			}
			text, err := os.ReadFile(filepath.Join(root, "objects", revision.ObjectRef))
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(text)
			if hex.EncodeToString(digest[:]) != revision.ContentDigest {
				t.Fatal("existing immutable source digest mismatch")
			}
			inputs = append(inputs, app.SourceSnapshot{Ref: app.SourceRef{MaterialID: revision.MaterialID, Revision: revision.Revision, Locator: revision.SourceLocator}, SourceKey: revision.SourceKey, Title: material.Material.Title, Text: string(text), Summary: revision.Summary, ContentDigest: revision.ContentDigest, Provenance: revision.Provenance})
			t.Logf("read-only source=%s revision=%d bytes=%d digest=%s", revision.SourceKey, revision.Revision, len(text), revision.ContentDigest)
		}
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(inputs) != 2 {
		t.Fatal(fmt.Sprintf("expected exactly the two approved source heads, got %d", len(inputs)))
	}
	verifyCombinedPrompt(t, inputs)
}
