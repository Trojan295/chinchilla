package agent

import (
	"os"
	"testing"
)

func TestParseMeminfo(t *testing.T) {
	fp, err := os.Open("../mockdata/meminfo")
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	memStats := parseMeminfo(fp)

	if memStats.Available != 13122188 {
		t.Errorf("Available mem is %d, should be %d", memStats.Available, 13122188)
	}
	if memStats.Total != 16330360 {
		t.Errorf("Total mem is %d, should be %d", memStats.Total, 16330360)
	}
}
