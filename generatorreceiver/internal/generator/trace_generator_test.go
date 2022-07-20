package generator

import (
	"testing"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/stretchr/testify/require"

	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
)

func TestLatencyPercentiles(t *testing.T) {
	percentiles := &topology.LatencyPercentiles{
		P50:  "10ms",
		P95:  "100ms",
		P99:  "1000ms",
		P999: "10000ms",
	}

	p50, p95, p99, p999, err := percentiles.ParseDurations()
	require.NoError(t, err)

	p50expected := time.Duration(10 * time.Millisecond)
	p95expected := time.Duration(100 * time.Millisecond)
	p99expected := time.Duration(1000 * time.Millisecond)
	p999expected := time.Duration(10000 * time.Millisecond)

	require.Equal(t, p50expected, p50)
	require.Equal(t, p95expected, p95)
	require.Equal(t, p99expected, p99)
	require.Equal(t, p999expected, p999)

	var samples []float64
	sampleSize := 1000
	for i := 0; i < sampleSize; i++ {
		samples = append(samples, float64(calculateLatencyBasedOnPercentiles(percentiles)))
	}

	p50actual, err := stats.Percentile(samples, 50)
	require.NoError(t, err)
	require.Equal(t, p50expected.Microseconds(), int64(p50actual))

	p95actual, err := stats.Percentile(samples, 95)
	require.NoError(t, err)
	require.Equal(t, p95expected.Microseconds(), int64(p95actual))

	p99actual, err := stats.Percentile(samples, 99)
	require.NoError(t, err)
	require.Equal(t, p99expected.Microseconds(), int64(p99actual))

	p999actual, err := stats.Percentile(samples, 99.9)
	require.NoError(t, err)
	require.Equal(t, p999expected.Microseconds(), int64(p999actual))
}
