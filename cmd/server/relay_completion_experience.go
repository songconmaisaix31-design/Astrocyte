package main

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type relayCompletionProjectAccess struct {
	projects workspace.LocalProjects
}

type relayAcceptedArtifactCompletionSource interface {
	attention.ExperienceCompletionReader
	HasAcceptedArtifactCompletion(context.Context, attention.Principal, string, string) (bool, error)
}

type relayCompletionSources struct {
	db       *sqlite.DB
	projects workspace.LocalProjects
	relay    workspace.RelayRuntime
	reader   attention.ExperienceCompletionReader
	accepted relayAcceptedArtifactCompletionSource
}

func (s *relayCompletionSources) ReadExperienceCompletion(ctx context.Context, principal attention.Principal, projectID, eventID string) (attention.ExperienceCompletionEvidence, error) {
	var empty attention.ExperienceCompletionEvidence
	transaction, err := sqlite.WorkspaceTransaction(ctx, s.db)
	if err != nil {
		return empty, err
	}
	if transaction != nil {
		return empty, relayContextFailure(apierrors.ScopeDenied, "completion source must be selected outside the database transaction")
	}
	if _, err := (relayCompletionProjectAccess{projects: s.projects}).currentOwner(ctx, principal, projectID); err != nil {
		return empty, err
	}
	var relayFound bool
	if s.relay != nil {
		tasks, err := s.relay.ListTasks(ctx, domain.Caller{Kind: principal.Kind, ID: principal.ID})
		if err != nil {
			return empty, err
		}
		for _, task := range tasks {
			if task.Completion == nil || task.Completion.ID != eventID {
				continue
			}
			if relayFound || task.Task.ProjectID != projectID {
				return empty, relayContextFailure(apierrors.ScopeDenied, "completion source identity is ambiguous or belongs to another project")
			}
			relayFound = true
		}
	}
	var acceptedFound bool
	if s.accepted != nil {
		acceptedFound, err = s.accepted.HasAcceptedArtifactCompletion(ctx, principal, projectID, eventID)
		if err != nil {
			return empty, err
		}
	}
	if relayFound && acceptedFound {
		return empty, relayContextFailure(apierrors.EvidenceMissing, "completion identity exists in multiple source namespaces")
	}
	var selected attention.ExperienceCompletionReader
	var kind string
	switch {
	case relayFound:
		selected, kind = s.reader, attention.ExperienceSourceRelayCompletion
	case acceptedFound:
		selected, kind = s.accepted, attention.ExperienceSourceAcceptedArtifact
	default:
		return empty, &apierrors.ServiceError{Code: apierrors.NotFound, Message: "original completion source is not registered"}
	}
	if selected == nil {
		return empty, relayContextFailure(apierrors.UnsupportedCapability, "selected completion source reader is unavailable")
	}
	evidence, err := selected.ReadExperienceCompletion(ctx, principal, projectID, eventID)
	if err != nil {
		return empty, err
	}
	if evidence.SourceKind != kind {
		return empty, relayContextFailure(apierrors.EvidenceMissing, "completion evidence differs from its registered source namespace")
	}
	return evidence, nil
}

func (s *relayCompletionSources) AuthorizeExperienceCompletion(ctx context.Context, principal attention.Principal, evidence attention.ExperienceCompletionEvidence) error {
	transaction, err := sqlite.WorkspaceTransaction(ctx, s.db)
	if err != nil {
		return err
	}
	if transaction == nil {
		return relayContextFailure(apierrors.ScopeDenied, "completion authorization requires the current database transaction")
	}
	if _, err := (relayCompletionProjectAccess{projects: s.projects}).currentOwner(ctx, principal, evidence.LocalProjectID); err != nil {
		return err
	}
	switch evidence.SourceKind {
	case attention.ExperienceSourceRelayCompletion:
		if s.reader != nil {
			return s.reader.AuthorizeExperienceCompletion(ctx, principal, evidence)
		}
	case attention.ExperienceSourceAcceptedArtifact:
		if s.accepted != nil {
			return s.accepted.AuthorizeExperienceCompletion(ctx, principal, evidence)
		}
	default:
		return relayContextFailure(apierrors.ScopeDenied, "completion source kind is not authorized")
	}
	return relayContextFailure(apierrors.UnsupportedCapability, "selected completion authorization reader is unavailable")
}

