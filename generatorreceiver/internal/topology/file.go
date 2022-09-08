package topology

import (
	"fmt"
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/flags"
)

type File struct {
	Topology   *Topology          `json:"topology" yaml:"topology"`
	Flags      []flags.FlagConfig `json:"flags" yaml:"flags"`
	RootRoutes []RootRoute        `json:"rootRoutes" yaml:"rootRoutes"`
}

type RootRoute struct {
	Service             string `json:"service" yaml:"service"`
	Route               string `json:"route" yaml:"route"`
	TracesPerHour       int    `json:"tracesPerHour" yaml:"tracesPerHour"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
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
		if rr.TracesPerHour <= 0 {
			return fmt.Errorf("rootRoute %s must have a positive, non-zero tracesPerHour defined", rr.Route)
		}
	}
	return nil
}
