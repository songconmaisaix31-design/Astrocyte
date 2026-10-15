package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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

func generationPath(index domain.RepositoryIndex) string {
	return filepath.Join(*evidence, "projections", index.Snapshot.WorktreeID, index.Generation)
}

func waitGenerationRemoved(ctx context.Context, path string, finished <-chan error) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	for {
		_, err := os.Stat(path)
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		must(err)
		select {
		case err := <-finished:
			must(err)
			panic("cleanup worker returned before removing the retired generation")
		case <-ctx.Done():
			panic(ctx.Err())
		case <-timer.C:
			panic("cleanup worker did not remove the retired generation within 30 seconds")
		case <-ticker.C:
		}
	}
}

func startCleanupWorker(ctx context.Context, service *repositorycontext.Service) (context.CancelFunc, <-chan error) {
	workerCtx, cancel := context.WithCancel(ctx)
	finished := make(chan error, 1)
	go func() { finished <- service.RunRepositoryCleanup(workerCtx, 1) }()
	return cancel, finished
}

func stopCleanupWorker(cancel context.CancelFunc, finished <-chan error) {
	cancel()
	select {
	case err := <-finished:
		must(err)
	case <-time.After(10 * time.Second):
		panic("cleanup worker did not stop before SQLite shutdown")
	}
}

func verifyCleanupWorker(ctx context.Context, db *sqlite.DB, service *repositorycontext.Service, human domain.Caller, qa, qb domain.RepositoryQuery) {
	other := qa
	other.ProjectID = "s4-cleanup-same-worktree"
	previous, err := service.SearchRepositoryCode(ctx, human, other)
	must(err)
	old := []domain.RepositoryIndex{previous.Index}
	for range 2 {
		index, err := service.UpdateRepositoryIndex(ctx, human, other, "sync")
		must(err)
		old = append(old, index)
	}
	// 冷启动实例没有前一个实例的内存通知，全部待办来自已保存事实。
	service, err = repositorycontext.New(db, db, repositorycontext.Options{
		Git:   repositorycontext.Git{EnvironmentID: "s4-public-windows"},
		Graph: repositorycontext.CodeGraph{Node: *node, SDK: *sdk, Script: *script},
		AOCI:  *aoci, CacheRoot: filepath.Join(*evidence, "projections"), SnapshotBudget: qa.Budget,
	})
	must(err)
	cancel, finished := startCleanupWorker(ctx, service)
	defer cancel()
	for _, index := range old[:2] {
		waitGenerationRemoved(ctx, generationPath(index), finished)
	}
	code(service.RunRepositoryCleanup(ctx, 1), apierrors.VersionConflict)
	emit("worker_startup_without_wake_removes_two_real_superseded_generations_in_bounded_passes_and_rejects_second_owner")
	current := old[2]
	must(os.WriteFile(filepath.Join(generationPath(current), ".codegraph", "codegraph.db"), []byte("This owned projection is deliberately damaged to verify actual rebuild and diagnostic retention."), 0600))
	_, err = service.SearchRepositoryCode(ctx, human, other)
	code(err, apierrors.ProviderUnavailable)
	diagnostics, err := filepath.Glob(filepath.Join(generationPath(current), ".codegraph", "astrocyte-error-*.log"))
	must(err)
	assert(len(diagnostics) == 1, "actual damaged graph diagnostic was not preserved")
	original, err := os.ReadFile(diagnostics[0])
	must(err)
	index, err := service.UpdateRepositoryIndex(ctx, human, other, "rebuild")
	must(err)
	waitGenerationRemoved(ctx, generationPath(current), finished)
	archived, err := os.ReadFile(filepath.Join(*evidence, "projections", "diagnostics", filepath.Base(diagnostics[0])))
	must(err)
	assert(bytes.Equal(original, archived), "cleanup changed or lost the actual graph failure diagnostic")
	current = index
	checkGraph(ctx, service, human, other)
	emit("real_corrupt_graph_rebuild_wakes_worker_and_preserves_original_diagnostic_while_removing_old_generation")

	p, err := db.LoadRepositoryProject(ctx, other.ProjectID)
	must(err)
	expected := p.Settings.Revision
	p.Settings.Revision++
	p.Settings.AllowDirectory = false
	must(db.SaveProject(ctx, p, expected))
	_, err = service.SearchRepositoryCode(ctx, human, other)
	code(err, apierrors.ScopeDenied)
	for range 1000 {
		service.WakeRepositoryCleanup()
	}
	waitGenerationRemoved(ctx, generationPath(current), finished)
	stopCleanupWorker(cancel, finished)
	checkGraph(ctx, service, human, qb)
	emit("committed_project_revocation_and_merged_wakes_cleanup_actual_projection_then_cancel_and_wait_before_db_close")

	expected = p.Settings.Revision
	p.Settings.Revision++
	p.Settings.AllowDirectory = true
	must(db.SaveProject(ctx, p, expected))
	index, err = service.UpdateRepositoryIndex(ctx, human, other, "rebuild")
	must(err)
	expected = p.Settings.Revision
	p.Settings.Revision++
	p.Settings.AllowDirectory = false
	must(db.SaveProject(ctx, p, expected))
	_, err = service.SearchRepositoryCode(ctx, human, other)
	code(err, apierrors.ScopeDenied)
	_, err = os.Stat(generationPath(index))
	must(err)
	emit("committed_revocation_without_wake_preserves_actual_pending_directory_for_next_process_recovery")
}

