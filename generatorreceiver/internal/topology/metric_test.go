package topology

import (
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/flags"
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
				Name: "moot",
				Type: "Gauge",
				Min:  0,
				Max:  100,
				EmbeddedFlags: flags.EmbeddedFlags{
					FlagSet: "yesFlag",
				},
			},
			enabledFlags: []string{"yesFlag"},
			want:         true,
		},
		{
			name: "one set flag, disabled",
			metric: Metric{
				Name: "moot",
				Type: "Gauge",
				Min:  0,
				Max:  100,
				EmbeddedFlags: flags.EmbeddedFlags{
					FlagSet: "yesFlag",
				},
			},
			disabledFlags: []string{"yesFlag"},
			want:          false,
		},
		{
			name: "one unset flag, disabled",
			metric: Metric{
				Name: "moot",
				Type: "Gauge",
				Min:  0,
				Max:  100,
				EmbeddedFlags: flags.EmbeddedFlags{
					FlagUnset: "yesFlag",
				},
			},
			disabledFlags: []string{"yesFlag"},
			want:          true,
		},
		{
			name: "one unset flag, enabled",
			metric: Metric{
				Name: "moot",
				Type: "Gauge",
				Min:  0,
				Max:  100,
				EmbeddedFlags: flags.EmbeddedFlags{
					FlagUnset: "yesFlag",
				},
			},
			enabledFlags: []string{"yesFlag"},
			want:         false,
		},
		{
			name: "one set on, one unset off, should generate",
			metric: Metric{
				Name: "moot",
				Type: "Gauge",
				Min:  0,
				Max:  100,
				EmbeddedFlags: flags.EmbeddedFlags{
					FlagSet:   "yesFlag",
					FlagUnset: "noFlag",
				},
			},
			enabledFlags:  []string{"yesFlag"},
			disabledFlags: []string{"noFlag"},
			want:          true,
		},
		{
			name: "one set on, one unset on, should not generate",
			metric: Metric{
				Name: "moot",
				Type: "Gauge",
				Min:  0,
				Max:  100,
				EmbeddedFlags: flags.EmbeddedFlags{
					FlagSet:   "yesFlag",
					FlagUnset: "noFlag",
				},
			},
			enabledFlags: []string{"yesFlag", "noFlag"},
			want:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags.Manager.Clear()
			theFlags := make([]flags.FlagConfig, 0, len(tt.enabledFlags)+len(tt.disabledFlags))
			for _, name := range tt.enabledFlags {
				theFlags = append(theFlags, flags.FlagConfig{Name: name})
			}
			for _, name := range tt.disabledFlags {
				theFlags = append(theFlags, flags.FlagConfig{Name: name})
			}
			flags.Manager.LoadFlags(theFlags, zap.NewNop())
			for _, name := range tt.enabledFlags {
				flags.Manager.GetFlag(name).Enable()
			}
			for _, name := range tt.disabledFlags {
				flags.Manager.GetFlag(name).Disable()
			}
			if got := tt.metric.ShouldGenerate(); got != tt.want {
				t.Errorf("ShouldGenerate() = %v, want %v", got, tt.want)
			}
		})
	}
}
