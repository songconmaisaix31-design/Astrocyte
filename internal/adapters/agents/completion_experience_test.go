package agents

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 验证缺失来源与冷 Registry 的真实拒绝，不创建原生进程或完成收据。
func TestCompletionExperienceRejectsMissingPrerequisitesAndColdReplay(t *testing.T) {
	registry := NewRegistry("unopened-completion-root")
	if interpreter, err := NewCompletionExperienceInterpreter(registry, CompletionExperienceOptions{}); interpreter != nil || err == nil {
		t.Fatal("missing authority readers produced a configured Interpreter")
	}
	options := CompletionExperienceOptions{
		DeadlineSeconds: 30,
		ReadContext: func(context.Context, attention.ExperienceCompletionEvidence) (domain.LocalProject, domain.ContextPacket, error) {
			panic("rejected completion must not read project context")
		},
		BeforeFirstInput: func(context.Context, domain.NativeSession, *domain.RelaySessionArchive, string) error {
			panic("rejected completion must not reach an input checkpoint")
		},
		ModelVersion: func(context.Context, domain.NativeSession, string, string) (string, error) {
			panic("rejected completion must not reach a model provenance reader")
		},
	}
	interpreter, err := NewCompletionExperienceInterpreter(registry, options)
	if err != nil {
		t.Fatal(err)
	}
	assert := func(request attention.CompletionExperienceRequest, code apierrors.Code) {
		t.Helper()
		result, err := interpreter.DraftCompletionExperience(context.Background(), request)
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != code || result.Refused != nil || result.Truncated != nil || result.Provenance.OperationID != "" {
			t.Fatal("missing completion evidence produced a draft or changed its refusal")
		}
	}
	assert(attention.CompletionExperienceRequest{}, apierrors.ValidationFailed)
	assert(attention.CompletionExperienceRequest{OperationID: uuid.NewString()}, apierrors.EvidenceMissing)
	op := uuid.NewString()
	for range 2 {
		assert(attention.CompletionExperienceRequest{OperationID: op, ResumeOnly: true}, apierrors.DeliveryUnknown)
	}
	native := registry.adapters["codex"].(*Native)
	if len(native.processes) != 0 || len(native.relayCaptures) != 0 {
		t.Fatal("cold completion replay created a process or source")
	}
}

// JSON 库拒绝不完整、拒绝回答及外来字段；错误不能含原生正文。
func TestCompletionExperienceRejectsIncompleteAndPrivateOutput(t *testing.T) {
	for _, output := range []string{
		"PRIVATE_COMPLETION_CANARY",
		`{"refused":true}`,
		`{"pattern":"PRIVATE_COMPLETION_CANARY","verdict":"prefer","condition":"","evidence":"","refused":false}`,
		`{"pattern":"PRIVATE_COMPLETION_CANARY","approved":true,"refused":false}`,
		`{} {}`,
	} {
		result, err := decodeCompletionExperienceDraft(completionExperienceOperation{text: output})
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != apierrors.EvidenceMissing || result.Pattern != "" || strings.Contains(err.Error(), "PRIVATE_COMPLETION_CANARY") {
			t.Fatal("incomplete response produced a candidate or disclosed its private body")
		}
	}
}

// 缺少原 Start 与受理回合的值只能拒绝，不能据完成 JSON 生成原生结果。
func TestCompletionExperienceRejectsUnrecordedOriginalTurn(t *testing.T) {
	for _, process := range []*nativeProcess{
		{},
		{originalSession: &domain.NativeSession{}},
		{relayEvidence: domain.RelayControllerEvidence{ControllerID: "unowned-controller", StartupOperationID: "unowned-operation"}},
	} {
		err := process.checkCompletionExperienceTurn("unowned-operation", completionExperienceOperation{operationID: "unowned-operation"})
		var failure *apierrors.ServiceError
		if !errors.As(err, &failure) || failure.Code != apierrors.EvidenceMissing || len(process.relayEvidence.Facts) != 0 || len(process.relayEvidence.Turns) != 0 || process.inputOperationID != "" {
			t.Fatal("unrecorded original controller acquired an accepted completion turn")
		}
	}
	process := &nativeProcess{originalSession: &domain.NativeSession{Ownership: "external_observed"}, relayEvidence: domain.RelayControllerEvidence{ControllerID: "unowned-controller", StartupOperationID: "unowned-operation"}}
	err := process.checkCompletionExperienceTurn("unowned-operation", completionExperienceOperation{operationID: "unowned-operation"})
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.ScopeDenied || len(process.relayEvidence.Facts) != 0 || len(process.relayEvidence.Turns) != 0 {
		t.Fatal("unowned session acquired an accepted completion turn")
	}
}
