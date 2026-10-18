package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var _ workspace.NativeMessageRepository = (*DB)(nil)

const privateNativeMessagesKey = "private_native_messages"

type storedNativeMessage struct {
	Session            domain.NativeSession `json:"session"`
	OperationID        string               `json:"operation_id"`
	ActorID            string               `json:"actor_id"`
	ActorKind          string               `json:"actor_kind"`
	ActorProjectID     string               `json:"actor_project_id"`
	ControllerID       string               `json:"controller_id"`
	StartupOperationID string               `json:"startup_operation_id"`
	WorkspaceRoot      string               `json:"workspace_root"`
	PreparedAt         time.Time            `json:"prepared_at"`
	ObjectRef          string               `json:"object_ref"`
	Phase              string               `json:"phase"`
	ScopeKind          string               `json:"scope_kind"`
	BindingID          string               `json:"binding_id"`
	LeaseWorkItemID    string               `json:"lease_work_item_id"`
	LeaseMissionID     string               `json:"lease_mission_id"`
	LeaseFence         int64                `json:"lease_fence"`
}

func storeNativeMessage(record workspace.NativeMessagePreparation) storedNativeMessage {
	i := record.Input
	return storedNativeMessage{Session: localNativeSessionForStorage(i.Session), OperationID: i.OperationID,
		ActorID: i.ActorID, ActorKind: i.ActorKind, ActorProjectID: i.ActorProjectID, ControllerID: i.ControllerID,
		StartupOperationID: i.StartupOperationID, WorkspaceRoot: i.WorkspaceRoot, PreparedAt: i.PreparedAt,
		ObjectRef: record.ObjectRef, Phase: record.Phase, ScopeKind: record.Scope.Kind, BindingID: record.Scope.BindingID,
		LeaseWorkItemID: record.Lease.WorkItemID, LeaseMissionID: record.Lease.MissionID, LeaseFence: record.Lease.Fence}
}

func (record storedNativeMessage) preparation() workspace.NativeMessagePreparation {
	return workspace.NativeMessagePreparation{Input: domain.RelayPreparedInput{
		Session: record.Session, OperationID: record.OperationID, ActorID: record.ActorID, ActorKind: record.ActorKind,
		ActorProjectID: record.ActorProjectID, ControllerID: record.ControllerID, StartupOperationID: record.StartupOperationID,
		WorkspaceRoot: record.WorkspaceRoot, PreparedAt: record.PreparedAt,
	}, ObjectRef: record.ObjectRef, Phase: record.Phase,
		Scope: workspace.NativeMessageScope{Kind: record.ScopeKind, BindingID: record.BindingID},
		Lease: workspace.NativeMessageLease{WorkItemID: record.LeaseWorkItemID, MissionID: record.LeaseMissionID, Fence: record.LeaseFence}}
}

func (db *DB) RequireNativeMessageOutsideTx(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := WorkspaceTransaction(ctx, db)
	if err != nil {
		return err
	}
	if conn != nil {
		return errors.New("native input preparation and sending require the caller's database transaction to have ended")
	}
	return nil
}

func (db *DB) WithNativeMessageTx(ctx context.Context, fn func(context.Context) error) error {
	return NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error { return fn(tx.Context()) })
}

func nativeMessageRecords(data []byte) (map[string]json.RawMessage, map[string]storedNativeMessage, error) {
	var row map[string]json.RawMessage
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, nil, err
	}
	if row == nil {
		return nil, nil, errors.New("actual native session JSON object is required")
	}
	records := map[string]storedNativeMessage{}
	if encoded, exists := row[privateNativeMessagesKey]; exists {
		if err := json.Unmarshal(encoded, &records); err != nil {
			return nil, nil, err
		}
		if records == nil {
			return nil, nil, errors.New("private native input records must be a JSON object")
		}
	}
	return row, records, nil
}

func (db *DB) LoadNativeMessagePreparation(ctx context.Context, sessionID, operationID string) (workspace.NativeMessagePreparation, error) {
	conn, err := db.localProjectConnection(ctx)
	if err != nil {
		return workspace.NativeMessagePreparation{}, err
	}
	var data []byte
	if err := conn.QueryRowContext(ctx, "SELECT data FROM local_agent_sessions WHERE id=?", sessionID).Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return workspace.NativeMessagePreparation{}, apierrors.NewNotFound("native session", sessionID)
		}
		return workspace.NativeMessagePreparation{}, err
	}
	_, records, err := nativeMessageRecords(data)
	if err != nil {
		return workspace.NativeMessagePreparation{}, err
	}
	record, exists := records[operationID]
	if !exists {
		return workspace.NativeMessagePreparation{}, apierrors.NewNotFound("prepared native input", operationID)
	}
	if err := workspace.ValidateNativeMessageScope(record.preparation().Scope); err != nil {
		return workspace.NativeMessagePreparation{}, err
	}
	if record.OperationID != operationID || record.Session.ID != sessionID || record.ControllerID == "" || record.StartupOperationID == "" || record.ActorID == "" || (record.ActorKind != "human" && record.ActorKind != "agent") || record.ObjectRef == "" || record.PreparedAt.IsZero() || (record.Phase != "prepared" && record.Phase != "send_attempted") {
		return workspace.NativeMessagePreparation{}, errors.New("stored native input lacks its original fixed identity or delivery phase")
	}
	return record.preparation(), nil
}

