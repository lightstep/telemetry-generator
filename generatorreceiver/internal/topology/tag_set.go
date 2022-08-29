package topology

import (
	"math/rand"

	"github.com/lightstep/demo-environment/generatorreceiver/internal/flags"
)

type TagSet struct {
	Weight              float64        `json:"weight" yaml:"weight"`
	Tags                TagMap         `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators       []TagGenerator `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit             []string       `json:"inherit,omitempty" yaml:"inherit,omitempty"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func pickBasedOnWeight(tagSets []TagSet) TagSet {
	totalWeight := 0.0
	for _, ts := range tagSets {
		if !ts.ShouldGenerate() {
			continue // ignore disabled tagSets from calculations
		}
		totalWeight += ts.Weight
	}
	choice := rand.Float64() * totalWeight
	current := 0.0
	for _, ts := range tagSets {
		if !ts.ShouldGenerate() {
			continue
		}
		current += ts.Weight
		if choice < current {
			return ts
		}
	}
	return TagSet{}
}
