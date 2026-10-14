package domain

import (
	"fmt"
	"math"
	"path"
	"strings"
	"time"
)

type Principal struct {
	ID        string `json:"-"`
	Kind      string `json:"-"`
	ProjectID string `json:"-"`
}

type Meta struct {
	SchemaVersion   int    `json:"schema_version"`
	RequestID       string `json:"request_id"`
	IdempotencyKey  string `json:"-"`
	ExpectedVersion int    `json:"expected_version"`
}

// 金额使用最小货币单位；未知费用使用空指针保存。
type Resources struct {
	Calls    int64  `json:"calls"`
	Tokens   int64  `json:"tokens"`
	Money    *int64 `json:"money"`
	Currency string `json:"currency"`
}

type Budget struct {
	Calls    *int64 `json:"calls"`
	Tokens   *int64 `json:"tokens"`
	Money    *int64 `json:"money"`
	Currency string `json:"currency"`
}

func (b Budget) Validate() error {
	for _, v := range []*int64{b.Calls, b.Tokens, b.Money} {
		if v != nil && *v < 0 {
			return Failure("validation_failed", "resource limits must be nonnegative")
		}
	}
	if b.Money != nil && b.Currency == "" || b.Money == nil && b.Currency != "" {
		return Failure("validation_failed", "money budget requires its currency")
	}
	if b.Calls == nil && b.Tokens == nil && b.Money == nil {
		return Failure("validation_failed", "an explicit resource limit is required")
	}
	return nil
}

func (r Resources) Within(b Budget) bool {
	return (b.Calls == nil || r.Calls <= *b.Calls) && (b.Tokens == nil || r.Tokens <= *b.Tokens) && (b.Money == nil || (r.Money != nil && r.Currency == b.Currency && *r.Money <= *b.Money))
}

type RuleError struct {
	Code    string
	Message string
}

func (e *RuleError) Error() string { return e.Code + ": " + e.Message }

func Failure[C ~string](code C, message string) error {
	return &RuleError{Code: string(code), Message: message}
}

func (r Resources) Validate() error {
	if r.Calls < 0 || r.Tokens < 0 || (r.Money != nil && *r.Money < 0) || (r.Money != nil && r.Currency == "") || (r.Money == nil && r.Currency != "") {
		return Failure("validation_failed", "resource quantities and currency must be explicit and nonnegative")
	}
	return nil
}

func (r Resources) Add(v Resources) (Resources, error) {
	if err := r.Validate(); err != nil {
		return Resources{}, err
	}
	if err := v.Validate(); err != nil {
		return Resources{}, err
	}
	if r.Calls > math.MaxInt64-v.Calls || r.Tokens > math.MaxInt64-v.Tokens {
		return Resources{}, Failure("budget_exhausted", "resource quantity overflow")
	}
	u := Resources{Calls: r.Calls + v.Calls, Tokens: r.Tokens + v.Tokens}
	if r.Money != nil || v.Money != nil {
		if r.Money == nil || v.Money == nil || r.Currency != v.Currency {
			return Resources{}, Failure("validation_failed", "money units must match")
		}
		if *r.Money > math.MaxInt64-*v.Money {
			return Resources{}, Failure("budget_exhausted", "money quantity overflow")
		}
		n := *r.Money + *v.Money
		u.Money = &n
		u.Currency = r.Currency
	}
	return u, nil
}

func (r Resources) Fits(limit Resources) bool {
	return r.Calls <= limit.Calls && r.Tokens <= limit.Tokens && (limit.Money == nil || (r.Money != nil && r.Currency == limit.Currency && *r.Money <= *limit.Money))
}

func (r Resources) Sub(v Resources) (Resources, error) {
	if err := r.Validate(); err != nil {
		return Resources{}, err
	}
	if err := v.Validate(); err != nil {
		return Resources{}, err
	}
	if !v.Fits(r) || (r.Money == nil) != (v.Money == nil) {
		return Resources{}, Failure("validation_failed", "invalid resource release")
	}
	u := Resources{Calls: r.Calls - v.Calls, Tokens: r.Tokens - v.Tokens}
	if r.Money != nil {
		n := *r.Money - *v.Money
		u.Money = &n
		u.Currency = r.Currency
	}
	return u, nil
}

