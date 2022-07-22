package topology

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"go.uber.org/zap"
	"testing"
)

func TestMetric_ShouldGenerate(t *testing.T) {
	tests := []struct {
		name          string
		metric        Metric
		enabledFlags  []string
		disabledFlags []string
		want          bool
	}{
		{
			name: "one set flag, enabled",
			metric: Metric{
				Name:    "moot",
				Type:    "Gauge",
				Min:     0,
				Max:     100,
				FlagSet: "yesFlag",
			},
			enabledFlags: []string{"yesFlag"},
			want:         true,
		},
		{
			name: "one set flag, disabled",
			metric: Metric{
				Name:    "moot",
				Type:    "Gauge",
				Min:     0,
				Max:     100,
				FlagSet: "yesFlag",
			},
			disabledFlags: []string{"yesFlag"},
			want:          false,
		},
		{
			name: "one unset flag, disabled",
			metric: Metric{
				Name:      "moot",
				Type:      "Gauge",
				Min:       0,
				Max:       100,
				FlagUnset: "yesFlag",
			},
			disabledFlags: []string{"yesFlag"},
			want:          true,
		},
		{
			name: "one unset flag, enabled",
			metric: Metric{
				Name:      "moot",
				Type:      "Gauge",
				Min:       0,
				Max:       100,
				FlagUnset: "yesFlag",
			},
			enabledFlags: []string{"yesFlag"},
			want:         false,
		},
		{
			name: "one set on, one unset off, should generate",
			metric: Metric{
				Name:      "moot",
				Type:      "Gauge",
				Min:       0,
				Max:       100,
				FlagSet:   "yesFlag",
				FlagUnset: "noFlag",
			},
			enabledFlags:  []string{"yesFlag"},
			disabledFlags: []string{"noFlag"},
			want:          true,
		},
		{
			name: "one set on, one unset on, should not generate",
			metric: Metric{
				Name:      "moot",
				Type:      "Gauge",
				Min:       0,
				Max:       100,
				FlagSet:   "yesFlag",
				FlagUnset: "noFlag",
			},
			enabledFlags: []string{"yesFlag", "noFlag"},
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theFlags := make([]flags.Flag, 0, len(tt.enabledFlags)+len(tt.disabledFlags))
			for _, name := range tt.enabledFlags {
				theFlags = append(theFlags, flags.Flag{Name: name})
			}
			for _, name := range tt.disabledFlags {
				theFlags = append(theFlags, flags.Flag{Name: name})
			}
			fm := flags.NewFlagManager(flags.NewIncidentManager(), theFlags, zap.NewNop())
			for _, name := range tt.enabledFlags {
				fm.GetFlag(name).Enable()
			}
			for _, name := range tt.disabledFlags {
				fm.GetFlag(name).Disable()
			}
			if got := tt.metric.ShouldGenerate(fm); got != tt.want {
				t.Errorf("ShouldGenerate() = %v, want %v", got, tt.want)
			}
		})
	}
}
