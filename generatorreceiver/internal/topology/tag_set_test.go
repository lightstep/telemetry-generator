package topology

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestInsertTag(t *testing.T) {
	tags := map[string]interface{}{
		"key1": true,
		"key2": "hi",
		"key3": 123.123,
		"key4": 10,
	}

	ts := &TagSet{
		Tags: tags,
	}

	attr := pdata.NewAttributeMap()

	ts.Tags.InsertTags(&attr)

	expectedAttr := pdata.NewAttributeMap()
	expectedAttr.InsertBool("key1", true)
	expectedAttr.InsertString("key2", "hi")
	expectedAttr.InsertDouble("key3", 123.123)
	expectedAttr.InsertInt("key4", 10)
	require.Equal(t, attr.Sort().AsRaw(), expectedAttr.Sort().AsRaw())
}
