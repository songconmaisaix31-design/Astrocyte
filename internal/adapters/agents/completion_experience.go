package agents

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

const completionExperiencePolicy = "completion-experience-fixed-json-v2"

type CompletionExperienceOptions struct {
	ReadContext      func(context.Context, attention.ExperienceCompletionEvidence) (domain.LocalProject, domain.ContextPacket, error)
	BeforeFirstInput RelayBeforeFirstInput
	// 只接收受信组装核实的实际模型版本；不从 CLI 版本或别名补造。
	ModelVersion    func(context.Context, domain.NativeSession, string, string) (string, error)
	DeadlineSeconds int
}

type CompletionExperienceInterpreter struct {
	registry *Registry
	options  CompletionExperienceOptions
}

var _ attention.CompletionExperienceInterpreter = (*CompletionExperienceInterpreter)(nil)

func NewCompletionExperienceInterpreter(registry *Registry, options CompletionExperienceOptions) (*CompletionExperienceInterpreter, error) {
	if registry == nil || options.ReadContext == nil || options.BeforeFirstInput == nil {
		return nil, nativeError(apierrors.EvidenceMissing, "completion Interpreter requires current authorized context, source and tools readers")
	}
	if options.DeadlineSeconds < 1 || options.DeadlineSeconds > 1800 {
		return nil, nativeError(apierrors.ValidationFailed, "completion Interpreter requires an explicit deadline within 1 to 1800 seconds")
	}
	adapter, err := registry.Adapter("codex")
	if err != nil {
		return nil, err
	}
	if _, ok := adapter.(*Native); !ok {
		return nil, nativeError(apierrors.UnsupportedCapability, "completion Interpreter requires the configured Codex native controller")
	}
	return &CompletionExperienceInterpreter{registry: registry, options: options}, nil
}

type completionExperienceOperation struct {
	evidence                      attention.ExperienceCompletionEvidence
	operationID                   string
	model, provider, modelVersion string
	turnID, itemID, text          string
	completed                     bool
	failure                       error
}

