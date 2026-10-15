package app_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 原生产服务与真实 SQLite 保留 ready 拒绝，不制造已执行任务或模型结果。
func TestRelayCompletionActualReadyCannotCompleteOrRelease(t *testing.T) {
	ctx := context.Background()
	service, store, command := relayRegistrationService(t)
	if err := service.(app.RelayTrustedAgentRegistrar).RegisterTrustedAgents(ctx, []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}}}); err != nil {
		t.Fatal(err)
	}
	intent, err := service.CreateIntent(ctx, caller(), command)
	if err != nil {
		t.Fatal(err)
	}
	for _, channel := range []string{"", domain.RelayUserConfirmedComplete, domain.RelayVerifiedComplete} {
		operation := caller()
		cmd := app.RelayTransitionCommand{From: domain.RelayReady, To: domain.RelayDone, CompletionConfirmed: true, CompletionChannel: channel}
		_, err = service.Transition(ctx, operation, intent.Task.ID, cmd)
		checkCode(t, err, apierrors.VersionConflict)
		before, err := store.Snapshot(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = service.Transition(ctx, operation, intent.Task.ID, cmd)
		checkCode(t, err, apierrors.VersionConflict)
		after, err := store.Snapshot(ctx)
		if err != nil || !reflect.DeepEqual(before, after) {
			t.Fatal("duplicate completion rejection rewrote original SQLite evidence")
		}
	}
	_, err = service.Transition(ctx, caller(), intent.Task.ID, app.RelayTransitionCommand{From: domain.RelayReady, To: domain.RelayReleased})
	checkCode(t, err, apierrors.VersionConflict)
	view, err := service.GetTask(ctx, caller(), intent.Task.ID)
	if err != nil || view.Task.State != domain.RelayReady || view.Completion != nil {
		t.Fatal("undispatched intent acquired completion or release")
	}
	state, err := store.Snapshot(ctx)
	if err != nil || len(state.Completions) != 0 || len(state.Sessions) != 0 || len(state.ControllerEvidence) != 0 || len(state.Workspaces) != 0 {
		t.Fatal("rejected completion created actual execution or completion evidence")
	}
	for _, fact := range state.RuntimeFacts {
		if fact.Kind == "human_target_confirmed" || fact.Kind == "turn_started" || fact.Kind == "resources_released" {
			t.Fatal("completion rejection fabricated a goal, start or release fact")
		}
	}
	_, err = service.(app.RelayCompletionEvidenceReader).ReadRelayCompletion(ctx, command.ProjectID, uuid.NewString())
	checkCode(t, err, apierrors.NotFound)
	after, err := store.Snapshot(ctx)
	if err != nil || !reflect.DeepEqual(state, after) {
		t.Fatal("missing completion read changed actual SQLite state")
	}
}