func verifyCleanupWorkerRestart(ctx context.Context, db *sqlite.DB, service *repositorycontext.Service, human domain.Caller, qa, qb domain.RepositoryQuery) {
	other := qa
	other.ProjectID = "s4-cleanup-same-worktree"
	data, err := os.ReadFile(filepath.Join(*evidence, "initial-snapshot.json"))
	must(err)
	var initial domain.ProjectContextSnapshot
	must(json.Unmarshal(data, &initial))
	index, err := db.LoadRepositoryIndex(ctx, other.ProjectID, initial.WorktreeID, "codegraph")
	must(err)
	_, err = os.Stat(generationPath(index))
	must(err)
	cancel, finished := startCleanupWorker(ctx, service)
	defer cancel()
	waitGenerationRemoved(ctx, generationPath(index), finished)
	stopCleanupWorker(cancel, finished)
	_, err = db.LoadRepositoryIndex(ctx, other.ProjectID, index.Snapshot.WorktreeID, "codegraph")
	code(err, apierrors.NotFound)
	checkGraph(ctx, service, human, qb)
	emit("independent_process_startup_recovers_committed_revocation_with_lost_wake_using_real_metadata_and_directories")
}

type cleanupGitObservation struct {
	project domain.LocalProject
	index   domain.RepositoryIndex
	input   io.WriteCloser
}

func prepareWithWaitingGit(ctx context.Context, db *sqlite.DB, budget domain.RepositoryBudget) (<-chan cleanupGitObservation, <-chan error) {
	started := make(chan cleanupGitObservation, 1)
	finished := make(chan error, 1)
	go func() {
		waiting := false
		_, err := db.PrepareRepositoryContextCleanup(ctx, func(ctx context.Context, project domain.LocalProject, index domain.RepositoryIndex) (bool, error) {
			if !waiting {
				waiting = true
				// 官方Git在真实stdin上等待，观察期间不能占用SQLite写事务。
				cmd := exec.CommandContext(ctx, "git", "-C", project.Root, "hash-object", "--stdin")
				cmd.Stdout = io.Discard
				input, err := cmd.StdinPipe()
				if err != nil {
					return false, err
				}
				defer input.Close()
				if err := cmd.Start(); err != nil {
					return false, err
				}
				started <- cleanupGitObservation{project, index, input}
				err = cmd.Wait()
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				if err != nil {
					return false, err
				}
			}
			observed, err := (repositorycontext.Git{EnvironmentID: "s4-public-windows"}).Capture(ctx, project, index.Snapshot.Roots, budget)
			return observed.WorktreeID == index.Snapshot.WorktreeID, err
		})
		finished <- err
	}()
	return started, finished
}

func waitCleanupGit(started <-chan cleanupGitObservation, finished <-chan error) cleanupGitObservation {
	select {
	case observation := <-started:
		return observation
	case err := <-finished:
		must(err)
		panic("actual Git observation did not start")
	case <-time.After(30 * time.Second):
		panic("actual Git observation did not start within 30 seconds")
	}
}

func waitCleanupObservation(finished <-chan error) error {
	select {
	case err := <-finished:
		return err
	case <-time.After(30 * time.Second):
		panic("actual Git observation did not finish within 30 seconds")
	}
}

func verifyCleanupSQLBoundary(ctx context.Context, db *sqlite.DB, service *repositorycontext.Service, human domain.Caller, qb domain.RepositoryQuery) {
	observationCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	started, finished := prepareWithWaitingGit(observationCtx, db, qb.Budget)
	observation := waitCleanupGit(started, finished)
	p := observation.project
	expected := p.Settings.Revision
	p.Settings.Revision++
	p.Settings.AllowDirectory = false
	written := make(chan error, 1)
	go func() { written <- db.SaveProject(ctx, p, expected) }()
	select {
	case err := <-written:
		must(err)
	case <-time.After(10 * time.Second):
		panic("SQLite writer was blocked by the waiting actual Git process")
	}
	_, err := service.SearchRepositoryCode(ctx, human, qb)
	code(err, apierrors.ScopeDenied)
	must(observation.input.Close())
	code(waitCleanupObservation(finished), apierrors.VersionConflict)
	saved, err := db.LoadRepositoryIndex(ctx, observation.index.Snapshot.ProjectID, observation.index.Snapshot.WorktreeID, observation.index.Backend)
	must(err)
	assert(saved.Generation == observation.index.Generation, "conflicting cleanup changed the observed index metadata")
	emit("actual_git_wait_does_not_hold_sqlite_writer_and_concurrent_authorization_change_rejects_cleanup_commit")

	expected = p.Settings.Revision
	p.Settings.Revision++
	p.Settings.AllowDirectory = true
	must(db.SaveProject(ctx, p, expected))
	_, err = service.UpdateRepositoryIndex(ctx, human, qb, "rebuild")
	must(err)
	canceledCtx, cancelGit := context.WithCancel(ctx)
	defer cancelGit()
	started, finished = prepareWithWaitingGit(canceledCtx, db, qb.Budget)
	observation = waitCleanupGit(started, finished)
	cancelGit()
	assert(errors.Is(waitCleanupObservation(finished), context.Canceled), "canceled real Git observation was accepted")
	_ = observation.input.Close()
	_, err = db.LoadRepositoryProject(ctx, qb.ProjectID)
	must(err)
	checkGraph(ctx, service, human, qb)
	emit("canceling_actual_git_observation_leaves_sqlite_and_current_graph_usable")
}
