package generatorreceiver

import (
	"encoding/json"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
	"io/ioutil"
	"os"
)

func parseTopoFile(topoPath string) (*topology.File, error){
	var topoFile topology.File
	jsonFile, err := os.Open(topoPath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &topoFile)
	if err != nil {
		return nil, err
	}
	return &topoFile, nil
}