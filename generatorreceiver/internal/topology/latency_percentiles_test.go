package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLatencyPercentiles_LoadDurations(t *testing.T) {
	percentiles := &LatencyPercentiles{
		P0Cfg:   "100ms",
		P50Cfg:  "200ms",
		P95Cfg:  "1000ms",
		P99Cfg:  "10000ms",
		P999Cfg: "11000ms",
		P100Cfg: "12000ms",
	}

	err := percentiles.loadDurations()
	require.NoError(t, err)

	p0expected := time.Duration(100 * time.Millisecond)
	p50expected := time.Duration(200 * time.Millisecond)
	p95expected := time.Duration(1000 * time.Millisecond)
	p99expected := time.Duration(10000 * time.Millisecond)
	p999expected := time.Duration(11000 * time.Millisecond)
	p100expected := time.Duration(12000 * time.Millisecond)

	require.Equal(t, p0expected, percentiles.durations.p0)
	require.Equal(t, p50expected, percentiles.durations.p50)
	require.Equal(t, p95expected, percentiles.durations.p95)
	require.Equal(t, p99expected, percentiles.durations.p99)
	require.Equal(t, p999expected, percentiles.durations.p999)
	require.Equal(t, p100expected, percentiles.durations.p100)
}
