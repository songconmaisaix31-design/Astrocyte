package sqlite

import (
	"context"
	"reflect"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	swarmdomain "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type AcceptedArtifactCompletionSource struct {
	db       *DB
	reader   swarm.AcceptedArtifactEvidenceReader
	projects workspace.LocalProjects
}

var _ attention.ExperienceCompletionReader = (*AcceptedArtifactCompletionSource)(nil)

func NewAcceptedArtifactCompletionSource(db *DB, reader swarm.AcceptedArtifactEvidenceReader, projects workspace.LocalProjects) (*AcceptedArtifactCompletionSource, error) {
	if db == nil || db.conn == nil || reader == nil || projects == nil {
		return nil, acceptedArtifactCompletionError(apierrors.UnsupportedCapability, "actual SQLite, accepted artifact and current project permission readers are required")
	}
	return &AcceptedArtifactCompletionSource{db: db, reader: reader, projects: projects}, nil
}

func acceptedArtifactCompletionError(code apierrors.Code, message string) error {
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: "inspect_actual_artifact_acceptance_and_current_project_permissions"}
}

func (s *AcceptedArtifactCompletionSource) validateRequest(ctx context.Context, principal attention.Principal, projectID, eventID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	// 当前调用者来自认证入口，历史采用者只保留为来源事实。
	if principal.Kind != "human" || strings.TrimSpace(principal.ID) == "" || strings.TrimSpace(principal.ID) != principal.ID {
		return acceptedArtifactCompletionError(apierrors.ScopeDenied, "accepted artifact evidence requires the current authenticated human")
	}
	if projectID == "" || strings.TrimSpace(projectID) != projectID || eventID == "" || strings.TrimSpace(eventID) != eventID {
		return acceptedArtifactCompletionError(apierrors.ValidationFailed, "exact registered project and original acceptance event references are required")
	}
	if s == nil || s.db == nil || s.db.conn == nil || s.reader == nil || s.projects == nil {
		return acceptedArtifactCompletionError(apierrors.UnsupportedCapability, "actual accepted artifact completion source is unavailable")
	}
	if tx, ok := ctx.Value(swarmTransactionKey{}).(*swarmTransaction); ok {
		conn, err := WorkspaceTransaction(ctx, s.db)
		if err != nil {
			return err
		}
		if tx.db != s.db || conn == nil {
			return acceptedArtifactCompletionError(apierrors.ScopeDenied, "accepted artifact source cannot open a Workspace transaction inside a standalone Swarm transaction")
		}
	}
	return nil
}

// 只定位实际 outbox 来源类型；完整 schema 和固定成果由来源读取端口核查。
func (s *AcceptedArtifactCompletionSource) HasAcceptedArtifactCompletion(ctx context.Context, principal attention.Principal, projectID, eventID string) (bool, error) {
	if err := s.validateRequest(ctx, principal, projectID, eventID); err != nil {
		return false, err
	}
	var found bool
	err := NewWorkspaceRepository(s.db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		ctx := tx.Context()
		if err := s.currentProject(ctx, principal, projectID); err != nil {
			return err
		}
		conn, err := WorkspaceTransaction(ctx, s.db)
		if err != nil {
			return err
		}
		if conn == nil {
			return acceptedArtifactCompletionError(apierrors.ScopeDenied, "source location requires the current owned SQLite transaction")
		}
		// 不过滤 schema，已登记但不支持的采用事件仍归原来源处理并失败。
		return conn.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM foundation_outbox WHERE event_id=? AND event_type='swarm.artifact_accepted')", eventID).Scan(&found)
	})
	if err != nil {
		return false, err
	}
	return found, nil
}

// 读取和当前项目权限核查共用本数据库的短事务，不读取对象或原生正文。
func (s *AcceptedArtifactCompletionSource) ReadExperienceCompletion(ctx context.Context, principal attention.Principal, projectID, eventID string) (attention.ExperienceCompletionEvidence, error) {
	var result attention.ExperienceCompletionEvidence
	if err := s.validateRequest(ctx, principal, projectID, eventID); err != nil {
		return result, err
	}
	err := NewWorkspaceRepository(s.db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		var err error
		result, err = s.readCurrent(tx.Context(), principal, projectID, eventID)
		return err
	})
	if err != nil {
		return attention.ExperienceCompletionEvidence{}, err
	}
	return result, nil
}

