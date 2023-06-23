package cmd

import (
	"fmt"
	"time"

	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib"
)

var packageRegistryBaseURL string
var configFile string
var totEvents uint64
var timeNowAsString string
var randSeed int64

func getTimeNowFromFlag(timeNowAsString string) (time.Time, error) {
	if len(timeNowAsString) > 0 {
		if timeNow, err := time.Parse(genlib.FieldTypeTimeLayout, timeNowAsString); err != nil {
			return timeNow, fmt.Errorf("wrong --now flag: %s (%w)", timeNowAsString, err)
		} else {
			return timeNow, nil
		}
	}

	return time.Now(), nil
}
