package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestFixedArxivImportReusesJobAndPreservesVersions(t *testing.T) {
	ctx := context.Background()
	s, repo, objects := fixture(t)
	calls := 0
	s.sources = sourceFunc(func(_ context.Context, c ImportMaterialCommand) (ImportedSource, error) {
		calls++
		version := "v1"
		if strings.HasSuffix(c.SourceLocator, "v2") {
			version = "v2"
		}
		return ImportedSource{SourceKey: "arxiv:2401.01234", SourceLocator: c.SourceLocator, Kind: "paper", Text: "paper " + version,
			Attachments: []SourceAttachment{{Name: "original.pdf", MediaType: "application/pdf", Data: []byte("PDF " + version), SourceLocator: c.SourceLocator}}}, nil
	})
	c := ImportMaterialCommand{CommandMeta: meta("fixed-first", 1), Kind: "paper", SourceLocator: "https://arxiv.org/abs/2401.01234v1", Adapter: "arxiv"}
	first, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	firstJob, err := s.GetJob(ctx, human, first.JobID)
	if err != nil || firstJob.Status != "succeeded" || firstJob.MaterialID == nil || firstJob.MaterialRevision == nil || *firstJob.MaterialRevision != 1 {
		t.Fatalf("first import: %+v, %v", firstJob, err)
	}
	before, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil {
		t.Fatal(err)
	}
	publications := objects.publishes.Load()
	c.CommandMeta = meta("fixed-new-form", 1)
	second, err := s.ImportMaterial(ctx, human, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	if second.JobID != first.JobID || second.Status != "succeeded" || calls != 1 || len(repo.state.Jobs) != 1 {
		t.Fatalf("fixed revision repeated source work: same_job=%t status=%s reader_calls=%d jobs=%d", second.JobID == first.JobID, second.Status, calls, len(repo.state.Jobs))
	}
	after, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil || !reflect.DeepEqual(before, after) || objects.publishes.Load() != publications {
		t.Fatal("reuse modified material history or published objects again", err)
	}
	// A receipt retains its original response, even after the job completed.
	original := c
	original.CommandMeta = meta("fixed-first", 1)
	replayed, err := s.ImportMaterial(ctx, human, original)
	if err != nil || replayed != first {
		t.Fatal("original queued receipt changed", err)
	}
	original.SourceLocator = "https://arxiv.org/abs/2401.01234v2"
	_, err = s.ImportMaterial(ctx, human, original)
	errorCode(t, err, apierrors.VersionConflict)
	c.CommandMeta = meta("fixed-v2", 1)
	c.SourceLocator = "https://arxiv.org/abs/2401.01234v2"
	next, err := s.ImportMaterial(ctx, human, c)
	if err != nil || next.JobID == first.JobID {
		t.Fatal("explicit v2 did not create a new job", err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	nextJob, err := s.GetJob(ctx, human, next.JobID)
	if err != nil || nextJob.Status != "succeeded" || nextJob.MaterialID == nil || *nextJob.MaterialID != *firstJob.MaterialID || nextJob.MaterialRevision == nil || *nextJob.MaterialRevision != 2 {
		t.Fatalf("v2 must append to the same material: %+v, %v", nextJob, err)
	}
	after, err = s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil || calls != 2 || len(repo.state.Jobs) != 2 || len(repo.state.Materials) != 1 || len(after.Revisions) != 2 || after.Material.CurrentRevision != 2 || !reflect.DeepEqual(after.Revisions[0], before.Revisions[0]) {
		t.Fatal("v2 replaced history or duplicated material", err)
	}
	oldContent, err := s.GetContent(ctx, human, *firstJob.MaterialID, 1)
	if err != nil || oldContent.Text != "paper v1" {
		t.Fatal("old source text lost", err)
	}
	oldPDF, err := s.GetAttachment(ctx, human, *firstJob.MaterialID, 1, "original.pdf")
	if err != nil || string(oldPDF.Data) != "PDF v1" {
		t.Fatal("old original attachment lost", err)
	}
	// Returning to an older fixed job must not decide the pending A-B-A head rule.
	c.CommandMeta = meta("fixed-v1-again", 1)
	c.SourceLocator = "https://arxiv.org/abs/2401.01234v1"
	reused, err := s.ImportMaterial(ctx, human, c)
	if err != nil || reused.JobID != first.JobID {
		t.Fatal("older fixed job was replaced", err)
	}
	final, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
	if err != nil || !reflect.DeepEqual(final, after) || calls != 2 {
		t.Fatal("job reuse changed the current source head", err)
	}
}

func TestImportFixedArxivBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, locator, adapter, kind, file string
		fixed                              bool
	}{
		{name: "implicit arxiv", locator: "https://arxiv.org/abs/2401.01234v1", kind: "paper", fixed: true},
		{name: "official pdf", locator: "https://arxiv.org/pdf/2401.01234v1.pdf", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "official html", locator: "https://www.arxiv.org/html/2401.01234v1?tracking=1#s2", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "official export", locator: "http://export.arxiv.org/abs/2401.01234v1", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "bare fixed ID", locator: "2401.01234v12", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "prefixed legacy ID", locator: "arXiv:hep-th/9901001v2", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "legacy category ID", locator: "https://arxiv.org/abs/math.GT/0309136v1", adapter: "arxiv", kind: "paper", fixed: true},
		{name: "unversioned arxiv", locator: "https://arxiv.org/abs/2401.01234", adapter: "arxiv", kind: "paper"},
		{name: "unversioned with query", locator: "https://arxiv.org/abs/2401.01234?version=v1", adapter: "arxiv", kind: "paper"},
		{name: "nonofficial versioned URL", locator: "https://example.test/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "host suffix", locator: "https://arxiv.org.example.test/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "userinfo", locator: "https://reader@arxiv.org/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "invalid scheme", locator: "ftp://arxiv.org/abs/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "unrecognized path", locator: "https://arxiv.org/other/2401.01234v1", adapter: "arxiv", kind: "paper"},
		{name: "invalid ID", locator: "https://arxiv.org/abs/arbitraryv1", adapter: "arxiv", kind: "paper"},
		{name: "invalid version zero", locator: "https://arxiv.org/abs/2401.01234v0", adapter: "arxiv", kind: "paper"},
		{name: "invalid version leading zero", locator: "https://arxiv.org/abs/2401.01234v01", adapter: "arxiv", kind: "paper"},
		{name: "manual source", locator: "https://arxiv.org/abs/2401.01234v1", adapter: "manual", kind: "paper"},
		{name: "summarize local export", locator: "https://arxiv.org/abs/2401.01234v1", adapter: "summarize_json", kind: "paper", file: "mutable.json"},
		{name: "implicit local file", locator: "https://arxiv.org/abs/2401.01234v1", kind: "paper", file: "mutable.md"},
		{name: "implicit text", locator: "https://arxiv.org/abs/2401.01234v1", kind: "text"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, _, _ := fixture(t)
			c := ImportMaterialCommand{CommandMeta: meta("one", 1), SourceLocator: tc.locator, Adapter: tc.adapter, Kind: tc.kind, LocalFileRef: tc.file}
			first, err := s.ImportMaterial(context.Background(), human, c)
			if err != nil {
				t.Fatal(err)
			}
			c.CommandMeta = meta("two", 1)
			second, err := s.ImportMaterial(context.Background(), human, c)
			if err != nil || first.JobID != second.JobID || (fixedArxivVersion(c) != "") != tc.fixed {
				t.Fatalf("fixed=%t same_job=%t error=%v", tc.fixed, first.JobID == second.JobID, err)
			}
		})
	}
}

func TestFixedArxivAliasesReuseSingleSourceRead(t *testing.T) {
	ctx := context.Background()
	s, repo, _ := fixture(t)
	calls := 0
	s.sources = sourceFunc(func(context.Context, ImportMaterialCommand) (ImportedSource, error) {
		calls++
		return ImportedSource{SourceKey: "arxiv:2401.01234", SourceLocator: "https://arxiv.org/abs/2401.01234v1", Kind: "paper", Text: "fixed paper bytes"}, nil
	})
	aliases := []string{
		"2401.01234v1", "arXiv:2401.01234v1",
		"https://arxiv.org/abs/2401.01234v1", "https://arxiv.org/pdf/2401.01234v1",
		"https://arxiv.org/pdf/2401.01234v1.pdf", "https://arxiv.org/html/2401.01234v1",
		"https://www.arxiv.org/abs/2401.01234v1?tracking=1#section", "http://export.arxiv.org/abs/2401.01234v1",
	}
	var first ImportJobResult
	for i, locator := range aliases {
		c := ImportMaterialCommand{CommandMeta: meta(locator, 1), Kind: "paper", SourceLocator: locator}
		if i%2 != 0 {
			c.Adapter = "arxiv"
		}
		result, err := s.ImportMaterial(ctx, human, c)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = result
		}
		if result != first {
			t.Fatalf("equivalent fixed source created another job: %s", locator)
		}
	}
	if _, err := s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	c := ImportMaterialCommand{CommandMeta: meta("completed-alias", 1), Kind: "paper", SourceLocator: aliases[4], Adapter: "arxiv"}
	completed, err := s.ImportMaterial(ctx, human, c)
	if err != nil || completed.JobID != first.JobID || completed.Status != "succeeded" {
		t.Fatal("alias did not reuse completed result", err)
	}
	if worked, err := s.ProcessNextJob(ctx); err != nil || worked || calls != 1 || len(repo.state.Jobs) != 1 || len(repo.state.Materials) != 1 || len(repo.state.Receipts) != len(aliases)+1 {
		t.Fatalf("aliases duplicated work or lost receipts: calls=%d jobs=%d materials=%d error=%v", calls, len(repo.state.Jobs), len(repo.state.Materials), err)
	}
}

