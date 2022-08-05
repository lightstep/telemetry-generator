package topology

import (
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
)

type File struct {
	Topology   *Topology          `json:"topology" yaml:"topology"`
	Flags      []flags.FlagConfig `json:"flags" yaml:"flags"`
	RootRoutes []struct {
		Service       string `json:"service" yaml:"service"`
		Route         string `json:"route" yaml:"route"`
		TracesPerHour int    `json:"tracesPerHour" yaml:"tracesPerHour"`
	} `json:"rootRoutes" yaml:"rootRoutes"`
}

func (file *File) ValidateRootRoutes() error {
	for _, rr := range file.RootRoutes {
		st := file.Topology.GetServiceTier(rr.Service)
		if st == nil {
			return fmt.Errorf("service %s does not exist", rr.Service)
		}
		if st.GetRoute(rr.Route) == nil {
			return fmt.Errorf("service %s does not have route %s defined", rr.Service, rr.Route)
		}
		if rr.TracesPerHour == 0 {
			return fmt.Errorf("rootRoute %s must have a non-zero tracesPerHour specified", rr.Route)
		}
	}
	return nil
}

func (file *File) ValidateServices() error {
	for _, s := range file.Topology.Services {
		for _, m := range s.Metrics {
			err := validateFlags(m.FlagSet, m.FlagUnset)
			if err != nil {
				return fmt.Errorf("error with metric %s in service %s: %v", m.Name, s.ServiceName, err)
			}
		}
		for _, r := range s.Routes {
			err := file.validateRoute(&r)
			if err != nil {
				return fmt.Errorf("error with route %s in service %s: %v", r.Route, s.ServiceName, err)
			}
		}
		// todo- check for loops in service graph
	}
	return nil
}

func (file *File) validateRoute(route *ServiceRoute) error {
	for k, v := range route.DownstreamCalls {
		st := file.Topology.GetServiceTier(k)
		if st == nil {
			return fmt.Errorf("downstream service %s does not exist", k)
		}
		if st.GetRoute(v) == nil {
			return fmt.Errorf("downstream service %s does not have route %s", k, v)
		}
	}
	return nil
}

func validateFlags(flagSet string, flagUnset string) error {
	// this should just verify that they exist, doesn't look for cycles
	if flagSet != "" && flags.Manager.GetFlag(flagSet) == nil {
		return fmt.Errorf("flag %v does not exist", flagSet)
	}
	if flagUnset != "" && flags.Manager.GetFlag(flagUnset) == nil {
		return fmt.Errorf("flag %v does not exist", flagUnset)
	}
	return nil
}
