package topology

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/flags"
)

type TagSet struct {
	Tags                TagMap            `json:"tags,omitempty" yaml:"tags,omitempty"`
	TagGenerators       []TagGenerator    `json:"tagGenerators,omitempty" yaml:"tagGenerators,omitempty"`
	Inherit             []string          `json:"inherit,omitempty" yaml:"inherit,omitempty"`
	CsvTags             map[string]string `json:"csv_tags,omitempty" yaml:"csv_tags,omitempty"`
	EmbeddedWeight      `json:",inline" yaml:",inline"`
	flags.EmbeddedFlags `json:",inline" yaml:",inline"`
}

func (ts *TagSet) loadCsvTags() error {
	if ts.Tags == nil && ts.CsvTags != nil {
		ts.Tags = make(TagMap)
	}

	for name, path := range ts.CsvTags {
		if ts.Tags[name] != nil {
			return fmt.Errorf("tag %s in csv file %s was already defined in yaml", name, path)
		}
		tags, err := readCsv(path)
		if err != nil {
			return err
		}
		ts.Tags[name] = tags
	}
	return nil
}

func readCsv(file string) ([]string, error) {
	csvFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	data, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("csv file %s cannot be empty", file)
	}

	tags := make([]string, 0, len(data))
	for _, tag := range data {
		if len(tag) != 1 {
			return nil, fmt.Errorf("each row in csv file %s must contain exactly one string", file)
		}
		str := tag[0]
		tags = append(tags, str)
	}

	return tags, nil
}
