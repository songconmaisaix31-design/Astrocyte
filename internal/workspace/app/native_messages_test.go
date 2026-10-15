package app

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type nativeMessageRuleParameters struct {
	project  domain.LocalProject
	session  domain.NativeSession
	owned    domain.RelayOwnedNativeSession
	prepared NativeMessagePreparation
}

func explicitNativeMessageRuleParameters() nativeMessageRuleParameters {
	// 全部值是纯规则的明确输入，不代表实际项目、准备或控制器来源。
	now := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	operation := "explicit-operation-parameter"
	project := domain.LocalProject{ID: "explicit-project-parameter", Root: "explicit-root-parameter", Settings: domain.ProjectSettings{Revision: 3}}
	session := domain.NativeSession{
		ID: "explicit-session-parameter", NativeID: "explicit-native-parameter", ProjectID: project.ID,
		CLI: "explicit-provider-parameter", Version: "explicit-version-parameter", Ownership: "owned",
		PendingOperation: operation, LastOperationID: operation,
		Operations: map[string]domain.NativeOperation{operation: {Action: "send", Status: "pending", CreatedAt: now}},
		ContextPacket: domain.ContextPacket{
			ID: "explicit-packet-parameter", ProjectID: project.ID, SettingsRevision: project.Settings.Revision,
			Materials: []domain.ContextMaterial{}, Files: []domain.ContextFile{{Path: "explicit-file-parameter", Version: "explicit-file-version-parameter"}},
		},
	}
	owned := domain.RelayOwnedNativeSession{Session: session, Evidence: domain.RelayControllerEvidence{
		SessionID: session.ID, NativeID: session.NativeID, Provider: session.CLI,
		ControllerID: "explicit-controller-parameter", StartupOperationID: "explicit-startup-parameter", WorkspaceRoot: project.Root,
	}}
	owned.Session.ContextPacket = nativeContextMetadata(session.ContextPacket)
	owned.Session.ContextPacket.Files[0].Text = "explicit-text-parameter"
	prepared := NativeMessagePreparation{Input: domain.RelayPreparedInput{
		Session: session, OperationID: operation, ActorID: "explicit-actor-parameter", ActorKind: "human",
		ControllerID: owned.Evidence.ControllerID, StartupOperationID: owned.Evidence.StartupOperationID,
		WorkspaceRoot: project.Root, PreparedAt: now,
	}, ObjectRef: "explicit-object-reference-parameter", Phase: "prepared"}
	return nativeMessageRuleParameters{project: project, session: session, owned: owned, prepared: prepared}
}

func TestNativeMessageExplicitPreparationRules(t *testing.T) {
	parameters := explicitNativeMessageRuleParameters()
	if err := validatePreparedNativeMessage(parameters.project, parameters.session, parameters.prepared, parameters.owned); err != nil {
		t.Fatal("matching explicit rule parameters were rejected", err)
	}
	if parameters.owned.Session.ContextPacket.Files[0].Text != "explicit-text-parameter" {
		t.Fatal("metadata comparison changed the caller's text parameter")
	}
	for _, test := range []struct {
		name string
		edit func(*nativeMessageRuleParameters)
		code apierrors.Code
	}{
		{"send already attempted", func(p *nativeMessageRuleParameters) { p.prepared.Phase = "send_attempted" }, apierrors.DeliveryUnknown},
		{"missing object reference", func(p *nativeMessageRuleParameters) { p.prepared.ObjectRef = "" }, apierrors.DeliveryUnknown},
		{"original controller exited", func(p *nativeMessageRuleParameters) { p.owned.Evidence.ExitConfirmed = true }, apierrors.DeliveryUnknown},
		{"session stopped", func(p *nativeMessageRuleParameters) { p.session.StopConfirmed = true }, apierrors.DeliveryUnknown},
		{"different pending operation", func(p *nativeMessageRuleParameters) { p.session.PendingOperation = "other-explicit-operation" }, apierrors.DeliveryUnknown},
		{"unknown receipt", func(p *nativeMessageRuleParameters) {
			op := p.prepared.Input.OperationID
			receipt := p.session.Operations[op]
			receipt.Status = "unknown"
			p.session.Operations[op] = receipt
		}, apierrors.DeliveryUnknown},
		{"different native identity", func(p *nativeMessageRuleParameters) { p.owned.Evidence.NativeID = "other-explicit-native" }, apierrors.ScopeDenied},
		{"different project root", func(p *nativeMessageRuleParameters) { p.owned.Evidence.WorkspaceRoot = "other-explicit-root" }, apierrors.ScopeDenied},
		{"different settings revision", func(p *nativeMessageRuleParameters) { p.project.Settings.Revision++ }, apierrors.ContextStale},
		{"different context version", func(p *nativeMessageRuleParameters) {
			p.owned.Session.ContextPacket.Files[0].Version = "other-explicit-file-version"
		}, apierrors.ContextStale},
		{"different controller", func(p *nativeMessageRuleParameters) { p.prepared.Input.ControllerID = "other-explicit-controller" }, apierrors.ContextStale},
		{"different startup", func(p *nativeMessageRuleParameters) { p.prepared.Input.StartupOperationID = "other-explicit-startup" }, apierrors.ContextStale},
		{"different fixed provider", func(p *nativeMessageRuleParameters) { p.prepared.Input.Session.CLI = "other-explicit-provider" }, apierrors.ContextStale},
	} {
		t.Run(test.name, func(t *testing.T) {
			parameters := explicitNativeMessageRuleParameters()
			test.edit(&parameters)
			err := validatePreparedNativeMessage(parameters.project, parameters.session, parameters.prepared, parameters.owned)
			var failure *apierrors.ServiceError
			if !errors.As(err, &failure) || failure.Code != test.code {
				t.Fatal("explicit parameter mismatch did not produce its required refusal", err)
			}
		})
	}
}

func TestNativeMessageExplicitCallerAndPrivateJSONRules(t *testing.T) {
	parameters := explicitNativeMessageRuleParameters()
	input := parameters.prepared.Input
	caller := domain.Caller{ID: input.ActorID, Kind: input.ActorKind, ProjectID: input.ActorProjectID, OperationID: input.OperationID}
	if err := preparedNativeMessageCaller(parameters.prepared, caller); err != nil {
		t.Fatal(err)
	}
	for _, different := range []domain.Caller{
		{ID: "different-explicit-actor", Kind: caller.Kind, OperationID: caller.OperationID},
		{ID: caller.ID, Kind: "agent", OperationID: caller.OperationID},
		{ID: caller.ID, Kind: caller.Kind, ProjectID: "different-explicit-project", OperationID: caller.OperationID},
		{ID: caller.ID, Kind: caller.Kind, OperationID: "different-explicit-operation"},
	} {
		if err := preparedNativeMessageCaller(parameters.prepared, different); err == nil {
			t.Fatal("different explicit caller tuple was accepted")
		}
	}
	for _, value := range []any{parameters.prepared, parameters.prepared.Input, parameters.owned} {
		encoded, err := json.Marshal(value)
		if err != nil || string(encoded) != "{}" {
			t.Fatal("private rule parameters escaped JSON serialization", err)
		}
	}
}
