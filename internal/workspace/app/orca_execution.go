package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// OrcaCommand 只包含产品允许的动作；身份和原生调用参数由宿主补充。
type OrcaCommand struct {
	Action     string `json:"action"`
	ProjectID  string `json:"project_id,omitempty"`
	Controller string `json:"controller,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
	DispatchID string `json:"dispatch_id,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Workdir    string `json:"workdir,omitempty"`
	SpecFile   string `json:"spec_file,omitempty"`
	InputFile  string `json:"input_file,omitempty"`
	ResultFile string `json:"result_file,omitempty"`
	OutputDir  string `json:"output_dir,omitempty"`
	Name       string `json:"name,omitempty"`
	Agent      string `json:"agent,omitempty"`
	Model      string `json:"model,omitempty"`
	Source     string `json:"source,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

type OrcaTaskSpec struct {
	Goal            string                `json:"goal"`
	SourceRepo      string                `json:"source_repo"`
	SourceRevision  string                `json:"source_revision"`
	SuccessCriteria []string              `json:"success_criteria"`
	MaxWorkers      int                   `json:"max_workers"`
	Context         domain.ContextRequest `json:"context"`
}

// OrcaScope 来自实际原生记录；调用者不能从 HTTP 正文提供这些证明。
type OrcaScope struct {
	ProjectID       string               `json:"project_id"`
	Controller      string               `json:"controller"`
	RunID           string               `json:"run_id"`
	RootTaskID      string               `json:"root_task_id"`
	AuthorityDomain string               `json:"authority_domain"`
	Context         domain.ContextPacket `json:"context"`
}

type OrcaInvocation struct {
	Command OrcaCommand           `json:"command"`
	Caller  OrcaExecutionCaller   `json:"caller"`
	Project domain.LocalProject   `json:"project"`
	Scope   OrcaScope             `json:"scope"`
	Spec    *OrcaTaskSpec         `json:"spec,omitempty"`
	Context *domain.ContextPacket `json:"context,omitempty"`
	Input   *domain.ContextFile   `json:"input,omitempty"`
}

type OrcaExecutionCaller struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	ProjectID        string `json:"project_id"`
	SettingsRevision int    `json:"settings_revision"`
}

type OrcaNativeError struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	Retryable      bool   `json:"retryable"`
	RequiredAction string `json:"required_action"`
	Data           any    `json:"data,omitempty"`
}

func (e *OrcaNativeError) Error() string { return e.Code + ": " + e.Message }

// Result 保留原生结果，不从进程存在或传输成功生成执行成功状态。
type OrcaEnvelope struct {
	OK              bool             `json:"ok"`
	RequestID       string           `json:"request_id,omitempty"`
	RuntimeID       string           `json:"runtime_id,omitempty"`
	AuthorityDomain string           `json:"authority_domain,omitempty"`
	Controller      string           `json:"controller,omitempty"`
	RunID           string           `json:"run_id,omitempty"`
	TaskID          string           `json:"task_id,omitempty"`
	DispatchID      string           `json:"dispatch_id,omitempty"`
	Result          any              `json:"result,omitempty"`
	Error           *OrcaNativeError `json:"error,omitempty"`
}

type OrcaExecutionRuntime interface {
	Target() string
	ResolveScope(context.Context, string, OrcaCommand) (OrcaScope, error)
	Execute(context.Context, OrcaInvocation) (OrcaEnvelope, error)
}

// 关联只有来源和项目授权引用，执行状态继续由原生记录拥有。
type OrcaExecutionLink struct {
	ProjectID        string    `json:"project_id"`
	GrantAgentID     string    `json:"grant_agent_id"`
	SettingsRevision int       `json:"settings_revision"`
	AuthorityDomain  string    `json:"authority_domain"`
	RunID            string    `json:"run_id"`
	RootTaskID       string    `json:"root_task_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type OrcaExecutionLinks interface {
	SaveOrcaExecutionLink(context.Context, OrcaExecutionLink) error
	LoadOrcaExecutionLink(context.Context, string, string) (OrcaExecutionLink, error)
}

type OrcaExecution interface {
	ExecuteOrca(context.Context, domain.Caller, OrcaCommand) (OrcaEnvelope, error)
}

type OrcaExecutionService struct {
	projects LocalProjectRepository
	files    ProjectFiles
	contexts LocalProjects
	runtime  OrcaExecutionRuntime
	links    OrcaExecutionLinks
}

func NewOrcaExecutionService(projects LocalProjectRepository, files ProjectFiles, contexts LocalProjects, runtime OrcaExecutionRuntime, links OrcaExecutionLinks) (*OrcaExecutionService, error) {
	if projects == nil || files == nil || contexts == nil || runtime == nil || links == nil {
		return nil, errors.New("Orca execution requires project, file, context, native runtime and association ports")
	}
	return &OrcaExecutionService{projects: projects, files: files, contexts: contexts, runtime: runtime, links: links}, nil
}

