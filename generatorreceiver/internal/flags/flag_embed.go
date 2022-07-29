package flags

type EmbeddedFlags struct {
	FlagSet   string `json:"flag_set" yaml:"flag_set"`
	FlagUnset string `json:"flag_unset" yaml:"flag_unset"`
}

func (f *EmbeddedFlags) ShouldGenerate(fm *FlagManager) bool {
	// TODO: use the set flag's _value_... somehow
	if f.FlagSet != "" {
		if set := fm.GetFlag(f.FlagSet); !(set != nil && set.Enabled()) {
			return false
		}
	}
	if f.FlagUnset != "" {
		if unset := fm.GetFlag(f.FlagUnset); unset != nil && unset.Enabled() {
			return false
		}
	}
	return true
}
