package topology

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"math"
	"time"
)

const DefaultPeriod = 60 * time.Minute

type Metric struct {
	Name      string         `json:"name" yaml:"name"`
	Type      string         `json:"type" yaml:"type"`
	Min       float64        `json:"min" yaml:"min"`
	Max       float64        `json:"max" yaml:"max"`
	Period    *time.Duration `json:"period" yaml:"period"`
	FlagSet   string         `json:"flag_set" yaml:"flag_set"`
	FlagUnset string         `json:"flag_unset" yaml:"flag_unset"`
	Tags      map[string]string
	// TODO: add "shape"
}

func (m *Metric) GetValue() float64 {
	if m.Period == nil {
		period := DefaultPeriod
		m.Period = &period
	}

	now := time.Now()
	since := now.Sub(now.Truncate(*m.Period))
	phase := float64(since) / float64(*m.Period)
	sin := math.Sin(2 * math.Pi * phase)
	return m.Min + (m.Max-m.Min)*(sin+1)/2
}

func (m *Metric) ShouldGenerate(fm *flags.FlagManager) bool {
	// TODO: use the set flag's _value_... somehow
	if m.FlagSet != "" {
		if set := fm.GetFlag(m.FlagSet); !set.Enabled() {
			return false
		}
	}
	if m.FlagUnset != "" {
		if unset := fm.GetFlag(m.FlagUnset); unset.Enabled() {
			return false
		}
	}
	return true
}