func orcaAction(action string) (permission string, mutation bool, ok bool) {
	switch action {
	case "status", "capabilities", "controller.show", "task.show", "task.list", "task.result", "worker.show", "worker.list", "worker.read", "request.show":
		return "observe", false, true
	case "controller.open", "task.create", "worker.start":
		return "start", true, true
	case "worker.send", "task.complete":
		return "send", true, true
	case "controller.close", "worker.stop", "worker.release":
		return "stop", true, true
	default:
		return "", false, false
	}
}

func validateOrcaCommand(c OrcaCommand) error {
	_, mutation, known := orcaAction(c.Action)
	if !known {
		return projectError(apierrors.ValidationFailed, "unknown Swarm CLI action")
	}
	if mutation {
		id, err := uuid.Parse(c.RequestID)
		if err != nil || id.String() != c.RequestID {
			return projectError(apierrors.ValidationFailed, "a canonical UUID request_id is required for mutations")
		}
	}
	if c.Action == "controller.open" {
		if c.ProjectID == "" || !filepath.IsAbs(c.Workdir) || c.Controller != "" {
			return projectError(apierrors.ValidationFailed, "controller open requires a project and absolute workdir")
		}
		return nil
	}
	if c.Action == "request.show" {
		if c.RequestID == "" || (c.Controller == "" && c.ProjectID == "") || (c.Controller != "" && c.ProjectID != "") {
			return projectError(apierrors.ValidationFailed, "request show requires request_id and exactly one of controller/project")
		}
		return nil
	}
	if c.Action == "status" || c.Action == "capabilities" {
		return nil
	}
	if c.Controller == "" {
		return projectError(apierrors.ValidationFailed, "controller is required")
	}
	if strings.HasPrefix(c.Action, "worker.") && c.Action != "worker.start" && c.Action != "worker.list" && c.DispatchID == "" {
		return projectError(apierrors.ValidationFailed, "dispatch_id is required")
	}
	if (strings.HasPrefix(c.Action, "task.") && c.Action != "task.create" && c.Action != "task.list") || c.Action == "worker.start" || c.Action == "worker.list" {
		if c.TaskID == "" {
			return projectError(apierrors.ValidationFailed, "task_id is required")
		}
	}
	if c.Action == "worker.start" && (c.Agent != "pi" || c.Model != "openai-codex/gpt-6.1-sol" || c.Name == "" || c.InputFile == "") {
		return projectError(apierrors.UnsupportedCapability, "worker start requires a named Pi SDK worker, fixed approved model and input file")
	}
	if c.Action == "worker.send" && c.InputFile == "" {
		return projectError(apierrors.ValidationFailed, "input_file is required")
	}
	if c.Action == "worker.read" && (c.Source != "transcript" || c.Limit < 1 || c.Limit > 1000) {
		return projectError(apierrors.ValidationFailed, "read requires transcript source and limit between 1 and 1000")
	}
	if c.Action == "task.create" && c.SpecFile == "" || c.Action == "task.complete" && c.ResultFile == "" || c.Action == "task.result" && !filepath.IsAbs(c.OutputDir) {
		return projectError(apierrors.ValidationFailed, "the action requires its explicit file or absolute output directory")
	}
	return nil
}

func (s *OrcaExecutionService) authorize(ctx context.Context, caller domain.Caller, projectID, permission string, roots []string, model bool) (domain.LocalProject, error) {
	p, err := s.projects.LoadProject(ctx, projectID)
	if err != nil {
		return p, err
	}
	var grant *domain.ProjectGrant
	if caller.Kind == "agent" {
		g, err := s.projects.LoadGrant(ctx, projectID, caller.ID)
		if err != nil {
			return p, err
		}
		grant = &g
	}
	if err := AuthorizeProjectActionCaller(p, caller, grant, permission, time.Now().UTC()); err != nil {
		return p, err
	}
	scope := ProjectActionScope{ProjectID: p.ID, SettingsRevision: p.Settings.Revision, Action: permission, Roots: roots}
	if model {
		scope.ExternalTargets = []string{s.runtime.Target()}
	}
	return p, ValidateProjectActionScope(p, scope)
}

func (s *OrcaExecutionService) readInput(ctx context.Context, p domain.LocalProject, path string) (domain.ContextFile, error) {
	var result domain.ContextFile
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(p.Root, path)
		if err != nil {
			return result, err
		}
		path = rel
	}
	files, err := s.files.ReadProjectFiles(ctx, p, []string{path})
	if err != nil {
		return result, err
	}
	if len(files) != 1 || files[0].Version == "" || !utf8.ValidString(files[0].Text) || strings.TrimSpace(files[0].Text) == "" {
		return result, projectError(apierrors.ValidationFailed, "the explicit UTF-8 input file is empty or unavailable")
	}
	return files[0], nil
}

