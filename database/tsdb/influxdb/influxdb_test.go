package influxdb

import (
	"testing"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestLineProtocol(t *testing.T) {
	data := telemetry.ExtDataStore{
		Output: "influx1::ifcounters",
		DS: telemetry.DataStore{
			"key":       "out-octets",
			"labels":    map[string]string{"name": "Ethernet3"},
			"prefix":    "/interfaces/interface/state/counters/",
			"system_id": "core1.bur",
			"timestamp": 1595768623436661269,
			"value":     5587651,
		},
	}

	l, err := getLineProtocol(data)
	require.Equal(t, err, nil)
	assert.Equal(t, l, "ifcounters,prefix=/interfaces/interface/state/counters/,system_id=core1.bur,name=Ethernet3 out-octets=5587651 1595768623436661269")
}

func BenchmarkLineProtocol(b *testing.B) {
	data := telemetry.ExtDataStore{
		Output: "influx1::ifcounters",
		DS: telemetry.DataStore{
			"key":       "out-octets",
			"labels":    map[string]string{"name": "Ethernet3"},
			"prefix":    "/interfaces/interface/state/counters/",
			"system_id": "core1.bur",
			"timestamp": 1595768623436661269,
			"value":     5587651,
		},
	}

	for i := 0; i < b.N; i++ {
		getLineProtocol(data)
	}
}