type Limits struct {
	MaxWorkItems      int       `json:"max_work_items"`
	MaxAttempts       int       `json:"max_attempts"`
	MaxNoProgress     int       `json:"max_no_progress"`
	NoProgressSeconds int       `json:"no_progress_seconds"`
	MaxActive         int       `json:"max_active"`
	Deadline          time.Time `json:"deadline"`
}

func (l Limits) Validate() error {
	if l.MaxWorkItems < 1 || l.MaxAttempts < 1 || (l.MaxNoProgress < 1 && l.NoProgressSeconds < 1) || l.MaxActive < 1 || l.Deadline.IsZero() {
		return Failure("validation_failed", "all mission stop limits are required")
	}
	return nil
}

type Grant struct {
	ID               string    `json:"id"`
	Epoch            int       `json:"epoch"`
	ProposalID       string    `json:"proposal_id"`
	ProposalRevision int       `json:"proposal_revision"`
	ProjectID        string    `json:"project_id"`
	Root             string    `json:"root"`
	Goal             string    `json:"goal"`
	Actions          []string  `json:"actions"`
	WritePaths       []string  `json:"write_paths"`
	Budget           Budget    `json:"budget"`
	Limits           Limits    `json:"limits"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type Mission struct {
	ID             string    `json:"id"`
	Version        int       `json:"version"`
	Grant          Grant     `json:"grant"`
	State          string    `json:"state"`
	Reason         string    `json:"reason"`
	Reserved       Resources `json:"reserved"`
	Used           Resources `json:"used"`
	WorkItemCount  int       `json:"work_item_count"`
	NoProgress     int       `json:"no_progress"`
	LastProgressAt time.Time `json:"last_progress_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewMission(id string, g Grant, now time.Time) (Mission, error) {
	if id == "" || g.ID == "" || g.Epoch < 1 || g.ProposalID == "" || g.ProposalRevision < 1 || g.ProjectID == "" || g.Root == "" || g.Goal == "" || len(g.WritePaths) == 0 || !now.Before(g.ExpiresAt) {
		return Mission{}, Failure("approval_required", "a current complete approval is required")
	}
	if err := g.Budget.Validate(); err != nil {
		return Mission{}, err
	}
	if err := g.Limits.Validate(); err != nil {
		return Mission{}, err
	}
	if !now.Before(g.Limits.Deadline) {
		return Mission{}, Failure("budget_exhausted", "mission deadline has passed")
	}
	for _, p := range g.WritePaths {
		if err := ValidateWritePath(p); err != nil {
			return Mission{}, err
		}
	}
	z := Resources{}
	if g.Budget.Money != nil {
		n := int64(0)
		z.Money = &n
		z.Currency = g.Budget.Currency
	}
	return Mission{ID: id, Version: 1, Grant: g, State: "ready", Reserved: z, Used: z, CreatedAt: now, LastProgressAt: now}, nil
}

func (m Mission) CanExecute(now time.Time) error {
	if m.State != "ready" && m.State != "running" {
		return Failure("version_conflict", "mission is "+m.State)
	}
	if m.StopReached(now) {
		return Failure("budget_exhausted", "mission stop condition reached")
	}
	return nil
}

func (m Mission) StopReached(now time.Time) bool {
	l := m.Grant.Limits
	return !now.Before(l.Deadline) || (l.MaxNoProgress > 0 && m.NoProgress >= l.MaxNoProgress) || (l.NoProgressSeconds > 0 && !now.Before(m.LastProgressAt.Add(time.Duration(l.NoProgressSeconds)*time.Second)))
}

func ValidateWritePath(p string) error {
	if p == "" || strings.ContainsAny(p, "\\:*?[]\x00") || strings.HasPrefix(p, "/") || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../") {
		return Failure("validation_failed", fmt.Sprintf("write path %q must be a normalized relative file or directory", p))
	}
	return nil
}

