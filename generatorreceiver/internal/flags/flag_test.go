package flags

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
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
				{Name: "flag_a", Incident: &IncidentConfig{ParentFlag: "flag_b"}},
				{Name: "flag_b", Incident: &IncidentConfig{ParentFlag: "flag_c"}},
				{Name: "flag_c"},
			},
			error: false,
		},
		{
			name: "All parent flags exist but there is a cycle",
			flagCfgs: []FlagConfig{
				{Name: "flag_a", Incident: &IncidentConfig{ParentFlag: "flag_b"}},
				{Name: "flag_b", Incident: &IncidentConfig{ParentFlag: "flag_c"}},
				{Name: "flag_c", Incident: &IncidentConfig{ParentFlag: "flag_d"}},
				{Name: "flag_d", Incident: &IncidentConfig{ParentFlag: "flag_b"}},
			},
			error: true,
		},
		{
			name: "Flag has parent that does not exist",
			flagCfgs: []FlagConfig{
				{Name: "flag_a", Incident: &IncidentConfig{ParentFlag: "fake_parent"}},
				{Name: "flag_b", Incident: &IncidentConfig{ParentFlag: "flag_c"}},
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
