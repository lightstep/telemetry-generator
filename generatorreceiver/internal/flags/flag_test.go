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

func TestEmbeddedFlags_GenerateTime(t *testing.T) {
	Manager.Clear()

	Manager.LoadFlags([]FlagConfig{
		{
			Name: "flag_a",
		},
		{
			Name: "flag_b",
		},
	}, zap.NewNop())

	a := Manager.GetFlag("flag_a")
	b := Manager.GetFlag("flag_b")

	ef := EmbeddedFlags{
		FlagSet:   "flag_a",
		FlagUnset: "flag_b",
	}

	a.Enable()
	b.Enable()
	if st := ef.GenerateStartTime().UnixNano(); st != 0 {
		assert.Fail(t, "'b' is enabled, start time should be 0")
	}

	a.Disable()
	b.Disable()
	if st := ef.GenerateStartTime().UnixNano(); st != 0 {
		assert.Fail(t, "'a' is disabled, start time should be 0")
	}

	a.Disable()
	b.Enable()
	if st := ef.GenerateStartTime().UnixNano(); st != 0 {
		assert.Fail(t, "'a' is disabled and 'b' is enabled, start time should be 0")
	}

	a.Enable()
	time.Sleep(10 * time.Millisecond)
	b.Disable()
	if st := ef.GenerateStartTime().UnixNano(); st != b.updated.UnixNano() {
		assert.Fail(t, "start time should be the same as the time that the last flag was set.")
	}

	a.Disable()
	b.Enable()

	b.Disable()
	time.Sleep(10 * time.Millisecond)
	a.Enable()
	if st := ef.GenerateStartTime().UnixNano(); st != a.updated.UnixNano() {
		assert.Fail(t, "start time should be the same as the time that the last flag was set.")
	}

	ef = EmbeddedFlags{
		FlagSet: "flag_a",
	}

	a.Enable()
	if st := ef.GenerateStartTime().UnixNano(); st != a.updated.UnixNano() {
		assert.Fail(t, "generate start time should be equal 'a'.")
	}

	a.Disable()
	if st := ef.GenerateStartTime().UnixNano(); st != 0 {
		assert.Fail(t, "'a' is disabled, generate start time should be 0.")
	}

	ef = EmbeddedFlags{
		FlagUnset: "flag_b",
	}

	b.Disable()
	if st := ef.GenerateStartTime().UnixNano(); st != b.updated.UnixNano() {
		assert.Fail(t, "generate start time should be equal 'b'.")
	}

	b.Enable()
	if st := ef.GenerateStartTime().UnixNano(); st != 0 {
		assert.Fail(t, "'b' is enabled, generate start time should be 0.")
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

			// Populate incidentConfig with parsed durations
			parsedStart := make([]time.Duration, 0, len(tt.start))
			for _, s := range tt.start {
				time, _ := time.ParseDuration(s)
				parsedStart = append(parsedStart, time)
			}
			parsedDuration, _ := time.ParseDuration(tt.duration)
			incidentCfg := IncidentConfig{
				ParentFlag: tt.parent,
				Start:      parsedStart,
				Duration:   parsedDuration,
			}

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

func TestFlag_shouldBeActive(t *testing.T) {
	tests := []struct {
		name             string
		incidentDuration string
		start            []string
		duration         string
		result           bool
	}{
		{
			name:             "Incident duration equals child start time",
			incidentDuration: "0m",
			start:            []string{"0m"},
			duration:         "5m",
			result:           false,
		},
		{
			name:             "Incident duration is just past child start time",
			incidentDuration: "1ms",
			start:            []string{"0m"},
			duration:         "5m",
			result:           true,
		},
		{
			name:             "Incident duration equals child end time",
			incidentDuration: "15m",
			start:            []string{"5m"},
			duration:         "10m",
			result:           false,
		},
		{
			name:             "Incident duration is just past child end time",
			incidentDuration: "15m1ms",
			start:            []string{"5m"},
			duration:         "10m",
			result:           false,
		},
		{
			name:             "Incident duration is between child start times and within child duration",
			incidentDuration: "13m",
			start:            []string{"4m", "12m", "20m"},
			duration:         "2m",
			result:           true,
		},
		{
			name:             "Incident duration is between child start times and not within child duration",
			incidentDuration: "7m",
			start:            []string{"4m", "12m", "20m"},
			duration:         "2m",
			result:           false,
		},
		{
			name:             "Incident duration falls within overlapping child durations",
			incidentDuration: "12m",
			start:            []string{"5m", "10m"},
			duration:         "10m",
			result:           true,
		},
		{
			name:             "Incident duration does not fall within overlapping child durations",
			incidentDuration: "30m",
			start:            []string{"5m", "10m"},
			duration:         "10m",
			result:           false,
		},
		{
			name:             "Incident duration is after child start time and no child duration is specified",
			incidentDuration: "10m",
			start:            []string{"3m"},
			duration:         "",
			result:           true,
		},
		{
			name:             "Incident duration is before child start time and no child duration is specified",
			incidentDuration: "2m",
			start:            []string{"3m"},
			duration:         "",
			result:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedStart := make([]time.Duration, 0, len(tt.start))
			for _, s := range tt.start {
				time, _ := time.ParseDuration(s)
				parsedStart = append(parsedStart, time)
			}
			parsedDuration, _ := time.ParseDuration(tt.duration)
			f := Flag{cfg: FlagConfig{Incident: &IncidentConfig{Start: parsedStart, Duration: parsedDuration}}}
			parsedIncidentDuration, _ := time.ParseDuration(tt.incidentDuration)
			assert.Equal(t, tt.result, f.shouldBeActive(parsedIncidentDuration))
		})
	}
}
