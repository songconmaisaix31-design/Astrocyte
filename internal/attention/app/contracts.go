package app

import (
	"context"
	"encoding/json"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

// Principal is set by the trusted transport, never decoded from a command body.
// S1 has a local human session and an optional read-only agent credential.
type Principal struct {
	ID   string
	Kind string
}
type CommandMeta struct {
	SchemaVersion   int    `json:"schema_version"`
	RequestID       string `json:"request_id"`
	ExpectedVersion int    `json:"expected_version"`
	IdempotencyKey  string `json:"-"`
}
type SourceRef struct {
	MaterialID string  `json:"material_id"`
	Revision   int     `json:"revision"`
	Locator    string  `json:"locator"`
	Span       *string `json:"span,omitempty"`
}
type DimensionScore struct {
	Value  *float64 `json:"value"`
	Reason string   `json:"reason"`
}
type Dimensions struct {
	GoalProgress       DimensionScore `json:"goal_progress"`
	CurrentInterest    DimensionScore `json:"current_interest"`
	ProjectImprovement DimensionScore `json:"project_improvement"`
	Originality        DimensionScore `json:"originality"`
}
type Provenance struct {
	Processor string `json:"processor"`
	Version   string `json:"version"`
	Mode      string `json:"mode"`
	Source    string `json:"source"`
}
type Material struct {
	ID               string    `json:"id"`
	SourceLocator    string    `json:"source_locator"`
	Kind             string    `json:"kind"`
	CurrentRevision  int       `json:"current_revision"`
	Lifecycle        string    `json:"lifecycle"`
	Title            string    `json:"title"`
	CollectionReason *string   `json:"collection_reason"`
	SourceSpans      []string  `json:"source_spans"`
	ImportStatus     string    `json:"import_status"`
	HumanUsageCount  int       `json:"human_usage_count"`
	AgentUsageCount  int       `json:"agent_usage_count"`
	AttentionScore   float64   `json:"attention_score"`
	LongTermValue    float64   `json:"long_term_value"`
	CreatedAt        time.Time `json:"created_at"`
}
type MaterialRevision struct {
	MaterialID    string          `json:"material_id"`
	Revision      int             `json:"revision"`
	SourceKey     string          `json:"source_key"`
	SourceLocator string          `json:"source_locator"`
	ContentDigest string          `json:"content_digest"`
	ObjectRef     string          `json:"object_ref"`
	SourceSpans   []string        `json:"source_spans"`
	Provenance    Provenance      `json:"provenance"`
	CreatedAt     time.Time       `json:"created_at"`
	Summary       string          `json:"summary,omitempty"`
	Attachments   []AttachmentRef `json:"attachments,omitempty"`
}
type AttachmentRef struct {
	Name          string `json:"name"`
	MediaType     string `json:"media_type"`
	ObjectRef     string `json:"object_ref"`
	SourceLocator string `json:"source_locator"`
}
type SourceAttachment struct {
	Name          string
	MediaType     string
	Data          []byte
	SourceLocator string
}
type Usage struct {
	ID         string    `json:"id"`
	MaterialID string    `json:"material_id"`
	ActorID    string    `json:"actor_id"`
	ActorKind  string    `json:"actor_kind"`
	Action     string    `json:"action"`
	OccurredAt time.Time `json:"occurred_at"`
}
type MaterialDetail struct {
	SchemaVersion int                `json:"schema_version"`
	Material      Material           `json:"material"`
	Revisions     []MaterialRevision `json:"revisions"`
	Distillations []Distillation     `json:"distillations"`
	Uses          []Usage            `json:"uses"`
}
type ImportMaterialCommand struct {
	CommandMeta
	SourceLocator    string   `json:"source_locator"`
	SourceKey        string   `json:"source_key"`
	Kind             string   `json:"kind"`
	ContentDigest    string   `json:"content_digest"`
	ExportText       string   `json:"export_text,omitempty"`
	LocalFileRef     string   `json:"local_file_ref,omitempty"`
	CollectionReason *string  `json:"collection_reason,omitempty"`
	SourceSpans      []string `json:"source_spans,omitempty"`
	Title            string   `json:"title,omitempty"`
	Adapter          string   `json:"adapter,omitempty"`
}
type ImportedSource struct {
	SourceKey     string
	SourceLocator string
	Kind          string
	Title         string
	Text          string
	SourceSpans   []string
	Provenance    Provenance
	Summary       string
	Attachments   []SourceAttachment
}
type SourceReader interface {
	ReadSource(context.Context, ImportMaterialCommand) (ImportedSource, error)
}

// Publish returns a digest-addressed immutable object reference after atomic publication.
// Read accepts object references only, never arbitrary filesystem paths.
type ObjectStore interface {
	Publish(context.Context, []byte) (string, error)
	Read(context.Context, string) ([]byte, error)
}
type Distillation struct {
	ID                  string      `json:"id"`
	InputRefs           []SourceRef `json:"input_refs"`
	Stage               string      `json:"stage"`
	OutputRef           string      `json:"output_ref"`
	OutputText          string      `json:"output_text"`
	Status              string      `json:"status"`
	NextQuestion        *string     `json:"next_question"`
	Question            string      `json:"question"`
	ProcessingConfig    string      `json:"processing_config"`
	RelatedRefs         []SourceRef `json:"related_refs"`
	RelatedIdeas        []string    `json:"related_ideas"`
	Conflicts           []string    `json:"conflicts"`
	PendingQuestions    []string    `json:"pending_questions"`
	GoalRefs            []string    `json:"goal_refs"`
	ExistingAssets      []string    `json:"existing_assets"`
	ExpectedImprovement string      `json:"expected_improvement"`
	MinimumArtifact     string      `json:"minimum_artifact"`
	MissingEvidence     []string    `json:"missing_evidence"`
	Provenance          Provenance  `json:"provenance"`
	ReuseKey            string      `json:"reuse_key"`
	CreatedAt           time.Time   `json:"created_at"`
}
type RecordDistillationCommand struct {
	CommandMeta
	InputRefs           []SourceRef `json:"input_refs"`
	Stage               string      `json:"stage"`
	OutputText          string      `json:"output_text"`
	NextQuestion        *string     `json:"next_question,omitempty"`
	Question            string      `json:"question"`
	ProcessingConfig    string      `json:"processing_config"`
	RelatedRefs         []SourceRef `json:"related_refs,omitempty"`
	RelatedIdeas        []string    `json:"related_ideas,omitempty"`
	Conflicts           []string    `json:"conflicts,omitempty"`
	PendingQuestions    []string    `json:"pending_questions,omitempty"`
	GoalRefs            []string    `json:"goal_refs,omitempty"`
	ExistingAssets      []string    `json:"existing_assets,omitempty"`
	ExpectedImprovement string      `json:"expected_improvement,omitempty"`
	MinimumArtifact     string      `json:"minimum_artifact,omitempty"`
	MissingEvidence     []string    `json:"missing_evidence,omitempty"`
}
type DistillationResult struct {
	SchemaVersion int          `json:"schema_version"`
	Distillation  Distillation `json:"distillation"`
	Reused        bool         `json:"reused"`
}
type Opportunity struct {
	ID              string      `json:"id"`
	Revision        int         `json:"revision"`
	State           string      `json:"state"`
	Title           string      `json:"title"`
	EvidenceRefs    []SourceRef `json:"evidence_refs"`
	GoalRefs        []string    `json:"goal_refs"`
	Dimensions      Dimensions  `json:"dimensions"`
	NextStep        string      `json:"next_step"`
	MissingEvidence []string    `json:"missing_evidence"`
	Purpose         string      `json:"purpose"`
	DistillationIDs []string    `json:"distillation_ids"`
	CreatedAt       time.Time   `json:"created_at"`
}
type OpportunityCommand struct {
	CommandMeta
	Title           string      `json:"title"`
	EvidenceRefs    []SourceRef `json:"evidence_refs"`
	GoalRefs        []string    `json:"goal_refs,omitempty"`
	Dimensions      Dimensions  `json:"dimensions"`
	NextStep        string      `json:"next_step"`
	MissingEvidence []string    `json:"missing_evidence,omitempty"`
	Purpose         string      `json:"purpose"`
	DistillationIDs []string    `json:"distillation_ids"`
}
type ReviewOpportunityCommand struct {
	CommandMeta
	Feedback   string      `json:"feedback"`
	Reason     string      `json:"reason"`
	Dimensions *Dimensions `json:"dimensions,omitempty"`
}
type Review struct {
	ID            string      `json:"id"`
	OpportunityID string      `json:"opportunity_id"`
	Revision      int         `json:"revision"`
	Feedback      string      `json:"feedback"`
	Reason        string      `json:"reason"`
	Dimensions    *Dimensions `json:"dimensions,omitempty"`
	ActorID       string      `json:"actor_id"`
	CreatedAt     time.Time   `json:"created_at"`
}
type OpportunityDetail struct {
	SchemaVersion int           `json:"schema_version"`
	Opportunity   Opportunity   `json:"opportunity"`
	Revisions     []Opportunity `json:"revisions"`
	Reviews       []Review      `json:"reviews"`
}
type RecordUseCommand struct {
	CommandMeta
	Action string `json:"action"`
}
type Job struct {
	SchemaVersion    int                     `json:"schema_version"`
	JobID            string                  `json:"job_id"`
	Status           string                  `json:"status"`
	Kind             string                  `json:"kind"`
	Version          int                     `json:"version"`
	DedupeKey        string                  `json:"dedupe_key"`
	OperationID      string                  `json:"operation_id"`
	MaterialID       *string                 `json:"material_id"`
	MaterialRevision *int                    `json:"material_revision"`
	Error            *apierrors.ServiceError `json:"error"`
	Attempts         int                     `json:"attempts"`
	MaxAttempts      int                     `json:"max_attempts"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
	DeadlineAt       time.Time               `json:"deadline_at"`
	CancelRequested  bool                    `json:"cancel_requested"`
	// Payload and caller are durable internal state, never serialized to clients.
	Payload json.RawMessage `json:"-"`
	Caller  Principal       `json:"-"`
}
type ImportJobResult struct {
	SchemaVersion int    `json:"schema_version"`
	JobID         string `json:"job_id"`
	Status        string `json:"status"`
}
type VersionResult struct {
	SchemaVersion int    `json:"schema_version"`
	ID            string `json:"id"`
	Version       int    `json:"version"`
}
type ContentResult struct {
	SchemaVersion int        `json:"schema_version"`
	MaterialID    string     `json:"material_id"`
	Revision      int        `json:"revision"`
	ContentDigest string     `json:"content_digest"`
	Text          string     `json:"text"`
	Provenance    Provenance `json:"provenance"`
}

// AttentionService is the S1 boundary. Plain queries never count human attention;
// explicit RecordUse receives the authenticated principal. Agent credentials are read-only.
type AttentionService interface {
	ListMaterials(context.Context) (apierrors.ListResult, error)
	ListOpportunities(context.Context) (apierrors.ListResult, error)
	ImportMaterial(context.Context, Principal, ImportMaterialCommand) (ImportJobResult, error)
	GetMaterial(context.Context, Principal, string) (MaterialDetail, error)
	GetContent(context.Context, Principal, string, int) (ContentResult, error)
	RecordDistillation(context.Context, Principal, RecordDistillationCommand) (DistillationResult, error)
	ListDistillations(context.Context, Principal) (apierrors.ListResult, error)
	CreateOpportunity(context.Context, Principal, OpportunityCommand) (OpportunityDetail, error)
	ReviseOpportunity(context.Context, Principal, string, OpportunityCommand) (OpportunityDetail, error)
	GetOpportunity(context.Context, Principal, string) (OpportunityDetail, error)
	ReviewOpportunity(context.Context, Principal, string, ReviewOpportunityCommand) (VersionResult, error)
	RecordUse(context.Context, Principal, string, RecordUseCommand) (VersionResult, error)
	ListJobs(context.Context, Principal) (apierrors.ListResult, error)
	GetJob(context.Context, Principal, string) (Job, error)
	RetryJob(context.Context, Principal, string, CommandMeta) (Job, error)
	CancelJob(context.Context, Principal, string, CommandMeta) (Job, error)
}

// Repository serializes each callback in one transaction. Do not call external
// providers or objects inside callbacks. All methods are scoped to Attention.
type Repository interface {
	WithTx(context.Context, func(AttentionTx) error) error
}
type Receipt struct {
	Digest string
	Result json.RawMessage
}
type OutboxEvent struct {
	ID               string
	Type             string
	AggregateID      string
	AggregateVersion int
	OccurredAt       time.Time
	CorrelationID    string
	CausationID      string
	Payload          json.RawMessage
}
type AttentionTx interface {
	ListMaterials() ([]MaterialDetail, error)
	LoadMaterial(string) (MaterialDetail, error)
	FindMaterialBySourceKey(string) (MaterialDetail, error)
	SaveMaterial(MaterialDetail, int) error
	ListDistillations() ([]Distillation, error)
	FindDistillationByReuseKey(string) (Distillation, error)
	SaveDistillation(Distillation) error
	ListOpportunities() ([]OpportunityDetail, error)
	LoadOpportunity(string) (OpportunityDetail, error)
	SaveOpportunity(OpportunityDetail, int) error
	ListJobs() ([]Job, error)
	LoadJob(string) (Job, error)
	FindJobByDedupeKey(string) (Job, error)
	SaveJob(Job, int) error
	LoadReceipt(caller, command, key string) (Receipt, error)
	SaveReceipt(caller, command, key string, receipt Receipt) error
	AppendEvent(OutboxEvent) error
}
