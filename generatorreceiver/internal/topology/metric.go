package topology

type Metric struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
	Min float64 `json:"min" yaml:"min"`
	Max float64 `json:"max" yaml:"max"`
	FlagSet string `json:"flag_set" yaml:"flag_set"`
	FlagUnset string `json:"flag_unset" yaml:"flag_unset"`
}