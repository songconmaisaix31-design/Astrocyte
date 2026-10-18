package sqlite

import (
	"context"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
)

func (db *DB) requireProjectNativeMessageScope(ctx context.Context, sessionID string) error {
	conn, err := WorkspaceTransaction(ctx, db)
	if err != nil {
		return err
	}
	if conn == nil {
		return errors.New("project input purpose must be checked in its final transaction")
	}
	var data []byte
	var projectID string
	if err := conn.QueryRowContext(ctx, "SELECT project_id,data FROM local_agent_sessions WHERE id=?", sessionID).Scan(&projectID, &data); err != nil {
		return err
	}
	_, records, err := nativeMessageRecords(data)
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := workspace.ValidateNativeMessageScope(record.preparation().Scope); err != nil {
			return err
		}
		if record.ScopeKind == "swarm" {
			return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "original Swarm session cannot become a project send"}
		}
	}
	if err := db.requireProjectNativeMessageBindings(ctx, projectID, sessionID); err != nil {
		return err
	}
	var leased bool
	if err := conn.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM swarm_leases WHERE json_extract(data,'$.session_id')=?)", sessionID).Scan(&leased); err != nil {
		return err
	}
	if leased {
		return &apierrors.ServiceError{Code: apierrors.LeaseLost, Message: "leased Swarm session requires its original explicit Swarm preparation"}
	}
	return nil
}

func (db *DB) BindSwarmNativeMessageLease(ctx context.Context, sessionID, operationID string, ref workspace.NativeMessageLease) error {
	conn, err := WorkspaceTransaction(ctx, db)
	if err != nil {
		return err
	}
	if conn == nil {
		return errors.New("original native preparation lease binding requires the claim transaction")
	}
	prepared, err := db.LoadNativeMessagePreparation(ctx, sessionID, operationID)
	if err != nil {
		return err
	}
	lease, err := swarmLoadLease(conn.QueryRowContext(ctx, "SELECT data FROM swarm_leases WHERE work_item_id=?", ref.WorkItemID), ref.WorkItemID)
	if err != nil {
		return err
	}
	i := prepared.Input
	if prepared.Scope.Kind != "swarm" || prepared.Scope.BindingID != lease.BindingID || prepared.Phase != "prepared" || prepared.Lease != (workspace.NativeMessageLease{}) || lease.Fence != ref.Fence || lease.MissionID != ref.MissionID || lease.SessionID != sessionID || lease.OperationID != operationID || lease.HolderID != i.ActorID || i.ActorKind != "agent" || i.ActorProjectID != lease.ProjectID || i.Session.ProjectID != lease.ProjectID || i.Session.NativeID != lease.NativeSessionID || i.Session.CLI != lease.AdapterID || i.ControllerID != lease.ControllerID || i.StartupOperationID != lease.StartupOperationID || i.WorkspaceRoot != lease.Worktree {
		return &apierrors.ServiceError{Code: apierrors.LeaseLost, Message: "claim cannot replace the original prepared Swarm input or lease"}
	}
	session, err := db.LoadSession(ctx, sessionID)
	if err != nil {
		return err
	}
	prepared.Lease = ref
	return db.SaveNativeMessagePreparation(ctx, session, prepared, "lease_bound")
}
