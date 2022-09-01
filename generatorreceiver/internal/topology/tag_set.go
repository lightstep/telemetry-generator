package topology

import (
	"encoding/csv"
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

func (ts *TagSet) Load() error {
	for name, path := range ts.CsvTags {
		ts.Tags[name] = readFile(path)
		//validate that name isn't already in Tags
	}
	//if any of the files its looking for is not there, it should also return an error
	//using fmt: https://stackoverflow.com/questions/37194739/how-check-whether-a-file-contains-a-string-or-not
	//look up the name in the map and if it's already there, could probably just overwrite

	return nil
}

//figure out where to call this load function - in load tree & validation

func readFile(file string) [][]string {
	csvfile, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer csvfile.Close()

	//read lines from file
	csvReader := csv.NewReader(csvfile)
	data, err := csvReader.ReadAll()
	//if err != nil { //add the log library? }

	//helper function? - make a slice for strings, one for each line
	//strip the whitespace/all the formatting

	return data
}
