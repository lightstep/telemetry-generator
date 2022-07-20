package topology

import "time"

type LatencyPercentiles struct {
	P50  string `json:"p50" yaml:"p50"`
	P95  string `json:"p95" yaml:"p95"`
	P99  string `json:"p99" yaml:"p99"`
	P999 string `json:"p99.9" yaml:"p99.9"`
}

func (l *LatencyPercentiles) ParseDurations() (time.Duration, time.Duration, time.Duration, time.Duration, error) {
	// TODO/future things:
	// 		normalize function for config parsing
	// 		maybe enforce either MaxLatencyMillis or LatencyPercentiles but not both?
	//			either way which overrides which? for now LatencyPercentiles will override MaxLatencyMillis
	p50, err := time.ParseDuration(l.P50)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	p95, err := time.ParseDuration(l.P95)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	p99, err := time.ParseDuration(l.P99)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	p999, err := time.ParseDuration(l.P999)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return p50, p95, p99, p999, nil
}

type ServiceRoute struct {
	Route                 string                 `json:"route" yaml:"route"`
	DownstreamCalls       map[string]string      `json:"downstreamCalls,omitempty" yaml:"downstreamCalls,omitempty"`
	MaxLatencyMillis      int64                  `json:"maxLatencyMillis" yaml:"maxLatencyMillis"`
	LatencyPercentiles    *LatencyPercentiles    `json:"latencyPercentiles" yaml:"latencyPercentiles"`
	TagSets               []TagSet               `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	FlagSet               string                 `json:"flag_set" yaml:"flag_set"`
	FlagUnset             string                 `json:"flag_unset" yaml:"flag_unset"`
}
