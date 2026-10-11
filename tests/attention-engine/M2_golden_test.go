package attention_engine_test

import (
	_ "embed"
	"encoding/json"
	"math"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// M2_golden_stream.json is the 30-day fake recomputation stream: deterministic
// input (Unix-ms timestamps, per-day weighted signal sum and confidence delta)
// plus per-step A/M/confidence output. The output was produced by an
// independent Python reference computation, not by the Go code under test, so
// the root can re-verify it with a separate JS formula.
//
//go:embed M2_golden_stream.json
var goldenStream []byte

type goldenDoc struct {
	Lambda  float64 `json:"lambda"`
	Alpha   float64 `json:"alpha"`
	Floor   float64 `json:"floor"`
	Ceiling float64 `json:"ceiling"`
	DayMs   int64   `json:"day_ms"`
	Initial struct {
		Attention  float64 `json:"attention"`
		Momentum   float64 `json:"momentum"`
		Confidence float64 `json:"confidence"`
	} `json:"initial"`
	Input []struct {
		TsMs            int64   `json:"ts_ms"`
		SignalSum       float64 `json:"signal_sum"`
		ConfidenceDelta float64 `json:"confidence_delta"`
	} `json:"input"`
	Output []struct {
		TsMs       int64   `json:"ts_ms"`
		Attention  float64 `json:"attention"`
		Momentum   float64 `json:"momentum"`
		Confidence float64 `json:"confidence"`
	} `json:"output"`
}

// TestM2GoldenThirtyDayStream replays the deterministic 30-day stream through
// the production domain functions and compares A(t)/M(t)/confidence per point
// against the independent golden output. The scenario exercises decay, signal
// accumulation, the attention ceiling (days 3, 5, 12, 15), the anti-Matthew
// floor (day 20 onward), momentum smoothing, and confidence (success +0.1 on
// day 12, failure -0.1 on day 18, which never enters attention).
func TestM2GoldenThirtyDayStream(t *testing.T) {
	var doc goldenDoc
	if err := json.Unmarshal(goldenStream, &doc); err != nil {
		t.Fatalf("decode golden stream: %v", err)
	}
	if doc.Alpha != domain.MomentumAlpha || doc.Floor != domain.AttentionFloor || doc.Ceiling != domain.AttentionCeiling {
		t.Fatalf("golden constants drifted from domain: alpha=%v floor=%v ceiling=%v",
			doc.Alpha, doc.Floor, doc.Ceiling)
	}
	if doc.DayMs != 86400000 {
		t.Fatalf("golden day_ms = %d, want 86400000", doc.DayMs)
	}
	if len(doc.Input) != 30 || len(doc.Output) != 30 {
		t.Fatalf("golden stream must be 30 days, got input=%d output=%d", len(doc.Input), len(doc.Output))
	}

	attention := doc.Initial.Attention
	momentum := doc.Initial.Momentum
	confidence := doc.Initial.Confidence
	dayMs := float64(doc.DayMs)

	for i, in := range doc.Input {
		dt := 1.0
		if i > 0 {
			dt = float64(in.TsMs-doc.Input[i-1].TsMs) / dayMs
		}
		newA := domain.UpdateAttention(attention, doc.Lambda, dt, in.SignalSum)
		newM := domain.UpdateMomentum(momentum, attention, newA)
		newC := domain.UpdateConfidence(confidence, in.ConfidenceDelta)
		want := doc.Output[i]
		if in.TsMs != want.TsMs {
			t.Fatalf("day %d: output ts_ms mismatch", i)
		}
		if math.Abs(newA-want.Attention) > 1e-12 {
			t.Fatalf("day %d: A = %.16f, want %.16f", i, newA, want.Attention)
		}
		if math.Abs(newM-want.Momentum) > 1e-12 {
			t.Fatalf("day %d: M = %.16f, want %.16f", i, newM, want.Momentum)
		}
		if math.Abs(newC-want.Confidence) > 1e-12 {
			t.Fatalf("day %d: C = %.16f, want %.16f", i, newC, want.Confidence)
		}
		attention, momentum, confidence = newA, newM, newC
	}
}
