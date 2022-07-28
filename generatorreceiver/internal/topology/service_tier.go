package topology

import "math/rand"

type ServiceTier struct {
	ServiceName           string                 `json:"serviceName" yaml:"serviceName"`
	Routes                []ServiceRoute         `json:"routes" yaml:"routes"`
	TagSets               []TagSet               `json:"tagSets" yaml:"tagSets"`
	ResourceAttributeSets []ResourceAttributeSet `json:"resourceAttrSets" yaml:"resourceAttrSets"`
	Metrics               []Metric               `json:"metrics" yaml:"metrics"`
	Kubernetes            Kubernetes             `json:"kubernetes" yaml:"kubernetes"`
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
	// TODO: also support resource attributes on routes
	// TODO: support weight
	return &st.ResourceAttributeSets[st.Random.Intn(len(st.ResourceAttributeSets))]
}

func (st *ServiceTier) GetRoute(routeName string) *ServiceRoute {
	for _, v := range st.Routes {
		if v.Route == routeName {
			return &v
		}
	}
	return nil
}
