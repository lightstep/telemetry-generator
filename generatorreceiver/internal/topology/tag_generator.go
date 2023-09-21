package topology

import (
	"math/rand"
)

type TagGenerator struct {
	// NumTags is used to generate a random tag set if Tags is empty
	NumTags   int `json:"numTags,omitempty" yaml:"numTags,omitempty"`
	NumVals   int `json:"numVals,omitempty" yaml:"numVals,omitempty"`
	// Tags is hardcoded tags for this generator. eg - you may want a couple tags "foo,bar" with 5 values each.
	// If Tags is non-empty, NumTags is ignored
	Tags []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	ValueVariability int `json:"valueVariability,omitempty" yaml:"valueVariability,omitempty"`
	Random    *rand.Rand

	tags map[string]string
	values []string
}

func (t *TagGenerator) Init(random *rand.Rand) {
	t.Random = random
	t.tags = make(map[string]string)
	t.values = make([]string, t.NumVals)

	rtg := &RandomTagGenerator{random: random}
	for i := 0; i < t.NumVals; i++ {
		t.values[i] = rtg.GenerateValue()
	}

	if len(t.Tags) == 0 {
		t.Tags = make([]string, t.NumTags)
		for i:=0; i< t.NumTags; i++ {
			t.Tags[i] = rtg.GenerateKey()
		}
	}

	for _, tag :=range t.Tags {
		t.tags[tag] = rtg.randomString(t.values)
	}
}

func (t *TagGenerator) GetTags() map[string]string {
	return t.tags
}

func (t *TagGenerator) GetRefreshedTags() map[string]string{
	rtg := &RandomTagGenerator{random: t.Random}
	for k := range t.tags {
		n := t.Random.Intn(100)

		if n < t.ValueVariability {
			t.tags[k] = rtg.GenerateValue()
		}
	}

	return t.tags
}
