package topology

import "fmt"

type Topology struct {
	Services map[string]*ServiceTier `json:"services" yaml:"services"`
}

func (t *Topology) GetServiceTier(serviceName string) *ServiceTier {
	return t.Services[serviceName]
}

func (t *Topology) ValidateServiceGraph() error {
	for _, st := range t.Services {
		for _, r := range st.Routes {
			err := t.validateDownstreamCalls(st.ServiceName, r.Route)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Topology) validateDownstreamCalls(service string, route string) error {
	seenCalls := map[string]bool{service + route: true}
	orderedCalls := []string{service + route}
	err := t.traverseServiceGraph(service, route, seenCalls, orderedCalls)
	// todo- optimize so that you dont check same service/route combo twice
	if err != nil {
		return err
	}
	return nil
}

func (t *Topology) traverseServiceGraph(service string, route string, seenCalls map[string]bool, orderedCalls []string) error {
	downstreamCalls := t.GetServiceTier(service).GetRoute(route).DownstreamCalls
	// already validated existence of all services/routes, so ^ is safe
	for ds, dr := range downstreamCalls {
		if seenCalls[ds+dr] {
			return fmt.Errorf(printServiceCycle(orderedCalls, ds+dr))
		}

		// make a copy of seenCalls and add current call to it (will be passed to recursive func)
		currentSeenCalls := make(map[string]bool)
		for k, v := range seenCalls {
			currentSeenCalls[k] = v
		}
		currentSeenCalls[ds+dr] = true

		// make a copy of orderedCalls and add current call to it (will be passed to recursive func)
		// this slice is needed for printing calls in-order if cycle is detected
		var currentOrderedCalls []string
		for _, name := range orderedCalls {
			currentOrderedCalls = append(currentOrderedCalls, name)
		}
		currentOrderedCalls = append(currentOrderedCalls, ds+dr)

		err := t.traverseServiceGraph(ds, dr, currentSeenCalls, currentOrderedCalls)
		if err != nil {
			return err
		}
	}
	return nil
}

func printServiceCycle(seenCalls []string, repeated string) string {
	var s string
	for _, call := range seenCalls {
		s += fmt.Sprintf("%s -> ", call)
	}
	s += repeated
	return s
}

func (t *Topology) Load() error {
	for name, service := range t.Services {
		service.ServiceName = name
		err := service.loadRoutes()
		if err != nil {
			return err
		}
	}
	return nil
}
