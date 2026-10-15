package judgment

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type QuestionType string
type Decision string
type Backend string

const (
	Choice       QuestionType = "choice"
	Score        QuestionType = "score"
	Truth        QuestionType = "truth"
	Recommend    Decision     = "recommend"
	Abstain      Decision     = "abstain"
	NeedEvidence Decision     = "need_evidence"
	RuleBackend  Backend      = "rule"
	JevBackend   Backend      = "jev"
	LayaBackend  Backend      = "laya"
)

// Actor 只能由已经鉴权的调用入口提供。
type Actor struct {
	ID        string
	Kind      string
	ProjectID string
}

type EvidenceRef struct {
	MaterialID string `json:"material_id"`
	Revision   int    `json:"revision"`
	Digest     string `json:"digest"`
}

type Evidence struct {
	Reference EvidenceRef
	Text      string
}

type Option struct {
	ID               string        `json:"id"`
	Description      string        `json:"description"`
	RequiredEvidence []EvidenceRef `json:"required_evidence"`
}

type Question struct {
	ID           string       `json:"question_id"`
	Type         QuestionType `json:"question_type"`
	Instructions string       `json:"instructions"`
	Options      []Option     `json:"options"`
	Levels       []string     `json:"levels"`
	Rule         string       `json:"rule"`
}

type Versions struct {
	Goal        string `json:"goal_version"`
	Taste       string `json:"taste_version"`
	Policy      string `json:"policy_version"`
	Model       string `json:"model_revision"`
	AccessEpoch string `json:"access_epoch"`
}

type Budget struct {
	TimeoutMS       int64 `json:"timeout_ms"`
	MaxInputBytes   int   `json:"max_input_bytes"`
	MaxInputTokens  int   `json:"max_input_tokens"`
	MaxOutputTokens int   `json:"max_output_tokens"`
}

type JudgmentRequest struct {
	SchemaVersion string           `json:"schema_version"`
	ProjectID     string           `json:"project_id"`
	WorktreeID    string           `json:"worktree_id"`
	SnapshotID    string           `json:"snapshot_id"`
	Question      Question         `json:"question"`
	EvidenceRefs  []EvidenceRef    `json:"evidence_refs"`
	Language      string           `json:"language"`
	Backend       Backend          `json:"backend"`
	ExternalScope string           `json:"external_scope"`
	Versions      Versions         `json:"versions"`
	Budget        Budget           `json:"budget"`
	Context       *DecisionContext `json:"context,omitempty"`
}

type OptionCoverage struct {
	Total           int `json:"total"`
	Distinct        int `json:"distinct"`
	TokensPerOption int `json:"tokens_per_option"`
}

type Usage struct {
	InputTokens        int                       `json:"input_tokens"`
	OutputTokens       int                       `json:"output_tokens"`
	StateTokens        *int                      `json:"state_tokens"`
	StateTokensDropped *int                      `json:"state_tokens_dropped"`
	Truncated          *bool                     `json:"truncated"`
	TruncatedQuestions []string                  `json:"truncated_questions"`
	Options            map[string]OptionCoverage `json:"options"`
}

type JudgmentResult struct {
	SchemaVersion      string                    `json:"schema_version"`
	ID                 string                    `json:"decision_id"`
	Version            int                       `json:"result_version"`
	QuestionID         string                    `json:"question_id"`
	QuestionType       QuestionType              `json:"question_type"`
	Selected           *string                   `json:"selected"`
	Value              *float64                  `json:"value"`
	Distribution       map[string]float64        `json:"distribution"`
	ProviderConfidence *float64                  `json:"provider_confidence"`
	Backend            Backend                   `json:"provider"`
	ProviderModel      *string                   `json:"provider_model"`
	ModelRevision      *string                   `json:"model_revision"`
	InputDigest        string                    `json:"input_digest"`
	Truncated          bool                      `json:"truncated"`
	Options            map[string]OptionCoverage `json:"options"`
	Usage              *Usage                    `json:"usage"`
	Decision           Decision                  `json:"decision"`
	ReasonCode         string                    `json:"reason_code"`
	Versions           Versions                  `json:"versions"`
	EvidenceRefs       []EvidenceRef             `json:"evidence_refs"`
	CreatedAt          time.Time                 `json:"created_at"`
	CacheHit           bool                      `json:"cache_hit"`
	Context            *DecisionContext          `json:"context,omitempty"`
}

