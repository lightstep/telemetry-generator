package topology

import (
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
)

type ServiceRoute struct {
	Route                 string                 `json:"route" yaml:"route"`
	DownstreamCalls       map[string]string      `json:"downstreamCalls,omitempty" yaml:"downstreamCalls,omitempty"`
	MaxLatencyMillis      int64                  `json:"maxLatencyMillis" yaml:"maxLatencyMillis"`
	LatencyPercentiles    *LatencyPercentiles    `json:"latencyPercentiles" yaml:"latencyPercentiles"`
	TagSets               []TagSet               `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	flags.EmbeddedFlags   `json:",inline" yaml:",inline"`
	// TODO: rename all references from `tag` to `attribute`, to follow the otel standard.
}

func (r *ServiceRoute) validate(t Topology) error {
	err := r.Validate()
	if err != nil {
		return err
	}

	for service, route := range r.DownstreamCalls {
		st := t.GetServiceTier(service)
		if st == nil {
			return fmt.Errorf("downstream service %s does not exist", service)
		}
		if st.GetRoute(route) == nil {
			return fmt.Errorf("downstream service %s does not have route %s defined", service, route)
		}
	}

	if r.MaxLatencyMillis <= 0 {
		return fmt.Errorf("must have a positive, non-zero maxLatencyMillis defined")
	}
	return nil
}

func (r *ServiceRoute) load(route string) error {
	r.Route = route
	if r.LatencyPercentiles != nil {
		err := r.LatencyPercentiles.loadDurations()
		if err != nil {
			return fmt.Errorf("error parsing latencyPercentiles: %v", err)
		}
	}
	return nil
}
