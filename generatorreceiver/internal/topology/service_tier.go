package topology

import (
	"math/rand"
)

type ServiceTier struct {
	ServiceName           string                 `json:"serviceName" yaml:"serviceName"`
	Routes                []*ServiceRoute        `json:"routes" yaml:"routes"`
	TagSets               []TagSet               `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	Metrics               []Metric               `json:"metrics" yaml:"metrics"`
	Random                *rand.Rand
	RouteMap              map[string]*ServiceRoute
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
	return st.RouteMap[routeName]
}

func (st *ServiceTier) LoadRouteMap() {
	st.RouteMap = make(map[string]*ServiceRoute)
	for _, r := range st.Routes {
		st.RouteMap[r.Route] = r
	}
}
