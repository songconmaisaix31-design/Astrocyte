package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// ServiceOptions configures bounded local workers and attention decay. Clock
// is injectable for deterministic rules; it must be safe for concurrent use.
type ServiceOptions struct {
	WorkerConcurrency           int
	MaxAttempts                 int
	JobTimeout                  time.Duration
	AttentionHalfLife           time.Duration
	AttentionWeights            map[string]float64
	Distiller                   Distiller
	ProjectDistillers           ProjectDistillerFactory
	ListingReader               PublicListingReader
	CollectionReader            PublicCollectionReader
	ListingRecommender          ListingRecommender
	ScholarSearcher             PaperSearcher
	AllowedProcessingSourceKeys []string
	Clock                       func() time.Time
}

type Service struct {
	repo               Repository
	sources            SourceReader
	objects            ObjectStore
	options            ServiceOptions
	wake               chan struct{}
	runMu              sync.Mutex
	running            bool
	activeMu           sync.Mutex
	activeJobs         map[string]*activeWork
	configurationError error
}

type activeWork struct{ cancel context.CancelFunc }

var _ AttentionService = (*Service)(nil)

func NewAttentionService(repo Repository, sources SourceReader, objects ObjectStore, options ServiceOptions) *Service {
	if options.WorkerConcurrency <= 0 {
		options.WorkerConcurrency = 2
	}
	if options.MaxAttempts <= 0 {
		options.MaxAttempts = 3
	}
	if options.JobTimeout <= 0 {
		options.JobTimeout = 5 * time.Minute
	}
	if options.AttentionHalfLife <= 0 {
		options.AttentionHalfLife = 7 * 24 * time.Hour
	}
	if options.Clock == nil {
		options.Clock = time.Now
	}
	weights := map[string]float64{"reread": 1, "annotate": 2, "adopt": 3, "classify": 1, "mention": 1, "project_reuse": 3}
	for name, weight := range options.AttentionWeights {
		weights[name] = weight
	}
	options.AttentionWeights = weights
	options.AllowedProcessingSourceKeys = slices.Clone(options.AllowedProcessingSourceKeys)
	return &Service{repo: repo, sources: sources, objects: objects, options: options, wake: make(chan struct{}, 1), activeJobs: make(map[string]*activeWork), configurationError: ValidateAttentionWeights(weights)}
}

// ValidateAttentionWeights lets composition fail invalid configuration before
// startup; invalid weights must never produce a plausible zero attention score.
func ValidateAttentionWeights(weights map[string]float64) error {
	for name, weight := range weights {
		if !slices.Contains([]string{"reread", "annotate", "adopt", "classify", "mention", "project_reuse"}, name) || math.IsNaN(weight) || math.IsInf(weight, 0) || weight < 0 {
			return serviceError(apierrors.ValidationFailed, "Human activity weights must be finite nonnegative supported event weights", "correct_attention_configuration")
		}
	}
	return nil
}

func serviceError(code apierrors.Code, message, action string) *apierrors.ServiceError {
	return &apierrors.ServiceError{Code: code, Message: message, RequiredAction: action}
}

func mapError(err error, requestID string) error {
	if err == nil {
		return nil
	}
	var e *apierrors.ServiceError
	if errors.As(err, &e) {
		copy := *e
		copy.RequestID = requestID
		return &copy
	}
	switch {
	case errors.Is(err, domain.ErrEvidence):
		e = serviceError(apierrors.EvidenceMissing, "Required evidence is missing", "provide_evidence")
	case errors.Is(err, domain.ErrVersion):
		e = serviceError(apierrors.VersionConflict, "The version changed", "reload_and_retry")
	case errors.Is(err, domain.ErrUnknown):
		e = serviceError(apierrors.DeliveryUnknown, "The original operation has an unknown outcome", "reconcile_original_operation")
	case errors.Is(err, domain.ErrInvalid):
		e = serviceError(apierrors.ValidationFailed, "Invalid command input", "correct_input")
	default:
		e = serviceError(apierrors.InternalError, "Attention operation could not be completed", "check_service_and_retry")
	}
	e.RequestID = requestID
	return e
}

func isMissing(err error) bool {
	var e *apierrors.ServiceError
	return errors.As(err, &e) && e.Code == apierrors.NotFound
}

