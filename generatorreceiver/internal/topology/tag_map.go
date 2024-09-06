package topology

import (
	"fmt"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"math/rand"
	"strconv"
)

type TagMap map[string]interface{}

func (tm *TagMap) InsertTags(attr *pcommon.Map, random *rand.Rand) {
	for key, val := range *tm {
		switch val := val.(type) {
		case float64:
			attr.PutDouble(key, val)
		case int:
			attr.PutInt(key, int64(val))
		case string:
			_, err := strconv.Atoi(val)
			if err != nil {
				attr.PutStr(key, val)
			}
		case bool:
			attr.PutBool(key, val)
		case []string:
			r := random.Intn(len(val))
			attr.PutStr(key, val[r])
		default:
			attr.PutStr(key, fmt.Sprint(val))
		}
	}
}
