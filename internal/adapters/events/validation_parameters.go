package events

import (
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Explicit validation bindings; these do not change product defaults or math.
const ValidationSignalStrength = 1.0
const ValidationSeedConfidence = .5
const ValidationLambdaRPerSecond = .05 / 86400
const ValidationReferenceIncrement = .1

func NormalizeValidationSignal(row app.StoredEvent) (*domain.Signal, error) {
	if row.Event == nil {
		return nil, nil
	}
	return domain.NormalizeSignal(row.OriginKey, *row.Event, row.Observation.Outcome, row.Observation.Depth, ValidationSignalStrength)
}

func ValidationKnowledgeQuery(topic, text string, now int64, basis app.KnowledgeAgeBasis) app.KnowledgeQuery {
	lambda := ValidationLambdaRPerSecond
	return app.KnowledgeQuery{Topic: topic, Text: text, Now: now, LambdaR: &lambda, AgeBasis: basis, ReferenceBoost: ValidationReferenceIncrement}
}

// Only the two decided seed fields are bound here; other existing dynamics
// options are explicitly supplied by the caller and are left unchanged.
func ValidationDynamicsOptions(options app.DynamicsOptions) app.DynamicsOptions {
	decay, confidence := .1, ValidationSeedConfidence
	options.Decay, options.InitialConfidence = &decay, &confidence
	return options
}