func TestMutableImportCommandsReadUpdatedContent(t *testing.T) {
	for _, local := range []bool{false, true} {
		name := "unversioned remote"
		if local {
			name = "local export with fixed remote locator"
		}
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			s, _, _ := fixture(t)
			text := "first source bytes"
			calls := 0
			file := filepath.Join(t.TempDir(), "export.md")
			if err := os.WriteFile(file, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			s.sources = sourceFunc(func(_ context.Context, c ImportMaterialCommand) (ImportedSource, error) {
				calls++
				actual := text
				if c.LocalFileRef != "" {
					data, err := os.ReadFile(c.LocalFileRef)
					if err != nil {
						return ImportedSource{}, err
					}
					actual = string(data)
				}
				return ImportedSource{SourceKey: "arxiv:2401.01234", SourceLocator: c.SourceLocator, Kind: "paper", Text: actual}, nil
			})
			c := ImportMaterialCommand{CommandMeta: meta("mutable-first", 1), Kind: "paper", SourceLocator: "https://arxiv.org/abs/2401.01234"}
			if local {
				c.SourceLocator += "v1"
				c.LocalFileRef = file
			}
			first, err := s.ImportMaterial(ctx, human, c)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.ProcessNextJob(ctx); err != nil {
				t.Fatal(err)
			}
			firstJob, err := s.GetJob(ctx, human, first.JobID)
			if err != nil || firstJob.MaterialID == nil {
				t.Fatal("first source was not imported", err)
			}
			text = "updated source bytes"
			if err = os.WriteFile(file, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			c.CommandMeta = meta("mutable-again", 1)
			c.Refresh = true
			second, err := s.ImportMaterial(ctx, human, c)
			if err != nil || second.JobID == first.JobID {
				t.Fatal("new command failed to schedule a mutable source read", err)
			}
			if _, err = s.ProcessNextJob(ctx); err != nil {
				t.Fatal(err)
			}
			row, err := s.GetMaterial(ctx, human, *firstJob.MaterialID)
			if err != nil || calls != 2 || len(row.Revisions) != 2 || row.Material.CurrentRevision != 2 {
				t.Fatal("mutable update failed to append revision", err)
			}
			old, err := s.GetContent(ctx, human, *firstJob.MaterialID, 1)
			if err != nil || old.Text != "first source bytes" {
				t.Fatal("mutable update lost prior content", err)
			}
			current, err := s.GetContent(ctx, human, *firstJob.MaterialID, 2)
			if err != nil || current.Text != text {
				t.Fatal("new revision is not the updated source bytes", err)
			}
		})
	}
}
