package app

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

type scopedTestDistiller struct {
	*testDistiller
	spaceID string
}

func (s *scopedTestDistiller) ProjectSpaceID(context.Context) (string, error) { return s.spaceID, nil }

type scopedTestFactory struct {
	t         *testing.T
	processor *scopedTestDistiller
	denied    bool
}

func (f *scopedTestFactory) Resolve(_ context.Context, p Principal, project, cli string) (Distiller, error) {
	if p != human || project != "selected-project" || cli != "selected-cli" {
		f.t.Fatal("factory did not preserve real caller selection")
	}
	if f.denied {
		return nil, serviceError(apierrors.ScopeDenied, "project model permission revoked", "review_project_permission")
	}
	return f.processor, nil
}

func TestProjectAutomaticChecksFixedMembershipBeforeObjectsAndNeverFallsBack(t *testing.T) {
	ctx := context.Background()
	s, repo, objects := fixture(t)
	m := importFixture(t, s, "project-body", "private selected fixed body")
	space, err := s.CreateProjectSpace(ctx, human, ProjectSpaceCommand{CommandMeta: meta("project-space", 1), Title: "Project"})
	if err != nil {
		t.Fatal(err)
	}
	d := &testDistiller{config: "contract-local-selected-project"}
	d.call = func(_ context.Context, input DistillationInput) (DistillationOutput, error) {
		if input.Inputs[0].Text != "private selected fixed body" {
			t.Fatal("wrong fixed selected body")
		}
		return DistillationOutput{OutputText: "contract-local project summary", Provenance: Provenance{Processor: "selected-cli", Version: "test", Mode: "contract_local"}}, nil
	}
	f := &scopedTestFactory{t: t, processor: &scopedTestDistiller{testDistiller: d, spaceID: space.Space.ID}}
	s.options.ProjectDistillers = f
	c := autoCommand(m, "before-reference")
	c.ProjectID = "selected-project"
	c.CLI = "selected-cli"
	c.ProcessingConfig = ""
	_, err = s.RequestDistillation(ctx, human, c)
	errorCode(t, err, apierrors.ScopeDenied)
	space, err = s.ReferenceMaterial(ctx, human, space.Space.ID, ReferenceMaterialCommand{CommandMeta: meta("project-include", space.Space.Version), MaterialID: m.Material.ID, Revision: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.CommandMeta = meta("project-automatic", 1)
	job := runAutomatic(t, s, c)
	if job.Status != "succeeded" || d.calls != 1 || len(repo.state.Distillations) != 1 {
		t.Fatal("project-selected processing did not run", job)
	}
	// Permission is rechecked on a fresh request; the legacy route is unavailable
	// and cannot replace a denied selected processor.
	f.denied = true
	c.CommandMeta = meta("project-revoked", 1)
	c.Question = "new question"
	_, err = s.RequestDistillation(ctx, human, c)
	errorCode(t, err, apierrors.ScopeDenied)
	if d.calls != 1 {
		t.Fatal("revoked permission fell back")
	}
	_ = objects
}
