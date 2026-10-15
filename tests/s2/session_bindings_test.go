package s2_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func verifyObservationBoundary(t *testing.T, ctx context.Context, repo *sqlite.WorkspaceRepository, service *workspace.S2Service, human domain.Caller, grant domain.ApprovalGrant) {
	t.Helper()
	request, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	check := domain.GrantCheck{Caller: human, GrantID: grant.ID, Epoch: grant.Epoch, ProposalRevision: grant.ProposalRevision, ProjectID: grant.Scope.ProjectID, Root: grant.Scope.Roots[0], Action: workspace.OwnedPurposeClaim}
	err := repo.WithWorkspaceTx(request, func(tx workspace.WorkspaceTx) error {
		_, err := service.ObserveSessionBindings(tx.Context(), check, []string{uuid.NewString()})
		requireCode(t, err, apierrors.ValidationFailed)
		_, err = service.BuildContextPacket(tx.Context(), human, workspace.BuildContextPacketCommand{WorkspaceCommandMeta: workspaceMeta(0), GrantID: grant.ID, GrantEpoch: grant.Epoch, MaxBytes: 128 * 1024})
		requireCode(t, err, apierrors.ValidationFailed)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.Err() != nil {
		t.Fatal("transaction boundary rejection consumed the original request deadline")
	}
	for _, observation := range []*workspace.SessionBindingObservation{nil, {}} {
		err = service.WithObservedSessionBindings(request, check, observation, func(context.Context, domain.ApprovalGrant, []domain.SessionBinding) error {
			t.Fatal("unobserved binding authorized a callback")
			return nil
		})
		requireCode(t, err, apierrors.EvidenceMissing)
	}
	canceled, finish := context.WithCancel(ctx)
	finish()
	if err = repo.RequireWorkspaceObservation(canceled); !errors.Is(err, context.Canceled) {
		t.Fatal("actual observation guard did not preserve request cancellation")
	}
}
