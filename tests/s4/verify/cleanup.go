package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func verifyCleanup(ctx context.Context, db *sqlite.DB, service *repositorycontext.Service, human domain.Caller, qa, qb domain.RepositoryQuery) {
	checkGraph(ctx, service, human, qa)
	checkGraph(ctx, service, human, qb)
	_, err := service.CleanupRepositoryContext(ctx, 0)
	code(err, apierrors.BudgetExhausted)
	verifyCleanupTransaction(ctx, db, service)
	emit("cleanup_rejects_unbounded_pass_and_request_transaction")

	agent := domain.Caller{Kind: "agent", ID: "s4-cleanup-agent", ProjectID: qa.ProjectID}
	grant := domain.ProjectGrant{ProjectID: qa.ProjectID, AgentID: agent.ID, Actions: []string{"read_context"}}
	must(db.SaveGrant(ctx, grant))
	checkGraph(ctx, service, agent, qa)
	now := time.Now().UTC()
	grant.RevokedAt = &now
	must(db.SaveGrant(ctx, grant))
	_, err = service.SearchRepositoryCode(ctx, agent, qa)
	code(err, apierrors.ApprovalRevoked)
	cleaned, err := service.CleanupRepositoryContext(ctx, 100)
	must(err)
	assert(cleaned.RetiredIndexes == 0 && cleaned.DeletedGenerations == 0 && !cleaned.HasMore, "single Agent revocation removed shared project projections")
	checkGraph(ctx, service, human, qa)
	checkGraph(ctx, service, human, qb)
	emit("actual_agent_grant_revocation_preserves_other_authorized_readers")

	p, err := db.LoadRepositoryProject(ctx, qa.ProjectID)
	must(err)
	alias := p
	alias.ID = "s4-cleanup-same-worktree"
	alias.Name = "Public cleanup same worktree"
	alias.Settings.Revision = 1
	must(db.SaveProject(ctx, alias, 0))
	other := qa
	other.ProjectID = alias.ID
	_, err = service.UpdateRepositoryIndex(ctx, human, other, "index")
	must(err)
	_, err = service.UpdateRepositoryIndex(ctx, human, qa, "sync")
	must(err)
	_, err = service.RefreshRepositoryMap(ctx, human, qa)
	must(err)
	_, err = service.CaptureProjectSnapshot(ctx, human, qa.ProjectID, qa.Roots)
	must(err)
	current, err := service.SearchRepositoryCode(ctx, human, qa)
	must(err)
	currentPath := filepath.Join(*evidence, "projections", current.Index.Snapshot.WorktreeID, current.Index.Generation)
	before, err := directoryState(filepath.Join(*evidence, "projections"))
	must(err)
	expected := p.Settings.Revision
	p.Settings.Revision++
	p.Settings.AllowedSubdirs = []string{"cmd"}
	must(db.SaveProject(ctx, p, expected))
	_, err = service.SearchRepositoryCode(ctx, human, qa)
	code(err, apierrors.ScopeDenied)
	_, err = service.ReadRepositoryMap(ctx, human, qa)
	code(err, apierrors.ScopeDenied)
	emit("actual_project_scope_narrowing_immediately_blocks_saved_graph_and_map")

	cache := filepath.Join(*evidence, "projections")
	link := filepath.Join(cache, strings.Repeat("0", 64))
	target := filepath.Join(cache, current.Index.Snapshot.WorktreeID)
	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "cmd", "/c", "mklink", "/J", link, target)
		must(cmd.Run())
	} else {
		must(os.Symlink(target, link))
	}
	defer func() { _ = os.Remove(link) }()
	_, err = service.CleanupRepositoryContext(ctx, 1)
	code(err, apierrors.ScopeDenied)
	_, err = db.LoadRepositoryIndex(ctx, qa.ProjectID, current.Index.Snapshot.WorktreeID, "codegraph")
	code(err, apierrors.NotFound)
	_, err = db.LoadRepositoryIndex(ctx, qa.ProjectID, current.Index.Snapshot.WorktreeID, "aoci")
	code(err, apierrors.NotFound)
	_, err = os.Stat(currentPath)
	must(err)
	must(os.Remove(link))
	after, err := directoryState(filepath.Join(*evidence, "projections"))
	must(err)
	assert(before == after, "unsafe directory link allowed projection deletion")
	emit("real_directory_link_refused_after_metadata_commit_and_all_graph_files_preserved")

	cleaned, err = service.CleanupRepositoryContext(ctx, 1)
	must(err)
	assert(cleaned.DeletedGenerations == 1 && cleaned.HasMore, "bounded cleanup did not expose remaining generations")
	checkGraph(ctx, service, human, other)
	checkGraph(ctx, service, human, qb)
	cleaned, err = service.CleanupRepositoryContext(ctx, 100)
	must(err)
	assert(cleaned.DeletedGenerations == 1 && !cleaned.HasMore, "interrupted cleanup did not finish remaining old generation")
	_, err = os.Stat(currentPath)
	assert(errors.Is(err, os.ErrNotExist), "withdrawn source projection remains on disk")
	checkGraph(ctx, service, human, other)
	checkGraph(ctx, service, human, qb)
	emit("bounded_resumed_cleanup_removes_withdrawn_generations_preserves_same_worktree_other_project_and_worktree_b")
}

func verifyCleanupTransaction(ctx context.Context, db *sqlite.DB, service *repositorycontext.Service) {
	must(sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
		_, err := service.CleanupRepositoryContext(tx.Context(), 1)
		code(err, apierrors.ValidationFailed)
		return nil
	}))
}

func verifyCleanupRestart(ctx context.Context, db *sqlite.DB, service *repositorycontext.Service, human domain.Caller, qa, qb domain.RepositoryQuery) {
	verifyCleanupTransaction(ctx, db, service)
	before, err := directoryState(filepath.Join(*evidence, "projections"))
	must(err)
	cleaned, err := service.CleanupRepositoryContext(ctx, 100)
	must(err)
	assert(cleaned.RetiredIndexes == 0 && cleaned.DeletedGenerations == 0 && !cleaned.HasMore, "completed cleanup is not idempotent after SQLite reopen")
	_, err = service.SearchRepositoryCode(ctx, human, qa)
	code(err, apierrors.ScopeDenied)
	other := qa
	other.ProjectID = "s4-cleanup-same-worktree"
	checkGraph(ctx, service, human, other)
	checkGraph(ctx, service, human, qb)
	after, err := directoryState(filepath.Join(*evidence, "projections"))
	must(err)
	assert(before == after, "idempotent cleanup or reads modified retained projections")
	emit("actual_sqlite_reopen_cleanup_idempotence_and_remaining_graph_queries")
}
