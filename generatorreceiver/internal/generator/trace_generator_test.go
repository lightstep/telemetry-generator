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
		P50:  "50ms",
		P90:  "100ms",
		P999: "1000ms",
	}

	p50, p90, p999, err := percentiles.ParseDurations()
	require.NoError(t, err)

	p50expected := time.Duration(50 * time.Millisecond)
	p90expected := time.Duration(100 * time.Millisecond)
	p999expected := time.Duration(1000 * time.Millisecond)

	require.Equal(t, p50expected, p50)
	require.Equal(t, p90expected, p90)
	require.Equal(t, p999expected, p999)

	var samples []float64
	sampleSize := 100
	for i := 0; i < sampleSize; i++ {
		samples = append(samples, float64(calculateLatencyBasedOnPercentiles(percentiles)))
	}

	p50actual, err := stats.Percentile(samples, 50)
	require.NoError(t, err)
	require.Equal(t, p50expected.Microseconds(), int64(p50actual))

	p90actual, err := stats.Percentile(samples, 90)
	require.NoError(t, err)
	require.Equal(t, p90expected.Microseconds(), int64(p90actual))

	p999actual, err := stats.Percentile(samples, 99.9)
	require.NoError(t, err)
	require.Equal(t, p999expected.Microseconds(), int64(p999actual))
}
