package domain

import "time"

// RegisteredProject observations grant no context, write, execution or model
// permission. Only manual registration binds a candidate to an approved space.
type RegisteredProject struct {
	ProjectKey   string                     `json:"project_key"`
	Contributors []ProjectContributor       `json:"contributors"`
	Root         string                     `json:"root"`
	Name         string                     `json:"name"`
	RepoID       string                     `json:"repo_id"`
	Source       string                     `json:"source"`
	ParentRoot   string                     `json:"parent_root"`
	Git          ProjectGitObservation      `json:"git"`
	Activity     ProjectActivityObservation `json:"activity"`
	Limitations  []string                   `json:"limitations"`
}

type ProjectGitObservation struct {
	Status       string     `json:"status"`
	Reason       string     `json:"reason"`
	Branch       *string    `json:"branch"`
	Head         *string    `json:"head"`
	LastCommitAt *time.Time `json:"last_commit_at"`
	Dirty        *bool      `json:"dirty"`
}

type ProjectActivityObservation struct {
	Status         string     `json:"status"`
	Source         string     `json:"source"`
	CreatedWithCLI *string    `json:"created_with_cli"`
	LastActivityAt *time.Time `json:"last_activity_at"`
}

type ProjectDiscoveryFailure struct {
	Root   string `json:"root"`
	Reason string `json:"reason"`
}

// Unknown observation time stays null, including before the first refresh.
// Errors remain per root; a partial discovery never replaces them with success.
type ProjectDiscoverySnapshot struct {
	Board      []ProjectSummary           `json:"board"`
	Sources    []ProjectSourceObservation `json:"sources"`
	Status     string                     `json:"status"`
	ObservedAt *time.Time                 `json:"observed_at"`
	Projects   []RegisteredProject        `json:"projects"`
	Failures   []ProjectDiscoveryFailure  `json:"failures"`
}

// A session header establishes historical association, never current execution,
// completion, model readiness or project read permission.
type ProjectContributor struct {
	CLI        string     `json:"cli"`
	Source     string     `json:"source"`
	Root       string     `json:"root"`
	SessionID  string     `json:"session_id"`
	ObservedAt *time.Time `json:"observed_at"`
	CreatedAt  *time.Time `json:"created_at"`
	ActivityAt *time.Time `json:"activity_at"`
}

type ProjectSourceObservation struct {
	CLI                  string `json:"cli"`
	Source               string `json:"source"`
	Status               string `json:"status"`
	Reason               string `json:"reason"`
	EntriesExamined      int    `json:"entries_examined"`
	MatchedHeaders       int    `json:"matched_headers"`
	HeadersExamined      int    `json:"headers_examined"`
	MatchedRoots         int    `json:"matched_roots"`
	RetainedAssociations int    `json:"retained_associations"`
}

type ProjectHumanMetadata struct {
	Notes     string     `json:"notes"`
	Review    string     `json:"review"`
	Group     string     `json:"group"`
	Intent    string     `json:"intent"`
	Archived  bool       `json:"archived"`
	Revision  int        `json:"revision"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type ProjectSummary struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Roots          []string             `json:"roots"`
	Observations   []RegisteredProject  `json:"observations"`
	Contributors   []ProjectContributor `json:"contributors"`
	LastActivityAt *time.Time           `json:"last_activity_at"`
	Limitations    []string             `json:"limitations"`
	Human          ProjectHumanMetadata `json:"human"`
}
