package topology

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"math/rand"
	"time"
)

type ServiceRoute struct {
	Route                 string                 `json:"route" yaml:"route"`
	DownstreamCalls       map[string]string      `json:"downstreamCalls,omitempty" yaml:"downstreamCalls,omitempty"`
	MaxLatencyMillis      int64                  `json:"maxLatencyMillis" yaml:"maxLatencyMillis"`
	LatencyPercentiles    *LatencyPercentiles    `json:"latencyPercentiles" yaml:"latencyPercentiles"`
	TagSets               []TagSet               `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	flags.EmbeddedFlags   `json:",inline" yaml:",inline"`
}

type LatencyPercentiles struct {
	P0   string `json:"p0" yaml:"p0"`
	P50  string `json:"p50" yaml:"p50"`
	P95  string `json:"p95" yaml:"p95"`
	P99  string `json:"p99" yaml:"p99"`
	P999 string `json:"p99.9" yaml:"p99.9"`
	P100 string `json:"p100" yaml:"p100"`
}

type parsedLatencyPercentiles struct {
	p0   time.Duration
	p50  time.Duration
	p95  time.Duration
	p99  time.Duration
	p999 time.Duration
	p100 time.Duration
}

func (l *LatencyPercentiles) Sample() float64 {
	uniform := func(timeA, timeB time.Duration) float64 {
		min := float64(timeA.Microseconds())
		max := float64(timeB.Microseconds())
		return (min + (max-min)*rand.Float64()) * 1000
	}
	percentiles, err := l.parseDurations()
	if err != nil {
		return 0
	}
	genNumber := rand.Float64()
	switch {
	case genNumber <= 0.001:
		// 0.1% of requests
		return uniform(percentiles.p99, percentiles.p999)
	case genNumber <= 0.01:
		// 1% of requests
		return uniform(percentiles.p95, percentiles.p99)
	case genNumber <= 0.05:
		// 5% of requests
		return uniform(percentiles.p50, percentiles.p95)
	case genNumber <= 0.5:
		// 50% of requests
		return uniform(percentiles.p0, percentiles.p50)
	default:
		return uniform(percentiles.p0, percentiles.p50)
		// not sure if --> is better, seems to skew it too high generally, return uniform(percentiles.p50, percentiles.p100)
	}
	/*
		TODO: the above is still not perfect - it is a bit off on the p50m the logic for default is prob wrong, should be reorderd like below -
		Trying the below also makes it off on p50 by more (I think because its getting more skewed from the p95), so its maybe not exactly right either
		case genNumber <= 0.5:
			return uniform(percentiles.p0, percentiles.p50)
		case genNumber <= 0.95:
			return uniform(percentiles.p50, percentiles.p95)
		case genNumber <= 0.99:
			return uniform(percentiles.p95, percentiles.p99)
		default:
			return uniform(percentiles.p99, percentiles.p999)
	*/
}

func (l *LatencyPercentiles) parseDurations() (parsedLatencyPercentiles, error) {
	// TODO/future things:
	// 		normalize function for config parsing
	// 		maybe enforce either MaxLatencyMillis or LatencyPercentiles but not both?
	//			either way which overrides which? for now LatencyPercentiles will override MaxLatencyMillis
	p0, err := time.ParseDuration(l.P0)
	if err != nil {
		return parsedLatencyPercentiles{}, err
	}
	p50, err := time.ParseDuration(l.P50)
	if err != nil {
		return parsedLatencyPercentiles{}, err
	}
	p95, err := time.ParseDuration(l.P95)
	if err != nil {
		return parsedLatencyPercentiles{}, err
	}
	p99, err := time.ParseDuration(l.P99)
	if err != nil {
		return parsedLatencyPercentiles{}, err
	}
	p999, err := time.ParseDuration(l.P999)
	if err != nil {
		return parsedLatencyPercentiles{}, err
	}
	p100, err := time.ParseDuration(l.P100)
	if err != nil {
		return parsedLatencyPercentiles{}, err
	}
	return parsedLatencyPercentiles{
		p0:   p0,
		p50:  p50,
		p95:  p95,
		p99:  p99,
		p999: p999,
		p100: p100,
	}, nil
}