func ContainsPath(parent, child string) bool {
	return parent == "." || parent == child || strings.HasPrefix(child, parent+"/")
}
func Overlaps(a, b []string) bool {
	for _, p := range a {
		for _, q := range b {
			if ContainsPath(p, q) || ContainsPath(q, p) {
				return true
			}
		}
	}
	return false
}
func Within(paths, scope []string) bool {
	for _, p := range paths {
		found := false
		for _, s := range scope {
			if ContainsPath(s, p) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type SourceVersion struct {
	ID       string `json:"id"`
	Revision string `json:"revision"`
}

type WorkItem struct {
	RequiredCapabilities []string        `json:"required_capabilities"`
	ID                   string          `json:"id"`
	MissionID            string          `json:"mission_id"`
	Version              int             `json:"version"`
	State                string          `json:"state"`
	Title                string          `json:"title"`
	Inputs               []SourceVersion `json:"inputs"`
	Dependencies         []string        `json:"dependencies"`
	WritePaths           []string        `json:"write_paths"`
	Attempts             int             `json:"attempts"`
	Fence                int64           `json:"fence"`
	ArtifactID           string          `json:"artifact_id"`
	Reason               string          `json:"reason"`
}

func (w WorkItem) Validate(m Mission) error {
	if w.ID == "" || w.MissionID != m.ID || w.Title == "" || len(w.Inputs) == 0 || len(w.WritePaths) == 0 {
		return Failure("validation_failed", "work item requires mission, input versions and write paths")
	}
	for _, v := range w.Inputs {
		if v.ID == "" || v.Revision == "" {
			return Failure("validation_failed", "input versions must be fixed")
		}
	}
	for _, p := range w.WritePaths {
		if err := ValidateWritePath(p); err != nil {
			return err
		}
	}
	if !Within(w.WritePaths, m.Grant.WritePaths) {
		return Failure("scope_denied", "work item exceeds approved write paths")
	}
	seen := map[string]bool{}
	for _, d := range w.Dependencies {
		if d == "" || d == w.ID || seen[d] {
			return Failure("validation_failed", "dependencies must be distinct earlier work items")
		}
		seen[d] = true
	}
	return nil
}

type Lease struct {
	ContextPacketID string    `json:"context_packet_id"`
	BindingID       string    `json:"binding_id"`
	PolicyVersion   string    `json:"policy_version"`
	WorkItemID      string    `json:"work_item_id"`
	MissionID       string    `json:"mission_id"`
	ProjectID       string    `json:"project_id"`
	HolderID        string    `json:"holder_id"`
	Fence           int64     `json:"fence"`
	OperationID     string    `json:"operation_id"`
	SessionID       string    `json:"session_id"`
	MemberVersion   string    `json:"member_version"`
	Worktree        string    `json:"worktree"`
	BaseCommit      string    `json:"base_commit"`
	WritePaths      []string  `json:"write_paths"`
	Reserved        Resources `json:"reserved"`
	ExpiresAt       time.Time `json:"expires_at"`
}

func (l Lease) Check(holder string, fence int64, now time.Time) error {
	if l.HolderID != holder || l.Fence != fence || !now.Before(l.ExpiresAt) {
		return Failure("lease_lost", "lease holder, fence or expiry no longer matches")
	}
	return nil
}

type Artifact struct {
	ExperienceStatus string          `json:"experience_status"`
	ContextPacketID  string          `json:"context_packet_id"`
	ProjectChecksRef string          `json:"project_checks_ref"`
	BaseCommit       string          `json:"base_commit"`
	ReviewReason     string          `json:"review_reason"`
	ReviewedBy       string          `json:"reviewed_by"`
	ReviewedAt       *time.Time      `json:"reviewed_at"`
	ID               string          `json:"id"`
	Version          int             `json:"version"`
	WorkItemID       string          `json:"work_item_id"`
	OperationID      string          `json:"operation_id"`
	SessionID        string          `json:"session_id"`
	Fence            int64           `json:"fence"`
	Inputs           []SourceVersion `json:"inputs"`
	Commit           string          `json:"commit"`
	ObjectRef        string          `json:"object_ref"`
	ChecksRef        string          `json:"checks_ref"`
	ChangedPaths     []string        `json:"changed_paths"`
	State            string          `json:"state"`
	AcceptanceID     string          `json:"acceptance_id"`
	AcceptedBy       string          `json:"accepted_by"`
	ExperienceID     string          `json:"experience_id"`
	CreatedAt        time.Time       `json:"created_at"`
}
