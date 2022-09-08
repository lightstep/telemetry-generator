package topology

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestTagMap_InsertTag(t *testing.T) {
	csvTags, _ := readCsv("testdata/color_tags.csv")
	tags := map[string]interface{}{
		"key1": true,
		"key2": "hi",
		"key3": 123.123,
		"key4": 10,
		"key5": csvTags,
	}

	ts := &TagSet{
		Tags: tags,
	}

	attr := pdata.NewAttributeMap()

	rand.Seed(123)
	ts.Tags.InsertTags(&attr)

	expectedAttr := pdata.NewAttributeMap()
	expectedAttr.InsertBool("key1", true)
	expectedAttr.InsertString("key2", "hi")
	expectedAttr.InsertDouble("key3", 123.123)
	expectedAttr.InsertInt("key4", 10)
	rand.Seed(123)
	expectedAttr.InsertString("key5", csvTags[rand.Intn(len(csvTags))])

	require.Equal(t, attr.Sort().AsRaw(), expectedAttr.Sort().AsRaw())
}

func TestTagSet_loadCsvTags(t *testing.T) {
	tests := []struct {
		name     string
		tagSet   TagSet
		expected TagMap
		error    bool
	}{
		{
			name: "valid csv files",
			tagSet: TagSet{
				Tags: TagMap{"version": "v35"},
				CsvTags: map[string]string{
					"color": "testdata/color_tags.csv",
					"shape": "testdata/shape_tags.csv",
				},
			},
			expected: TagMap{
				"version": "v35",
				"color":   []string{"blue", "red", "yellow", "purple", "pink", "black", "orange", "green", "brown"},
				"shape":   []string{"circle", "square", "pentagon", "rectangle", "triangle", "hexagon"},
			},
			error: false,
		},
		{
			name: "nonexistent csv path",
			tagSet: TagSet{
				CsvTags: map[string]string{"some_tag": "fake_path.csv"},
			},
			expected: TagMap{},
			error:    true,
		},
		{
			name: "csv file contains multiple columns",
			tagSet: TagSet{
				CsvTags: map[string]string{"color": "testdata/invalid_tags.csv"},
			},
			expected: TagMap{},
			error:    true,
		},
		{
			name: "csv file is empty",
			tagSet: TagSet{
				CsvTags: map[string]string{"another_tag": "testdata/empty_tags.csv"},
			},
			expected: TagMap{},
			error:    true,
		},
		{
			name: "csv tag already defined in config",
			tagSet: TagSet{
				Tags: TagMap{"color": "magenta"},
				CsvTags: map[string]string{
					"color": "testdata/color_tags.csv",
				},
			},
			expected: TagMap{"color": "magenta"},
			error:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tagSet.loadCsvTags()
			if err != nil && !tt.error {
				assert.Fail(t, fmt.Sprintf("did not expect validation error but got: %v", err))
			}
			if err == nil && tt.error {
				assert.Fail(t, "expected validation error")
			}
			require.Equal(t, tt.expected, tt.tagSet.Tags)
		})
	}
}
