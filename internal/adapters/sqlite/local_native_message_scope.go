package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"slices"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 历史绑定和固定批准共同决定用途；撤销与租约删除不删除这个依据。
func (db *DB) requireProjectNativeMessageBindings(ctx context.Context, projectID, sessionID string) (err error) {
	defer func() { err = workspace.NormalizeWorkspaceError(err) }()
	conn, err := WorkspaceTransaction(ctx, db)
	if err != nil {
		return err
	}
	if conn == nil {
		return errors.New("native member purpose requires the final primary database transaction")
	}
	rows, err := conn.QueryContext(ctx, `SELECT b.id,b.version,b.project_id,b.session_id,b.grant_id,b.packet_id,b.data,
        a.epoch,a.proposal_id,a.proposal_revision,a.data,p.data,c.data,
        EXISTS(SELECT 1 FROM swarm_missions m WHERE m.grant_id=b.grant_id)
        FROM workspace_session_bindings b
        LEFT JOIN workspace_approvals a ON a.id=b.grant_id
        LEFT JOIN workspace_proposal_revisions p ON p.id=a.proposal_id AND p.revision=a.proposal_revision
        LEFT JOIN workspace_context_packets c ON c.id=b.packet_id
        WHERE b.session_id=?`, sessionID)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	for rows.Next() {
		var id, boundProject, boundSession, grantID, packetID, bindingJSON string
		var version int
		var epoch, revision sql.NullInt64
		var proposalID, approvalJSON, proposalJSON, packetJSON sql.NullString
		var mission bool
		if err := rows.Scan(&id, &version, &boundProject, &boundSession, &grantID, &packetID, &bindingJSON, &epoch, &proposalID, &revision, &approvalJSON, &proposalJSON, &packetJSON, &mission); err != nil {
			return err
		}
		if !epoch.Valid || !revision.Valid || !proposalID.Valid || !approvalJSON.Valid || !proposalJSON.Valid || !packetJSON.Valid {
			return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "original member binding lacks its approved proposal or fixed packet"}
		}
		var binding storedSessionBinding
		var grant domain.ApprovalGrant
		var proposal domain.Proposal
		var packet domain.ImmutableContextPacket
		for _, source := range []struct {
			data  string
			value any
		}{{bindingJSON, &binding}, {approvalJSON.String, &grant}, {proposalJSON.String, &proposal}, {packetJSON.String, &packet}} {
			if err := json.Unmarshal([]byte(source.data), source.value); err != nil {
				return err
			}
		}
		b := binding.SessionBinding
		if id == "" || version < 1 || boundProject != projectID || boundSession != sessionID || b.ID != id || b.Version != version || b.ProjectID != boundProject || b.SessionID != boundSession || b.GrantID != grantID || b.PacketID != packetID || b.GrantEpoch < 1 || b.MemberID == "" {
			return &apierrors.ServiceError{Code: apierrors.ContextStale, Message: "original member binding identity differs from its stored reference"}
		}
		if grant.ID != grantID || grant.Epoch != int(epoch.Int64) || grant.Epoch < b.GrantEpoch || grant.ProposalID != proposalID.String || grant.ProposalRevision != int(revision.Int64) || grant.ProposalRevision < 1 || grant.ApproverID == "" || grant.ApprovedAt.IsZero() || !grant.ExpiresAt.After(grant.ApprovedAt) || !slices.Contains([]string{"active", "revoked", "expired", "completed"}, grant.Status) {
			return &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "original approval identity or epoch cannot establish member purpose"}
		}
		if proposal.ID != grant.ProposalID || proposal.Revision != grant.ProposalRevision || proposal.GrantID != grant.ID || proposal.Status != domain.ProposalApproved && proposal.Status != domain.ProposalSuperseded || packet.ID != packetID || packet.ProjectID != projectID || packet.GrantID != grantID || packet.GrantEpoch != b.GrantEpoch || packet.ProposalID != grant.ProposalID || packet.ProposalRevision != grant.ProposalRevision || grant.Scope.ProjectID != projectID || !reflect.DeepEqual(grant.Scope, proposal.Content.Scope) || !reflect.DeepEqual(grant.Scope, packet.Scope) {
			return &apierrors.ServiceError{Code: apierrors.ContextStale, Message: "member purpose differs from its fixed approved proposal or historical packet"}
		}
		if err := proposal.Content.Validate(grant.ApprovedAt, true); err != nil {
			return err
		}
		swarm, err := workspace.NativeMessageRequiresSwarmLease(proposal.Content.Scope.Actions)
		if err != nil {
			return err
		}
		if swarm || mission {
			return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "original approved Swarm member requires an explicit Swarm preparation and current lease"}
		}
	}
	return rows.Err()
}
