package topology

import (
	"fmt"
	"go.opentelemetry.io/collector/model/pdata"
	"math/rand"
	"strconv"
	"time"
)

type TagMap map[string]interface{}

func (tm *TagMap) InsertTags(attr *pdata.AttributeMap) {
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
		case []string:
			rand.Seed(time.Now().Unix())
			choice := rand.Intn(len(val))
			attr.InsertString(key, val[choice])
		default:
			attr.InsertString(key, fmt.Sprint(val))
		}
	}
}
