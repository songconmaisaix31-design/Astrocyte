package domain

import "time"

// GitHubMetadata is an anonymous public API observation, never a code checkout.
type GitHubMetadata struct {
	MetadataSource string    `json:"metadata_source"`
	UnknownFields  []string  `json:"unknown_fields"`
	GitHubID       int64     `json:"github_id"`
	FullName       string    `json:"full_name"`
	HTMLURL        string    `json:"html_url"`
	CloneURL       string    `json:"clone_url"`
	Description    string    `json:"description"`
	DefaultBranch  string    `json:"default_branch"`
	Language       string    `json:"language"`
	Stars          int       `json:"stars"`
	Archived       bool      `json:"archived"`
	Fork           bool      `json:"fork"`
	UpdatedAt      time.Time `json:"updated_at"`
	PushedAt       time.Time `json:"pushed_at"`
}

// Legacy caches and test sources cannot acquire REST provenance by default.
func NormalizeGitHubMetadata(m GitHubMetadata) GitHubMetadata {
	if m.MetadataSource == "" {
		m.MetadataSource = "unknown"
	}
	if m.UnknownFields == nil {
		m.UnknownFields = []string{}
		if m.MetadataSource == "unknown" {
			m.UnknownFields = []string{"description", "default_branch", "language", "stars", "archived", "fork", "updated_at", "pushed_at"}
		}
	}
	return m
}

type GitHubRepository struct {
	ID               string         `json:"id"`
	Revision         int            `json:"revision"`
	MetadataRevision int            `json:"metadata_revision"`
	Metadata         GitHubMetadata `json:"metadata"`
	SyncStatus       string         `json:"sync_status"`
	SyncedAt         time.Time      `json:"synced_at"`
	SpaceID          string         `json:"space_id"`
	ProjectID        string         `json:"project_id"`
	CloneStatus      string         `json:"clone_status"`
	Root             string         `json:"root"`
	Head             string         `json:"head"`
	CloneAttempts    int            `json:"clone_attempts"`
	LastError        string         `json:"last_error"`
}

type RepositoryClone struct {
	Root string `json:"root"`
	Head string `json:"head"`
}

type RepositorySyncCommand struct {
	Input string `json:"input"`
}

type RepositoryPlacementCommand struct {
	SpaceID          string `json:"space_id"`
	ExpectedRevision int    `json:"expected_revision"`
}
