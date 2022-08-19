package topology

import (
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"math/rand"
)

type ServiceRoute struct {
	Route                 string                 `json:"route" yaml:"route"`
	DownstreamCalls       map[string]string      `json:"downstreamCalls,omitempty" yaml:"downstreamCalls,omitempty"`
	MaxLatencyMillis      int64                  `json:"maxLatencyMillis" yaml:"maxLatencyMillis"`
	LatencyConfigs        LatencyConfigs         `json:"latencyConfigs" yaml:"latencyConfigs"`
	TagSets               []TagSet               `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	flags.EmbeddedFlags   `json:",inline" yaml:",inline"`
	// TODO: rename all references from `tag` to `attribute`, to follow the otel standard.
}

func (r *ServiceRoute) validate(t Topology) error {
	err := r.ValidateFlags()
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

	if r.LatencyConfigs == nil && r.MaxLatencyMillis <= 0 {
		return fmt.Errorf("must have either latencyPercentiles or positive, non-zero maxLatencyMillis defined")
	}
	return nil
}

func (r *ServiceRoute) load(route string) error {
	r.Route = route
	if r.LatencyConfigs == nil {
		if r.MaxLatencyMillis == 0 {
			return fmt.Errorf("route must include maxLatencyMillis or latencyConfigs")
		} else {
			return nil
		}
	}
	hasDefault := false
	for _, cfg := range r.LatencyConfigs {
		err := cfg.loadDurations()
		if err != nil {
			return fmt.Errorf("error parsing latencyPercentiles: %v", err)
		}
		err = cfg.ValidateFlags()
		if err != nil {
			return err
		}
		if cfg.IsDefault() {
			if hasDefault {
				return fmt.Errorf("latencyConfigs must include exactly one default config (no flag_set or flag_unset)")
			}
			hasDefault = true
		}
	}
	if !hasDefault {
		return fmt.Errorf("latencyConfigs must include exactly one default config (no flag_set or flag_unset)")
	}
	return nil
}

func (r *ServiceRoute) SampleLatency() int64 {
	if r.LatencyConfigs == nil {
		return rand.Int63n(r.MaxLatencyMillis * 1000000)
	} else {
		return r.LatencyConfigs.Sample()
	}
}
