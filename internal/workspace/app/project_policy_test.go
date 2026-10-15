package app_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestProjectActionScopeExplicitPolicyInputs(t *testing.T) {
	// 参数与期望值仅描述纯权限规则，不声明项目、批准或外部来源事实。
	root := t.TempDir()
	project := domain.LocalProject{ID: "project-input", Root: root, Settings: domain.ProjectSettings{
		Revision: 3, AllowDirectory: true, AllowedSubdirs: []string{"internal"},
		AllowedActions: []string{"claim_work_item"}, AllowAgentControl: true,
		AllowedTools: []string{"input-tool"}, ExternalModelCLI: "codex",
	}}
	scope := app.ProjectActionScope{ProjectID: project.ID, SettingsRevision: 3, Action: "claim_work_item", Roots: []string{filepath.Join(root, "internal")}, Tools: []string{"input-tool"}, ExternalTargets: []string{"codex"}}
	if err := app.ValidateProjectActionScope(project, scope); err != nil {
		t.Fatal("explicit authorized policy input was rejected", err)
	}
	for _, item := range []struct {
		name   string
		change func(*domain.LocalProject, *app.ProjectActionScope)
		code   apierrors.Code
	}{
		{"revision", func(p *domain.LocalProject, s *app.ProjectActionScope) { p.Settings.Revision++ }, apierrors.ContextStale},
		{"unknown_revision", func(p *domain.LocalProject, s *app.ProjectActionScope) { s.SettingsRevision = 0 }, apierrors.ContextStale},
		{"action", func(p *domain.LocalProject, s *app.ProjectActionScope) { p.Settings.AllowedActions = nil }, apierrors.ApprovalRequired},
		{"no_alias", func(p *domain.LocalProject, s *app.ProjectActionScope) { s.Action = "claim" }, apierrors.ScopeDenied},
		{"control", func(p *domain.LocalProject, s *app.ProjectActionScope) { p.Settings.AllowAgentControl = false }, apierrors.ScopeDenied},
		{"tool", func(p *domain.LocalProject, s *app.ProjectActionScope) { s.Tools = []string{"unapproved-tool"} }, apierrors.ScopeDenied},
		{"external", func(p *domain.LocalProject, s *app.ProjectActionScope) {
			s.ExternalTargets = []string{"https://caller-input.invalid"}
		}, apierrors.ScopeDenied},
		{"directory", func(p *domain.LocalProject, s *app.ProjectActionScope) { p.Settings.AllowDirectory = false }, apierrors.ScopeDenied},
		{"subdirectory", func(p *domain.LocalProject, s *app.ProjectActionScope) {
			s.Roots = []string{filepath.Join(root, "outside-approved-subdirectory")}
		}, apierrors.ScopeDenied},
		{"empty_subdirectories", func(p *domain.LocalProject, s *app.ProjectActionScope) { p.Settings.AllowedSubdirs = nil }, apierrors.ScopeDenied},
		{"escaped_subdirectory", func(p *domain.LocalProject, s *app.ProjectActionScope) { p.Settings.AllowedSubdirs = []string{".."} }, apierrors.ScopeDenied},
	} {
		t.Run(item.name, func(t *testing.T) {
			p, s := project, scope
			item.change(&p, &s)
			var service *apierrors.ServiceError
			if !errors.As(app.ValidateProjectActionScope(p, s), &service) || service.Code != item.code {
				t.Fatal("explicit policy violation did not retain its expected error code")
			}
		})
	}
}

func TestProjectActionCallerExplicitPolicyInputs(t *testing.T) {
	now := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	project := domain.LocalProject{ID: "project-input"}
	caller := domain.Caller{Kind: "agent", ID: "agent-input", ProjectID: project.ID}
	grant := domain.ProjectGrant{ProjectID: project.ID, AgentID: caller.ID, Actions: []string{"claim_work_item", "accept_artifact", "reject_artifact"}}
	if err := app.AuthorizeProjectActionCaller(project, caller, &grant, "claim_work_item", now); err != nil {
		t.Fatal("explicit Agent permission input was rejected", err)
	}
	for _, action := range []string{"create_mission", "accept_artifact", "reject_artifact", "pause_mission", "cancel_mission", "recover_mission"} {
		var service *apierrors.ServiceError
		if !errors.As(app.AuthorizeProjectActionCaller(project, caller, &grant, action, now), &service) || service.Code != apierrors.ScopeDenied {
			t.Fatal("Agent input obtained a human-only action")
		}
	}
	grant.RevokedAt = &now
	var service *apierrors.ServiceError
	if !errors.As(app.AuthorizeProjectActionCaller(project, caller, &grant, "claim_work_item", now), &service) || service.Code != apierrors.ApprovalRevoked {
		t.Fatal("revoked project permission input was accepted")
	}
	if err := app.AuthorizeProjectActionCaller(project, domain.Caller{Kind: "human", ID: "human-input"}, nil, "stop", now); err != nil {
		t.Fatal("authenticated human safety stop depended on an execution grant", err)
	}
}
