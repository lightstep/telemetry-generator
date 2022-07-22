package topology

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"math"
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
)

type Metric struct {
	Name      string         `json:"name" yaml:"name"`
	Type      string         `json:"type" yaml:"type"`
	Min       float64        `json:"min" yaml:"min"`
	Max       float64        `json:"max" yaml:"max"`
	Period    *time.Duration `json:"period" yaml:"period"`
	Offset    *time.Duration `json:"offset" yaml:"offset"`
	FlagSet   string         `json:"flag_set" yaml:"flag_set"`
	FlagUnset string         `json:"flag_unset" yaml:"flag_unset"`
	Tags      map[string]string
	Shape     Shape `json:"shape" yaml:"shape"`
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

func (m *Metric) GetValue() float64 {
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
		default:
			// TODO: what would be a reasonable default? Maybe just sine?
			return SineValue(phase)
		}
	}(phase)

	return m.Min + (m.Max-m.Min)*factor
}

func (m *Metric) ShouldGenerate(fm *flags.FlagManager) bool {
	// TODO: use the set flag's _value_... somehow
	if m.FlagSet != "" {
		if set := fm.GetFlag(m.FlagSet); !(set != nil && set.Enabled()) {
			return false
		}
	}
	if m.FlagUnset != "" {
		if unset := fm.GetFlag(m.FlagUnset); unset != nil && unset.Enabled() {
			return false
		}
	}
	return true
}
