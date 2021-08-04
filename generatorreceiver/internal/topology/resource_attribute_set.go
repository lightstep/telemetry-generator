package topology

type ResourceAttributeSet struct {
	Weight int `json:"weight"`
	ResourceAttributes map[string]interface{} `json:"resourceAttrs,omitempty"`
}