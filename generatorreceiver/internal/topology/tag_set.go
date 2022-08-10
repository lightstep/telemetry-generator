package topology

import (
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
)

type TagSet struct {
	Weight              int            `json:"weight" yaml:"weight"`
	Tags                TagMap         `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators       []TagGenerator `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit             []string       `json:"inherit,omitempty" yaml:"inherit,omitempty"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func (ts *TagSet) validate() error {
	if ts.FlagSet != "" && flags.Manager.GetFlag(ts.FlagSet) == nil {
		return fmt.Errorf("flag %v does not exist", ts.FlagSet)
	}
	if ts.FlagUnset != "" && flags.Manager.GetFlag(ts.FlagUnset) == nil {
		return fmt.Errorf("flag %v does not exist", ts.FlagUnset)
	}
	return nil
}
