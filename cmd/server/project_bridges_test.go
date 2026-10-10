package main

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type modelProjectProbe struct {
	workspaceapp.LocalProjects
	revoked bool
	callers []domain.Caller
}

func (p *modelProjectProbe) CheckProjectModel(_ context.Context, c domain.Caller, project, cli string) (domain.LocalProject, error) {
	p.callers = append(p.callers, c)
	if p.revoked || project != "chosen-project" || cli != "chosen-cli" {
		return domain.LocalProject{}, &apierrors.ServiceError{Code: apierrors.ScopeDenied}
	}
	return domain.LocalProject{SpaceID: "chosen-space"}, nil
}

type selectedTextProbe struct{ configs, processes int }

func (p *selectedTextProbe) ConfigurationID(context.Context, string) (string, error) {
	p.configs++
	return "factual-native-config", nil
}
func (p *selectedTextProbe) ProcessSelectedText(context.Context, domain.TextRequest) (domain.TextResult, error) {
	p.processes++
	return domain.TextResult{Text: "processor-output", Version: "actual-version"}, nil
}
func TestProjectModelBridgeRechecksConsentAfterConfiguration(t *testing.T) {
	project := &modelProjectProbe{}
	processor := &selectedTextProbe{}
	b := &selectedTextBridge{projects: project, processor: processor}
	ctx := context.Background()
	p := attentionapp.Principal{ID: "actual-agent", Kind: "agent"}
	space, err := b.ProjectSpaceID(ctx, p, "chosen-project", "chosen-cli")
	if err != nil || space != "chosen-space" {
		t.Fatal(space, err)
	}
	if _, err = b.ConfigurationID(ctx, p, "chosen-project", "chosen-cli"); err != nil {
		t.Fatal(err)
	}
	project.revoked = true
	if _, err = b.ProcessSelectedText(ctx, p, attentionapp.SelectedTextRequest{ProjectID: "chosen-project", CLI: "chosen-cli", Prompt: "selected-only"}); err == nil {
		t.Fatal("revoked permission processed")
	}
	if processor.processes != 0 || processor.configs != 1 || len(project.callers) != 3 {
		t.Fatalf("consent bypass: %#v %#v", project, processor)
	}
	for _, caller := range project.callers {
		if caller.Kind != "agent" || caller.ID != "actual-agent" || caller.ProjectID != "chosen-project" {
			t.Fatalf("identity changed: %+v", caller)
		}
	}
}

type materialReferenceProbe struct {
	attentionapp.AttentionProjectReferences
	caller   attentionapp.Principal
	expanded bool
	space    string
}

func (p *materialReferenceProbe) ReadProjectReference(_ context.Context, c attentionapp.Principal, space string, ref attentionapp.SourceRef, expanded bool) (attentionapp.SourceSnapshot, error) {
	p.caller = c
	p.expanded = expanded
	p.space = space
	return attentionapp.SourceSnapshot{Ref: ref, Title: "source title", Text: "fixed text"}, nil
}
func TestReferenceBridgePreservesIdentityVersionAndExpansion(t *testing.T) {
	p := &materialReferenceProbe{}
	b := &projectReferenceBridge{attention: p}
	material, err := b.ReadSelected(context.Background(), domain.Caller{ID: "scoped-agent", Kind: "agent", ProjectID: "P"}, "S", domain.FixedReference{MaterialID: "M", Revision: 2}, true)
	if err != nil || p.caller.Kind != "agent" || p.caller.ID != "scoped-agent" || p.space != "S" || !p.expanded || material.Reference.MaterialID != "M" || material.Reference.Revision != 2 || material.Text != "fixed text" {
		t.Fatalf("wrong scope mapping: %+v %+v %v", p, material, err)
	}
}