var _ attention.ExperienceCompletionReader = (*relaySessionRuntime)(nil)

func (s *relaySessionRuntime) ReadExperienceCompletion(ctx context.Context, principal attention.Principal, projectID, completionID string) (attention.ExperienceCompletionEvidence, error) {
	if s == nil || s.completionSource == nil {
		return attention.ExperienceCompletionEvidence{}, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "actual Relay completion source is unavailable"}
	}
	return s.completionSource.ReadExperienceCompletion(ctx, principal, projectID, completionID)
}

func (s *relaySessionRuntime) AuthorizeExperienceCompletion(ctx context.Context, principal attention.Principal, evidence attention.ExperienceCompletionEvidence) error {
	if s == nil || s.completionSource == nil {
		return &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "actual Relay completion source is unavailable"}
	}
	return s.completionSource.AuthorizeExperienceCompletion(ctx, principal, evidence)
}

// 复用当前主数据库的项目读取和事务 context；Agent action 尚未授权。
func (a relayCompletionProjectAccess) currentOwner(ctx context.Context, principal attention.Principal, projectID string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	denied := &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "completion requires the current authenticated human and registered project"}
	if principal.Kind != "human" || strings.TrimSpace(principal.ID) == "" || projectID == "" || strings.TrimSpace(projectID) != projectID {
		return "", denied
	}
	if a.projects == nil {
		return "", &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "current completion project authorization is unavailable"}
	}
	projects, err := a.projects.ListProjects(ctx, domain.Caller{ID: principal.ID, Kind: principal.Kind})
	if err != nil {
		return "", err
	}
	for _, project := range projects {
		if project.ID == projectID {
			return project.ID, nil
		}
	}
	return "", denied
}

func assembleRelayCompletionSource(services *httpapi.Services) error {
	if services == nil || services.LocalProjects == nil || services.Relay == nil && services.SwarmS3 == nil {
		return errors.New("completion source requires the original project and execution services")
	}
	bridge, ok := services.ExperienceGovernance.(*experienceHTTPBridge)
	if !ok || bridge == nil || bridge.db == nil || bridge.projects != services.LocalProjects {
		return errors.New("completion source requires the original production experience bridge")
	}
	sources := &relayCompletionSources{db: bridge.db, projects: services.LocalProjects, relay: services.Relay}
	var runtime *relaySessionRuntime
	if services.Relay != nil {
		runtime, ok = services.Relay.(*relaySessionRuntime)
		if !ok || runtime == nil {
			return errors.New("completion source requires the original production Relay instance")
		}
		reader, ok := runtime.RelayRuntime.(workspace.RelayCompletionEvidenceReader)
		if !ok {
			return errors.New("original Relay completion reader is unavailable")
		}
		source, err := agents.NewRelayCompletionExperienceSource(reader, relayCompletionProjectAccess{projects: services.LocalProjects}.currentOwner)
		if err != nil {
			return err
		}
		sources.reader = source
	}
	if services.SwarmS3 != nil {
		source, err := sqlite.NewAcceptedArtifactCompletionSource(bridge.db, services.SwarmS3, services.LocalProjects)
		if err != nil {
			return err
		}
		sources.accepted = source
	}
	if runtime != nil {
		runtime.completionSource = sources.reader
	}
	bridge.completionSources = sources
	return nil
}

type relayCompletionRequestKey struct{}

// 当前身份只在一次可信 Capture 请求内保存，不从完成记录取操作者。
type relayCompletionRequest struct {
	mu          sync.Mutex
	principal   attention.Principal
	operationID string
	evidence    attention.ExperienceCompletionEvidence
	project     domain.LocalProject
	packet      domain.ContextPacket
}

type relayCompletionExperienceHTTP struct {
	attention.ExperienceGovernanceService
	completion attention.CompletionExperienceGovernance
}

