package flags

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestFlagManager_ValidateFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagCfgs []FlagConfig
		error    bool
	}{
		{
			name: "All parent flags exist and there are no cycles",
			flagCfgs: []FlagConfig{
				{Name: "flag_a", Incident: &IncidentConfig{ParentFlag: "flag_b", Start: []time.Duration{time.Duration(0)}}},
				{Name: "flag_b", Incident: &IncidentConfig{ParentFlag: "flag_c", Start: []time.Duration{time.Duration(0)}}},
				{Name: "flag_c"},
			},
			error: false,
		},
		{
			name: "All parent flags exist but there is a cycle",
			flagCfgs: []FlagConfig{
				{Name: "flag_a", Incident: &IncidentConfig{ParentFlag: "flag_b", Start: []time.Duration{time.Duration(0)}}},
				{Name: "flag_b", Incident: &IncidentConfig{ParentFlag: "flag_c", Start: []time.Duration{time.Duration(0)}}},
				{Name: "flag_c", Incident: &IncidentConfig{ParentFlag: "flag_d", Start: []time.Duration{time.Duration(0)}}},
				{Name: "flag_d", Incident: &IncidentConfig{ParentFlag: "flag_b", Start: []time.Duration{time.Duration(0)}}},
			},
			error: true,
		},
		{
			name: "Flag has parent that does not exist",
			flagCfgs: []FlagConfig{
				{Name: "flag_a", Incident: &IncidentConfig{ParentFlag: "fake_parent", Start: []time.Duration{time.Duration(0)}}},
				{Name: "flag_b", Incident: &IncidentConfig{ParentFlag: "flag_c", Start: []time.Duration{time.Duration(0)}}},
				{Name: "flag_c"},
			},
			error: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Manager.Clear()
			Manager.LoadFlags(tt.flagCfgs, zap.NewNop())

			err := Manager.ValidateFlags()
			if err != nil && !tt.error {
				assert.Fail(t, fmt.Sprintf("did not expect validation error but got: %v", err))
			}
			if err == nil && tt.error {
				assert.Fail(t, "expected validation error")
			}
		})
	}
}

func TestEmbeddedFlags_Validate(t *testing.T) {
	tests := []struct {
		name          string
		embeddedFlags EmbeddedFlags
		flags         []string
		error         bool
	}{
		{
			name: "FlagSet specified but does not exist",
			embeddedFlags: EmbeddedFlags{
				FlagSet:   "fakeFlag",
				FlagUnset: "",
			},
			flags: []string{"realFlag"},
			error: true,
		},
		{
			name: "FlagUnset specified but does not exist",
			embeddedFlags: EmbeddedFlags{
				FlagSet:   "",
				FlagUnset: "anotherFakeFlag",
			},
			flags: []string{"anotherRealFlag"},
			error: true,
		},
		{
			name: "No flags specified",
			embeddedFlags: EmbeddedFlags{
				FlagSet:   "",
				FlagUnset: "",
			},
			flags: []string{"someFlag", "someOtherFlag"},
			error: false,
		},
		{
			name: "Both flags specified and exist",
			embeddedFlags: EmbeddedFlags{
				FlagSet:   "realFlag",
				FlagUnset: "anotherRealFlag",
			},
			flags: []string{"realFlag", "anotherRealFlag"},
			error: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Manager.Clear()
			theFlags := make([]FlagConfig, 0, len(tt.flags))
			for _, name := range tt.flags {
				theFlags = append(theFlags, FlagConfig{Name: name})
			}
			Manager.LoadFlags(theFlags, zap.NewNop())

			err := tt.embeddedFlags.ValidateFlags()
			if err != nil && !tt.error {
				assert.Fail(t, fmt.Sprintf("did not expect validation error but got: %v", err))
			}
			if err == nil && tt.error {
				assert.Fail(t, "expected validation error")
			}
		})
	}
}

func TestIncidentConfig_validate(t *testing.T) {
	tests := []struct {
		name     string
		parent   string
		start    []string
		duration string
		flags    []string
		error    bool
	}{
		{
			name:     "Multiple start times and duration",
			parent:   "someFlag",
			start:    []string{"0m", "5m"},
			duration: "5m",
			flags:    []string{"someFlag", "someOtherFlag"},
			error:    false,
		},
		{
			name:     "Multiple start times but no duration",
			parent:   "someOtherFlag",
			start:    []string{"0m", "5m"},
			duration: "",
			flags:    []string{"someFlag", "someOtherFlag"},
			error:    true,
		},
		{
			name:     "One start time but no duration",
			parent:   "someFlag",
			start:    []string{"10m"},
			duration: "",
			flags:    []string{"someFlag", "someOtherFlag"},
			error:    false,
		},
		{
			name:     "No start times",
			parent:   "someOtherFlag",
			start:    []string{},
			duration: "20m",
			flags:    []string{"someFlag", "someOtherFlag"},
			error:    true,
		},
		{
			name:     "Duplicate start times",
			parent:   "someFlag",
			start:    []string{"5m", "300s"},
			duration: "5m",
			flags:    []string{"someFlag", "someOtherFlag"},
			error:    true,
		},
		{
			name:     "Start times not strictly increasing",
			parent:   "someOtherFlag",
			start:    []string{"5m", "10m", "15m", "8m"},
			duration: "1m",
			flags:    []string{"someFlag", "someOtherFlag"},
			error:    true,
		},
		{
			name:     "Parent flag does not exist",
			parent:   "fakeFlag",
			start:    []string{"0m", "15m"},
			duration: "10m",
			flags:    []string{"someFlag", "someOtherFlag"},
			error:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure flag manager and its flag map
			Manager.Clear()
			theFlags := make([]FlagConfig, 0, len(tt.flags))
			for _, name := range tt.flags {
				theFlags = append(theFlags, FlagConfig{Name: name})
			}
			Manager.LoadFlags(theFlags, zap.NewNop())

			// Populate incidentConfig with parsed times
			parsedStart := make([]time.Duration, 0, len(tt.start))
			for _, s := range tt.start {
				duration, _ := time.ParseDuration(s)
				parsedStart = append(parsedStart, duration)
			}
			parsedDuration, _ := time.ParseDuration(tt.duration)
			incidentCfg := IncidentConfig{
				ParentFlag: tt.parent,
				Start:      parsedStart,
				Duration:   parsedDuration,
			}

			// Run validate() method and check if error is expected
			err := incidentCfg.validate()
			if err != nil && !tt.error {
				assert.Fail(t, fmt.Sprintf("did not expect validation error but got: %v", err))
			}
			if err == nil && tt.error {
				assert.Fail(t, "expected validation error")
			}
		})
	}
}