// 发布候选时必须使用同一主数据库的当前事务，重新读取来源及操作者权限。
func (s *AcceptedArtifactCompletionSource) AuthorizeExperienceCompletion(ctx context.Context, principal attention.Principal, fixed attention.ExperienceCompletionEvidence) error {
	if err := s.validateRequest(ctx, principal, fixed.LocalProjectID, fixed.EventID); err != nil {
		return err
	}
	conn, err := WorkspaceTransaction(ctx, s.db)
	if err != nil {
		return err
	}
	if conn == nil {
		return acceptedArtifactCompletionError(apierrors.ScopeDenied, "accepted artifact publication requires the current owned SQLite transaction")
	}
	if fixed.SourceKind != attention.ExperienceSourceAcceptedArtifact {
		return acceptedArtifactCompletionError(apierrors.ScopeDenied, "completion evidence does not belong to the accepted artifact source")
	}
	current, err := s.readCurrent(ctx, principal, fixed.LocalProjectID, fixed.EventID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, fixed) {
		return acceptedArtifactCompletionError(apierrors.ContextStale, "fixed experience source differs from the original accepted artifact event")
	}
	return nil
}

func (s *AcceptedArtifactCompletionSource) currentProject(ctx context.Context, principal attention.Principal, projectID string) error {
	projects, err := s.projects.ListProjects(ctx, domain.Caller{ID: principal.ID, Kind: principal.Kind})
	if err != nil {
		return err
	}
	for _, project := range projects {
		if project.ID == projectID {
			return nil
		}
	}
	return acceptedArtifactCompletionError(apierrors.ScopeDenied, "accepted artifact source requires a currently accessible registered local project")
}

func (s *AcceptedArtifactCompletionSource) readCurrent(ctx context.Context, principal attention.Principal, projectID, eventID string) (attention.ExperienceCompletionEvidence, error) {
	var empty attention.ExperienceCompletionEvidence
	if err := s.currentProject(ctx, principal, projectID); err != nil {
		return empty, err
	}
	event, err := s.reader.ReadAcceptedArtifactEvidence(ctx, swarmdomain.Principal{ID: principal.ID, Kind: principal.Kind, ProjectID: projectID}, projectID, eventID)
	if err != nil {
		return empty, err
	}
	a := event.Artifact
	if event.EventID != eventID || event.ProjectID != projectID || event.MissionID == "" || event.OccurredAt.IsZero() || event.ActorKind != "human" || event.ActorID == "" || a.State != "accepted" || a.ID == "" || a.Version < 1 || a.WorkItemID == "" || a.AcceptanceID == "" || a.AcceptedBy != event.ActorID || a.AcceptedActorKind != event.ActorKind || strings.TrimSpace(a.Commit) == "" || strings.TrimSpace(a.ObjectRef) == "" || strings.TrimSpace(a.ChecksRef) == "" || strings.TrimSpace(a.ProjectChecksRef) == "" {
		return empty, acceptedArtifactCompletionError(apierrors.EvidenceMissing, "accepted artifact event lacks its immutable human adoption and fixed artifact references")
	}
	// 使用 outbox 中采用时的成果版本，不读取后续经验关联后的当前版本。
	value := attention.ExperienceCompletionEvidence{
		SourceKind: attention.ExperienceSourceAcceptedArtifact, EventID: event.EventID,
		Version: a.Version, OccurredAt: event.OccurredAt.UnixMilli(),
		LocalProjectID: projectID, OwnerProject: projectID,
		ActorID: event.ActorID, ActorKind: event.ActorKind, TaskID: a.WorkItemID,
		Outcome: "success", OriginEventIDs: []string{},
		Artifacts: []attention.CompletionArtifactReference{{
			ID: a.ID, Version: a.Version, Commit: a.Commit,
			ObjectRefs: []string{a.ObjectRef, a.ChecksRef, a.ProjectChecksRef},
			// 通用 SourceVersion 可能指向提案或目标，不能推断为资料引用。
			Inputs: []attention.SourceRef{},
		}},
	}
	if err := attention.ValidateCompletionSource(value, projectID, eventID, projectID); err != nil {
		return empty, err
	}
	return value, nil
}
