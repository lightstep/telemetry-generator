package topology


type TagSet struct {
	Weight int `json:"weight"`
	Tags map[string]interface{} `json:"tags,omitempty"`
	TagGenerators []TagGenerator `json:"tagGenerators,omitempty"`
	Inherit []string `json:"inherit,omitempty"`
}
