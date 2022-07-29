package topology

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"strconv"

	"go.opentelemetry.io/collector/model/pdata"
)

type TagSet struct {
	Weight              int                    `json:"weight" yaml:"weight"`
	Tags                map[string]interface{} `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators       []TagGenerator         `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit             []string               `json:"inherit,omitempty" yaml:"inherit,omitempty"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func (ts *TagSet) InsertTags(attr *pdata.AttributeMap) {
	for key, val := range ts.Tags {
		switch val := val.(type) {
		case float64:
			attr.InsertDouble(key, val)
		case int:
			attr.InsertInt(key, int64(val))
		case string:
			_, err := strconv.Atoi(val)
			if err != nil {
				attr.InsertString(key, val)
			}
		case bool:
			attr.InsertBool(key, val)
		default:
			// just insert empty string if it's an unsupported type instead of returning error (todo decide if we want error handling somewhere above this)
			attr.InsertString(key, "")
		}
	}
}