func isRestricted(err error) bool {
	var e *apierrors.ServiceError
	return errors.As(err, &e) && e.Code == apierrors.ScopeDenied
}

func authorize(p Principal, write bool) error {
	if p.ID == "" || (p.Kind != "human" && p.Kind != "agent") || (write && p.Kind != "human") {
		return serviceError(apierrors.ScopeDenied, "This operation requires an authenticated human session", "use_human_session")
	}
	if p.Kind == "agent" {
		// A bearer credential identifies an Agent; it does not grant access to
		// any material. Scope/selection rules await the user's explicit policy.
		return serviceError(apierrors.ScopeDenied, "No material access has been granted to this Agent", "ask_user_to_select_materials")
	}
	return nil
}

func validateMeta(m CommandMeta) error {
	if m.SchemaVersion != 1 || strings.TrimSpace(m.RequestID) == "" || strings.TrimSpace(m.IdempotencyKey) == "" || m.ExpectedVersion < 1 {
		return serviceError(apierrors.ValidationFailed, "Command metadata is incomplete", "provide_schema_request_id_idempotency_key_and_expected_version")
	}
	return nil
}

func digestBytes(bytes []byte) string { h := sha256.Sum256(bytes); return hex.EncodeToString(h[:]) }

func semanticDigest(input any) (string, error) {
	bytes, err := json.Marshal(input)
	if err != nil {
		return "", err
	}
	var object map[string]json.RawMessage
	if err = json.Unmarshal(bytes, &object); err != nil {
		return "", err
	}
	delete(object, "request_id")
	bytes, err = json.Marshal(object)
	if err != nil {
		return "", err
	}
	return digestBytes(bytes), nil
}

// command commits state, the complete replay result, and events in the same
// transaction. Any callback error rolls them all back. Request IDs are tracing
// metadata and do not turn an identical semantic retry into a different input.
func command[T any](s *Service, ctx context.Context, p Principal, m CommandMeta, name string, input any, mutate func(AttentionTx) (T, error)) (T, error) {
	var result T
	if err := authorize(p, true); err != nil {
		return result, mapError(err, m.RequestID)
	}
	if s.configurationError != nil {
		return result, mapError(s.configurationError, m.RequestID)
	}
	if err := validateMeta(m); err != nil {
		return result, mapError(err, m.RequestID)
	}
	digest, err := semanticDigest(input)
	if err != nil {
		return result, mapError(err, m.RequestID)
	}
	err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
		receipt, err := tx.LoadReceipt(p.Kind+":"+p.ID, name, m.IdempotencyKey)
		if err == nil {
			if receipt.Digest != digest {
				return serviceError(apierrors.VersionConflict, "Idempotency key was used with different input", "use_new_idempotency_key")
			}
			return json.Unmarshal(receipt.Result, &result)
		}
		if !isMissing(err) {
			return err
		}
		result, err = mutate(tx)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			return err
		}
		return tx.SaveReceipt(p.Kind+":"+p.ID, name, m.IdempotencyKey, Receipt{Digest: digest, Result: bytes})
	})
	return result, mapError(err, m.RequestID)
}

func (s *Service) event(tx AttentionTx, kind, id string, version int, m CommandMeta, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return tx.AppendEvent(OutboxEvent{ID: rand.Text(), Type: "attention." + kind, AggregateID: id, AggregateVersion: version, OccurredAt: s.options.Clock(), CorrelationID: m.RequestID, CausationID: m.IdempotencyKey, Payload: bytes})
}

func listResult[T any](values []T) apierrors.ListResult {
	result := apierrors.EmptyList()
	for _, value := range values {
		result.Items = append(result.Items, value)
	}
	return result
}

