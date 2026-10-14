package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/relayworktree"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

// 两个生产 SQLite 连接共同处理实际被权限拒绝的动作，不启动模型。
func relayFactsAcrossSQLiteConnections(t *testing.T, ctx context.Context, store *sqlite.RelayStore, stateRoot string, service app.RelayRuntime, works app.Works, db *sqlite.DB, registry *agents.Registry, root, id string) {
	t.Helper()
	other, err := sqlite.OpenRelayStore(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	second := app.NewRelayService(other, works, db, relayworktree.New(filepath.Join(root, "worktrees"), relayworktree.Options{AllowCommittedGit: true}), agents.NewRelayNativeDriver(registry), nil, nil, app.RelayServiceOptions{Agents: []domain.AgentSpec{{Name: "codex", DriveMode: "structured", LaunchCmd: []string{"codex"}}}})
	op := caller()
	op.OperationID = uuid.NewString()
	var wait sync.WaitGroup
	results := make(chan error, 16)
	for i := 0; i < 16; i++ {
		wait.Add(1)
		runtime := service
		if i%2 == 1 {
			runtime = second
		}
		go func() {
			defer wait.Done()
			_, e := runtime.(app.RelayLifecycle).ReconcileEvents(ctx, op, id)
			results <- e
		}()
	}
	wait.Wait()
	close(results)
	for e := range results {
		var known *apierrors.ServiceError
		if !errors.As(e, &known) || known.Code != apierrors.ApprovalRequired && known.Code != apierrors.DeliveryUnknown {
			t.Fatal("duplicate action did not preserve its actual permission or pending result")
		}
	}
	snapshot, err := store.Snapshot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var key string
	count := 0
	for candidate, r := range snapshot.Receipts {
		if r.OperationID == op.OperationID {
			key = candidate
			count++
			if r.Status != "rejected" || r.Error == nil || r.Error.Code != string(apierrors.ApprovalRequired) {
				t.Fatal("original rejection was changed")
			}
		}
	}
	if count != 1 {
		t.Fatal("duplicate operation created more than one receipt")
	}
	factCount := 0
	factIndex := -1
	for i, fact := range snapshot.RuntimeFacts {
		if fact.Kind == "action_claimed" || fact.Kind == "turn_started" || fact.Kind == "human_target_confirmed" {
			t.Fatal("permission rejection created execution or human completion proof")
		}
		if fact.OperationID == op.OperationID {
			factCount++
			factIndex = i
			if fact.ReceiptID != key || fact.Kind != "action_result" || fact.Status != "rejected" {
				t.Fatal("runtime fact differs from its actual rejected action")
			}
		}
	}
	if factCount != 1 || len(snapshot.ControllerEvidence) != 0 || len(snapshot.SessionRelations) != 0 {
		t.Fatal("duplicate action created duplicate facts or false session relationships")
	}
	if err = store.Update(ctx, func(st *app.RelayState) error { st.RuntimeFacts[factIndex].Status = "confirmed"; return nil }); err == nil {
		t.Fatal("final runtime fact mutation was accepted")
	}
	if err = store.Update(ctx, func(st *app.RelayState) error { st.RuntimeFacts = nil; return nil }); err == nil {
		t.Fatal("runtime facts were removed")
	}
	if err = store.Update(ctx, func(st *app.RelayState) error {
		r := st.Receipts[key]
		r.Status = "accepted"
		st.Receipts[key] = r
		return nil
	}); err == nil {
		t.Fatal("final receipt mutation was accepted")
	}
	read, relations, err := second.(app.RelayRuntimeFactsReader).ListRuntimeFacts(ctx, caller(), id)
	if err != nil || len(read) != len(snapshot.RuntimeFacts) || len(relations) != 0 {
		t.Fatal("persisted read-only facts were lost")
	}
	history, err := second.(app.RelayControllerHistoryReader).ListOwnedControllerHistory(ctx, id)
	if err != nil || len(history) != 0 {
		t.Fatal("a task without a controller acquired historical ownership")
	}
	ownership, err := second.(app.RelayCurrentOwnershipReader).GetCurrentOwnership(ctx, caller(), id)
	if err != nil || ownership.Status != "unknown" || ownership.Reason != "controller_binding_unrecorded" {
		t.Fatal("ready task acquired current ownership from persisted state")
	}
	_, err = second.(app.RelayCurrentOwnershipReader).GetCurrentOwnership(ctx, domain.Caller{Kind: "agent", ID: "codex"}, id)
	var denied *apierrors.ServiceError
	if !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied {
		t.Fatal("Agent accessed the human ownership view")
	}
	after, err := other.Snapshot(ctx)
	if err != nil || !reflect.DeepEqual(snapshot, after) {
		t.Fatal("fact read or rejected mutation changed persisted state")
	}
	encoded, err := json.Marshal(domain.RelayControllerEvidence{WorkspaceRoot: "private-pointer-canary", NativeID: "private-native-canary"})
	if err != nil || string(encoded) != "{}" {
		t.Fatal("private controller evidence was exposed by JSON")
	}
	t.Log("two SQLite connections: one rejected receipt and one immutable result fact; no controller, start, completion or relationship proof")
}
