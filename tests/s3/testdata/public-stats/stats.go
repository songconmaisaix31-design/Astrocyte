package stats

import (
	"errors"
	"math"
	"math/big"
)

var ErrEmptyInput = errors.New("input must contain at least one value")
var ErrNonFiniteInput = errors.New("input must contain only finite values")

// Mean 使用精确有理数累加，避免有限输入的中间求和溢出。
func Mean(values []float64) (float64, error) {
	if len(values) == 0 {
		return 0, ErrEmptyInput
	}
	sum := new(big.Rat)
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return 0, ErrNonFiniteInput
		}
		sum.Add(sum, new(big.Rat).SetFloat64(value))
	}
	sum.Quo(sum, new(big.Rat).SetInt64(int64(len(values))))
	result, _ := sum.Float64()
	return result, nil
}
