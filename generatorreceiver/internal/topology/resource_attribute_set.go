package topology

import "github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"

type ResourceAttributeSet struct {
	Weight              int        `json:"weight" yaml:"weight"`
	Kubernetes          Kubernetes `json:"kubernetes" yaml:"kubernetes"`
	ResourceAttributes  TagMap     `json:"resourceAttrs,omitempty" yaml:"resourceAttrs,omitempty"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}
