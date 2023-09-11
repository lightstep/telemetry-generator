package topology

import (
	"math/rand"
)

type TagGenerator struct {
	NumTags   int `json:"numTags,omitempty" yaml:"numTags,omitempty"`
	NumVals   int `json:"numVals,omitempty" yaml:"numVals,omitempty"`
	ValueVariability int `json:"valueVariability,omitempty" yaml:"valueVariability,omitempty"`
	Random    *rand.Rand

	tags map[string]string
	values []string
}

func (t *TagGenerator) Init(random *rand.Rand) {
	t.Random = random
	t.tags = make(map[string]string, t.NumTags)
	t.values = make([]string, t.NumVals)

	rtg := &RandomTagGenerator{random: random}
	for i := 0; i < t.NumVals; i++ {
		t.values[i] = rtg.GenerateValue()
	}

	for i:=0; i< t.NumTags; i++ {
		t.tags[rtg.GenerateKey()] = rtg.randomString(t.values)
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int, r *rand.Rand) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
