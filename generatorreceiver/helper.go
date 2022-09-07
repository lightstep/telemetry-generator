package generatorreceiver

import (
	"fmt"
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/topology"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

func hasAnySuffix(s string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}

	return false
}

func parseTopoFile(topoPath string) (*topology.File, error) {
	var topo topology.File
	topoFile, err := os.Open(topoPath)
	if err != nil {
		return nil, err
	}
	defer topoFile.Close()

	byteValue, _ := ioutil.ReadAll(topoFile)
	lowerTopoPath := strings.ToLower(topoPath)
	if hasAnySuffix(lowerTopoPath, []string{".yaml", ".yml"}) {
		err = yaml.Unmarshal(byteValue, &topo)
	} else {
		err = fmt.Errorf("Unrecognized topology file type: %s", topoPath)
	}

	if err != nil {
		return nil, err
	}
	return &topo, nil
}
