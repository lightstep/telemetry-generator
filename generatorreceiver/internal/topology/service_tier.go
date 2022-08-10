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

func (st *ServiceTier) loadRoutes() (err error) {
	for name, route := range st.Routes {
		route.Route = name
		route.LatencyPercentiles = &LatencyPercentiles{}
		err = route.LatencyPercentiles.loadLatencyPercentiles(route.LatencyPercentilesConfig)
		if err != nil {
			return fmt.Errorf("error parsing latencyPercentiles for route %s in service %s: %v", name, st.ServiceName, err)
		}
	}
	return nil
}