func (b *relayCompletionExperienceHTTP) CaptureCompletionExperience(ctx context.Context, principal attention.Principal, command attention.CaptureCompletionExperienceCommand) (attention.StoredExperience, error) {
	if principal.Kind != "human" || strings.TrimSpace(principal.ID) == "" {
		return attention.StoredExperience{}, relayContextFailure(apierrors.ScopeDenied, "completion interpretation requires the current authenticated human")
	}
	request := &relayCompletionRequest{principal: principal}
	return b.completion.CaptureCompletionExperience(context.WithValue(ctx, relayCompletionRequestKey{}, request), principal, command)
}

func (b *relayCompletionExperienceHTTP) ResolveExperienceEvidence(ctx context.Context, principal attention.Principal, id, projectID string) ([]attention.SafeExperienceEvidence, error) {
	return b.completion.ResolveExperienceEvidence(ctx, principal, id, projectID)
}

type relayCompletionInterpretation struct {
	backend  *agents.CompletionExperienceInterpreter
	db       *sqlite.DB
	sources  attention.ExperienceCompletionReader
	projects workspace.LocalProjects
	fixed    *relayPublicContext
	guard    *relaySourceGuard
}

func completionRequest(ctx context.Context) (*relayCompletionRequest, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	request, ok := ctx.Value(relayCompletionRequestKey{}).(*relayCompletionRequest)
	if !ok || request == nil || request.principal.Kind != "human" || strings.TrimSpace(request.principal.ID) == "" {
		return nil, relayContextFailure(apierrors.ScopeDenied, "current completion request identity is unavailable")
	}
	return request, nil
}

func (i *relayCompletionInterpretation) DraftCompletionExperience(ctx context.Context, input attention.CompletionExperienceRequest) (attention.InterpretedExperienceDraft, error) {
	request, err := completionRequest(ctx)
	if err != nil {
		return attention.InterpretedExperienceDraft{}, err
	}
	request.mu.Lock()
	if request.operationID != "" && (request.operationID != input.OperationID || !reflect.DeepEqual(request.evidence, input.Evidence)) {
		request.mu.Unlock()
		return attention.InterpretedExperienceDraft{}, relayContextFailure(apierrors.ContextStale, "completion request already binds different original evidence")
	}
	request.operationID, request.evidence = input.OperationID, input.Evidence
	request.mu.Unlock()
	return i.backend.DraftCompletionExperience(ctx, input)
}

func (i *relayCompletionInterpretation) readCurrentContext(ctx context.Context, principal attention.Principal, evidence attention.ExperienceCompletionEvidence) (domain.LocalProject, domain.ContextPacket, error) {
	var project domain.LocalProject
	var packet domain.ContextPacket
	transaction, err := sqlite.WorkspaceTransaction(ctx, i.db)
	if err != nil {
		return project, packet, err
	}
	if transaction != nil {
		return project, packet, relayContextFailure(apierrors.ScopeDenied, "completion source and files must be read outside the database transaction")
	}
	current, err := i.sources.ReadExperienceCompletion(ctx, principal, evidence.LocalProjectID, evidence.EventID)
	if err != nil {
		return project, packet, err
	}
	if !reflect.DeepEqual(current, evidence) {
		return project, packet, relayContextFailure(apierrors.ContextStale, "fixed completion source changed before interpretation")
	}
	caller := domain.Caller{Kind: principal.Kind, ID: principal.ID}
	project, err = i.projects.CheckProjectModel(ctx, caller, evidence.LocalProjectID, "codex")
	if err != nil {
		return project, packet, err
	}
	packet, err = i.fixed.ReadRelayLaunchContext(ctx, caller, workspace.RelayLaunch{Project: project})
	return project, packet, err
}

