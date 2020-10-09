//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package mdt

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/mock"
)

func TestHandler(t *testing.T) {
	buf := new(bytes.Buffer)
	ch := make(telemetry.ExtDSChan, 10)

	m := &MDT{
		pathOutput: map[string]string{"Sub3": "console::stdout"},
		outChan:    ch,
		systemID:   "127.0.0.1",
	}

	tm := mock.MDTInterfaceII()
	m.handler(buf, tm)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tt := map[string]struct {
		key       string
		timestamp uint64
		labels    map[string]string
		value     interface{}
	}{
		"GigabitEthernet0/0/0/0state/counters/in-octets": {
			key:       "state/counters/in-octets",
			timestamp: 1597098791076,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/0", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(1023),
		},
		"GigabitEthernet0/0/0/0state/counters/out-octets": {
			key:       "state/counters/out-octets",
			timestamp: 1597098791076,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/0", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(872),
		},
		"GigabitEthernet0/0/0/1state/counters/in-octets": {
			key:       "state/counters/in-octets",
			timestamp: 1597098791086,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/1", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(1223),
		},
		"GigabitEthernet0/0/0/1state/counters/out-octets": {
			key:       "state/counters/out-octets",
			timestamp: 1597098791086,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/1", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(8172),
		},
	}

	for i := 0; i < 4; i++ {
		select {
		case r := <-ch:
			exp := tt[r.DS["labels"].(map[string]string)["name"]+r.DS["key"].(string)]

			assert.Equal(t, "console::stdout", r.Output)
			assert.Equal(t, exp.value, r.DS["value"])
			assert.Equal(t, exp.labels, r.DS["labels"])
			assert.Equal(t, "127.0.0.1", r.DS["system_id"])
			assert.Equal(t, exp.timestamp, r.DS["timestamp"])

		case <-ctx.Done():
			assert.Fail(t, "deadline exceeded")
		}

	}
}
