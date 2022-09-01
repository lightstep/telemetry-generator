package topology

import (
	"encoding/csv"
	"fmt"
	"github.com/lightstep/demo-environment/generatorreceiver/internal/flags"
	"os"
)

type TagSet struct {
	Weight              float64           `json:"weight" yaml:"weight"`
	Tags                TagMap            `json:"tags,omitempty" yaml:"tags,omitempty"`
	CsvTags             map[string]string `json:"csv_tags,omitempty" yaml:"csv_tags,omitempty"`
	TagGenerators       []TagGenerator    `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit             []string          `json:"inherit,omitempty" yaml:"inherit,omitempty"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func (ts *TagSet) LoadFromCSV() error {
	for name, path := range ts.CsvTags {
		if ts.Tags[name] != nil { //if tag name already exists
			return fmt.Errorf("csv tag %s was already defined in the yaml", name)
		}
		ts.Tags[name] = getTagsFromCSV(path)
	}
	return nil
}

//figure out where to call this load function - in load tree & validation

func getTagsFromCSV(file string) []string {
	csvfile, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer csvfile.Close()

	csvReader := csv.NewReader(csvfile)
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil
	}

	tagValues := make([]string, 0, len(data))

	//loop through each slice and convert to string
	for _, tagValue := range data {
		str := tagValue[0]
		tagValues = append(tagValues, str)
	}
	fmt.Println(tagValues)

	return tagValues
}
