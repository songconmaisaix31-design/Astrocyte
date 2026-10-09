// Package contracts embeds the fixed processor output schema for adapters.
package contracts

import _ "embed"

// DistillationOutputSchema is the business output only. The model cannot supply
// processor identity, permissions, approval or provenance fields.
//
//go:embed distillation-output.schema.json
var DistillationOutputSchema []byte
