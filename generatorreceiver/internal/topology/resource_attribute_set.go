package topology

import (
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/flags"
	"math/rand"
)

type ResourceAttributeSet struct {
	Kubernetes          *Kubernetes `json:"kubernetes" yaml:"kubernetes"`
	ResourceAttributes  TagMap      `json:"resourceAttrs,omitempty" yaml:"resourceAttrs,omitempty"`
	EmbeddedWeight      `json:",inline" yaml:",inline"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func (r *ResourceAttributeSet) GetAttributes(random *rand.Rand) *TagMap {
	tm := make(TagMap)
	for k, v := range r.ResourceAttributes {
		tm[k] = v
	}
	if k8s := r.Kubernetes; k8s != nil {
		for k, v := range k8s.GetRandomK8sTags(random) {
			tm[k] = v
		}
	}
	return &tm
}
