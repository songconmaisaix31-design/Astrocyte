package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

var phase = flag.String("phase", "", "initial, updated, deleted, restart, corrupt-rebuild, permissions, map")
var rootA = flag.String("a", "", "owned public worktree A")
var rootB = flag.String("b", "", "owned public worktree B")
var evidence = flag.String("evidence", "", "ignored local evidence directory")
var sdk = flag.String("sdk", "", "installed official CodeGraph SDK directory")
var script = flag.String("script", "", "repository CodeGraph adapter script")
var node = flag.String("node", "node", "trusted Node executable")
var aoci = flag.String("aoci", "", "trusted AOCI executable")

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func assert(ok bool, message string) {
	if !ok {
		panic(errors.New(message))
	}
}
func code(err error, want apierrors.Code) {
	var e *apierrors.ServiceError
	assert(errors.As(err, &e) && e.Code == want, fmt.Sprintf("expected %s, got %v", want, err))
}
func emit(step string) { fmt.Println("PASS", step) }

func main() {
	flag.Parse()
	defer func() {
		if failure := recover(); failure != nil {
			_ = os.WriteFile(filepath.Join(*evidence, *phase+"-failure-"+time.Now().UTC().Format("20060102T150405.000000000")+".txt"), []byte(fmt.Sprint(failure)), 0600)
			fmt.Fprintln(os.Stderr, "FAIL", *phase, "see ignored evidence directory")
			os.Exit(1)
		}
	}()
	assert(*phase != "" && *rootA != "" && *rootB != "" && *evidence != "", "explicit owned roots, phase and evidence are required")
	assert(slices.Contains([]string{"initial", "updated", "deleted", "deleted-check", "restart", "query-readonly", "uncheckpointed-wal", "sync-current", "expect-stale", "result-bounds", "ambient-git", "same-project-worktrees", "corrupt-rebuild", "permissions", "map", "map-readonly", "map-missing", "cleanup", "cleanup-restart"}, *phase), "unsupported verification phase")
	must(os.MkdirAll(*evidence, 0700))
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
	defer cancel()
	for _, root := range []string{*rootA, *rootB} {
		cmd := exec.CommandContext(ctx, "git", "-C", root, "remote", "get-url", "origin")
		data, err := cmd.Output()
		must(err)
		assert(strings.TrimSpace(string(data)) == "https://github.com/songconmaisaix31-design/astrocyte-s4-public-context-20261015.git", "verification is restricted to the owned public project")
	}
	db, err := sqlite.Open(filepath.Join(*evidence, "state.sqlite"))
	must(err)
	defer db.Close()
	must(db.RunMigrations(ctx, migrations.FS))
	budget := domain.RepositoryBudget{MaxFiles: 100, MaxBytes: 8 * 1024 * 1024, MaxResults: 100, MaxDepth: 3}
	service, err := repositorycontext.New(db, db, repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: "s4-public-windows"}, Graph: repositorycontext.CodeGraph{Node: *node, SDK: *sdk, Script: *script}, AOCI: *aoci, CacheRoot: filepath.Join(*evidence, "projections"), SnapshotBudget: budget})
	must(err)
	human := domain.Caller{Kind: "human", ID: "s4-public-control"}
	query := func(id string) domain.RepositoryQuery {
		return domain.RepositoryQuery{ProjectID: id, Purpose: "verify_owned_public_repository_context", Roots: []string{"."}, Text: "Count", Budget: budget}
	}
	qa, qb := query("s4-public-a"), query("s4-public-b")
	if *phase == "cleanup" {
		verifyCleanup(ctx, db, service, human, qa, qb)
	}
	if *phase == "cleanup-restart" {
		verifyCleanupRestart(ctx, db, service, human, qa, qb)
	}
	if *phase == "initial" {
		for id, root := range map[string]string{qa.ProjectID: *rootA, qb.ProjectID: *rootB} {
			must(db.SaveProject(ctx, domain.LocalProject{ID: id, Name: "Public Unicode statistics", Root: root, Settings: domain.ProjectSettings{Revision: 1, AllowDirectory: true, AllowedSubdirs: []string{"."}, AllowedActions: []string{"read_context"}}, CreatedAt: time.Now().UTC()}, 0))
		}
		sa, err := service.CaptureProjectSnapshot(ctx, human, qa.ProjectID, []string{*rootA})
		must(err)
		sb, err := service.CaptureProjectSnapshot(ctx, human, qb.ProjectID, []string{*rootB})
		must(err)
		assert(sa.Head == sb.Head && sa.RepoID == sb.RepoID && sa.WorktreeID != sb.WorktreeID, "same commit worktree isolation")
		must(sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error { return service.ValidateProjectSnapshot(tx.Context(), human, sa) }))
		emit("same_commit_distinct_worktrees_and_same_transaction_snapshot_validation")
		data, err := json.Marshal(sa)
		must(err)
		must(os.WriteFile(filepath.Join(*evidence, "initial-snapshot.json"), data, 0600))
		sa.Head = "untrusted-head"
		code(service.ValidateProjectSnapshot(ctx, human, sa), apierrors.ContextStale)
		emit("snapshot_tampering_rejected")
		_, err = service.UpdateRepositoryIndex(ctx, human, qa, "index")
		must(err)
		_, err = service.UpdateRepositoryIndex(ctx, human, qb, "index")
		must(err)
		emit("official_codegraph_indexes_created_for_both_worktrees")
		checkGraph(ctx, service, human, qa)
		checkGraph(ctx, service, human, qb)
	}
	if *phase == "updated" || *phase == "deleted" {
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		code(err, apierrors.IndexStale)
		emit("dirty_source_rejects_old_index")
		_, err = service.SearchRepositoryCode(ctx, human, qb)
		must(err)
		emit("other_worktree_index_remains_current")
		var before domain.ProjectContextSnapshot
		data, err := os.ReadFile(filepath.Join(*evidence, "initial-snapshot.json"))
		must(err)
		must(json.Unmarshal(data, &before))
		code(service.ValidateProjectSnapshot(ctx, human, before), apierrors.ContextStale)
		_, err = service.UpdateRepositoryIndex(ctx, human, qa, "sync")
		must(err)
		emit("real_incremental_sync")
		qa.Text = "path:textstats/paragraphs.go name:CountParagraphs"
		result, err := service.SearchRepositoryCode(ctx, human, qa)
		must(err)
		if *phase == "updated" {
			assert(len(result.Symbols) > 0, "new paragraph feature absent")
		} else {
			assert(len(result.Symbols) == 0, "deleted paragraph feature still indexed")
		}
		qb.Text = qa.Text
		other, err := service.SearchRepositoryCode(ctx, human, qb)
		must(err)
		assert(len(other.Symbols) == 0, "worktree symbol crossed isolation")
		emit("added_or_deleted_symbol_and_worktree_isolation")
	}
	if *phase == "deleted-check" {
		qa.Text = "path:textstats/paragraphs.go name:CountParagraphs"
		result, err := service.SearchRepositoryCode(ctx, human, qa)
		must(err)
		assert(len(result.Symbols) == 0, "deleted source file still has indexed symbols")
		qa.Text = "name:countParagraphs"
		result, err = service.SearchRepositoryCode(ctx, human, qa)
		must(err)
		assert(len(result.Symbols) > 0, "preserved paragraph feature is absent after source-file deletion")
		qb.Text = qa.Text
		other, err := service.SearchRepositoryCode(ctx, human, qb)
		must(err)
		assert(len(other.Symbols) == 0, "private paragraph implementation crossed worktrees")
		emit("deleted_file_absent_preserved_feature_present_other_worktree_isolated")
	}
	if *phase == "restart" {
		checkGraph(ctx, service, human, qa)
		checkGraph(ctx, service, human, qb)
		emit("actual_sqlite_and_adapter_restart")
	}
	if *phase == "query-readonly" {
		graphBefore, err := directoryState(filepath.Join(*evidence, "projections"))
		must(err)
		mapBefore, err := directoryState(filepath.Join(*rootA, ".aoci"))
		must(err)
		checkGraph(ctx, service, human, qa)
		checkGraph(ctx, service, human, qb)
		result, err := service.ReadRepositoryMap(ctx, human, qa)
		must(err)
		assert(len(result.MapEntries) == 6, "actual saved authored map is incomplete")
		graphAfter, err := directoryState(filepath.Join(*evidence, "projections"))
		must(err)
		mapAfter, err := directoryState(filepath.Join(*rootA, ".aoci"))
		must(err)
		assert(graphBefore == graphAfter, "repository queries changed the actual graph projection")
		assert(mapBefore == mapAfter, "repository queries changed AOCI runtime files")
		emit("real_graph_search_impact_and_map_leave_all_projection_files_unchanged")
	}
	if *phase == "uncheckpointed-wal" {
		result, err := service.SearchRepositoryCode(ctx, human, qa)
		must(err)
		projection := filepath.Join(*evidence, "projections", result.Index.Snapshot.WorktreeID, result.Index.Generation, ".codegraph", "codegraph.db")
		writer := exec.CommandContext(ctx, *node, "--liftoff-only", filepath.Join(filepath.Dir(*script), "../../../tests/s4/wal.mjs"), projection)
		var diagnostics bytes.Buffer
		writer.Stderr = &diagnostics
		output, err := writer.StdoutPipe()
		must(err)
		input, err := writer.StdinPipe()
		must(err)
		must(writer.Start())
		closed := false
		defer func() {
			if !closed {
				_ = input.Close()
				_ = writer.Wait()
			}
		}()
		var ready struct {
			Ready bool `json:"ready"`
		}
		must(json.NewDecoder(output).Decode(&ready))
		assert(ready.Ready, "actual SQLite writer did not become ready")
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		code(err, apierrors.IndexStale)
		_, err = service.SearchRepositoryCode(ctx, human, qb)
		must(err)
		must(input.Close())
		must(writer.Wait())
		closed = true
		checkGraph(ctx, service, human, qa)
		emit("real_uncheckpointed_wal_rejected_other_worktree_unaffected_and_closed_writer_recovers")
	}
	if *phase == "sync-current" {
		_, err = service.UpdateRepositoryIndex(ctx, human, qa, "sync")
		must(err)
		checkGraph(ctx, service, human, qa)
		checkGraph(ctx, service, human, qb)
		emit("explicit_sync_to_current_source_snapshot")
	}
	if *phase == "expect-stale" {
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		code(err, apierrors.IndexStale)
		_, err = service.ReadRepositoryMap(ctx, human, qa)
		code(err, apierrors.IndexStale)
		_, err = service.SearchRepositoryCode(ctx, human, qb)
		must(err)
		emit("changed_git_index_or_head_rejects_saved_map_and_graph_only_in_worktree_a")
	}
	if *phase == "result-bounds" {
		bad := qa
		bad.Budget.MaxResults = 1
		_, err = service.SearchRepositoryCode(ctx, human, bad)
		code(err, apierrors.BudgetExhausted)
		_, err = service.ReadRepositoryMap(ctx, human, bad)
		code(err, apierrors.BudgetExhausted)
		emit("result_count_budget_rejects_without_silent_truncation")
	}
	if *phase == "ambient-git" {
		before, err := service.CaptureProjectSnapshot(ctx, human, qa.ProjectID, qa.Roots)
		must(err)
		cmd := exec.CommandContext(ctx, "git", "-C", *rootB, "rev-parse", "--path-format=absolute", "--git-path", "index")
		data, err := cmd.Output()
		must(err)
		previous, existed := os.LookupEnv("GIT_INDEX_FILE")
		must(os.Setenv("GIT_INDEX_FILE", strings.TrimSpace(string(data))))
		defer func() {
			if existed {
				_ = os.Setenv("GIT_INDEX_FILE", previous)
			} else {
				_ = os.Unsetenv("GIT_INDEX_FILE")
			}
		}()
		after, err := service.CaptureProjectSnapshot(ctx, human, qa.ProjectID, qa.Roots)
		must(err)
		assert(before.IndexDigest == after.IndexDigest && before.WorktreeDigest == after.WorktreeDigest && before.WorktreeID == after.WorktreeID, "ambient Git index redirected the registered repository")
		must(service.ValidateProjectSnapshot(ctx, human, before))
		emit("actual_other_worktree_index_environment_cannot_redirect_snapshot")
	}
	if *phase == "same-project-worktrees" {
		p, err := db.LoadProject(ctx, qa.ProjectID)
		must(err)
		originalRoot := p.Root
		original, err := service.CaptureProjectSnapshot(ctx, human, qa.ProjectID, qa.Roots)
		must(err)
		version := p.Settings.Revision
		p.Settings.Revision++
		p.Root = *rootB
		must(db.SaveProject(ctx, p, version))
		other, err := service.CaptureProjectSnapshot(ctx, human, qa.ProjectID, qa.Roots)
		must(err)
		assert(original.Head == other.Head && original.RepoID == other.RepoID && original.WorktreeID != other.WorktreeID, "same project did not distinguish its changed worktree")
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		code(err, apierrors.NotFound)
		_, err = service.UpdateRepositoryIndex(ctx, human, qa, "index")
		must(err)
		feature := qa
		feature.Text = "name:countParagraphs"
		result, err := service.SearchRepositoryCode(ctx, human, feature)
		must(err)
		assert(len(result.Symbols) == 0, "same-project worktree B reused A source")
		version = p.Settings.Revision
		p.Settings.Revision++
		p.Root = originalRoot
		must(db.SaveProject(ctx, p, version))
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		code(err, apierrors.IndexStale)
		_, err = service.UpdateRepositoryIndex(ctx, human, qa, "rebuild")
		must(err)
		result, err = service.SearchRepositoryCode(ctx, human, feature)
		must(err)
		assert(len(result.Symbols) > 0, "same-project worktree A did not retain its own source")
		_, err = service.RefreshRepositoryMap(ctx, human, qa)
		must(err)
		emit("same_project_same_commit_worktree_switch_never_reuses_other_graph")
	}
	if *phase == "corrupt-rebuild" {
		snapshot, err := service.CaptureProjectSnapshot(ctx, human, qa.ProjectID, qa.Roots)
		must(err)
		index, err := db.LoadRepositoryIndex(ctx, qa.ProjectID, snapshot.WorktreeID, "codegraph")
		must(err)
		projection := filepath.Join(*evidence, "projections", snapshot.WorktreeID, index.Generation, ".codegraph", "codegraph.db")
		must(os.WriteFile(projection, []byte("owned damaged projection"), 0600))
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		assert(err != nil, "corrupt index query succeeded")
		emit("actual_projection_corruption_rejected")
		_, err = service.UpdateRepositoryIndex(ctx, human, qa, "rebuild")
		must(err)
		checkGraph(ctx, service, human, qa)
		emit("explicit_rebuild_from_real_source")
	}
	if *phase == "permissions" {
		agent := domain.Caller{Kind: "agent", ID: "s4-public-agent", ProjectID: qa.ProjectID}
		grant := domain.ProjectGrant{ProjectID: qa.ProjectID, AgentID: agent.ID, Actions: []string{"read_context"}}
		must(db.SaveGrant(ctx, grant))
		_, err = service.SearchRepositoryCode(ctx, agent, qa)
		must(err)
		_, err = service.SearchRepositoryCode(ctx, agent, qb)
		code(err, apierrors.ScopeDenied)
		bad := qa
		bad.Roots = []string{"../worktree-b"}
		_, err = service.SearchRepositoryCode(ctx, agent, bad)
		code(err, apierrors.ScopeDenied)
		bad = qa
		bad.Purpose = ""
		_, err = service.SearchRepositoryCode(ctx, agent, bad)
		code(err, apierrors.ValidationFailed)
		bad = qa
		bad.Budget.MaxBytes = 1
		_, err = service.SearchRepositoryCode(ctx, agent, bad)
		code(err, apierrors.BudgetExhausted)
		bad = qa
		bad.Budget.MaxResults = 0
		_, err = service.SearchRepositoryCode(ctx, agent, bad)
		code(err, apierrors.BudgetExhausted)
		bad = qa
		bad.Roots = []string{".git"}
		_, err = service.SearchRepositoryCode(ctx, agent, bad)
		code(err, apierrors.ScopeDenied)
		bad = qa
		bad.Roots = []string{".env"}
		_, err = service.SearchRepositoryCode(ctx, agent, bad)
		code(err, apierrors.ScopeDenied)
		emit("real_project_cross_scope_private_paths_purpose_and_budget_rejection")
		now := time.Now().UTC()
		grant.RevokedAt = &now
		must(db.SaveGrant(ctx, grant))
		_, err = service.SearchRepositoryCode(ctx, agent, qa)
		code(err, apierrors.ApprovalRevoked)
		emit("current_sqlite_authorization_revocation")
		p, err := db.LoadProject(ctx, qa.ProjectID)
		must(err)
		revision := p.Settings.Revision
		p.Settings.Revision++
		p.Settings.AllowDirectory = false
		must(db.SaveProject(ctx, p, revision))
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		code(err, apierrors.ScopeDenied)
		p.Settings.Revision++
		p.Settings.AllowDirectory = true
		must(db.SaveProject(ctx, p, revision+1))
		_, err = service.SearchRepositoryCode(ctx, human, qa)
		code(err, apierrors.IndexStale)
		_, err = service.UpdateRepositoryIndex(ctx, human, qa, "rebuild")
		must(err)
		emit("directory_revocation_and_scope_revision_rebuild")
	}
	if *phase == "map" {
		_, err = service.RefreshRepositoryMap(ctx, human, qa)
		must(err)
		result, err := service.ReadRepositoryMap(ctx, human, qa)
		must(err)
		assert(len(result.MapEntries) > 0, "AOCI returned no authored entries")
		emit("actual_aoci_authored_map_persisted_and_read")
	}
	if *phase == "map-readonly" {
		before, err := directoryState(filepath.Join(*rootA, ".aoci"))
		must(err)
		withoutAOCI, err := repositorycontext.New(db, db, repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: "s4-public-windows"}, CacheRoot: filepath.Join(*evidence, "projections"), SnapshotBudget: budget})
		must(err)
		result, err := withoutAOCI.ReadRepositoryMap(ctx, human, qa)
		must(err)
		assert(len(result.MapEntries) == 6, "saved map did not preserve all actual authored entries")
		after, err := directoryState(filepath.Join(*rootA, ".aoci"))
		must(err)
		assert(before == after, "map GET changed AOCI runtime files")
		_, err = withoutAOCI.ReadRepositoryMap(ctx, human, qb)
		code(err, apierrors.NotFound)
		emit("map_get_uses_saved_projection_without_aoci_process_or_ledger_writes")
	}
	if *phase == "map-missing" {
		_, err = service.RefreshRepositoryMap(ctx, human, qb)
		code(err, apierrors.EvidenceMissing)
		emit("uninitialized_aoci_map_remains_missing")
	}
	fmt.Println("PASS phase", *phase)
}

func directoryState(root string) (string, error) {
	state := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		state[path] = fmt.Sprintf("%d:%d", info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(state)
	return string(data), err
}

func checkGraph(ctx context.Context, service *repositorycontext.Service, caller domain.Caller, q domain.RepositoryQuery) {
	result, err := service.SearchRepositoryCode(ctx, caller, q)
	must(err)
	id := ""
	for _, symbol := range result.Symbols {
		if symbol.Name == "Count" && symbol.Span.Path == "textstats/stats.go" {
			id = symbol.ID
		}
	}
	assert(id != "", "actual Count symbol is absent")
	q.SymbolID = id
	impact, err := service.QueryRepositoryImpact(ctx, caller, q)
	must(err)
	assert(len(impact.Relations) > 0, "actual Count callers are absent")
	for _, relation := range impact.Relations {
		assert(relation.EvidenceKind == "codegraph_static_heuristic" && relation.ExtractorVersion == "1.0.1", "relationship provenance is not explicit")
	}
	data, err := json.Marshal(impact)
	must(err)
	assert(!strings.Contains(string(data), *rootA) && !strings.Contains(string(data), *rootB), "absolute source roots escaped safe projection")
	emit("official_symbol_query_real_callers_and_safe_relative_spans")
}
