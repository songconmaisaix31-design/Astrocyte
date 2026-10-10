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
	Model     string `json:"model,omitempty"`
	Processor string `json:"processor"`
	Version   string `json:"version"`
	Mode      string `json:"mode"`
	Source    string `json:"source"`
}
type Material struct {
	DomainIDs                []string           `json:"domain_ids"`
	RankingStrategy          string             `json:"ranking_strategy,omitempty"`
	RankingReason            string             `json:"ranking_reason,omitempty"`
	AttentionHalfLifeSeconds float64            `json:"attention_half_life_seconds,omitempty"`
	AttentionWeights         map[string]float64 `json:"attention_weights,omitempty"`
	Pinned                   bool               `json:"pinned"`
	Version                  int                `json:"version"`
	ID                       string             `json:"id"`
	SourceLocator            string             `json:"source_locator"`
	Kind                     string             `json:"kind"`
	CurrentRevision          int                `json:"current_revision"`
	Lifecycle                string             `json:"lifecycle"`
	Title                    string             `json:"title"`
	CollectionReason         *string            `json:"collection_reason"`
	SourceSpans              []string           `json:"source_spans"`
	ImportStatus             string             `json:"import_status"`
	HumanUsageCount          int                `json:"human_usage_count"`
	AgentUsageCount          int                `json:"agent_usage_count"`
	AttentionScore           float64            `json:"attention_score"`
	LongTermValue            float64            `json:"long_term_value"`
	CreatedAt                time.Time          `json:"created_at"`
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
	Refresh          bool     `json:"refresh,omitempty"`
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
	PriorDistillationIDs []string             `json:"prior_distillation_ids,omitempty"`
	CandidateSuggestion  *CandidateSuggestion `json:"candidate_suggestion,omitempty"`
	ID                   string               `json:"id"`
	InputRefs            []SourceRef          `json:"input_refs"`
	Stage                string               `json:"stage"`
	OutputRef            string               `json:"output_ref"`
	OutputText           string               `json:"output_text"`
	Status               string               `json:"status"`
	NextQuestion         *string              `json:"next_question"`
	Question             string               `json:"question"`
	ProcessingConfig     string               `json:"processing_config"`
	RelatedRefs          []SourceRef          `json:"related_refs"`
	RelatedIdeas         []string             `json:"related_ideas"`
	Conflicts            []string             `json:"conflicts"`
	PendingQuestions     []string             `json:"pending_questions"`
	GoalRefs             []string             `json:"goal_refs"`
	ExistingAssets       []string             `json:"existing_assets"`
	ExpectedImprovement  string               `json:"expected_improvement"`
	MinimumArtifact      string               `json:"minimum_artifact"`
	MissingEvidence      []string             `json:"missing_evidence"`
	Provenance           Provenance           `json:"provenance"`
	ReuseKey             string               `json:"reuse_key"`
	CreatedAt            time.Time            `json:"created_at"`
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
	CompositeScore        *float64    `json:"composite_score,omitempty"`
	RankingStrategy       string      `json:"ranking_strategy,omitempty"`
	RankingReason         string      `json:"ranking_reason,omitempty"`
	RankingProfileVersion *int        `json:"ranking_profile_version,omitempty"`
	Version               int         `json:"version"`
	ID                    string      `json:"id"`
	Revision              int         `json:"revision"`
	State                 string      `json:"state"`
	Title                 string      `json:"title"`
	EvidenceRefs          []SourceRef `json:"evidence_refs"`
	GoalRefs              []string    `json:"goal_refs"`
	Dimensions            Dimensions  `json:"dimensions"`
	NextStep              string      `json:"next_step"`
	MissingEvidence       []string    `json:"missing_evidence"`
	Purpose               string      `json:"purpose"`
	DistillationIDs       []string    `json:"distillation_ids"`
	CreatedAt             time.Time   `json:"created_at"`
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
type UpdateMaterialCommand struct {
	CommandMeta
	Pinned           bool    `json:"pinned"`
	Lifecycle        string  `json:"lifecycle"`
	CollectionReason *string `json:"collection_reason"`
}
type Job struct {
	DistillationID   *string                 `json:"distillation_id,omitempty"`
	ExternalStarted  bool                    `json:"external_started"`
	DeliveryUnknown  bool                    `json:"delivery_unknown"`
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
type AttachmentContent struct {
	Data      []byte
	MediaType string
	Name      string
}

// MaterialDomain is a human classification; membership never grants Agent access.
type MaterialDomain struct {
	ID          string    `json:"id"`
	Version     int       `json:"version"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
type MaterialDomainCommand struct {
	CommandMeta
	Title       string `json:"title"`
	Description string `json:"description"`
}
type MaterialDomainResult struct {
	SchemaVersion int            `json:"schema_version"`
	Domain        MaterialDomain `json:"domain"`
}
type SetMaterialDomainsCommand struct {
	CommandMeta
	DomainIDs []string `json:"domain_ids"`
}

// ProjectSpace is an Attention projection of human-selected file references,
// not a Workspace project/approval or an executable work directory.
type ProjectSpace struct {
	ID           string      `json:"id"`
	Version      int         `json:"version"`
	Title        string      `json:"title"`
	MaterialRefs []SourceRef `json:"material_refs"`
	CreatedAt    time.Time   `json:"created_at"`
}
type ProjectSpaceCommand struct {
	CommandMeta
	Title string `json:"title"`
}
type ProjectSpaceResult struct {
	SchemaVersion int          `json:"schema_version"`
	Space         ProjectSpace `json:"space"`
}
type ReferenceMaterialCommand struct {
	CommandMeta
	MaterialID string `json:"material_id"`
	Revision   int    `json:"revision"`
}
type RemoveMaterialReferenceCommand struct {
	CommandMeta
	MaterialID string `json:"material_id"`
}

// Only selected, authorized snapshots enter the processor. There are no database
// handles, object paths, shell capabilities or authentication fields in this input.
type SourceSnapshot struct {
	Ref           SourceRef  `json:"ref"`
	SourceKey     string     `json:"source_key"`
	Title         string     `json:"title"`
	Text          string     `json:"text"`
	Summary       string     `json:"summary"`
	ContentDigest string     `json:"content_digest"`
	Provenance    Provenance `json:"provenance"`
}
type DistillationInput struct {
	PriorDistillations []Distillation   `json:"prior_distillations,omitempty"`
	JobID              string           `json:"job_id"`
	OperationID        string           `json:"operation_id"`
	Inputs             []SourceSnapshot `json:"inputs"`
	Stage              string           `json:"stage"`
	ProcessingConfig   string           `json:"processing_config"`
	Question           string           `json:"question"`
}
type DistillationOutput struct {
	CandidateSuggestion *CandidateSuggestion `json:"candidate_suggestion"`
	OutputText          string               `json:"output_text"`
	NextQuestion        *string              `json:"next_question"`
	RelatedRefs         []SourceRef          `json:"related_refs"`
	RelatedIdeas        []string             `json:"related_ideas"`
	Conflicts           []string             `json:"conflicts"`
	PendingQuestions    []string             `json:"pending_questions"`
	GoalRefs            []string             `json:"goal_refs"`
	ExistingAssets      []string             `json:"existing_assets"`
	ExpectedImprovement string               `json:"expected_improvement"`
	MinimumArtifact     string               `json:"minimum_artifact"`
	MissingEvidence     []string             `json:"missing_evidence"`
	Provenance          Provenance           `json:"provenance"`
}
type Distiller interface {
	// ConfigurationID describes actual frozen processor settings, including model
	// and isolation policy. It performs no paid model call or credential export.
	ConfigurationID(context.Context) (string, error)
	Distill(context.Context, DistillationInput) (DistillationOutput, error)
}

// DistillerStatusProvider is an optional factual native-runtime projection.
// It must not invoke model processing or return a guessed model/configuration.
type DistillerStatusProvider interface {
	Status(context.Context) (DistillerStatus, error)
}
type RequestDistillationCommand struct {
	CommandMeta
	ProjectID            string      `json:"project_id,omitempty"`
	CLI                  string      `json:"cli,omitempty"`
	PriorDistillationIDs []string    `json:"prior_distillation_ids,omitempty"`
	InputRefs            []SourceRef `json:"input_refs"`
	Stage                string      `json:"stage"`
	ProcessingConfig     string      `json:"processing_config"`
	Question             string      `json:"question"`
}

// Kept separate so a runtime can expose human recording while its automatic
// processor is blocked by missing native read isolation, with honest 501/jobs.
type AutomaticDistillationService interface {
	GetDistillerStatus(context.Context, Principal) (DistillerStatus, error)
	RequestDistillation(context.Context, Principal, RequestDistillationCommand) (ImportJobResult, error)
}
type DistillerStatus struct {
	SchemaVersion     int      `json:"schema_version"`
	Available         bool     `json:"available"`
	Processor         string   `json:"processor"`
	ConfigurationID   *string  `json:"configuration_id"`
	Model             *string  `json:"model"`
	Reason            string   `json:"reason"`
	RequiredAction    string   `json:"required_action"`
	AllowedSourceKeys []string `json:"allowed_source_keys"`
}
type CandidateSuggestion struct {
	Title           string      `json:"title"`
	EvidenceRefs    []SourceRef `json:"evidence_refs"`
	GoalRefs        []string    `json:"goal_refs"`
	Dimensions      Dimensions  `json:"dimensions"`
	NextStep        string      `json:"next_step"`
	MissingEvidence []string    `json:"missing_evidence"`
	Purpose         string      `json:"purpose"`
}

// A single explicit, versioned user profile controls candidate composite ordering.
// No profile means manual order; partial or unknown dimensions never become zero.
type RankingProfile struct {
	ID        string             `json:"id"`
	Version   int                `json:"version"`
	Enabled   bool               `json:"enabled"`
	Weights   map[string]float64 `json:"weights"`
	CreatedAt time.Time          `json:"created_at"`
}
type RankingProfileDetail struct {
	SchemaVersion int              `json:"schema_version"`
	Configured    bool             `json:"configured"`
	Profile       *RankingProfile  `json:"profile"`
	Versions      []RankingProfile `json:"versions"`
}
type UpdateRankingProfileCommand struct {
	CommandMeta
	Enabled bool               `json:"enabled"`
	Weights map[string]float64 `json:"weights"`
}
type RankingProfileService interface {
	GetRankingProfile(context.Context, Principal) (RankingProfileDetail, error)
	UpdateRankingProfile(context.Context, Principal, UpdateRankingProfileCommand) (RankingProfileDetail, error)
}
type RankingProfileTx interface {
	AttentionTx
	LoadRankingProfile() (RankingProfileDetail, error)
	SaveRankingProfile(RankingProfileDetail, int) error
}

type AttentionService interface {
	ListMaterials(context.Context) (apierrors.ListResult, error)
	ListOpportunities(context.Context) (apierrors.ListResult, error)
	ImportMaterial(context.Context, Principal, ImportMaterialCommand) (ImportJobResult, error)
	GetMaterial(context.Context, Principal, string) (MaterialDetail, error)
	UpdateMaterial(context.Context, Principal, string, UpdateMaterialCommand) (MaterialDetail, error)
	GetContent(context.Context, Principal, string, int) (ContentResult, error)
	GetAttachment(context.Context, Principal, string, int, string) (AttachmentContent, error)
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
	ListMaterialDomains(context.Context, Principal) (apierrors.ListResult, error)
	CreateMaterialDomain(context.Context, Principal, MaterialDomainCommand) (MaterialDomainResult, error)
	ReviseMaterialDomain(context.Context, Principal, string, MaterialDomainCommand) (MaterialDomainResult, error)
	SetMaterialDomains(context.Context, Principal, string, SetMaterialDomainsCommand) (MaterialDetail, error)
	ListProjectSpaces(context.Context, Principal) (apierrors.ListResult, error)
	CreateProjectSpace(context.Context, Principal, ProjectSpaceCommand) (ProjectSpaceResult, error)
	GetProjectSpace(context.Context, Principal, string) (ProjectSpaceResult, error)
	ReferenceMaterial(context.Context, Principal, string, ReferenceMaterialCommand) (ProjectSpaceResult, error)
	RemoveMaterialReference(context.Context, Principal, string, RemoveMaterialReferenceCommand) (ProjectSpaceResult, error)
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
	ListMaterialDomains() ([]MaterialDomain, error)
	LoadMaterialDomain(string) (MaterialDomain, error)
	SaveMaterialDomain(MaterialDomain, int) error
	ListProjectSpaces() ([]ProjectSpace, error)
	LoadProjectSpace(string) (ProjectSpace, error)
	SaveProjectSpace(ProjectSpace, int) error
}

// Tracking stores public metadata only. A binding never imports every item,
// grants credentials, or permits model processing outside a selected project.
type TrackingSource struct {
	ID            string                  `json:"id"`
	Version       int                     `json:"version"`
	Platform      string                  `json:"platform"`
	SourceKind    string                  `json:"source_kind"`
	ExternalID    string                  `json:"external_id"`
	OwnerID       string                  `json:"owner_id"`
	Locator       string                  `json:"locator"`
	Title         string                  `json:"title"`
	Status        string                  `json:"status"`
	LastSuccessAt *time.Time              `json:"last_success_at"`
	LastError     *apierrors.ServiceError `json:"last_error"`
	NextCursor    *string                 `json:"next_cursor"`
	HasMore       bool                    `json:"has_more"`
	Warnings      []string                `json:"warnings"`
}
type ListingMetadata struct {
	ExternalID        string `json:"external_id"`
	Locator           string `json:"locator"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Author            string `json:"author"`
	Cover             string `json:"cover"`
	PublishedAt       int64  `json:"published_at"`
	ProviderStatus    *int   `json:"provider_status"`
	UnavailableReason string `json:"unavailable_reason"`
}
type SourceRecommendation struct {
	Status           string                  `json:"status"`
	Text             string                  `json:"text"`
	Reason           string                  `json:"reason"`
	MetadataRevision int                     `json:"metadata_revision"`
	Provenance       Provenance              `json:"provenance"`
	ConfigurationID  string                  `json:"configuration_id"`
	Error            *apierrors.ServiceError `json:"error"`
}
type SourceItem struct {
	SourceID       string                `json:"source_id"`
	ExternalID     string                `json:"external_id"`
	Revision       int                   `json:"revision"`
	Metadata       ListingMetadata       `json:"metadata"`
	Stale          bool                  `json:"stale"`
	Selected       bool                  `json:"selected"`
	ImportJobID    *string               `json:"import_job_id"`
	MaterialID     *string               `json:"material_id"`
	Recommendation *SourceRecommendation `json:"recommendation"`
}
type TrackingSourceResult struct {
	SchemaVersion int               `json:"schema_version"`
	Source        TrackingSource    `json:"source"`
	Items         []SourceItem      `json:"items"`
	NextCursor    *string           `json:"next_cursor"`
	HasMore       bool              `json:"has_more"`
	Warnings      []string          `json:"warnings"`
	Jobs          []ImportJobResult `json:"jobs"`
}
type BindTrackingSourceCommand struct {
	CommandMeta
	Platform   string `json:"platform"`
	SourceKind string `json:"source_kind"`
	ExternalID string `json:"external_id"`
	OwnerID    string `json:"owner_id"`
	Locator    string `json:"locator"`
	Title      string `json:"title"`
}
type SyncTrackingSourceCommand struct {
	CommandMeta
	Limit  int    `json:"limit,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}
type SourceItemSelection struct {
	ExternalID string `json:"external_id"`
	Revision   int    `json:"revision"`
}
type RecommendSourceItemsCommand struct {
	CommandMeta
	ProjectID string                `json:"project_id"`
	CLI       string                `json:"cli"`
	Items     []SourceItemSelection `json:"items"`
}
type SelectSourceItemsCommand struct {
	CommandMeta
	Items            []SourceItemSelection `json:"items"`
	CollectionReason *string               `json:"collection_reason,omitempty"`
}
type TrackingService interface {
	ListTrackingSources(context.Context, Principal) (apierrors.ListResult, error)
	BindTrackingSource(context.Context, Principal, BindTrackingSourceCommand) (TrackingSourceResult, error)
	GetTrackingSource(context.Context, Principal, string) (TrackingSourceResult, error)
	SyncTrackingSource(context.Context, Principal, string, SyncTrackingSourceCommand) (TrackingSourceResult, error)
	RecommendSourceItems(context.Context, Principal, string, RecommendSourceItemsCommand) (TrackingSourceResult, error)
	SelectSourceItems(context.Context, Principal, string, SelectSourceItemsCommand) (TrackingSourceResult, error)
}

// TrackingTx is optional so pre-tracking repository implementations remain valid.
type TrackingTx interface {
	AttentionTx
	ListTrackingSources() ([]TrackingSource, error)
	LoadTrackingSource(string) (TrackingSource, error)
	SaveTrackingSource(TrackingSource, int) error
	ListSourceItems(string) ([]SourceItem, error)
	SaveSourceItem(SourceItem) error
}
type ListingPage struct {
	Observed   int
	Items      []ListingMetadata
	NextCursor *string
	HasMore    bool
	Warnings   []string
}
type PublicListingReader interface {
	ReadPage(context.Context, TrackingSource, string, int) (ListingPage, error)
}
type ListingRecommendationInput struct {
	Caller      Principal
	ProjectID   string
	CLI         string
	JobID       string
	OperationID string
	Items       []SourceItem
}
type ListingRecommender interface {
	ConfigurationID(context.Context, Principal, string, string) (string, error)
	Recommend(context.Context, ListingRecommendationInput) (map[string]SourceRecommendation, error)
}
type SourceCollection struct {
	ExternalID string `json:"external_id"`
	OwnerID    string `json:"owner_id"`
	Title      string `json:"title"`
	Locator    string `json:"locator"`
	ItemCount  *int   `json:"item_count"`
}
type SourceCollectionService interface {
	ListSourceCollections(context.Context, Principal, string, string) (apierrors.ListResult, error)
}
type PublicCollectionReader interface {
	ListCollections(context.Context, string, string) ([]SourceCollection, error)
}

// Called only after Workspace authorizes a project operation. Membership and
// fixed-version linked access remain checked in Attention on every read.
type AttentionProjectReferences interface {
	ReadProjectReference(context.Context, Principal, string, SourceRef, bool) (SourceSnapshot, error)
	LinkedProjectReferences(context.Context, Principal, string, SourceRef) ([]SourceRef, error)
}
type SelectedTextRequest struct {
	JobID           string
	OperationID     string
	ProjectID       string
	CLI             string
	Prompt          string
	DeadlineSeconds int
	OutputLimit     int
}
type SelectedTextResult struct {
	Text     string
	NativeID string
	Version  string
	Model    *string
}

// The composition bridge rechecks project model consent on config and processing.
// It never changes provider or converts denied scope into a fallback CLI.
type SelectedTextProcessor interface {
	ProjectSpaceID(context.Context, Principal, string, string) (string, error)
	ConfigurationID(context.Context, Principal, string, string) (string, error)
	ProcessSelectedText(context.Context, Principal, SelectedTextRequest) (SelectedTextResult, error)
}
type ProjectDistillerFactory interface {
	Resolve(context.Context, Principal, string, string) (Distiller, error)
}
