package httpapi

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type WorkspaceSnapshotDTO struct {
	EnvironmentID  string    `json:"environment_id"`
	RepoID         string    `json:"repo_id"`
	WorktreeID     string    `json:"worktree_id"`
	ProjectID      string    `json:"project_id"`
	Head           string    `json:"head"`
	IndexDigest    string    `json:"index_digest"`
	WorktreeDigest string    `json:"worktree_digest"`
	Generation     string    `json:"generation"`
	ObservedAt     time.Time `json:"observed_at"`
}

type WorkspaceContextPacketDTO struct {
	ID                      string                `json:"id"`
	SchemaVersion           int                   `json:"schema_version"`
	ProjectID               string                `json:"project_id"`
	ProjectSettingsRevision int                   `json:"project_settings_revision"`
	PredecessorID           string                `json:"predecessor_id,omitempty"`
	GrantID                 string                `json:"grant_id"`
	GrantEpoch              int                   `json:"grant_epoch"`
	ProposalID              string                `json:"proposal_id"`
	ProposalRevision        int                   `json:"proposal_revision"`
	Goal                    domain.Goal           `json:"goal"`
	Snapshot                WorkspaceSnapshotDTO  `json:"snapshot"`
	Scope                   domain.ExecutionScope `json:"scope"`
	Stops                   domain.StopConditions `json:"stops"`
	Sources                 []domain.PacketSource `json:"sources"`
	Decisions               []string              `json:"decisions"`
	Results                 []string              `json:"results"`
	Checks                  []string              `json:"checks"`
	Failures                []string              `json:"failures"`
	OpenQuestions           []string              `json:"open_questions"`
	NextSteps               []string              `json:"next_steps"`
	ObjectRef               string                `json:"object_ref"`
	CreatedAt               time.Time             `json:"created_at"`
}

func workspaceContextPacketDTO(p domain.ImmutableContextPacket) (WorkspaceContextPacketDTO, error) {
	if !filepath.IsAbs(p.Snapshot.Root) {
		return WorkspaceContextPacketDTO{}, domain.WorkspaceError(apierrors.EvidenceMissing, "context snapshot has no absolute local identity")
	}
	scope := p.Scope
	scope.Roots = make([]string, 0, len(p.Scope.Roots))
	for _, root := range p.Scope.Roots {
		rel, err := filepath.Rel(p.Snapshot.Root, root)
		if err != nil {
			return WorkspaceContextPacketDTO{}, domain.WorkspaceError(apierrors.ScopeDenied, "context root cannot be expressed within its fixed project")
		}
		if !filepath.IsAbs(root) || filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return WorkspaceContextPacketDTO{}, domain.WorkspaceError(apierrors.ScopeDenied, "context root is outside its fixed project")
		}
		scope.Roots = append(scope.Roots, filepath.ToSlash(rel))
	}
	snapshot := p.Snapshot
	return WorkspaceContextPacketDTO{ID: p.ID, SchemaVersion: p.SchemaVersion, ProjectID: p.ProjectID, ProjectSettingsRevision: p.ProjectSettingsRevision, PredecessorID: p.PredecessorID, GrantID: p.GrantID, GrantEpoch: p.GrantEpoch, ProposalID: p.ProposalID, ProposalRevision: p.ProposalRevision, Goal: p.Goal, Snapshot: WorkspaceSnapshotDTO{EnvironmentID: snapshot.EnvironmentID, RepoID: snapshot.RepoID, WorktreeID: snapshot.WorktreeID, ProjectID: snapshot.ProjectID, Head: snapshot.Head, IndexDigest: snapshot.IndexDigest, WorktreeDigest: snapshot.WorktreeDigest, Generation: snapshot.Generation, ObservedAt: snapshot.ObservedAt}, Scope: scope, Stops: p.Stops, Sources: p.Sources, Decisions: p.Decisions, Results: p.Results, Checks: p.Checks, Failures: p.Failures, OpenQuestions: p.OpenQuestions, NextSteps: p.NextSteps, ObjectRef: p.ObjectRef, CreatedAt: p.CreatedAt}, nil
}

type WorkspaceCapabilityDTO struct {
	Status    string     `json:"status"`
	CheckedAt *time.Time `json:"checked_at"`
}

type WorkspaceSessionBindingDTO struct {
	ID           string                            `json:"id"`
	Version      int                               `json:"version"`
	ProjectID    string                            `json:"project_id"`
	SessionID    string                            `json:"session_id"`
	AdapterID    string                            `json:"adapter_id"`
	WorktreeID   string                            `json:"worktree_id"`
	GrantID      string                            `json:"grant_id"`
	GrantEpoch   int                               `json:"grant_epoch"`
	PacketID     string                            `json:"packet_id"`
	MemberID     string                            `json:"member_id"`
	Status       string                            `json:"status"`
	Capabilities map[string]WorkspaceCapabilityDTO `json:"capabilities"`
	CreatedAt    time.Time                         `json:"created_at"`
	UpdatedAt    time.Time                         `json:"updated_at"`
}

func workspaceSessionBindingDTO(b domain.SessionBinding) WorkspaceSessionBindingDTO {
	capabilities := map[string]WorkspaceCapabilityDTO{}
	for name := range domain.UnknownNativeCapabilities() {
		if observed, ok := b.Capabilities[name]; ok {
			capabilities[name] = WorkspaceCapabilityDTO{Status: observed.Status, CheckedAt: observed.CheckedAt}
		}
	}
	return WorkspaceSessionBindingDTO{ID: b.ID, Version: b.Version, ProjectID: b.ProjectID, SessionID: b.SessionID, AdapterID: b.AdapterID, WorktreeID: b.WorktreeID, GrantID: b.GrantID, GrantEpoch: b.GrantEpoch, PacketID: b.PacketID, MemberID: b.MemberID, Status: b.Status, Capabilities: capabilities, CreatedAt: b.CreatedAt, UpdatedAt: b.UpdatedAt}
}
