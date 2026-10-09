package domain

import (
	"errors"
	"math"
	"slices"
	"strings"
	"time"
)

var (
	ErrInvalid  = errors.New("validation_failed")
	ErrEvidence = errors.New("evidence_missing")
	ErrVersion  = errors.New("version_conflict")
	ErrUnknown  = errors.New("delivery_unknown")
)

// NextMaterialRevision reuses an existing immutable content revision, including
// when a source returns to an older content digest. A new digest appends history.
func NextMaterialRevision(digests []string, digest string) (revision int, reused bool, err error) {
	if strings.TrimSpace(digest) == "" {
		return 0, false, ErrInvalid
	}
	for i, old := range digests {
		if old == digest {
			return i + 1, true, nil
		}
	}
	return len(digests) + 1, false, nil
}

type InputRef struct {
	Kind     string `json:"kind"`
	ID       string `json:"id"`
	Revision int    `json:"revision"`
}

func ValidateRefs(refs []InputRef) error {
	if len(refs) == 0 {
		return ErrEvidence
	}
	for _, ref := range refs {
		if ref.ID == "" || ref.Revision < 1 || (ref.Kind != "material" && ref.Kind != "distillation" && ref.Kind != "opportunity") {
			return ErrInvalid
		}
	}
	return nil
}

// WorkIdentity is a value describing actual distillation input, configuration,
// and an additional question. Ref order is irrelevant; versions are significant.
type WorkIdentity struct {
	Inputs        []InputRef
	Stage         string
	Configuration string
	Question      string
}

func (w WorkIdentity) Canonical() WorkIdentity {
	w.Inputs = slices.Clone(w.Inputs)
	slices.SortFunc(w.Inputs, func(a, b InputRef) int {
		if n := strings.Compare(a.Kind, b.Kind); n != 0 {
			return n
		}
		if n := strings.Compare(a.ID, b.ID); n != 0 {
			return n
		}
		return a.Revision - b.Revision
	})
	w.Inputs = slices.Compact(w.Inputs)
	w.Question = strings.TrimSpace(w.Question)
	return w
}

func (w WorkIdentity) Same(other WorkIdentity) bool {
	a, b := w.Canonical(), other.Canonical()
	return a.Stage == b.Stage && a.Configuration == b.Configuration && a.Question == b.Question && slices.Equal(a.Inputs, b.Inputs)
}

type DistillationOutput struct {
	SourcePreserved     bool
	Summary             string
	Related             []InputRef
	ExistingView        string
	Conflict            string
	OpenQuestions       []string
	Goal                string
	ExistingAsset       string
	ExpectedImprovement string
	MinimumOutcome      string
	MissingEvidence     []string
}

// ValidateDistillation never invents citations, collection reasons or model
// output. A theme may finish with explicit unanswered questions; a project
// association records missing evidence even when that list is empty.
func ValidateDistillation(stage string, inputs []InputRef, out DistillationOutput) error {
	if err := ValidateRefs(inputs); err != nil {
		return err
	}
	switch stage {
	case "content":
		if !out.SourcePreserved || strings.TrimSpace(out.Summary) == "" {
			return ErrEvidence
		}
	case "topic":
		if len(out.Related) == 0 && strings.TrimSpace(out.ExistingView) == "" && strings.TrimSpace(out.Conflict) == "" && !hasText(out.OpenQuestions) {
			return ErrEvidence
		}
		if len(out.Related) > 0 {
			return ValidateRefs(out.Related)
		}
	case "project":
		if strings.TrimSpace(out.Goal) == "" || strings.TrimSpace(out.ExistingAsset) == "" || strings.TrimSpace(out.ExpectedImprovement) == "" || strings.TrimSpace(out.MinimumOutcome) == "" || out.MissingEvidence == nil {
			return ErrEvidence
		}
	default:
		return ErrInvalid
	}
	return nil
}

func hasText(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return true
		}
	}
	return false
}

