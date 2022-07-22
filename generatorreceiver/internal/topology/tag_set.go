package topology

type TagSet struct {
	Weight        int               `json:"weight" yaml:"weight"`
	FlagSet       string            `json:"flag_set" yaml:"flag_set"`
	FlagUnset     string            `json:"flag_unset" yaml:"flag_unset"`
	Tags          map[string]string `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators []TagGenerator    `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit       []string          `json:"inherit,omitempty" yaml:"inherit,omitempty"`
}
