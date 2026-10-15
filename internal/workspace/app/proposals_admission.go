package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 只消费可信准入端口；HTTP和Agent不能提供自证准入快照。
func (s *S2Service) ConsumeOpportunityAdmission(ctx context.Context, eventID string) (proposal domain.Proposal, err error) {
	defer func() { err = NormalizeWorkspaceError(err) }()
	if s.admissions == nil {
		return proposal, domain.WorkspaceError(apierrors.UnsupportedCapability, "trusted admission reader is unavailable")
	}
	snapshot, err := s.admissions.ReadAdmittedOpportunity(ctx, eventID)
	if err != nil {
		return proposal, err
	}
	if !snapshot.Active || snapshot.EventID != eventID || snapshot.AdmissionID == "" || snapshot.OpportunityID == "" || snapshot.OpportunityRevision < 1 || snapshot.OwnerNamespace == "" || snapshot.Title == "" || len(snapshot.Inputs) == 0 || !has([]string{"human", "delegated_judgment"}, snapshot.Authority) || snapshot.Authority == "delegated_judgment" && snapshot.PolicyVersion == "" {
		return proposal, domain.WorkspaceError(apierrors.EvidenceMissing, "admission lacks current authority, fixed version or sources")
	}
	for _, ref := range snapshot.Inputs {
		if ref.MaterialID == "" || ref.Revision < 1 {
			return proposal, domain.WorkspaceError(apierrors.EvidenceMissing, "admission source revision is missing")
		}
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return proposal, err
	}
	digest := sha256.Sum256(data)
	requestDigest := hex.EncodeToString(digest[:])
	err = s.repo.WithWorkspaceTx(ctx, func(tx WorkspaceTx) error {
		current, err := s.admissions.ReadAdmittedOpportunity(tx.Context(), eventID)
		if err != nil {
			return err
		}
		currentData, err := json.Marshal(current)
		if err != nil {
			return err
		}
		if string(currentData) != string(data) {
			return domain.WorkspaceError(apierrors.ContextStale, "admission authority or fixed sources changed before draft commit")
		}
		old, err := tx.LoadReceipt("workspace.admission_consumer", "workspace.opportunity_admitted", eventID)
		if err == nil {
			if old.Digest != requestDigest {
				return domain.WorkspaceError(apierrors.VersionConflict, "admission event binds another snapshot")
			}
			return json.Unmarshal(old.Result, &proposal)
		}
		if !absent(err) {
			return err
		}
		proposal, err = tx.FindAdmission(snapshot.OpportunityID, snapshot.OpportunityRevision)
		if err != nil && !absent(err) {
			return err
		}
		if absent(err) {
			now := time.Now().UTC()
			proposal = domain.Proposal{ID: uuid.NewString(), Version: 1, Revision: 1, Status: domain.ProposalDraft, AdmissionID: snapshot.AdmissionID, OpportunityID: snapshot.OpportunityID, OpportunityRevision: snapshot.OpportunityRevision, OwnerNamespace: snapshot.OwnerNamespace, AssignmentStatus: "unassigned", Content: domain.ProposalContent{Title: snapshot.Title, Inputs: snapshot.Inputs, Scope: domain.ExecutionScope{ProjectID: snapshot.ProjectID, Roots: []string{}, Actions: []string{}, Tools: []string{}, ExternalTargets: []string{}, WritePaths: []string{}}}, CreatedAt: now, UpdatedAt: now}
			if snapshot.ProjectID != "" {
				if _, err = tx.LoadProject(snapshot.ProjectID); err != nil {
					return err
				}
			}
			if err = tx.SaveProposal(proposal, 0); err != nil {
				return err
			}
		} else if proposal.AdmissionID != snapshot.AdmissionID || proposal.OwnerNamespace != snapshot.OwnerNamespace {
			return domain.WorkspaceError(apierrors.VersionConflict, "opportunity version has another admission identity")
		}
		result, err := json.Marshal(proposal)
		if err != nil {
			return err
		}
		return tx.SaveReceipt("workspace.admission_consumer", "workspace.opportunity_admitted", eventID, WorkspaceReceipt{Digest: requestDigest, Result: result})
	})
	return
}