var DimensionNames = []string{"goal_progress", "current_interest", "project_improvement", "originality"}

type DimensionScore struct {
	Value  *float64
	Reason string
}

func ValidateDimensions(dimensions map[string]DimensionScore) error {
	for name, score := range dimensions {
		if !slices.Contains(DimensionNames, name) {
			return ErrInvalid
		}
		if score.Value != nil && (math.IsNaN(*score.Value) || math.IsInf(*score.Value, 0)) {
			return ErrInvalid
		}
	}
	return nil
}

// CompleteWeights allows automatic ranking only with an explicit, complete
// profile. All dimensions contribute and goal progress has primary weight.
func CompleteWeights(weights map[string]float64) bool {
	if len(weights) != len(DimensionNames) {
		return false
	}
	for _, name := range DimensionNames {
		w := weights[name]
		if math.IsNaN(w) || math.IsInf(w, 0) || w <= 0 {
			return false
		}
		if name != "goal_progress" && weights["goal_progress"] < w {
			return false
		}
	}
	return true
}

// CompositeScore returns nil for unknown dimensions or an incomplete profile,
// preserving manual order instead of silently replacing unknown by zero.
func CompositeScore(dimensions map[string]DimensionScore, weights map[string]float64) *float64 {
	if !CompleteWeights(weights) || ValidateDimensions(dimensions) != nil {
		return nil
	}
	var sum, total float64
	for _, name := range DimensionNames {
		score, ok := dimensions[name]
		if !ok || score.Value == nil {
			return nil
		}
		sum += *score.Value * weights[name]
		total += weights[name]
	}
	result := sum / total
	return &result
}

func OpportunityState(evidence []InputRef, purpose, nextStep string) (string, error) {
	if len(evidence) > 0 {
		if err := ValidateRefs(evidence); err != nil {
			return "", err
		}
	}
	if len(evidence) == 0 || strings.TrimSpace(purpose) == "" || strings.TrimSpace(nextStep) == "" {
		return "incubating", nil
	}
	return "ready_for_review", nil
}

// FeedbackTransition contains no admission/approval path. adopt is an outcome
// signal, not permission to register a project or start a mission.
func FeedbackTransition(state, kind string, reasons map[string]string) (newState string, negativeExample bool, err error) {
	if !slices.Contains([]string{"incubating", "ready_for_review", "deferred", "rejected", "withdrawn", "admitted"}, state) {
		return "", false, ErrInvalid
	}
	switch kind {
	case "later":
		return "deferred", false, nil
	case "reject":
		if len(reasons) == 0 {
			return "", false, ErrEvidence
		}
		for dimension, reason := range reasons {
			if !slices.Contains(DimensionNames, dimension) || strings.TrimSpace(reason) == "" {
				return "", false, ErrInvalid
			}
		}
		return "rejected", true, nil
	case "revise":
		return "incubating", false, nil
	case "adopt", "already_solved":
		return state, false, nil
	default:
		return "", false, ErrInvalid
	}
}

type AttentionEvent struct {
	Actor  string
	Kind   string
	At     time.Time
	Weight float64
}

// HumanActivity reads only explicit human signals. Read queries and machine
// events cannot implicitly become a human signal; pinned/value are separate.
func HumanActivity(events []AttentionEvent, now time.Time, halfLife time.Duration) (float64, error) {
	if halfLife <= 0 {
		return 0, ErrInvalid
	}
	var activity float64
	for _, event := range events {
		if event.Actor != "human" || !slices.Contains([]string{"annotate", "reread", "adopt", "classify", "mention", "project_reuse"}, event.Kind) {
			continue
		}
		if event.At.IsZero() || event.At.After(now) || math.IsNaN(event.Weight) || math.IsInf(event.Weight, 0) || event.Weight < 0 {
			return 0, ErrInvalid
		}
		activity += event.Weight * math.Exp2(-float64(now.Sub(event.At))/float64(halfLife))
	}
	return activity, nil
}
