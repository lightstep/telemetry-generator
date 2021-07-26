package topology

type Metric struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`

}