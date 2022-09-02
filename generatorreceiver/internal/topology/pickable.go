package topology

import "math/rand"

type Pickable interface { // currently TagSet and ResourceAttributeSet satisfy this interface
	GetWeight() float64
	ShouldGenerate() bool
}

type EmbeddedWeight struct {
	Weight float64 `json:"weight" yaml:"weight"`
}

func (w EmbeddedWeight) GetWeight() float64 {
	return w.Weight
}

func pickBasedOnWeight[P Pickable](ps []P) P {
	var activeSets []P
	totalWeight := 0.0
	for _, set := range ps {
		if set.ShouldGenerate() {
			activeSets = append(activeSets, set)
			totalWeight += set.GetWeight()
		}
	}

	choice := rand.Float64() * totalWeight
	current := 0.0
	for _, set := range activeSets {
		current += set.GetWeight()
		if choice < current {
			return set
		}
	}

	// if no set is chosen, return zero value
	var zeroP P
	return zeroP
}
