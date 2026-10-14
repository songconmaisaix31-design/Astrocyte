package stats

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func TestMeanFiniteInputs(t *testing.T) {
	for _, test := range []struct {
		name   string
		values []float64
		want   float64
	}{
		{"single", []float64{7}, 7}, {"even", []float64{1, 2, 3, 4}, 2.5},
		{"negative", []float64{-5, -1}, -3}, {"duplicates", []float64{2, 2, 2}, 2},
		{"mixed", []float64{-4, 1, 6}, 1}, {"large", []float64{math.MaxFloat64, math.MaxFloat64}, math.MaxFloat64},
		{"opposite-large", []float64{-math.MaxFloat64, math.MaxFloat64}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			before := slices.Clone(test.values)
			got, err := Mean(test.values)
			if err != nil || got != test.want {
				t.Fatalf("mean=%v error=%v want=%v", got, err, test.want)
			}
			if !slices.Equal(test.values, before) {
				t.Fatal("Mean changed its input")
			}
		})
	}
}

func TestMeanRejectsEmptyAndNonFiniteInputs(t *testing.T) {
	if _, err := Mean(nil); !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("empty error=%v", err)
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, err := Mean([]float64{value}); !errors.Is(err, ErrNonFiniteInput) {
			t.Fatalf("non-finite error=%v", err)
		}
	}
}
