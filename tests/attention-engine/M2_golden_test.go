package attentionengine_test

import (
	"math"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// TestM2GoldenThirtyDayStream replays a deterministic 30-day event stream and
// compares A(t)/M(t) per point against hand-recomputed golden values produced
// by an independent reference computation, not by the code under test. The
// scenario exercises decay, signal accumulation, the attention ceiling (days
// 3, 5, 12, 15), the anti-Matthew floor (day 20 onward), and momentum smoothing.
func TestM2GoldenThirtyDayStream(t *testing.T) {
	const lambda = 0.1
	signalSum := map[int]float64{
		0: 0.6, 3: 0.6, 5: 1.0, 9: 0.3, 10: 0.3, 12: 0.8, 15: 1.5, 20: -0.8,
	}
	expectedA := []float64{
		0.6, 0.5429024508215757, 0.49123845184678905, 1.0, 0.9048374180359595,
		1.0, 0.9048374180359595, 0.8187307530779817, 0.7408182206817177,
		0.9703200460356391, 1.0, 0.9048374180359595, 1.0, 0.9048374180359595,
		0.8187307530779817, 1.0, 0.9048374180359595, 0.8187307530779817,
		0.7408182206817177, 0.6703200460356391, 0.05, 0.05, 0.05, 0.05, 0.05,
		0.05, 0.05, 0.05, 0.05, 0.05,
	}
	expectedM := []float64{
		0.42, 0.08603171557510303, -0.010355284609819767, 0.3530264983243017,
		0.03929414212246218, 0.07840205001156698, -0.04309319237135824,
		-0.07320262318199193, -0.0764995596319824, 0.13770140985815033,
		0.06208639073249769, -0.047987890155079024, 0.05221744032830462,
		-0.05094857527633695, -0.07555923805348555, 0.10422070142936714,
		-0.03534759694601819, -0.07087894455438992, -0.07580245604370181,
		-0.07208945906536551, -0.455850869944557, -0.1367552609833671,
		-0.04102657829501013, -0.012307973488503038, -0.003692392046550911,
		-0.0011077176139652734, -0.000332315284189582, -9.969458525687459e-05,
		-2.9908375577062376e-05, -8.972512673118713e-06,
	}

	attention, momentum := 0.0, 0.0
	for day := 0; day < 30; day++ {
		newA := domain.UpdateAttention(attention, lambda, 1.0, signalSum[day])
		newM := domain.UpdateMomentum(momentum, attention, newA)
		if math.Abs(newA-expectedA[day]) > 1e-12 {
			t.Fatalf("day %d: A = %.16f, want %.16f", day, newA, expectedA[day])
		}
		if math.Abs(newM-expectedM[day]) > 1e-12 {
			t.Fatalf("day %d: M = %.16f, want %.16f", day, newM, expectedM[day])
		}
		attention, momentum = newA, newM
	}
}
