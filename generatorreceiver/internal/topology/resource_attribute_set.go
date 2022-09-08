package topology

import (
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/flags"
)

type ResourceAttributeSet struct {
	Kubernetes          Kubernetes `json:"kubernetes" yaml:"kubernetes"`
	ResourceAttributes  TagMap     `json:"resourceAttrs,omitempty" yaml:"resourceAttrs,omitempty"`
	EmbeddedWeight      `json:",inline" yaml:",inline"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func (r *ResourceAttributeSet) GetAttributes() *TagMap {
	tm := make(TagMap)
	for k, v := range r.ResourceAttributes {
		tm[k] = v
	}

	for k, v := range r.Kubernetes.GetK8sTags() {
		tm[k] = v
	}

	return &tm
}
