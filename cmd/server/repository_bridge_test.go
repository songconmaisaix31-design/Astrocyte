package main

import (
	"context"
	"errors"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type repositorySpaceProbe struct {
	calls int
	err   error
}

func (p *repositorySpaceProbe) GetProjectSpace(_ context.Context, c attentionapp.Principal, id string) (attentionapp.ProjectSpaceResult, error) {
	p.calls++
	if c.Kind != "human" || c.ID != "human-one" || id != "selected-space" {
		return attentionapp.ProjectSpaceResult{}, errors.New("wrong human or selected space")
	}
	return attentionapp.ProjectSpaceResult{}, p.err
}

func TestRepositorySpaceBridgeCannotApproveAgentPlacement(t *testing.T) {
	p := &repositorySpaceProbe{}
	b := &repositorySpaceBridge{attention: p}
	for _, caller := range []domain.Caller{{ID: "agent", Kind: "agent"}, {Kind: "human"}, {}} {
		if err := b.ValidateRepositorySpace(context.Background(), caller, "selected-space"); err == nil || p.calls != 0 {
			t.Fatalf("non-human expanded placement authority: %v %+v", err, p)
		}
	}
	if err := b.ValidateRepositorySpace(context.Background(), domain.Caller{ID: "human-one", Kind: "human"}, "selected-space"); err != nil || p.calls != 1 {
		t.Fatal(err, p)
	}
	p.err = &apierrors.ServiceError{Code: apierrors.NotFound}
	if err := b.ValidateRepositorySpace(context.Background(), domain.Caller{ID: "human-one", Kind: "human"}, "selected-space"); !errors.Is(err, p.err) {
		t.Fatal("missing space error did not propagate", err)
	}
}