func (s *OrcaExecutionService) ExecuteOrca(ctx context.Context, caller domain.Caller, command OrcaCommand) (OrcaEnvelope, error) {
	var reply OrcaEnvelope
	if err := validateOrcaCommand(command); err != nil {
		return reply, err
	}
	permission, _, _ := orcaAction(command.Action)
	projectID := command.ProjectID
	if projectID == "" {
		projectID = caller.ProjectID
	}
	if projectID == "" {
		return reply, projectError(apierrors.ScopeDenied, "a current project-scoped caller is required")
	}
	model := command.Action == "worker.start" || command.Action == "worker.send"
	p, err := s.authorize(ctx, caller, projectID, permission, nil, model)
	if err != nil {
		return reply, err
	}
	invocation := OrcaInvocation{Command: command, Project: p, Caller: OrcaExecutionCaller{ID: caller.ID, Kind: caller.Kind, ProjectID: projectID, SettingsRevision: p.Settings.Revision}}
	if command.Action == "controller.open" {
		root, err := s.files.CanonicalRoot(command.Workdir)
		if err != nil {
			return reply, err
		}
		p, err = s.authorize(ctx, caller, projectID, permission, []string{root}, false)
		if err != nil {
			return reply, err
		}
		invocation.Command.Workdir = root
		invocation.Project = p
	} else if command.Controller != "" {
		scope, err := s.runtime.ResolveScope(ctx, projectID, command)
		if err != nil {
			return reply, err
		}
		if scope.ProjectID != projectID || scope.Controller != command.Controller || scope.AuthorityDomain == "" {
			return reply, projectError(apierrors.ScopeDenied, "native controller scope does not match the current project")
		}
		invocation.Scope = scope
		if scope.Context.ProjectID != "" && (scope.Context.ProjectID != projectID || scope.Context.SettingsRevision != p.Settings.Revision) {
			return reply, projectError(apierrors.ContextStale, "the fixed task context no longer matches current project permissions")
		}
	}
	if command.Action == "task.create" {
		file, err := s.readInput(ctx, p, command.SpecFile)
		if err != nil {
			return reply, err
		}
		var spec OrcaTaskSpec
		if err := json.Unmarshal([]byte(file.Text), &spec); err != nil {
			return reply, fmt.Errorf("decode task specification: %w", err)
		}
		if strings.TrimSpace(spec.Goal) == "" || spec.SourceRepo == "" || spec.SourceRevision == "" || len(spec.SuccessCriteria) == 0 || spec.MaxWorkers != 2 {
			return reply, projectError(apierrors.ValidationFailed, "task specification requires goal, source revision, success criteria and max_workers=2")
		}
		packet, err := s.contexts.ReadProjectContext(ctx, caller, projectID, spec.Context)
		if err != nil {
			return reply, err
		}
		invocation.Spec, invocation.Context = &spec, &packet
	}
	inputPath := command.InputFile
	if command.Action == "task.complete" {
		inputPath = command.ResultFile
	}
	if inputPath != "" {
		file, err := s.readInput(ctx, p, inputPath)
		if err != nil {
			return reply, err
		}
		invocation.Input = &file
	}
	// 文件读取和原生观察之后再次核准，撤销不能沿旧预检继续输入。
	current, err := s.authorize(ctx, caller, projectID, permission, nil, model)
	if err != nil {
		return reply, err
	}
	if current.Settings.Revision != p.Settings.Revision {
		return reply, projectError(apierrors.ContextStale, "project settings changed while preparing the original command")
	}
	reply, err = s.runtime.Execute(ctx, invocation)
	if err != nil {
		return reply, err
	}
	if reply.Error != nil {
		return reply, reply.Error
	}
	if !reply.OK || reply.RuntimeID == "" || reply.AuthorityDomain == "" {
		return reply, projectError(apierrors.DeliveryUnknown, "native response lacks a verifiable runtime result")
	}
	if command.Action == "task.create" {
		if reply.RunID == "" || reply.TaskID == "" {
			return reply, projectError(apierrors.DeliveryUnknown, "native task creation lacks its exact Run/root Task identity")
		}
		link := OrcaExecutionLink{ProjectID: projectID, GrantAgentID: caller.ID, SettingsRevision: p.Settings.Revision, AuthorityDomain: reply.AuthorityDomain, RunID: reply.RunID, RootTaskID: reply.TaskID, CreatedAt: time.Now().UTC()}
		if err := s.links.SaveOrcaExecutionLink(ctx, link); err != nil {
			return reply, &apierrors.ServiceError{Code: apierrors.DeliveryUnknown, Message: "native task exists; saving the project association failed: " + err.Error(), RequestID: command.RequestID, RequiredAction: "inspect_original_request_and_native_run"}
		}
	}
	return reply, nil
}
