package flags

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
