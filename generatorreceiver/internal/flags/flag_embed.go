package flags

import "fmt"

type EmbeddedFlags struct {
	FlagSet   string `json:"flag_set" yaml:"flag_set"`
	FlagUnset string `json:"flag_unset" yaml:"flag_unset"`
}

func (f *EmbeddedFlags) ShouldGenerate() bool {
	// TODO: use the set flag's _value_... somehow
	if f.FlagSet != "" {
		if set := Manager.GetFlag(f.FlagSet); !(set != nil && set.Active()) {
			return false
		}
	}
	if f.FlagUnset != "" {
		if unset := Manager.GetFlag(f.FlagUnset); unset != nil && unset.Active() {
			return false
		}
	}
	return true
}

func (f *EmbeddedFlags) Validate() error {
	if f.FlagSet != "" && Manager.GetFlag(f.FlagSet) == nil {
		return fmt.Errorf("flag %v does not exist", f.FlagSet)
	}
	if f.FlagUnset != "" && Manager.GetFlag(f.FlagUnset) == nil {
		return fmt.Errorf("flag %v does not exist", f.FlagUnset)
	}
	return nil
}
