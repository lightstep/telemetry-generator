package topology

import (
	"math/rand"
	"time"
)

type LatencyPercentiles struct {
	P0Cfg     string `json:"p0" yaml:"p0"`
	P50Cfg    string `json:"p50" yaml:"p50"`
	P95Cfg    string `json:"p95" yaml:"p95"`
	P99Cfg    string `json:"p99" yaml:"p99"`
	P999Cfg   string `json:"p99.9" yaml:"p99.9"`
	P100Cfg   string `json:"p100" yaml:"p100"`
	durations struct {
		p0   time.Duration
		p50  time.Duration
		p95  time.Duration
		p99  time.Duration
		p999 time.Duration
		p100 time.Duration
	}
}

func (l *LatencyPercentiles) Sample() float64 {
	uniform := func(timeA, timeB time.Duration) float64 {
		min := float64(timeA.Microseconds())
		max := float64(timeB.Microseconds())
		return (min + (max-min)*rand.Float64()) * 1000
	}
	genNumber := rand.Float64()
	switch {
	case genNumber <= 0.001:
		// 0.1% of requests
		return uniform(l.durations.p99, l.durations.p999)
	case genNumber <= 0.01:
		// 1% of requests
		return uniform(l.durations.p95, l.durations.p99)
	case genNumber <= 0.05:
		// 5% of requests
		return uniform(l.durations.p50, l.durations.p95)
	case genNumber <= 0.5:
		// 50% of requests
		return uniform(l.durations.p0, l.durations.p50)
	default:
		return uniform(l.durations.p0, l.durations.p50)
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

func (l *LatencyPercentiles) loadDurations() error {
	// TODO/future things:
	// 		normalize function for config parsing
	// 		maybe enforce either MaxLatencyMillis or LatencyPercentiles but not both?
	//			either way which overrides which? for now LatencyPercentiles will override MaxLatencyMillis
	var err error
	l.durations.p0, err = time.ParseDuration(l.P0Cfg)
	if err != nil {
		return err
	}
	l.durations.p50, err = time.ParseDuration(l.P50Cfg)
	if err != nil {
		return err
	}
	l.durations.p95, err = time.ParseDuration(l.P95Cfg)
	if err != nil {
		return err
	}
	l.durations.p99, err = time.ParseDuration(l.P99Cfg)
	if err != nil {
		return err
	}
	l.durations.p999, err = time.ParseDuration(l.P999Cfg)
	if err != nil {
		return err
	}
	l.durations.p100, err = time.ParseDuration(l.P100Cfg)
	if err != nil {
		return err
	}
	return nil
}
