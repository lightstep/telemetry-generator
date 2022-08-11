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

func (st *ServiceTier) Validate(topology Topology) error {
	for _, m := range st.Metrics {
		err := m.EmbeddedFlags.Validate()
		if err != nil {
			return fmt.Errorf("error with metric %s in service %s: %v", m.Name, st.ServiceName, err)
		}
	}
	for _, r := range st.Routes {
		err := r.validate(topology)
		if err != nil {
			return fmt.Errorf("error with route %s in service %s: %v", r.Route, st.ServiceName, err)
		}
	}
	for _, t := range st.TagSets {
		err := t.EmbeddedFlags.Validate()
		if err != nil {
			return fmt.Errorf("error with tagSets in service %s: %v", st.ServiceName, err)
		}
	}
	for _, ra := range st.ResourceAttributeSets {
		err := ra.EmbeddedFlags.Validate()
		if err != nil {
			return fmt.Errorf("error with resourceAttributeSets in service %s: %v", st.ServiceName, err)
		}
	}
	return nil
}

func (st *ServiceTier) load(service string) error {
	st.ServiceName = service
	for name, route := range st.Routes {
		err := route.load(name)
		if err != nil {
			return fmt.Errorf("error loading route %s for service %s: %v", name, service, err)
		}
	}
	return nil
}