type Policy struct {
	Revision                int                           `json:"revision"`
	ProjectID               string                        `json:"project_id"`
	ProjectSettingsRevision int                           `json:"project_settings_revision"`
	WorktreeID              string                        `json:"worktree_id"`
	SnapshotID              string                        `json:"snapshot_id"`
	SnapshotFiles           []string                      `json:"snapshot_files"`
	Snapshot                domain.ProjectContextSnapshot `json:"snapshot"`
	QuestionType            QuestionType                  `json:"question_type"`
	Versions                Versions                      `json:"versions"`
	Backend                 Backend                       `json:"backend"`
	Languages               []string                      `json:"languages"`
	ExternalScope           string                        `json:"external_scope"`
	ExternalEvidence        []EvidenceRef                 `json:"external_evidence"`
	Budget                  Budget                        `json:"budget"`
	MinConfidence           *float64                      `json:"min_confidence"`
	TruthAcceptMin          *float64                      `json:"truth_accept_min"`
	TruthAcceptMax          *float64                      `json:"truth_accept_max"`
}

type AuthorizedInput struct {
	Policy        Policy
	AccessVersion string
	Evidence      []Evidence
	WorktreeID    string
	SnapshotID    string
}

// 权限、固定引用和策略从真实业务来源读取，传输内容不能自行批准。
type Authorizer interface {
	Resolve(context.Context, Actor, JudgmentRequest) (AuthorizedInput, error)
}

type RecordAuthorizer interface {
	AuthorizeRecord(context.Context, Actor, JudgmentRequest) error
}

type PolicyRepository interface {
	WithJudgmentTx(context.Context, func(context.Context) error) error
	LoadJudgmentPolicy(context.Context, string, QuestionType) (Policy, error)
	SaveJudgmentPolicy(context.Context, Policy, int) error
}

type BackendInput struct {
	Question      Question `json:"question"`
	State         string   `json:"state"`
	Language      string   `json:"language"`
	ModelRevision string   `json:"model_revision"`
	Budget        Budget   `json:"budget"`
}

type BackendResult struct {
	Selected      *string
	Value         *float64
	Distribution  map[string]float64
	Confidence    *float64
	ProviderModel string
	ModelRevision string
	Usage         Usage
	Truncated     bool
	Refused       bool
	Options       map[string]OptionCoverage
}

type Model interface {
	Backend() Backend
	Evaluate(context.Context, BackendInput) (BackendResult, error)
}

type Record struct {
	Actor     Actor
	Request   JudgmentRequest
	Result    JudgmentResult
	CacheKey  string
	Cacheable bool
}

type Feedback struct {
	ID            string    `json:"id"`
	ResultID      string    `json:"result_id"`
	ResultVersion int       `json:"result_version"`
	ActorID       string    `json:"actor_id"`
	ActorKind     string    `json:"actor_kind"`
	Kind          string    `json:"kind"`
	Reason        string    `json:"reason"`
	CreatedAt     time.Time `json:"created_at"`
}

type Repository interface {
	WithJudgmentTx(context.Context, func(context.Context) error) error
	FindCached(context.Context, string) (Record, bool, error)
	SaveResult(context.Context, Record) error
	LoadResult(context.Context, string) (Record, error)
	SaveFeedback(context.Context, Feedback) error
	ListFeedback(context.Context, string) ([]Feedback, error)
}
