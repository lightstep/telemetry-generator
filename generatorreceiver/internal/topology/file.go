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
		if rr.TracesPerHour <= 0 {
			return fmt.Errorf("rootRoute %s must have a positive, non-zero tracesPerHour defined", rr.Route)
		}
	}
	return nil
}

func (file *File) ValidateServices() error {
	var err error
	for _, s := range file.Topology.Services {
		for _, m := range s.Metrics {
			err = validateFlagsExist(m.FlagSet, m.FlagUnset)
			if err != nil {
				return fmt.Errorf("error with metric %s in service %s: %v", m.Name, s.ServiceName, err)
			}
		}
		for _, r := range s.Routes {
			err = file.validateRoute(r)
			if err != nil {
				return fmt.Errorf("error with route %s in service %s: %v", r.Route, s.ServiceName, err)
			}
		}
		for _, t := range s.TagSets {
			err = validateFlagsExist(t.FlagSet, t.FlagUnset)
			if err != nil {
				return fmt.Errorf("error with tagSets in service %s: %v", s.ServiceName, err)
			}
		}
	}
	// now that we've checked that all routes/services exist, safe to call validateDownstreamCalls()
	err = file.validateDownstreamCalls()
	if err != nil {
		return err
	}
	return nil
}

func (file *File) validateDownstreamCalls() error {
	for _, s := range file.Topology.Services {
		for _, r := range s.Routes {
			err := file.traverseServiceGraph(s.ServiceName, r.Route, map[string]bool{r.Route: true}, []string{r.Route})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (file *File) traverseServiceGraph(s string, r string, seenRoutes map[string]bool, orderedRoutes []string) error {
	downstreamCalls := file.Topology.GetServiceTier(s).GetRoute(r).DownstreamCalls
	for ds, dr := range downstreamCalls {
		if seenRoutes[dr] {
			return fmt.Errorf("cyclical service graph detected: %s", printServiceCycle(orderedRoutes, dr))
		}

		// make a copy of seenRoutes and add current route to it
		currentSeenRoutes := make(map[string]bool)
		for k, v := range seenRoutes {
			currentSeenRoutes[k] = v
		}
		currentSeenRoutes[dr] = true

		// make a copy of orderedRoutes and add current route to it
		// this slice is needed for printing routes in-order if cycle is detected
		var currentOrderedRoutes []string
		for _, name := range orderedRoutes {
			currentOrderedRoutes = append(currentOrderedRoutes, name)
		}
		currentOrderedRoutes = append(currentOrderedRoutes, dr)

		// check that everything downstream of the current route is valid
		err := file.traverseServiceGraph(ds, dr, currentSeenRoutes, currentOrderedRoutes)
		if err != nil {
			return err
		}
	}
	return nil
}

func printServiceCycle(seenOrdered []string, repeated string) string {
	var s string
	for _, route := range seenOrdered {
		s += fmt.Sprintf("%s -> ", route)
	}
	s += repeated
	return s
}

func (file *File) validateRoute(route *ServiceRoute) error {
	for s, r := range route.DownstreamCalls {
		st := file.Topology.GetServiceTier(s)
		if st == nil {
			return fmt.Errorf("downstream service %s does not exist", s)
		}
		if st.GetRoute(r) == nil {
			return fmt.Errorf("downstream service %s does not have route %s defined", s, r)
		}
		err := validateFlagsExist(route.FlagSet, route.FlagUnset)
		if err != nil {
			return err
		}
		if route.MaxLatencyMillis <= 0 {
			return fmt.Errorf("must have a positive, non-zero maxLatencyMillis defined")
		}
	}
	return nil
}

func validateFlagsExist(flagSet string, flagUnset string) error {
	// this should just verify that they exist, doesn't look for cycles
	if flagSet != "" && flags.Manager.GetFlag(flagSet) == nil {
		return fmt.Errorf("flag %v does not exist", flagSet)
	}
	if flagUnset != "" && flags.Manager.GetFlag(flagUnset) == nil {
		return fmt.Errorf("flag %v does not exist", flagUnset)
	}
	return nil
}