// 既有 session JSON 的原值作为 CAS；只允许首次固定输入或单次占用发送。
func (db *DB) SaveNativeMessagePreparation(ctx context.Context, session domain.NativeSession, prepared workspace.NativeMessagePreparation, expectedPhase string) error {
	conn, err := WorkspaceTransaction(ctx, db)
	if err != nil {
		return err
	}
	if conn == nil {
		return errors.New("native input changes require the current primary database transaction")
	}
	var original []byte
	if err := conn.QueryRowContext(ctx, "SELECT data FROM local_agent_sessions WHERE id=? AND project_id=?", session.ID, session.ProjectID).Scan(&original); err != nil {
		return err
	}
	row, records, err := nativeMessageRecords(original)
	if err != nil {
		return err
	}
	var previousSession domain.NativeSession
	if err := json.Unmarshal(original, &previousSession); err != nil {
		return err
	}
	record := storeNativeMessage(prepared)
	if err := workspace.ValidateNativeMessageScope(prepared.Scope); err != nil {
		return err
	}
	op := record.OperationID
	if op == "" || record.Session.ID != session.ID || record.Session.ProjectID != session.ProjectID || record.Session.NativeID != session.NativeID || record.ActorID == "" || (record.ActorKind != "human" && record.ActorKind != "agent") || record.ControllerID == "" || record.StartupOperationID == "" || record.WorkspaceRoot == "" || record.PreparedAt.IsZero() || record.ObjectRef == "" {
		return errors.New("native input preparation requires complete original private metadata")
	}
	receipt, hasReceipt := session.Operations[op]
	if !hasReceipt || receipt.Action != "send" || receipt.Status != "pending" || receipt.FinishedAt != nil || !receipt.CreatedAt.Equal(record.PreparedAt) || session.PendingOperation != op || session.LastOperationID != op {
		return errors.New("native input preparation differs from its original pending send receipt")
	}
	previous, exists := records[op]
	switch expectedPhase {
	case "":
		if prepared.Lease != (workspace.NativeMessageLease{}) {
			return errors.New("new input cannot already carry an execution lease")
		}
		if exists || record.Phase != "prepared" || previousSession.PendingOperation != "" || previousSession.StopConfirmed || previousSession.Ownership != "owned" || previousSession.NativeID != session.NativeID || len(records) >= 64 {
			return conflict("native input cannot replace a prior operation or unsettled session")
		}
		if _, exists := previousSession.Operations[op]; exists {
			return conflict("original operation already has a native receipt")
		}
		expected := previousSession
		if expected.Operations == nil {
			expected.Operations = map[string]domain.NativeOperation{}
		}
		expected.Operations[op] = receipt
		expected.PendingOperation, expected.LastOperationID, expected.UpdatedAt = op, op, record.PreparedAt
		if !reflect.DeepEqual(localNativeSessionForStorage(expected), localNativeSessionForStorage(session)) || !reflect.DeepEqual(record.Session, localNativeSessionForStorage(session)) {
			return conflict("native input preparation changed the existing session or fixed context")
		}
	case "prepared":
		if !exists || previous.Phase != "prepared" || record.Phase != "send_attempted" || previousSession.PendingOperation != op {
			return conflict("original native input has already been claimed for sending")
		}
		previous.Phase = "send_attempted"
		if !reflect.DeepEqual(previous, record) || !reflect.DeepEqual(previousSession, session) {
			return conflict("original prepared native input or session changed before sending")
		}
	case "lease_bound":
		if !exists || previous.Phase != "prepared" || record.Phase != "prepared" || previous.ScopeKind != "swarm" || previous.BindingID == "" || previous.LeaseFence != 0 || previous.LeaseWorkItemID != "" || previous.LeaseMissionID != "" || record.LeaseFence < 1 || record.LeaseWorkItemID == "" || record.LeaseMissionID == "" {
			return conflict("original Swarm preparation cannot replace its first lease")
		}
		previous.LeaseFence, previous.LeaseWorkItemID, previous.LeaseMissionID = record.LeaseFence, record.LeaseWorkItemID, record.LeaseMissionID
		if !reflect.DeepEqual(previous, record) || !reflect.DeepEqual(previousSession, session) {
			return conflict("lease binding changed the original native preparation")
		}
	default:
		return errors.New("unsupported native preparation transition")
	}
	public, err := json.Marshal(localNativeSessionForStorage(session))
	if err != nil {
		return err
	}
	var publicFields map[string]json.RawMessage
	if err := json.Unmarshal(public, &publicFields); err != nil {
		return err
	}
	for key, value := range publicFields {
		row[key] = value
	}
	records[op] = record
	row[privateNativeMessagesKey], err = json.Marshal(records)
	if err != nil {
		return err
	}
	data, err := json.Marshal(row)
	if err != nil {
		return err
	}
	result, err := conn.ExecContext(ctx, "UPDATE local_agent_sessions SET data=? WHERE id=? AND project_id=? AND data=?", string(data), session.ID, session.ProjectID, string(original))
	return workspaceCAS(result, err)
}