func (s *Service) ListMaterials(ctx context.Context) (apierrors.ListResult, error) {
	if s.configurationError != nil {
		return apierrors.EmptyList(), mapError(s.configurationError, "")
	}
	var materials []Material
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		rows, err := tx.ListMaterials()
		if err != nil {
			return err
		}
		byID := make(map[string]Material, len(rows))
		ranks := make([]domain.MaterialRank, 0, len(rows))
		now := s.options.Clock()
		for _, row := range rows {
			row = s.projectAttentionAt(row, now)
			byID[row.Material.ID] = row.Material
			ranks = append(ranks, domain.MaterialRank{ID: row.Material.ID, Pinned: row.Material.Pinned, Activity: row.Material.AttentionScore, LongTermValue: row.Material.LongTermValue, CreatedAt: row.Material.CreatedAt})
		}
		for _, rank := range domain.RankMaterials(ranks) {
			row := byID[rank.ID]
			materials = append(materials, row)
		}
		return nil
	})
	return listResult(materials), mapError(err, "")
}

func (s *Service) ListOpportunities(ctx context.Context) (apierrors.ListResult, error) {
	var opportunities []Opportunity
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		rows, err := tx.ListOpportunities()
		if err != nil {
			return err
		}
		profile, err := rankingProfile(tx)
		if err != nil {
			return err
		}
		for _, row := range rows {
			// The retained S0 list boundary has no principal. Keep derived
			// withdrawn content out of this shared list; human detail queries
			// still retain restricted history.
			if err := validateOpportunityAccess(tx, row.Opportunity); err != nil {
				if isRestricted(err) {
					continue
				}
				return err
			}
			opportunities = append(opportunities, row.Opportunity)
		}
		opportunities = rankOpportunities(opportunities, profile)
		return nil
	})
	return listResult(opportunities), mapError(err, "")
}

func (s *Service) GetMaterial(ctx context.Context, p Principal, id string) (MaterialDetail, error) {
	var result MaterialDetail
	if err := authorize(p, false); err != nil {
		return result, err
	}
	if s.configurationError != nil {
		return result, mapError(s.configurationError, "")
	}
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		var err error
		result, err = tx.LoadMaterial(id)
		if err != nil {
			return err
		}
		result, err = materialProjection(tx, result)
		if err != nil {
			return err
		}
		if p.Kind == "agent" {
			if result.Material.Lifecycle == "withdrawn" {
				return serviceError(apierrors.ScopeDenied, "Material was withdrawn", "choose_active_material")
			}
			if err = s.machineRead(tx, &result, p); err != nil {
				return err
			}
		}
		return nil
	})
	return s.projectAttention(result), mapError(err, "")
}

func materialProjection(tx AttentionTx, row MaterialDetail) (MaterialDetail, error) {
	all, err := tx.ListDistillations()
	if err != nil {
		return row, err
	}
	row.Distillations = []Distillation{}
	for _, d := range all {
		for _, ref := range append(append([]SourceRef{}, d.InputRefs...), d.RelatedRefs...) {
			if ref.MaterialID == row.Material.ID {
				row.Distillations = append(row.Distillations, d)
				break
			}
		}
	}
	return row, nil
}

func (s *Service) GetContent(ctx context.Context, p Principal, id string, revision int) (ContentResult, error) {
	var result ContentResult
	if err := authorize(p, false); err != nil {
		return result, err
	}
	var row MaterialRevision
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		material, err := tx.LoadMaterial(id)
		if err != nil {
			return err
		}
		if material.Material.Lifecycle == "withdrawn" {
			return serviceError(apierrors.ScopeDenied, "Material was withdrawn", "choose_active_material")
		}
		if revision == 0 {
			revision = material.Material.CurrentRevision
		}
		for _, r := range material.Revisions {
			if r.Revision == revision {
				row = r
				if p.Kind == "agent" {
					return s.machineRead(tx, &material, p)
				}
				return nil
			}
		}
		return apierrors.NewNotFound("material_revision", id)
	})
	if err != nil {
		return result, mapError(err, "")
	}
	bytes, err := s.objects.Read(ctx, row.ObjectRef)
	if err != nil {
		return result, mapError(err, "")
	}
	if err = s.checkMaterialAccess(ctx, id); err != nil {
		return result, mapError(err, "")
	}
	return ContentResult{SchemaVersion: 1, MaterialID: id, Revision: revision, ContentDigest: row.ContentDigest, Text: string(bytes), Provenance: row.Provenance}, nil
}

