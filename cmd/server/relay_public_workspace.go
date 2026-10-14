package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayPublicWorkspace struct {
	workspace.RelayWorkspaces
	approved os.FileInfo
}

func relayWorkspaceProvider(root string) (workspace.RelayWorkspaces, error) {
	selected := os.Getenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT")
	if selected == "" {
		return relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{}), nil
	}
	canonical, err := workbenchCanonical(selected)
	if err != nil || !filepath.IsAbs(selected) || !relayInside(filepath.Join(root, "verification"), canonical) {
		return nil, errors.New("public verification project must remain inside isolated relay verification state")
	}
	info, err := os.Stat(canonical)
	if err != nil || !info.IsDir() {
		return nil, errors.New("approved public verification project unavailable")
	}
	return &relayPublicWorkspace{RelayWorkspaces: relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true}), approved: info}, nil
}

func (s *relayPublicWorkspace) Prepare(ctx context.Context, work domain.Work, state domain.RelayWorkspace) (domain.RelayWorkspace, error) {
	canonical, err := workbenchCanonical(work.Root)
	if err != nil {
		return state, &apierrors.ServiceError{Code: apierrors.ScopeDenied}
	}
	info, err := os.Stat(canonical)
	if err != nil || !os.SameFile(info, s.approved) {
		return state, &apierrors.ServiceError{Code: apierrors.ScopeDenied}
	}
	return s.RelayWorkspaces.Prepare(ctx, work, state)
}
