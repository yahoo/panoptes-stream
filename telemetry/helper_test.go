//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

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

func TestGetSensors(t *testing.T) {
	s := []*config.Sensor{
		{
			Path:   "/interfaces/interface/state/counters",
			Output: "console::stdout",
		},
		{
			Path:   "/interfaces/interface[name=lo]/state/counters",
			Output: "console::stderr",
		},
		{
			Path:   "/interfaces/interface[name=Ethernet1]/state/counters",
			Output: "console::stderr",
		},
		{
			Path:   "/network-instances/network-instance",
			Output: "console::stderr",
		},
	}

	deviceSensors := map[string][]*config.Sensor{"arista.gnmi": s}
	newSensors := getSensors(deviceSensors)
	assert.Len(t, newSensors, 3)
	assert.Len(t, newSensors["arista.gnmi"], 2)
	assert.Contains(t, newSensors, "arista.gnmi::ext0")
	assert.Contains(t, newSensors, "arista.gnmi::ext1")
}
