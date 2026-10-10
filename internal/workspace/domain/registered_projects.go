package domain

import "time"

// RegisteredProject observations grant no context, write, execution or model
// permission. Only manual registration binds a candidate to an approved space.
type RegisteredProject struct {
	Root        string                     `json:"root"`
	Name        string                     `json:"name"`
	RepoID      string                     `json:"repo_id"`
	Source      string                     `json:"source"`
	ParentRoot  string                     `json:"parent_root"`
	Git         ProjectGitObservation      `json:"git"`
	Activity    ProjectActivityObservation `json:"activity"`
	Limitations []string                   `json:"limitations"`
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
	Status     string                    `json:"status"`
	ObservedAt *time.Time                `json:"observed_at"`
	Projects   []RegisteredProject       `json:"projects"`
	Failures   []ProjectDiscoveryFailure `json:"failures"`
}
