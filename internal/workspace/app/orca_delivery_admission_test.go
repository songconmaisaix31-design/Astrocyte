package app_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	orcaadapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/orca"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type deliveryAdmissionInput struct {
	db         *sqlite.DB
	projects   workspace.LocalProjects
	service    *workspace.OrcaExecutionService
	human      domain.Caller
	project    domain.LocalProject
	invocation workspace.OrcaInvocation
	link       workspace.OrcaExecutionLink
}

func newDeliveryAdmissionInput(t *testing.T, creator string) *deliveryAdmissionInput {
	t.Helper()
	ctx := context.Background()
	directory := t.TempDir()
	db, err := sqlite.Open(filepath.Join(directory, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	human := domain.Caller{Kind: "human", ID: "delivery-current-human"}
	spaces := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	space, err := spaces.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "公开准入规则验证"})
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(directory, "public-inputs")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	for path, text := range map[string]string{"input.txt": "请继续核对公开接口，不调用工具。\n", "context.txt": "公开准入检查的明确上下文。\n"} {
		if err := os.WriteFile(filepath.Join(root, path), []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, agents.NewRegistry(filepath.Join(directory, "native")))
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "公开输入准入", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowDirectory, settings.AllowAgentControl = true, true
	settings.AllowedSubdirs = []string{"."}
	settings.AllowedActions = []string{"observe", "read_context", "start", "send", "stop"}
	settings.ExternalModelCLI = "pi"
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	human.ProjectID = project.ID
	packet, err := projects.ReadProjectContext(ctx, human, project.ID, domain.ContextRequest{Files: []string{"context.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := (agents.ProjectFiles{}).ReadProjectFiles(ctx, project, []string{"input.txt"})
	if err != nil || len(inputs) != 1 {
		t.Fatal("实际输入文件读取失败", err)
	}
	// 此处只验证持久化与效果前授权；选择与 Run 标识是规则输入，不声称原生来源已观察。
	sourceHome := filepath.Join(directory, "account-rule-input")
	selection := workspace.OrcaAgentSelection{AgentID: "pi", Adapter: "pi", ExecutionMode: "structured", Executable: filepath.Join(directory, "uninvoked-native-entry"), Version: "1.0.1", Model: "openai/gpt-rule-input", Provider: "openai", ExternalTarget: "pi", ConfigurationVersion: "unobserved-admission-input", NativeLaunch: workspace.OrcaNativeLaunchPolicy{Schema: "swa.native-launch.v1", Account: workspace.OrcaNativeLaunchAccount{Kind: "pi-native-api-key", Provider: "openai", SourceHome: sourceHome, AuthSource: workspace.OrcaNativeConfigurationSource{Path: filepath.Join(sourceHome, "auth.json"), Version: "unobserved-admission-input"}, PrivateHome: filepath.Join(directory, "pi", project.ID), ModelTarget: workspace.OrcaNativeModelTarget{ModelID: "gpt-rule-input", API: "openai-responses", BaseURL: "https://api.openai.com"}}, Tools: workspace.OrcaNativeLaunchTools{Mode: "disabled", NativeContract: "pi-cli-1.0.1"}}}
	invocation := workspace.OrcaInvocation{Command: workspace.OrcaCommand{Action: "worker.send", Controller: "approval-input-controller", DispatchID: "approval-input-dispatch", RequestID: uuid.NewString(), InputFile: "input.txt"}, Caller: workspace.OrcaExecutionCaller{ID: human.ID, Kind: human.Kind, ProjectID: project.ID, SettingsRevision: project.Settings.Revision}, Project: project, Scope: workspace.OrcaScope{ProjectID: project.ID, Controller: "approval-input-controller", RunID: "approval-input-run", RootTaskID: "approval-input-root", AuthorityDomain: "approval-input-authority", Context: packet}, Input: &inputs[0], Selection: &selection}
	link := workspace.OrcaExecutionLink{ProjectID: project.ID, GrantAgentID: creator, SettingsRevision: project.Settings.Revision, AuthorityDomain: invocation.Scope.AuthorityDomain, RunID: invocation.Scope.RunID, RootTaskID: invocation.Scope.RootTaskID, CreatedAt: time.Now().UTC()}
	if err := db.SaveOrcaExecutionLink(ctx, link); err != nil {
		t.Fatal(err)
	}
	// 使用未调用的实际适配器满足构造契约；一旦错误进入原生路径就应失败，不提供替身返回。
	service, err := workspace.NewOrcaExecutionService(db, agents.ProjectFiles{}, projects, &orcaadapter.Runtime{}, db)
	if err != nil {
		t.Fatal(err)
	}
	return &deliveryAdmissionInput{db: db, projects: projects, service: service, human: human, project: project, invocation: invocation, link: link}
}

func (input *deliveryAdmissionInput) grant(t *testing.T, id string) domain.ProjectGrant {
	t.Helper()
	ctx := context.Background()
	if _, err := input.projects.GrantProjectAgent(ctx, input.human, input.project.ID, domain.ProjectGrant{AgentID: id, Actions: []string{"observe", "read_context", "start", "send"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := input.projects.IssueProjectAgentToken(ctx, input.human, input.project.ID, id); err != nil {
		t.Fatal(err)
	}
	grant, err := input.db.LoadGrant(ctx, input.project.ID, id)
	if err != nil {
		t.Fatal(err)
	}
	return grant
}

func (input *deliveryAdmissionInput) check(t *testing.T, allowed bool) {
	t.Helper()
	result, err := input.service.AdmitOrcaDelivery(context.Background(), input.invocation)
	if allowed && (err != nil || result.Schema != "swa.delivery-admission.v1" || result.RequestID != input.invocation.Command.RequestID || result.ProjectID != input.project.ID || input.invocation.Caller.Kind == "human" && result.GrantExpiresAt != nil) {
		t.Fatalf("实际当前准入不符: result=%+v error=%v", result, err)
	}
	if !allowed && err == nil {
		t.Fatal("越界的效果前准入被接受")
	}
	link, loadErr := input.db.LoadOrcaExecutionLink(context.Background(), input.link.AuthorityDomain, input.link.RunID)
	if loadErr != nil || !reflect.DeepEqual(link, input.link) {
		t.Fatal("准入改写了原创建者或原生关联", loadErr)
	}
	var count int
	if err := input.db.Conn().QueryRow("SELECT COUNT(*) FROM local_agent_sessions").Scan(&count); err != nil || count != 0 {
		t.Fatal("效果前准入产生了会话副作用", count, err)
	}
}

func TestOrcaDeliveryHumanIndependentSendSQLite(t *testing.T) {
	for _, creatorState := range []string{"absent", "current", "revoked", "expired"} {
		t.Run(creatorState, func(t *testing.T) {
			input := newDeliveryAdmissionInput(t, "original-agent")
			if creatorState != "absent" {
				grant := input.grant(t, "original-agent")
				if creatorState == "revoked" {
					if _, err := input.projects.RevokeProjectAgent(context.Background(), input.human, input.project.ID, grant.AgentID); err != nil {
						t.Fatal(err)
					}
				} else if creatorState == "expired" {
					past := time.Now().UTC().Add(-time.Minute)
					grant.ExpiresAt = &past
					if err := input.db.SaveGrant(context.Background(), grant); err != nil {
						t.Fatal(err)
					}
				}
			}
			input.check(t, true)
		})
	}
}

func TestOrcaDeliveryCurrentAuthorizationBoundariesSQLite(t *testing.T) {
	cases := []string{"cross-creator-start", "agent-wrong-creator-send", "agent-wrong-creator-start", "human-carries-grant", "empty-human", "unknown-kind", "wrong-project", "wrong-root", "wrong-run", "wrong-authority", "wrong-revision", "changed-input", "changed-context", "enabled-tools", "wrong-model", "disabled-control", "missing-send", "denied-directory", "wrong-external-target"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			input := newDeliveryAdmissionInput(t, "original-agent")
			invocation := &input.invocation
			if name == "cross-creator-start" || name == "agent-wrong-creator-start" {
				invocation.Command.Action, invocation.Command.DispatchID = "worker.start", ""
				invocation.Command.TaskID, invocation.Command.Name, invocation.Command.Agent, invocation.Command.Model = invocation.Scope.RootTaskID, "explicit-review", "pi", invocation.Selection.Model
			}
			switch name {
			case "agent-wrong-creator-send", "agent-wrong-creator-start":
				grant := input.grant(t, "different-agent")
				invocation.Caller.ID, invocation.Caller.Kind, invocation.Caller.GrantExpiresAt = grant.AgentID, "agent", grant.ExpiresAt
			case "human-carries-grant":
				future := time.Now().UTC().Add(time.Hour)
				invocation.Caller.GrantExpiresAt = &future
			case "empty-human":
				invocation.Caller.ID = ""
			case "unknown-kind":
				invocation.Caller.Kind = "browser-claimed-human"
			case "wrong-project":
				invocation.Scope.ProjectID = uuid.NewString()
			case "wrong-root":
				invocation.Scope.RootTaskID = "different-root"
			case "wrong-run":
				invocation.Scope.RunID = "different-run"
			case "wrong-authority":
				invocation.Scope.AuthorityDomain = "different-authority"
			case "wrong-revision":
				invocation.Caller.SettingsRevision++
			case "changed-input", "changed-context":
				file := "input.txt"
				if name == "changed-context" {
					file = "context.txt"
				}
				if err := os.WriteFile(filepath.Join(input.project.Root, file), []byte("公开文件已变更\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			case "enabled-tools":
				invocation.Selection.NativeLaunch.Tools.Mode = "enabled"
			case "wrong-model":
				invocation.Selection.Model = "openai/different-model"
			case "disabled-control", "missing-send", "denied-directory", "wrong-external-target":
				settings := input.project.Settings
				switch name {
				case "disabled-control":
					settings.AllowAgentControl = false
				case "missing-send":
					settings.AllowedActions = []string{"observe", "read_context", "start"}
				case "denied-directory":
					settings.AllowDirectory = false
				case "wrong-external-target":
					settings.ExternalModelCLI = "codex"
				}
				if _, err := input.projects.SetProjectSettings(context.Background(), input.human, input.project.ID, domain.SettingsCommand{ExpectedRevision: settings.Revision, Settings: settings}); err != nil {
					t.Fatal(err)
				}
			}
			input.check(t, false)
		})
	}
}

func TestOrcaDeliveryOriginalCreatorStillAuthorizedSQLite(t *testing.T) {
	for _, kind := range []string{"human", "agent"} {
		t.Run(kind, func(t *testing.T) {
			creator := "delivery-current-human"
			if kind == "agent" {
				creator = "original-agent"
			}
			input := newDeliveryAdmissionInput(t, creator)
			if kind == "agent" {
				grant := input.grant(t, creator)
				input.invocation.Caller.ID, input.invocation.Caller.Kind, input.invocation.Caller.GrantExpiresAt = creator, kind, grant.ExpiresAt
			}
			input.check(t, true)
			input.invocation.Command.Action, input.invocation.Command.DispatchID = "worker.start", ""
			input.invocation.Command.TaskID, input.invocation.Command.Name, input.invocation.Command.Agent, input.invocation.Command.Model = input.invocation.Scope.RootTaskID, "explicit-review", "pi", input.invocation.Selection.Model
			input.check(t, true)
			if kind == "agent" {
				if _, err := input.projects.RevokeProjectAgent(context.Background(), input.human, input.project.ID, creator); err != nil {
					t.Fatal(err)
				}
				input.check(t, false)
			}
		})
	}
}
