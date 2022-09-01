package topology

import (
	"github.com/lightstep/demo-environment/generatorreceiver/internal/flags"
)

type TagSet struct {
	Tags                TagMap         `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators       []TagGenerator `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit             []string       `json:"inherit,omitempty" yaml:"inherit,omitempty"`
	EmbeddedWeight      `json:",inline" yaml:",inline"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}
