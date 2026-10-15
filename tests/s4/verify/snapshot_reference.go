package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func verifySnapshotReference(ctx context.Context) {
	work, err := os.MkdirTemp(*evidence, "snapshot-reference-")
	must(err)
	root := filepath.Join(work, "public-project")
	_, err = exec.CommandContext(ctx, "git", "clone", "--no-hardlinks", *rootB, root).CombinedOutput()
	must(err)
	dbPath := filepath.Join(work, "state.sqlite")
	db, err := sqlite.Open(dbPath)
	must(err)
	defer func() { must(db.Close()) }()
	must(db.RunMigrations(ctx, migrations.FS))
	options := repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: "s4-reference-public-windows"}, CacheRoot: filepath.Join(work, "projections"), SnapshotBudget: domain.RepositoryBudget{MaxFiles: 100, MaxBytes: 8 * 1024 * 1024, MaxResults: 100, MaxDepth: 3}}
	service, err := repositorycontext.New(db, db, options)
	must(err)
	project := domain.LocalProject{ID: "s4-reference-public", Name: "Public Unicode statistics", Root: root, Settings: domain.ProjectSettings{Revision: 1, AllowDirectory: true, AllowedSubdirs: []string{"."}}, CreatedAt: time.Now().UTC()}
	must(db.SaveProject(ctx, project, 0))
	human := domain.Caller{Kind: "human", ID: "s4-reference-control"}
	agent := domain.Caller{Kind: "agent", ID: "s4-reference-agent", ProjectID: project.ID}
	grant := domain.ProjectGrant{ProjectID: project.ID, AgentID: agent.ID, Actions: []string{"read_context"}}
	must(db.SaveGrant(ctx, grant))
	snapshot, err := service.CaptureProjectSnapshot(ctx, human, project.ID, []string{"."})
	must(err)
	validate := func(caller domain.Caller, value domain.ProjectContextSnapshot) error {
		return sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
			return service.ValidateProjectSnapshotReference(tx.Context(), caller, value)
		})
	}
	stored, err := db.LoadRepositoryIndex(ctx, project.ID, snapshot.WorktreeID, "git-snapshot:"+snapshot.Generation)
	must(err)
	assert(stored.Snapshot.Root == "", "private root entered the stored public projection")
	data, err := json.Marshal(stored)
	must(err)
	assert(!bytes.Contains(data, []byte(`"root":`)), "private root leaked into the public projection")
	must(db.Close())
	db, err = sqlite.Open(dbPath)
	must(err)
	service, err = repositorycontext.New(db, db, options)
	must(err)
	must(validate(human, snapshot))
	must(validate(agent, snapshot))
	emit("snapshot_reference_actual_sqlite_restart_and_private_root")

	func() {
		parked := filepath.Join(work, "public-project-parked")
		must(os.Rename(root, parked))
		defer func() { must(os.Rename(parked, root)) }()
		emptyPath := filepath.Join(work, "empty-path")
		must(os.Mkdir(emptyPath, 0700))
		previousPath := os.Getenv("PATH")
		must(os.Setenv("PATH", emptyPath))
		defer func() { must(os.Setenv("PATH", previousPath)) }()
		must(sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
			conn, err := sqlite.WorkspaceTransaction(tx.Context(), db)
			if err != nil {
				return err
			}
			if _, err := conn.ExecContext(tx.Context(), "PRAGMA query_only=ON"); err != nil {
				return err
			}
			defer func() { _, err := conn.ExecContext(tx.Context(), "PRAGMA query_only=OFF"); must(err) }()
			return service.ValidateProjectSnapshotReference(tx.Context(), agent, snapshot)
		}))
		assert(service.ValidateProjectSnapshot(ctx, human, snapshot) != nil, "actual Git validation accepted the absent worktree")
	}()
	emit("snapshot_reference_readonly_same_connection_without_git_or_worktree")

	changes := map[string]func(*domain.ProjectContextSnapshot){
		"environment":     func(s *domain.ProjectContextSnapshot) { s.EnvironmentID += "-changed" },
		"repository":      func(s *domain.ProjectContextSnapshot) { s.RepoID += "-changed" },
		"worktree":        func(s *domain.ProjectContextSnapshot) { s.WorktreeID += "-changed" },
		"project":         func(s *domain.ProjectContextSnapshot) { s.ProjectID += "-changed" },
		"root":            func(s *domain.ProjectContextSnapshot) { s.Root = *rootB },
		"head":            func(s *domain.ProjectContextSnapshot) { s.Head += "-changed" },
		"index":           func(s *domain.ProjectContextSnapshot) { s.IndexDigest += "-changed" },
		"worktree_digest": func(s *domain.ProjectContextSnapshot) { s.WorktreeDigest += "-changed" },
		"generation":      func(s *domain.ProjectContextSnapshot) { s.Generation = "invalid" },
		"observed_at":     func(s *domain.ProjectContextSnapshot) { s.ObservedAt = s.ObservedAt.Add(time.Nanosecond) },
	}
	for name, change := range changes {
		value := snapshot
		change(&value)
		assert(validate(human, value) != nil, "tampered snapshot accepted: "+name)
	}
	otherOptions := options
	otherOptions.Git.EnvironmentID += "-changed"
	other, err := repositorycontext.New(db, db, otherOptions)
	must(err)
	code(other.ValidateProjectSnapshotReference(ctx, human, snapshot), apierrors.ContextStale)
	emit("snapshot_reference_full_identity_and_service_environment_rechecked")

	now := time.Now().UTC()
	grant.RevokedAt = &now
	must(db.SaveGrant(ctx, grant))
	code(validate(agent, snapshot), apierrors.ApprovalRevoked)
	grant.RevokedAt, grant.ExpiresAt = nil, &now
	must(db.SaveGrant(ctx, grant))
	code(validate(agent, snapshot), apierrors.ApprovalRevoked)
	grant.ExpiresAt, grant.Actions = nil, []string{"discover"}
	must(db.SaveGrant(ctx, grant))
	code(validate(agent, snapshot), apierrors.ScopeDenied)
	grant.Actions = []string{"read_context"}
	must(db.SaveGrant(ctx, grant))
	must(validate(agent, snapshot))
	outside := agent
	outside.ProjectID += "-changed"
	code(validate(outside, snapshot), apierrors.ScopeDenied)
	emit("snapshot_reference_current_grant_revocation_expiry_action_and_scope")

	err = sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
		conn, err := sqlite.WorkspaceTransaction(tx.Context(), db)
		if err != nil {
			return err
		}
		if _, err := conn.ExecContext(tx.Context(), "UPDATE local_agent_grants SET data=json_set(data,'$.revoked_at',?) WHERE project_id=? AND agent_id=?", time.Now().UTC().Format(time.RFC3339Nano), project.ID, agent.ID); err != nil {
			return err
		}
		return service.ValidateProjectSnapshotReference(tx.Context(), agent, snapshot)
	})
	code(err, apierrors.ApprovalRevoked)
	must(validate(agent, snapshot))
	err = sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
		conn, err := sqlite.WorkspaceTransaction(tx.Context(), db)
		if err != nil {
			return err
		}
		if _, err := conn.ExecContext(tx.Context(), "UPDATE local_agent_projects SET data=json_set(data,'$.root',?) WHERE id=?", *rootB, project.ID); err != nil {
			return err
		}
		return service.ValidateProjectSnapshotReference(tx.Context(), human, snapshot)
	})
	code(err, apierrors.ContextStale)
	must(validate(human, snapshot))
	emit("snapshot_reference_uncommitted_grant_and_registered_root_in_same_transaction")

	otherDB, err := sqlite.Open(filepath.Join(work, "other.sqlite"))
	must(err)
	err = sqlite.NewWorkspaceRepository(otherDB).WithWorkspaceTx(ctx, func(tx app.WorkspaceTx) error {
		return service.ValidateProjectSnapshotReference(tx.Context(), human, snapshot)
	})
	assert(err != nil, "foreign database transaction accepted")
	must(otherDB.Close())
	emit("snapshot_reference_foreign_database_transaction_rejected")

	storedChanges := map[string]func(*domain.RepositoryIndex){
		"version":     func(i *domain.RepositoryIndex) { i.Version += "-changed" },
		"state":       func(i *domain.RepositoryIndex) { i.State = "building" },
		"revision":    func(i *domain.RepositoryIndex) { i.Snapshot.SettingsRevision++ },
		"roots":       func(i *domain.RepositoryIndex) { i.Snapshot.Roots = []string{".ssh"} },
		"empty_roots": func(i *domain.RepositoryIndex) { i.Snapshot.Roots = nil },
	}
	for name, change := range storedChanges {
		value := stored
		change(&value)
		must(db.SaveRepositoryIndex(ctx, value, stored.Generation))
		assert(validate(human, snapshot) != nil, "invalid stored reference accepted: "+name)
		must(db.SaveRepositoryIndex(ctx, stored, stored.Generation))
	}
	_, err = db.Conn().ExecContext(ctx, "UPDATE repository_context_indexes SET generation=? WHERE project_id=? AND backend=?", "changed-column", project.ID, stored.Backend)
	must(err)
	assert(validate(human, snapshot) != nil, "generation column mismatch accepted")
	_, err = db.Conn().ExecContext(ctx, "UPDATE repository_context_indexes SET generation=? WHERE project_id=? AND backend=?", stored.Generation, project.ID, stored.Backend)
	must(err)
	emit("snapshot_reference_stored_version_state_roots_revision_and_generation_column")

	projectChanges := map[string]func(*domain.LocalProject){
		"directory_disabled":      func(p *domain.LocalProject) { p.Settings.AllowDirectory = false },
		"scope_narrowed":          func(p *domain.LocalProject) { p.Settings.AllowedSubdirs = []string{"cmd"} },
		"same_scope_new_revision": func(p *domain.LocalProject) {},
		"registered_root_changed": func(p *domain.LocalProject) { p.Root = *rootB },
	}
	for name, change := range projectChanges {
		current := project
		current.Settings.Revision++
		change(&current)
		must(db.SaveProject(ctx, current, project.Settings.Revision))
		assert(validate(human, snapshot) != nil, "changed project authorization accepted: "+name)
		project.Settings.Revision = current.Settings.Revision + 1
		must(db.SaveProject(ctx, project, current.Settings.Revision))
		snapshot, err = service.CaptureProjectSnapshot(ctx, human, project.ID, []string{"."})
		must(err)
		must(validate(human, snapshot))
	}
	emit("snapshot_reference_current_project_directory_roots_revision_and_registration")
	must(os.WriteFile(filepath.Join(root, "reference-observation.txt"), []byte("Actual public worktree mutation for snapshot validation.\n"), 0600))
	must(validate(human, snapshot))
	code(service.ValidateProjectSnapshot(ctx, human, snapshot), apierrors.ContextStale)
	emit("snapshot_reference_preserves_external_actual_git_validation_requirement")
}
