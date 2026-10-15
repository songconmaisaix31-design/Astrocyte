package admissionbridge

import (
	"context"
	"errors"

	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type bridge struct {
	admissions *attention.AdmissionService
	workspace  *workspace.S2Service
}

// 必须在启动请求处理之前装配；两个 SQLite repository 使用同一个实际数据库。
func Connect(core *attention.Service, repository attention.AdmissionRepository, proposals *workspace.S2Service, options attention.AdmissionOptions) (*attention.AdmissionService, error) {
	if core == nil || repository == nil || proposals == nil || options.Drafts != nil {
		return nil, errors.New("actual attention, admission repository and workspace service are required without a replacement consumer")
	}
	b := &bridge{workspace: proposals}
	options.Drafts = b
	b.admissions = attention.NewAdmissionService(core, repository, options)
	proposals.ConfigureAdmissions(b)
	return b.admissions, nil
}

func (b *bridge) ConsumeAdmittedOpportunity(ctx context.Context, eventID string) (attention.AdmissionDraft, error) {
	proposal, err := b.workspace.ConsumeOpportunityAdmission(ctx, eventID)
	if err != nil {
		return attention.AdmissionDraft{}, err
	}
	return attention.AdmissionDraft{ID: proposal.ID, Version: proposal.Version}, nil
}

func (b *bridge) ReadAdmittedOpportunity(ctx context.Context, eventID string) (domain.AdmittedOpportunitySnapshot, error) {
	snapshot, err := b.admissions.ReadAdmittedOpportunitySnapshot(ctx, eventID)
	if err != nil {
		return domain.AdmittedOpportunitySnapshot{}, err
	}
	inputs := make([]domain.FixedReference, len(snapshot.Inputs))
	for i, ref := range snapshot.Inputs {
		inputs[i] = domain.FixedReference{MaterialID: ref.MaterialID, Revision: ref.Revision}
	}
	return domain.AdmittedOpportunitySnapshot{
		EventID: snapshot.EventID, AdmissionID: snapshot.AdmissionID,
		OpportunityID: snapshot.OpportunityID, OpportunityRevision: snapshot.OpportunityRevision,
		OwnerNamespace: snapshot.OwnerNamespace, Title: snapshot.Title, Inputs: inputs,
		ProjectID: snapshot.ProjectID, Authority: string(snapshot.Authority),
		PolicyVersion: snapshot.PolicyVersion, Active: snapshot.Active,
	}, nil
}