func (s *Service) checkMaterialAccess(ctx context.Context, id string) error {
	return s.repo.WithTx(ctx, func(tx AttentionTx) error {
		row, err := tx.LoadMaterial(id)
		if err != nil {
			return err
		}
		if row.Material.Lifecycle == "withdrawn" {
			return serviceError(apierrors.ScopeDenied, "Material was withdrawn", "choose_active_material")
		}
		return nil
	})
}

func (s *Service) projectAttention(row MaterialDetail) MaterialDetail {
	return s.projectAttentionAt(row, s.options.Clock())
}

func (s *Service) projectAttentionAt(row MaterialDetail, now time.Time) MaterialDetail {
	row.Revisions = nonNil(row.Revisions)
	row.Distillations = nonNil(row.Distillations)
	row.Uses = nonNil(row.Uses)
	row.Material.SourceSpans = nonNil(row.Material.SourceSpans)
	events := make([]domain.AttentionEvent, 0, len(row.Uses))
	for _, use := range row.Uses {
		weight := s.options.AttentionWeights[use.Action]
		events = append(events, domain.AttentionEvent{Actor: use.ActorKind, Kind: use.Action, At: use.OccurredAt, Weight: weight})
	}
	row.Material.AttentionScore, _ = domain.HumanActivity(events, now, s.options.AttentionHalfLife)
	row.Material.DomainIDs = nonNil(row.Material.DomainIDs)
	row.Material.RankingStrategy = domain.MaterialRankingStrategy
	row.Material.RankingReason = fmt.Sprintf("固定=%t；人类关注=%.4f；长期价值=%.4f；固定优先，关注降序，长期价值及时间作同分排序", row.Material.Pinned, row.Material.AttentionScore, row.Material.LongTermValue)
	row.Material.AttentionHalfLifeSeconds = s.options.AttentionHalfLife.Seconds()
	row.Material.AttentionWeights = maps.Clone(s.options.AttentionWeights)
	return row
}

func domainRefs(refs []SourceRef) []domain.InputRef {
	result := make([]domain.InputRef, 0, len(refs))
	for _, ref := range refs {
		result = append(result, domain.InputRef{Kind: "material", ID: ref.MaterialID, Revision: ref.Revision})
	}
	return result
}

func dimensions(d Dimensions) map[string]domain.DimensionScore {
	return map[string]domain.DimensionScore{"goal_progress": {Value: d.GoalProgress.Value, Reason: d.GoalProgress.Reason}, "current_interest": {Value: d.CurrentInterest.Value, Reason: d.CurrentInterest.Reason}, "project_improvement": {Value: d.ProjectImprovement.Value, Reason: d.ProjectImprovement.Reason}, "originality": {Value: d.Originality.Value, Reason: d.Originality.Reason}}
}

func normalizeRefs(refs []SourceRef) []SourceRef {
	refs = slices.Clone(refs)
	slices.SortFunc(refs, func(a, b SourceRef) int {
		aBytes, _ := json.Marshal(a)
		bBytes, _ := json.Marshal(b)
		return strings.Compare(string(aBytes), string(bBytes))
	})
	return slices.CompactFunc(refs, func(a, b SourceRef) bool {
		return a.MaterialID == b.MaterialID && a.Revision == b.Revision && a.Locator == b.Locator && ((a.Span == nil && b.Span == nil) || (a.Span != nil && b.Span != nil && *a.Span == *b.Span))
	})
}

func validateSourceRefs(tx AttentionTx, refs []SourceRef) error {
	for _, ref := range refs {
		if ref.MaterialID == "" || ref.Revision < 1 || ref.Locator == "" {
			return domain.ErrInvalid
		}
		row, err := tx.LoadMaterial(ref.MaterialID)
		if err != nil {
			return err
		}
		if row.Material.Lifecycle == "withdrawn" {
			return serviceError(apierrors.ScopeDenied, "Input material was withdrawn", "choose_active_material")
		}
		found := false
		for _, revision := range row.Revisions {
			if revision.Revision != ref.Revision {
				continue
			}
			found = true
			if revision.SourceLocator != ref.Locator {
				return domain.ErrEvidence
			}
			if ref.Span != nil && !slices.Contains(revision.SourceSpans, *ref.Span) {
				return domain.ErrEvidence
			}
		}
		if !found {
			return apierrors.NewNotFound("material_revision", ref.MaterialID)
		}
	}
	return nil
}