func (i *CompletionExperienceInterpreter) DraftCompletionExperience(ctx context.Context, request attention.CompletionExperienceRequest) (attention.InterpretedExperienceDraft, error) {
	var empty attention.InterpretedExperienceDraft
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	if _, err := uuid.Parse(request.OperationID); err != nil {
		return empty, nativeError(apierrors.ValidationFailed, "completion interpretation requires its original UUID operation")
	}
	adapter, err := i.registry.Adapter("codex")
	if err != nil {
		return empty, err
	}
	native := adapter.(*Native)
	native.completionMu.Lock()
	defer native.completionMu.Unlock()
	if err := ctx.Err(); err != nil {
		return empty, err
	}
	native.mu.Lock()
	held := native.processes[request.OperationID]
	native.mu.Unlock()
	if held != nil {
		held.mu.Lock()
		original := held.completionExperience
		retained := original != nil
		matches := retained && original.operationID == request.OperationID && reflect.DeepEqual(original.evidence, request.Evidence)
		held.mu.Unlock()
		if !retained {
			return empty, nativeError(apierrors.DeliveryUnknown, "original completion evidence was not retained; do not resend")
		}
		if !matches {
			return empty, nativeError(apierrors.ContextStale, "original completion operation binds different evidence")
		}
		// 已受理或未知的原操作只读取；ResumeOnly 不发送也不恢复进程。
		return i.readCompletionExperience(ctx, held, request.OperationID)
	}
	if request.ResumeOnly {
		return empty, nativeError(apierrors.DeliveryUnknown, "original completion controller is unavailable; do not resend after restart")
	}
	if err := validateCompletionExperienceEvidence(request.Evidence); err != nil {
		return empty, err
	}
	encoded, err := json.Marshal(request.Evidence)
	if err != nil {
		return empty, nativeError(apierrors.ValidationFailed, "completion evidence cannot be encoded")
	}
	var selected attention.ExperienceCompletionEvidence
	if err := json.Unmarshal(encoded, &selected); err != nil {
		return empty, nativeError(apierrors.ValidationFailed, "completion evidence cannot be retained")
	}
	project, packet, err := i.options.ReadContext(ctx, selected)
	if err != nil {
		return empty, completionExperienceCallbackFailure(err, "authorized completion context is unavailable")
	}
	if project.ID != request.Evidence.LocalProjectID || !filepath.IsAbs(project.Root) || packet.ID == "" || packet.SchemaVersion != 1 || packet.ProjectID != project.ID || packet.SettingsRevision != project.Settings.Revision || len(packet.Files)+len(packet.Materials) == 0 {
		return empty, nativeError(apierrors.ScopeDenied, "completion context differs from its current authorized project")
	}
	root, err := (RelayOwnedRoots{}).CanonicalRoot(project.Root)
	if err != nil || !sameRelayMetadataCWD(root, project.Root) {
		return empty, nativeError(apierrors.ScopeDenied, "completion Interpreter workspace is unavailable or redirected")
	}
	var fixed attention.ExperienceCompletionEvidence
	if err := json.Unmarshal(encoded, &fixed); err != nil {
		return empty, nativeError(apierrors.ValidationFailed, "completion evidence cannot be retained")
	}
	prompt := "Draft one experience for human review from the fixed completion evidence and authorized context. Return JSON only. Fields: pattern, verdict (avoid, prefer, or condition), condition, evidence (the exact original completion EventID), refused (boolean). Preserve completion_channel and completion_receipt_id as source facts. user_confirmed_complete records the actor's confirmation and does not establish independent goal verification or native success. verified_complete records the source's owned output and goal checks. Missing native trace remains absent; completion and transition receipt references are not native event identities. Do not invent checks, artifacts, usage or improved task performance. If the evidence cannot support a draft, set refused=true and leave pattern and condition empty. Do not approve, rank, score, execute tools, or claim that event references contain facts you have not received. Fixed completion evidence:\n" + string(encoded)
	wirePrompt := packetPrompt(packet, prompt)
	if len(wirePrompt) > 512*1024 || !utf8.ValidString(wirePrompt) {
		return empty, nativeError(apierrors.ValidationFailed, "fixed completion input is invalid UTF-8 or exceeds 512 KiB")
	}
	callCtx, cancel := context.WithTimeout(ctx, time.Duration(i.options.DeadlineSeconds)*time.Second)
	defer cancel()
	session, startErr := native.StartRelayOperationWithInputGuard(callCtx, domain.NativeRequest{
		SessionID: request.OperationID, Project: project,
		Session: domain.NativeSession{ID: request.OperationID, ProjectID: project.ID},
		Command: domain.NativeCommand{CLI: "codex", DeadlineSeconds: i.options.DeadlineSeconds}, Packet: packet,
	}, request.OperationID, i.options.BeforeFirstInput)
	if startErr != nil {
		return empty, completionExperienceCallbackFailure(startErr, "original completion initialization failed; inspect without resending")
	}
	process, err := native.get(session)
	if err != nil {
		return empty, err
	}
	process.mu.Lock()
	operation := &completionExperienceOperation{evidence: fixed, operationID: request.OperationID, model: process.model, provider: process.provider, modelVersion: "unknown"}
	process.completionExperience = operation
	process.mu.Unlock()
	refuse := func(err error) (attention.InterpretedExperienceDraft, error) {
		process.mu.Lock()
		operation.failure = err
		process.mu.Unlock()
		_, stopErr := native.Stop(context.Background(), session)
		if stopErr != nil {
			return empty, nativeError(apierrors.DeliveryUnknown, "completion preflight failed and original controller exit is unconfirmed")
		}
		return empty, err
	}
	if operation.model != "gpt-6.1-sol" || operation.provider == "" {
		return refuse(nativeError(apierrors.EvidenceMissing, "configured native model differs from the authorized completion Interpreter"))
	}
	if i.options.ModelVersion != nil {
		version, err := i.options.ModelVersion(callCtx, session, operation.model, operation.provider)
		if err != nil {
			return refuse(completionExperienceCallbackFailure(err, "actual completion model provenance is unavailable"))
		}
		version = strings.TrimSpace(version)
		if version != "" && version != "unknown" {
			if version == operation.model || version == session.Version {
				return refuse(nativeError(apierrors.EvidenceMissing, "independent model version cannot reuse model routing or CLI version"))
			}
			process.mu.Lock()
			operation.modelVersion = version
			process.mu.Unlock()
		}
	}
	if _, err := native.SendOperation(callCtx, session, prompt, request.OperationID); err != nil {
		safe := completionExperienceCallbackFailure(err, "original completion input was not confirmed; inspect without resending")
		var rejected *domain.RelayInputNotSubmitted
		if errors.As(err, &rejected) {
			process.mu.Lock()
			operation.failure = safe
			process.mu.Unlock()
		}
		return empty, safe
	}
	return i.readCompletionExperience(callCtx, process, request.OperationID)
}

func completionExperienceCallbackFailure(err error, message string) error {
	code := apierrors.EvidenceMissing
	var known *apierrors.ServiceError
	if errors.As(err, &known) {
		code = known.Code
	}
	return nativeError(code, message)
}

func validateCompletionExperienceEvidence(evidence attention.ExperienceCompletionEvidence) error {
	return attention.ValidateCompletionSource(evidence, evidence.LocalProjectID, evidence.EventID, evidence.OwnerProject)
}
