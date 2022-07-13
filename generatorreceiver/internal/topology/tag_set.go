package topology


type TagSet struct {
	Weight int `json:"weight" yaml:"weight"`
	Tags map[string]interface{} `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators []TagGenerator `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit []string `json:"inherit,omitempty" yaml:"inherit,omitempty"`
}
