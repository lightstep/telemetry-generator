package topology

import (
	"fmt"
	"math/rand"
)

type ServiceTier struct {
	ServiceName           string
	Routes                map[string]*ServiceRoute `json:"routes" yaml:"routes"`
	TagSets               []TagSet                 `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet   `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	Metrics               []Metric                 `json:"metrics" yaml:"metrics"`
	Random                *rand.Rand
}

func (st *ServiceTier) GetTagSet(routeName string) []TagSet {
	// TODO: support weight
	tags := st.TagSets
	routeTags := st.GetRoute(routeName).TagSets
	return append(tags, routeTags...)
}

func (st *ServiceTier) GetResourceAttributeSet() *ResourceAttributeSet {
	if len(st.ResourceAttributeSets) == 0 {
		return nil
	}
	var enabledResources []ResourceAttributeSet
	for _, resource := range st.ResourceAttributeSets {
		if resource.ShouldGenerate() {
			enabledResources = append(enabledResources, resource)
		}
	}

	if len(enabledResources) == 0 {
		return nil
	}

	// TODO: also support resource attributes on routes
	// TODO: support weight
	return &enabledResources[st.Random.Intn(len(enabledResources))]
}

func (st *ServiceTier) GetRoute(routeName string) *ServiceRoute {
	return st.Routes[routeName]
}

func (st *ServiceTier) Validate(topology Topology) (err error) {
	for _, m := range st.Metrics {
		err = m.validate()
		if err != nil {
			return fmt.Errorf("error with metric %s in service %s: %v", m.Name, st.ServiceName, err)
		}
	}
	for _, r := range st.Routes {
		err = r.validate(topology) //TODO- find better way to pass topology along
		if err != nil {
			return fmt.Errorf("error with route %s in service %s: %v", r.Route, st.ServiceName, err)
		}
	}
	for _, t := range st.TagSets {
		err = t.validate()
		if err != nil {
			return fmt.Errorf("error with tagSets in service %s: %v", st.ServiceName, err)
		}
	}
	return nil
}

func (st *ServiceTier) loadRoutes() (err error) {
	for name, route := range st.Routes {
		route.Route = name
		if route.LatencyPercentiles != nil {
			err = route.LatencyPercentiles.loadDurations()
			if err != nil {
				return fmt.Errorf("error parsing latencyPercentiles for route %s in service %s: %v", name, st.ServiceName, err)
			}
		}
	}
	return nil
}
