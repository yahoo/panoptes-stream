package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestGetPathOutput(t *testing.T) {
	sensors := []*config.Sensor{
		{
			Path:   "/tests/test/",
			Output: "console::stdout",
		},
		{
			Path:   "/interfaces/interface",
			Output: "console::stderr",
		},
	}

	po := GetPathOutput(sensors)
	assert.Equal(t, "console::stdout", po["/tests/test/"])
	assert.Equal(t, "console::stderr", po["/interfaces/interface/"])
}
