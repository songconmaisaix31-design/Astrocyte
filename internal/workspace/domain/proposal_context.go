package domain

import "time"

type ProjectContextSnapshot struct {
	EnvironmentID  string    `json:"environment_id"`
	RepoID         string    `json:"repo_id"`
	WorktreeID     string    `json:"worktree_id"`
	ProjectID      string    `json:"project_id"`
	Root           string    `json:"root"`
	Head           string    `json:"head"`
	IndexDigest    string    `json:"index_digest"`
	WorktreeDigest string    `json:"worktree_digest"`
	Generation     string    `json:"generation"`
	ObservedAt     time.Time `json:"observed_at"`
}

type PacketSource struct {
	Reference FixedReference `json:"reference"`
	ObjectRef string         `json:"object_ref"`
	Title     string         `json:"title"`
	Locations []string       `json:"locations"`
}

type ImmutableContextPacket struct {
	ID                      string                 `json:"id"`
	SchemaVersion           int                    `json:"schema_version"`
	ProjectID               string                 `json:"project_id"`
	ProjectSettingsRevision int                    `json:"project_settings_revision"`
	PredecessorID           string                 `json:"predecessor_id,omitempty"`
	GrantID                 string                 `json:"grant_id"`
	GrantEpoch              int                    `json:"grant_epoch"`
	ProposalID              string                 `json:"proposal_id"`
	ProposalRevision        int                    `json:"proposal_revision"`
	Goal                    Goal                   `json:"goal"`
	Snapshot                ProjectContextSnapshot `json:"snapshot"`
	Sources                 []PacketSource         `json:"sources"`
	Scope                   ExecutionScope         `json:"scope"`
	Stops                   StopConditions         `json:"stops"`
	Decisions               []string               `json:"decisions"`
	Results                 []string               `json:"results"`
	Checks                  []string               `json:"checks"`
	Failures                []string               `json:"failures"`
	OpenQuestions           []string               `json:"open_questions"`
	NextSteps               []string               `json:"next_steps"`
	ObjectRef               string                 `json:"object_ref"`
	CreatedAt               time.Time              `json:"created_at"`
}
