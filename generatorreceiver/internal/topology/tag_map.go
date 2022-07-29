package topology

import (
	"go.opentelemetry.io/collector/model/pdata"
	"strconv"
)

type tagMap map[string]interface{}

func (tm *tagMap) InsertTags(attr *pdata.AttributeMap) {
	for key, val := range *tm {
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
