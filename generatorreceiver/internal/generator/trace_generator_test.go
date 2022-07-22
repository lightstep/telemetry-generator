package generator

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/stretchr/testify/require"

	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
)

func TestLatencyPercentiles(t *testing.T) {
	percentiles := &topology.LatencyPercentiles{
		P0:   "100ms",
		P50:  "200ms",
		P95:  "1000ms",
		P99:  "10000ms",
		P999: "11000ms",
		P100: "12000ms",
	}

	p0, p50, p95, p99, p999, p100, err := percentiles.ParseDurations()
	require.NoError(t, err)

	p0expected := time.Duration(100 * time.Millisecond)
	p50expected := time.Duration(200 * time.Millisecond)
	p95expected := time.Duration(1000 * time.Millisecond)
	p99expected := time.Duration(10000 * time.Millisecond)
	p999expected := time.Duration(11000 * time.Millisecond)
	p100expected := time.Duration(12000 * time.Millisecond)

	require.Equal(t, p0expected, p0)
	require.Equal(t, p50expected, p50)
	require.Equal(t, p95expected, p95)
	require.Equal(t, p99expected, p99)
	require.Equal(t, p999expected, p999)
	require.Equal(t, p100expected, p100)

	rand.Seed(time.Now().UnixNano())
	var samples []float64
	sampleSize := 10
	for i := 0; i < sampleSize; i++ {
		samples = append(samples, calculateLatencyBasedOnPercentiles(percentiles))
	}

	p50actual, err := stats.Percentile(samples, 50)
	require.NoError(t, err)

	requireWithinRange := func(a time.Duration, b float64, rangee float64) {
		require.LessOrEqual(t, math.Abs(float64(a.Microseconds())-b), rangee, fmt.Sprintf("expected: %v, actual: %f, %f", a.Microseconds(), b, samples))
	}
	variance := rand.Float64() * float64(p50.Microseconds())
	requireWithinRange(p50expected, p50actual, variance)

	p95actual, err := stats.Percentile(samples, 95)
	require.NoError(t, err)
	requireWithinRange(p95expected, p95actual, variance)

	p99actual, err := stats.Percentile(samples, 99)
	require.NoError(t, err)
	requireWithinRange(p99expected, p99actual, variance)

	p999actual, err := stats.Percentile(samples, 99.9)
	require.NoError(t, err)
	requireWithinRange(p999expected, p999actual, variance)
}
