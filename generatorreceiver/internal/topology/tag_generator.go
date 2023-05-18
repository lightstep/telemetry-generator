package topology

import "math/rand"

type TagGenerator struct {
	ValLength int    `json:"valLength,omitempty" yaml:"valLength,omitempty"`
	NumTags   int    `json:"numTags,omitempty" yaml:"numTags,omitempty"`
	KeyName   string `json:"keyName,omitempty" yaml:"keyName,omitempty"`
	NumVals   int    `json:"numVals,omitempty" yaml:"numVals,omitempty"`
	Random    *rand.Rand
}

func (t *TagGenerator) GenerateTags() map[string]string {
	retVal := make(map[string]string)

	if len(t.KeyName) != 0 {
		retVal[t.KeyName] = randStringBytes(t.ValLength, t.Random)
		return retVal
	}

	nameGenerator := &TagNameGenerator{
		random: t.Random,
	}

	for i := 0; i < t.NumTags; i++ {
		retVal[nameGenerator.Generate()] = randStringBytes(t.ValLength, t.Random)
	}
	return retVal
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int, r *rand.Rand) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
