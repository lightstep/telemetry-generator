package topology

import (
	"encoding/binary"
	"math"

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

	// If no sets are generating, return zero value.
	var zeroP P
	if len(activeSets) == 0 {
		return zeroP
	}

	// Take out last 8 bytes from trace id
	secondHalf := traceID[8:16]
	// Transform them into a uint64
	traceUint := binary.BigEndian.Uint64(secondHalf)
	// Use the half of the traceID as a ratio.
	ratio := float64(traceUint) / float64(math.MaxUint64)
	// Search for the item by weight from N-1 items.
	chooseFrom := activeSets[:len(activeSets)-1]
	choice := ratio * totalWeight
	current := 0.0
	for _, set := range chooseFrom {
		current += set.GetWeight()
		if choice < current {
			return set
		}
	}

	// The last-weighted item was selected.  Floating point
	// rounding requires falling through here.
	return activeSets[len(activeSets)-1]
}
