package flags

import (
	"fmt"
	"time"
)

type EmbeddedFlags struct {
	FlagSet   string `json:"flag_set" yaml:"flag_set"`
	FlagUnset string `json:"flag_unset" yaml:"flag_unset"`
}

func (f *EmbeddedFlags) ShouldGenerate() bool {
	// TODO: use the set flag's _value_... somehow
	if f.FlagSet != "" {
		if set := Manager.GetFlag(f.FlagSet); !set.Active() {
			return false
		}
	}
	if f.FlagUnset != "" {
		if unset := Manager.GetFlag(f.FlagUnset); unset.Active() {
			return false
		}
	}
	return true
}

func (f *EmbeddedFlags) IsDefault() bool {
	return f.FlagSet == "" && f.FlagUnset == ""
}

func (f *EmbeddedFlags) GenerateStartTime() time.Time {
	if !f.ShouldGenerate() {
		return time.UnixMilli(0)
	}

	s, u := time.UnixMilli(0), time.UnixMilli(0)

	if f.FlagSet != "" {
		s = Manager.GetFlag(f.FlagSet).updated
	}

	if f.FlagUnset != "" {
		u = Manager.GetFlag(f.FlagUnset).updated
	}

	if s.After(u) {
		return s
	}

	return u
}

func (f *EmbeddedFlags) ValidateFlags() error {
	if f.FlagSet != "" && Manager.GetFlag(f.FlagSet) == nil {
		return fmt.Errorf("flag %v does not exist", f.FlagSet)
	}
	if f.FlagUnset != "" && Manager.GetFlag(f.FlagUnset) == nil {
		return fmt.Errorf("flag %v does not exist", f.FlagUnset)
	}
	return nil
}
