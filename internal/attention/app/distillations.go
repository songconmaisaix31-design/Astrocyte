package app

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func replay[T any](s *Service, ctx context.Context, p Principal, m CommandMeta, name string, input any) (T, bool, error) {
	var result T
	if err := authorize(p, true); err != nil {
		return result, false, mapError(err, m.RequestID)
	}
	if err := validateMeta(m); err != nil {
		return result, false, mapError(err, m.RequestID)
	}
	digest, err := semanticDigest(input)
	if err != nil {
		return result, false, mapError(err, m.RequestID)
	}
	found := false
	err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
		receipt, err := tx.LoadReceipt(p.Kind+":"+p.ID, name, m.IdempotencyKey)
		if isMissing(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if receipt.Digest != digest {
			return serviceError(apierrors.VersionConflict, "Idempotency key was used with different input", "use_new_idempotency_key")
		}
		found = true
		return json.Unmarshal(receipt.Result, &result)
	})
	return result, found, mapError(err, m.RequestID)
}

func (s *Service) ListDistillations(ctx context.Context, p Principal) (apierrors.ListResult, error) {
	if err := authorize(p, false); err != nil {
		return apierrors.EmptyList(), err
	}
	var rows []Distillation
	err := s.repo.WithTx(ctx, func(tx AttentionTx) error {
		all, err := tx.ListDistillations()
		if err != nil {
			return err
		}
		for _, d := range all {
			if p.Kind == "agent" {
				if err := validateSourceRefs(tx, d.InputRefs); err != nil {
					if isRestricted(err) {
						continue
					}
					return err
				}
				if err := validateSourceRefs(tx, d.RelatedRefs); err != nil {
					if isRestricted(err) {
						continue
					}
					return err
				}
			}
			rows = append(rows, d)
		}
		return nil
	})
	return listResult(rows), mapError(err, "")
}

func (s *Service) RecordDistillation(ctx context.Context, p Principal, c RecordDistillationCommand) (DistillationResult, error) {
	c.InputRefs = normalizeRefs(c.InputRefs)
	c.RelatedRefs = normalizeRefs(c.RelatedRefs)
	result, found, err := replay[DistillationResult](s, ctx, p, c.CommandMeta, "RecordDistillation", c)
	if err != nil || found {
		return result, err
	}
	if c.ExpectedVersion != 1 || strings.TrimSpace(c.ProcessingConfig) == "" || strings.TrimSpace(c.OutputText) == "" {
		return result, mapError(domain.ErrInvalid, c.RequestID)
	}
	out := domain.DistillationOutput{SourcePreserved: len(c.InputRefs) > 0, Summary: c.OutputText, Related: domainRefs(c.RelatedRefs), ExistingView: strings.Join(c.RelatedIdeas, "\n"), Conflict: strings.Join(c.Conflicts, "\n"), OpenQuestions: c.PendingQuestions, Goal: strings.Join(c.GoalRefs, "\n"), ExistingAsset: strings.Join(c.ExistingAssets, "\n"), ExpectedImprovement: c.ExpectedImprovement, MinimumOutcome: c.MinimumArtifact, MissingEvidence: c.MissingEvidence}
	if err = domain.ValidateDistillation(c.Stage, domainRefs(c.InputRefs), out); err != nil {
		return result, mapError(err, c.RequestID)
	}
	identity := struct {
		Inputs                  []SourceRef
		Related                 []SourceRef
		Stage, Config, Question string
	}{c.InputRefs, c.RelatedRefs, c.Stage, c.ProcessingConfig, strings.TrimSpace(c.Question)}
	bytes, _ := json.Marshal(identity)
	reuseKey := digestBytes(bytes)
	// Check reuse and source access before publishing the output. The final
	// transaction rechecks both to close races with another writer or withdrawal.
	err = s.repo.WithTx(ctx, func(tx AttentionTx) error {
		if err := validateSourceRefs(tx, append(append([]SourceRef{}, c.InputRefs...), c.RelatedRefs...)); err != nil {
			return err
		}
		old, err := tx.FindDistillationByReuseKey(reuseKey)
		if err == nil {
			result = DistillationResult{SchemaVersion: 1, Distillation: old, Reused: true}
			return nil
		}
		if isMissing(err) {
			return nil
		}
		return err
	})
	if err != nil {
		return result, mapError(err, c.RequestID)
	}
	var outputRef string
	if !result.Reused {
		outputRef, err = s.objects.Publish(ctx, []byte(c.OutputText))
		if err != nil {
			return DistillationResult{}, mapError(err, c.RequestID)
		}
	}
	return command(s, ctx, p, c.CommandMeta, "RecordDistillation", c, func(tx AttentionTx) (DistillationResult, error) {
		if err := validateSourceRefs(tx, append(append([]SourceRef{}, c.InputRefs...), c.RelatedRefs...)); err != nil {
			return DistillationResult{}, err
		}
		old, err := tx.FindDistillationByReuseKey(reuseKey)
		if err == nil {
			return DistillationResult{SchemaVersion: 1, Distillation: old, Reused: true}, nil
		}
		if !isMissing(err) {
			return DistillationResult{}, err
		}
		d := Distillation{ID: rand.Text(), InputRefs: nonNil(c.InputRefs), Stage: c.Stage, OutputRef: outputRef, OutputText: c.OutputText, Status: "succeeded", NextQuestion: c.NextQuestion, Question: strings.TrimSpace(c.Question), ProcessingConfig: c.ProcessingConfig, RelatedRefs: nonNil(c.RelatedRefs), RelatedIdeas: nonNil(c.RelatedIdeas), Conflicts: nonNil(c.Conflicts), PendingQuestions: nonNil(c.PendingQuestions), GoalRefs: nonNil(c.GoalRefs), ExistingAssets: nonNil(c.ExistingAssets), ExpectedImprovement: c.ExpectedImprovement, MinimumArtifact: c.MinimumArtifact, MissingEvidence: nonNil(c.MissingEvidence), Provenance: Provenance{Processor: "human", Version: "1", Mode: "manual", Source: p.ID}, ReuseKey: reuseKey, CreatedAt: s.options.Clock()}
		if err := tx.SaveDistillation(d); err != nil {
			return DistillationResult{}, err
		}
		if err := s.event(tx, "distillation_recorded", d.ID, 1, c.CommandMeta, map[string]any{"distillation_id": d.ID, "stage": d.Stage, "input_refs": d.InputRefs}); err != nil {
			return DistillationResult{}, err
		}
		return DistillationResult{SchemaVersion: 1, Distillation: d}, nil
	})
}

func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
