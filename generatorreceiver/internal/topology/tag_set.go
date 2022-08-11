package topology

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
)

type TagSet struct {
	Weight              int            `json:"weight" yaml:"weight"`
	Tags                TagMap         `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators       []TagGenerator `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit             []string       `json:"inherit,omitempty" yaml:"inherit,omitempty"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}
