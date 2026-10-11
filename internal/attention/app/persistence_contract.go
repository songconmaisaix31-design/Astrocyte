package app

import (
	"context"
	"encoding/json"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Observation retains source semantics even when no core Event can be mapped.
// Unknown token volume is nil. Outcome/depth are observations, never authority.
type Observation struct {
	Source     string `json:"source"`
	SourceKind string `json:"source_kind"`
	Session    string `json:"session"`
	Scope      string `json:"scope"`
	TS         int64  `json:"ts"`
	Tokens     *int   `json:"tokens"`
	Outcome    string `json:"outcome"`
	Depth      string `json:"depth"`
	Provenance string `json:"provenance"`
}

// StoredEvent is the original receipt, not an independently stored Evidence.
// OriginKey identifies a source record across replays; ID is assigned once by
// the event store. RawBytes must be bound as a SQLite BLOB, never JSON-marshaled.
// Event=nil means pending mapping; pending receipts cannot produce Signals.
type StoredEvent struct {
	ID          string        `json:"id"`
	OriginKey   string        `json:"origin_key"`
	ReceivedAt  int64         `json:"received_at"`
	Observation Observation   `json:"observation"`
	Event       *domain.Event `json:"event"`
	RawBytes    []byte        `json:"raw_bytes"`
}

type EventReader interface {
	GetEvent(context.Context, string) (StoredEvent, error)
	ListEvents(context.Context, int64, int64) ([]StoredEvent, error)
}

// Lifecycle/history extensions stay outside the task book's AttentionNode.
type StoredAttentionNode struct {
	Node      domain.AttentionNode `json:"node"`
	UpdatedAt int64                `json:"updated_at"`
	Lifecycle string               `json:"lifecycle"`
	History   json.RawMessage      `json:"history"`
	Version   int                  `json:"version"`
}

type ShareAudit struct {
	OwnerProject  string `json:"owner_project"`
	TargetProject string `json:"target_project"`
	ActorID       string `json:"actor_id"`
	ActorKind     string `json:"actor_kind"`
	TS            int64  `json:"ts"`
	RequestID     string `json:"request_id"`
}

type ActionAudit struct {
	Action    string `json:"action"`
	ActorID   string `json:"actor_id"`
	ActorKind string `json:"actor_kind"`
	TS        int64  `json:"ts"`
	RequestID string `json:"request_id"`
	Version   int    `json:"version"`
}

type KnowledgeVersion struct {
	Version    int             `json:"version"`
	Content    json.RawMessage `json:"content"`
	RecordedAt int64           `json:"recorded_at"`
}

type StoredKnowledge struct {
	Entry                domain.KnowledgeEntry `json:"entry"`
	OwnerProject         string                `json:"owner_project"`
	Version              int                   `json:"version"`
	CreatedAt            int64                 `json:"created_at"`
	UpdatedAt            int64                 `json:"updated_at"`
	ConsolidatedAt       *int64                `json:"consolidated_at"`
	CrystallizedVersions []KnowledgeVersion    `json:"crystallized_versions"`
	Shares               []ShareAudit          `json:"shares"`
	Actions              []ActionAudit         `json:"actions"`
}

type StoredExperience struct {
	Experience    domain.Experience `json:"experience"`
	OwnerProject  string            `json:"owner_project"`
	OriginEventID string            `json:"origin_event_id"`
	Version       int               `json:"version"`
	CreatedAt     int64             `json:"created_at"`
	UpdatedAt     int64             `json:"updated_at"`
	Shares        []ShareAudit      `json:"shares"`
	Actions       []ActionAudit     `json:"actions"`
}

// Sleep persistence records one atomic daily consolidation, its measured
// representation sizes and retained-fact counts. It is not a task scheduler.
type EngineSleepRun struct {
	ID          string             `json:"id"`
	Day         string             `json:"day"`
	StartedAt   int64              `json:"started_at"`
	CompletedAt int64              `json:"completed_at"`
	Metrics     EngineSleepMetrics `json:"metrics"`
	Candidates  json.RawMessage    `json:"candidates"`
	ConflictIDs []string           `json:"conflict_ids"`
}

type EngineSleepMetrics struct {
	RetentionMeasurable bool    `json:"retention_measurable"`
	StorageMeasurable   bool    `json:"storage_measurable"`
	Representation      string  `json:"representation"`
	KeyFactsBefore      int     `json:"key_facts_before"`
	KeyFactsRetained    int     `json:"key_facts_retained"`
	BytesBefore         int64   `json:"bytes_before"`
	BytesAfter          int64   `json:"bytes_after"`
	RetentionPrecision  float64 `json:"retention_precision"`
	StorageReduction    float64 `json:"storage_reduction"`
}
