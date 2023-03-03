package topology

import (
	"encoding/binary"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

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

func pickBasedOnWeight[P Pickable](ps []P, traceID pcommon.TraceID) P {
	var activeSets []P
	totalWeight := 0.0
	for _, set := range ps {
		if set.ShouldGenerate() {
			activeSets = append(activeSets, set)
			totalWeight += set.GetWeight()
		}
	}

	// Take out last 8 bytes from trace id
	secondHalf := traceID[8:16]
	// Transform them into a uint64
	traceUint := binary.BigEndian.Uint64(secondHalf)
	// Use the last two digits as percentage.
	choice := float64(traceUint%100) / 100.0

	choice = choice * totalWeight
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
