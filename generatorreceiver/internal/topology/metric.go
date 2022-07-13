package topology

type Metric struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
	Min float64 `json:"min" yaml:"min"`
	Max float64 `json:"max" yaml:"max"`

}