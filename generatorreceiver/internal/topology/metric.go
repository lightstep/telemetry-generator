package topology

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"math"
	"math/rand"
	"time"
)

const DefaultPeriod = 60 * time.Minute
const DefaultOffset = 0 * time.Minute

type Shape string

const (
	Sine     Shape = "sine"
	Sawtooth Shape = "sawtooth"
	Square   Shape = "square"
	Triangle Shape = "triangle"
	Average  Shape = "average"
)

type Metric struct {
	Name                string            `json:"name" yaml:"name"`
	Type                string            `json:"type" yaml:"type"`
	Min                 float64           `json:"min" yaml:"min"`
	Max                 float64           `json:"max" yaml:"max"`
	Period              *time.Duration    `json:"period" yaml:"period"`
	Offset              *time.Duration    `json:"offset" yaml:"offset"`
	Shape               Shape             `json:"shape" yaml:"shape"`
	Tags                map[string]string `json:"tags" yaml:"tags"`
	Jitter              float64           `json:"jitter" yaml:"jitter"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func SineValue(phase float64) float64 {
	return (math.Sin(2*math.Pi*phase) + 1) / 2
}

func SawtoothValue(phase float64) float64 {
	return phase
}

func SquareValue(phase float64) float64 {
	if phase < 0.5 {
		return 0.0
	}
	return 1.0
}

func TriangleValue(phase float64) float64 {
	return 1.0 - 2.0*math.Abs(0.5-phase)
}

func AverageValue(_ float64) float64 {
	return 0.5
}

func (m *Metric) GetValue(random *rand.Rand) float64 {
	if m.Period == nil {
		period := DefaultPeriod
		m.Period = &period
	}
	if m.Offset == nil {
		offset := DefaultOffset
		m.Offset = &offset
	}

	now := time.Now().Add(-*m.Offset)
	since := now.Sub(now.Truncate(*m.Period))
	phase := float64(since) / float64(*m.Period)

	factor := func(phase float64) float64 {
		switch m.Shape {
		case Sine:
			return SineValue(phase)
		case Sawtooth:
			return SawtoothValue(phase)
		case Square:
			return SquareValue(phase)
		case Triangle:
			return TriangleValue(phase)
		case Average:
			return AverageValue(phase)
		default:
			// TODO: what would be a reasonable default? Maybe just sine?
			return SineValue(phase)
		}
	}(phase)

	v := m.Min + (m.Max-m.Min)*factor

	// adds jitter that ranges from [-m.Jitter/2, m.Jitter/2]
	j := 1 + random.Float64()*m.Jitter - m.Jitter/2

	v = v * j

	// ensures value is on the [m.Min, m.Max] boundary
	v = math.Min(v, m.Max)
	v = math.Max(v, m.Min)

	return v
}