func (i *relayCompletionInterpretation) ReadContext(ctx context.Context, evidence attention.ExperienceCompletionEvidence) (domain.LocalProject, domain.ContextPacket, error) {
	request, err := completionRequest(ctx)
	if err != nil {
		return domain.LocalProject{}, domain.ContextPacket{}, err
	}
	request.mu.Lock()
	bound := request.operationID != "" && reflect.DeepEqual(request.evidence, evidence)
	request.mu.Unlock()
	if !bound {
		return domain.LocalProject{}, domain.ContextPacket{}, relayContextFailure(apierrors.EvidenceMissing, "completion context is not bound to its original interpretation operation")
	}
	project, packet, err := i.readCurrentContext(ctx, request.principal, evidence)
	if err != nil {
		return project, packet, err
	}
	request.mu.Lock()
	request.project, request.packet = project, packet
	request.mu.Unlock()
	return project, packet, nil
}

func (i *relayCompletionInterpretation) BeforeFirstInput(ctx context.Context, session domain.NativeSession, archive *domain.RelaySessionArchive, gap string) error {
	request, err := completionRequest(ctx)
	if err != nil {
		return err
	}
	request.mu.Lock()
	operation, evidence, selected, packet := request.operationID, request.evidence, request.project, request.packet
	request.mu.Unlock()
	if operation == "" || session.ID != operation || selected.ID == "" || session.ProjectID != selected.ID || packet.ID == "" {
		return relayContextFailure(apierrors.EvidenceMissing, "completion input lacks its original authorized operation and context")
	}
	current, actual, err := i.readCurrentContext(ctx, request.principal, evidence)
	if err != nil {
		return err
	}
	if current.ID != selected.ID || current.Root != selected.Root || !reflect.DeepEqual(current.Settings, selected.Settings) || actual.ProjectID != packet.ProjectID || actual.SettingsRevision != packet.SettingsRevision || actual.Mode != packet.Mode || !reflect.DeepEqual(actual.Files, packet.Files) || !reflect.DeepEqual(actual.Materials, packet.Materials) {
		return relayContextFailure(apierrors.ContextStale, "completion permission or fixed input changed before submission")
	}
	return i.guard.Check(ctx, session, archive, gap)
}

func assembleRelayCompletionInterpreter(services *httpapi.Services, registry *agents.Registry, contextReaders ...workspace.RelayLaunchContextReader) error {
	if services == nil || registry == nil || len(contextReaders) > 1 {
		return errors.New("completion Interpreter requires the original services and Registry")
	}
	bridge, ok := services.ExperienceGovernance.(*experienceHTTPBridge)
	if !ok || bridge == nil || bridge.db == nil || bridge.completionSources == nil || bridge.projects == nil || bridge.completionInterpreter != nil {
		return errors.New("completion Interpreter requires the original connected experience bridge")
	}
	guard := newRelaySourceGuard(registry)
	if services.Relay != nil {
		runtime, ok := services.Relay.(*relaySessionRuntime)
		if !ok || runtime == nil || runtime.sourceGuard == nil || runtime.sourceGuard.registry != registry {
			return errors.New("completion Interpreter requires the same production Relay source guard")
		}
		guard = runtime.sourceGuard
	}
	deadline, err := relayNativeDeadlineSeconds()
	if err != nil {
		return err
	}
	// 两项均未配置时保留既有 provider_unavailable，且不创建解释操作收据。
	if deadline != 0 || len(contextReaders) != 0 {
		if deadline == 0 || len(contextReaders) != 1 {
			return relayContextFailure(apierrors.ValidationFailed, "completion Interpreter requires an explicit deadline and fixed public context together")
		}
		fixed, ok := contextReaders[0].(*relayPublicContext)
		if !ok || fixed == nil || fixed.projects != bridge.projects {
			return errors.New("completion Interpreter requires the original fixed public project reader")
		}
		interpreter := &relayCompletionInterpretation{db: bridge.db, sources: bridge.completionSources, projects: bridge.projects, fixed: fixed, guard: guard}
		interpreter.backend, err = agents.NewCompletionExperienceInterpreter(registry, agents.CompletionExperienceOptions{ReadContext: interpreter.ReadContext, BeforeFirstInput: interpreter.BeforeFirstInput, DeadlineSeconds: deadline})
		if err != nil {
			return err
		}
		bridge.completionInterpreter = interpreter
	}
	services.ExperienceGovernance = &relayCompletionExperienceHTTP{ExperienceGovernanceService: bridge, completion: bridge}
	return nil
}
