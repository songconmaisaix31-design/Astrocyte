package judgment

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func invalid(message string) error {
	return &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: message}
}

func conflict(message string) error {
	return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: message}
}

func ValidateRequest(r JudgmentRequest) error {
	if r.SchemaVersion != "1.0" || r.ProjectID == "" || r.WorktreeID == "" || r.SnapshotID == "" {
		return invalid("judgment requires schema 1.0 and explicit project/worktree/snapshot")
	}
	q := r.Question
	if q.ID == "" || strings.TrimSpace(q.Instructions) == "" {
		return invalid("question id and instructions are required")
	}
	if r.Language == "" || r.Versions.Goal == "" || r.Versions.Taste == "" || r.Versions.Policy == "" || r.Versions.Model == "" || r.Versions.AccessEpoch == "" {
		return invalid("language and every judgment version must be explicit")
	}
	if r.Backend != LayaBackend && r.Backend != JevBackend && r.Backend != RuleBackend {
		return invalid("select exactly one supported backend")
	}
	if r.Budget.TimeoutMS <= 0 || r.Budget.MaxInputBytes <= 0 || r.Budget.MaxInputTokens <= 0 || r.Budget.MaxOutputTokens < 0 {
		return invalid("explicit nonzero input/time budgets are required")
	}
	switch q.Type {
	case Choice:
		if len(q.Options) < 2 || len(q.Options) > 255 || len(q.Levels) != 0 {
			return invalid("choice requires 2 to 255 options")
		}
		seen := map[string]bool{}
		for _, o := range q.Options {
			if o.ID == "" || seen[o.ID] {
				return invalid("option ids must be nonempty and unique")
			}
			seen[o.ID] = true
			for _, ref := range o.RequiredEvidence {
				if err := validRef(ref); err != nil {
					return err
				}
			}
		}
	case Score:
		if len(q.Levels) < 2 || len(q.Levels) > 10 || len(q.Options) != 0 {
			return invalid("score requires 2 to 10 ordered rubric levels")
		}
		for _, level := range q.Levels {
			if strings.TrimSpace(level) == "" {
				return invalid("score levels must be nonempty")
			}
		}
	case Truth:
		if len(q.Options) != 0 || len(q.Levels) != 0 {
			return invalid("truth does not accept choices or score levels")
		}
	default:
		return invalid("unsupported question type")
	}
	seen := map[string]bool{}
	for _, ref := range r.EvidenceRefs {
		if err := validRef(ref); err != nil {
			return err
		}
		key := fmt.Sprintf("%s:%d", ref.MaterialID, ref.Revision)
		if seen[key] {
			return invalid("duplicate evidence reference")
		}
		seen[key] = true
	}
	switch q.Rule {
	case "":
	case "all_evidence_present":
		if q.Type != Truth {
			return invalid("all_evidence_present requires truth")
		}
	case "evidence_coverage":
		if q.Type != Score {
			return invalid("evidence_coverage requires score")
		}
	case "single_supported_option":
		if q.Type != Choice {
			return invalid("single_supported_option requires choice")
		}
		for _, option := range q.Options {
			if len(option.RequiredEvidence) == 0 {
				return invalid("rule options require explicit evidence")
			}
		}
	default:
		return invalid("unsupported deterministic rule")
	}
	return nil
}

func validRef(ref EvidenceRef) error {
	decoded, err := hex.DecodeString(ref.Digest)
	if ref.MaterialID == "" || ref.Revision <= 0 || err != nil || len(decoded) != sha256.Size || strings.ToLower(ref.Digest) != ref.Digest {
		return invalid("evidence requires material id, fixed revision and lowercase sha256")
	}
	return nil
}

func DigestText(text string) string {
	value := sha256.Sum256([]byte(text))
	return hex.EncodeToString(value[:])
}

func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return DigestText(string(data)), nil
}

func ValidateBackendInputBytes(input BackendInput) error {
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	if input.Budget.MaxInputBytes <= 0 || len(data) > input.Budget.MaxInputBytes {
		return invalid("complete question, criteria and evidence input exceeds explicit byte budget")
	}
	return nil
}

func ValidatePolicy(p Policy) error {
	if p.ProjectID == "" || len(p.Languages) == 0 || p.Revision < 1 || p.ProjectSettingsRevision < 1 || p.WorktreeID == "" || p.SnapshotID == "" {
		return invalid("policy requires project and allowed languages")
	}
	if p.Snapshot.ProjectID != p.ProjectID || p.Snapshot.WorktreeID != p.WorktreeID || p.Snapshot.Generation != p.SnapshotID || p.Snapshot.EnvironmentID == "" || p.Snapshot.RepoID == "" || p.Snapshot.Root == "" || p.Snapshot.Head == "" || p.Snapshot.IndexDigest == "" || p.Snapshot.WorktreeDigest == "" || p.Snapshot.ObservedAt.IsZero() || len(p.SnapshotFiles) == 0 {
		return invalid("policy requires the actual persisted authorized project snapshot")
	}
	if p.Backend != RuleBackend && p.Backend != LayaBackend && p.Backend != JevBackend {
		return invalid("policy requires a single backend")
	}
	if p.QuestionType != Choice && p.QuestionType != Score && p.QuestionType != Truth {
		return invalid("policy question type is unsupported")
	}
	if p.Versions.Goal == "" || p.Versions.Taste == "" || p.Versions.Policy == "" || p.Versions.Model == "" || p.Versions.AccessEpoch == "" {
		return invalid("policy versions must be explicit")
	}
	if p.Budget.TimeoutMS <= 0 || p.Budget.MaxInputBytes <= 0 || p.Budget.MaxInputTokens <= 0 || p.Budget.MaxOutputTokens < 0 {
		return invalid("policy budgets must be explicit")
	}
	seen := map[string]bool{}
	for _, language := range p.Languages {
		if language == "" || seen[language] {
			return invalid("policy languages must be nonempty and unique")
		}
		seen[language] = true
	}
	for _, threshold := range []*float64{p.MinConfidence, p.TruthAcceptMin, p.TruthAcceptMax} {
		if threshold != nil && !unit(*threshold) {
			return invalid("policy threshold is outside [0,1]")
		}
	}
	if (p.TruthAcceptMin == nil) != (p.TruthAcceptMax == nil) {
		return invalid("truth acceptance needs both explicit bounds")
	}
	if p.TruthAcceptMin != nil && *p.TruthAcceptMin > *p.TruthAcceptMax {
		return invalid("truth acceptance bounds are reversed")
	}
	if p.Backend == JevBackend && p.ExternalScope == "" {
		return invalid("Jev policy requires explicit external data scope")
	}
	for _, ref := range p.ExternalEvidence {
		if err := validRef(ref); err != nil {
			return err
		}
	}
	return nil
}

func unit(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
