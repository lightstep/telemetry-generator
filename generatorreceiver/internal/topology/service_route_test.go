package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLatencyPercentilesParsing(t *testing.T) {
	percentiles := &LatencyPercentiles{
		P0:   "100ms",
		P50:  "200ms",
		P95:  "1000ms",
		P99:  "10000ms",
		P999: "11000ms",
		P100: "12000ms",
	}

	parsedDurations, err := percentiles.parseDurations()
	require.NoError(t, err)

	p0expected := time.Duration(100 * time.Millisecond)
	p50expected := time.Duration(200 * time.Millisecond)
	p95expected := time.Duration(1000 * time.Millisecond)
	p99expected := time.Duration(10000 * time.Millisecond)
	p999expected := time.Duration(11000 * time.Millisecond)
	p100expected := time.Duration(12000 * time.Millisecond)

	require.Equal(t, p0expected, parsedDurations.p0)
	require.Equal(t, p50expected, parsedDurations.p50)
	require.Equal(t, p95expected, parsedDurations.p95)
	require.Equal(t, p99expected, parsedDurations.p99)
	require.Equal(t, p999expected, parsedDurations.p999)
	require.Equal(t, p100expected, parsedDurations.p100)
}
