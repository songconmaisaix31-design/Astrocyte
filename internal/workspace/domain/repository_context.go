package domain

import "time"

// RepositoryVersion只由本机Git观察产生，根目录不进入公开投影。
type RepositoryVersion struct {
	EnvironmentID    string    `json:"environment_id"`
	RepoID           string    `json:"repo_id"`
	WorktreeID       string    `json:"worktree_id"`
	ProjectID        string    `json:"project_id"`
	Root             string    `json:"-"`
	Head             string    `json:"head"`
	IndexDigest      string    `json:"index_digest"`
	WorktreeDigest   string    `json:"worktree_digest"`
	SettingsRevision int       `json:"settings_revision"`
	Roots            []string  `json:"roots"`
	Dirty            bool      `json:"dirty"`
	ObservedAt       time.Time `json:"observed_at"`
}

type RepositoryBudget struct {
	MaxFiles   int   `json:"max_files"`
	MaxBytes   int64 `json:"max_bytes"`
	MaxResults int   `json:"max_results"`
	MaxDepth   int   `json:"max_depth"`
}

type RepositoryQuery struct {
	ProjectID string           `json:"project_id"`
	Purpose   string           `json:"purpose"`
	Roots     []string         `json:"roots"`
	Text      string           `json:"text"`
	SymbolID  string           `json:"symbol_id"`
	Budget    RepositoryBudget `json:"budget"`
}

type RepositorySpan struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type RepositorySymbol struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Kind  string         `json:"kind"`
	Span  RepositorySpan `json:"source_span"`
	Score float64        `json:"score"`
}

type RepositoryRelation struct {
	Source           string         `json:"source"`
	Target           string         `json:"target"`
	Kind             string         `json:"kind"`
	EvidenceKind     string         `json:"evidence_kind"`
	Span             RepositorySpan `json:"source_span"`
	ExtractorVersion string         `json:"extractor_version"`
}

type RepositoryIndex struct {
	Backend    string               `json:"backend"`
	Version    string               `json:"version"`
	Generation string               `json:"generation"`
	Snapshot   RepositoryVersion    `json:"snapshot"`
	State      string               `json:"state"`
	BuiltAt    time.Time            `json:"built_at"`
	MapEntries []RepositoryMapEntry `json:"map_entries,omitempty"`
}

type RepositoryMapEntry struct {
	Path           string `json:"path"`
	Responsibility string `json:"responsibility"`
	Relationships  string `json:"relationships"`
	API            string `json:"api"`
	Constraints    string `json:"constraints"`
	EvidenceKind   string `json:"evidence_kind"`
}

type RepositoryResult struct {
	Index      RepositoryIndex      `json:"index"`
	Symbols    []RepositorySymbol   `json:"symbols"`
	Relations  []RepositoryRelation `json:"relations"`
	Coverage   []string             `json:"coverage"`
	Gaps       []string             `json:"gaps"`
	MapEntries []RepositoryMapEntry `json:"map_entries,omitempty"`
}
