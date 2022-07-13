package topology

type ResourceAttributeSet struct {
	Weight int `json:"weight" yaml:"weight"`
	ResourceAttributes map[string]interface{} `json:"resourceAttrs,omitempty" yaml:"resourceAttrs,omitempty"`
}