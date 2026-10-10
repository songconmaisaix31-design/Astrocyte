package domain

import "time"

// Caller is supplied by the authenticated transport, never decoded from a body.
type Caller struct {
	OperationID string `json:"-"`
	Kind        string `json:"-"`
	ID          string `json:"-"`
	ProjectID   string `json:"-"`
}

type ProjectSettings struct {
	HistoryRoots      map[string]string `json:"history_roots"`
	Revision          int               `json:"revision"`
	AllowDirectory    bool              `json:"allow_directory"`
	AllowedSubdirs    []string          `json:"allowed_subdirs"`
	ExpandReferences  bool              `json:"expand_references"`
	AllowedActions    []string          `json:"allowed_actions"`
	AllowedTools      []string          `json:"allowed_tools"`
	ExternalModelCLI  string            `json:"external_model_cli"`
	AllowAgentControl bool              `json:"allow_agent_control"`
}

type LocalProject struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Root      string          `json:"root"`
	SpaceID   string          `json:"space_id"`
	Settings  ProjectSettings `json:"settings"`
	CreatedAt time.Time       `json:"created_at"`
}

type ProjectCandidate struct {
	Root string `json:"root"`
	Name string `json:"name"`
}

type ProjectGrant struct {
	TokenDigest string     `json:"-"`
	ExpiresAt   *time.Time `json:"expires_at"`
	ProjectID   string     `json:"project_id"`
	AgentID     string     `json:"agent_id"`
	Actions     []string   `json:"actions"`
	RevokedAt   *time.Time `json:"revoked_at"`
}

type FixedReference struct {
	MaterialID string `json:"material_id"`
	Revision   int    `json:"revision"`
}

// ContextMaterial contains a fixed source version; the reference port must check
// project membership and material access on every call, including linked nodes.
type ContextMaterial struct {
	Expanded  bool           `json:"expanded"`
	Reference FixedReference `json:"reference"`
	Title     string         `json:"title"`
	Text      string         `json:"text"`
}

type ContextFile struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Text    string `json:"text"`
}

type ContextPacket struct {
	ID               string            `json:"id"`
	SchemaVersion    int               `json:"schema_version"`
	ProjectID        string            `json:"project_id"`
	SettingsRevision int               `json:"settings_revision"`
	Mode             string            `json:"mode"`
	Materials        []ContextMaterial `json:"materials"`
	Files            []ContextFile     `json:"files"`
	CreatedAt        time.Time         `json:"created_at"`
}

type ContextRequest struct {
	References []FixedReference `json:"references"`
	Files      []string         `json:"files"`
}

type RegisterProjectCommand struct {
	Name    string `json:"name"`
	Root    string `json:"root"`
	SpaceID string `json:"space_id"`
}

type SettingsCommand struct {
	ExpectedRevision int             `json:"expected_revision"`
	Settings         ProjectSettings `json:"settings"`
}

type NativeCommand struct {
	OperationID string         `json:"-"`
	CLI         string         `json:"cli"`
	Message     string         `json:"message"`
	Context     ContextRequest `json:"context"`
	// Lifetime is bounded by the controller, covering idle time and all turns.
	DeadlineSeconds int    `json:"deadline_seconds"`
	Mode            string `json:"mode"`
	SourceSessionID string `json:"source_session_id"`
}

type NativeSession struct {
	Ownership        string        `json:"ownership"`
	HistoryPath      string        `json:"-"`
	LastOperationID  string        `json:"last_operation_id"`
	PendingOperation string        `json:"pending_operation"`
	ID               string        `json:"id"`
	ProjectID        string        `json:"project_id"`
	CLI              string        `json:"cli"`
	NativeID         string        `json:"native_id"`
	Version          string        `json:"version"`
	Mode             string        `json:"mode"`
	Status           string        `json:"status"`
	StopConfirmed    bool          `json:"stop_confirmed"`
	ContextPacket    ContextPacket `json:"context_packet"`
	SourceSessionID  string        `json:"source_session_id"`
	UpdatedAt        time.Time     `json:"updated_at"`
	// Cooperative CLI tools do not provide an OS read-isolation guarantee.
	Limitations []string `json:"limitations"`
}

type NativeEvent struct {
	Sequence int    `json:"sequence"`
	Kind     string `json:"kind"`
	Text     string `json:"text"`
}

type NativeObservation struct {
	Status          string        `json:"status"`
	Events          []NativeEvent `json:"events"`
	StopConfirmed   bool          `json:"stop_confirmed"`
	OutputTruncated bool          `json:"output_truncated"`
}

// NativeRequest is internal: executable, credentials, arbitrary args and native
// history paths are never selected by HTTP input.
type NativeRequest struct {
	SessionID string
	Project   LocalProject
	Session   NativeSession
	Command   NativeCommand
	Packet    ContextPacket
}

type AgentToken struct {
	Token     string    `json:"token"`
	ProjectID string    `json:"project_id"`
	AgentID   string    `json:"agent_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type TextRequest struct {
	CLI             string
	Prompt          string
	DeadlineSeconds int
	OutputLimit     int
}

type TextResult struct {
	Text     string
	NativeID string
	Version  string
	Model    *string
	Usage    any
}
