//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package telemetry

import (
	"bytes"
	"testing"

	"github.com/openconfig/gnmi/proto/gnmi"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/config"
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

func TestGetKey(t *testing.T) {
	buf := new(bytes.Buffer)
	path := &gnmi.Path{
		Elem: []*gnmi.PathElem{
			{Name: "interfaces"},
			{Name: "interface", Key: map[string]string{"name": "Ethernet1"}},
			{Name: "state"},
			{Name: "counters"},
			{Name: "out-octets"},
		},
	}
	key, labels := GetKey(buf, path.Elem)
	assert.Equal(t, "interfaces/interface/state/counters/out-octets", key)
	assert.Equal(t, map[string]string{"name": "Ethernet1"}, labels)

	path = &gnmi.Path{
		Elem: []*gnmi.PathElem{
			{Name: "interfaces"},
			{Name: "interface", Key: map[string]string{"name": "Ethernet1"}},
			{Name: "state", Key: map[string]string{"name": "test"}},
			{Name: "counters"},
			{Name: "out-octets"},
		},
	}

	buf.Reset()
	key, labels = GetKey(buf, path.Elem)
	assert.Equal(t, "interfaces/interface/state/counters/out-octets", key)
	assert.Equal(t, map[string]string{"name": "Ethernet1", "/interfaces/interface/state/name": "test"}, labels)
}

func TestGetDefaultOutput(t *testing.T) {
	sensors := []*config.Sensor{
		{
			Output: "kafka1::mytopic",
		},
		{
			Output: "kafka1::mytopic",
		},
	}
	assert.Equal(t, "kafka1::mytopic", GetDefaultOutput(sensors))

	sensors = []*config.Sensor{
		{
			Output: "kafka1::mytopic1",
		},
		{
			Output: "kafka1::mytopic2",
		},
	}
	assert.Equal(t, "", GetDefaultOutput(sensors))
}

func TestGetValue(t *testing.T) {
	tt := []struct {
		tv *gpb.TypedValue
		v  interface{}
	}{
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_BoolVal{BoolVal: true}},
			v:  true,
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_DecimalVal{
				DecimalVal: &gpb.Decimal64{Digits: 55},
			}},
			v: float64(55),
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_BytesVal{
				BytesVal: []byte("gnmi"),
			}},
			v: []byte("gnmi"),
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_IntVal{
				IntVal: 55,
			}},
			v: int64(55),
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{
				StringVal: "gnmi",
			}},
			v: "gnmi",
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_FloatVal{
				FloatVal: 0.55,
			}},
			v: float32(0.55),
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_AsciiVal{
				AsciiVal: "gnmi",
			}},
			v: "gnmi",
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{
				UintVal: 55,
			}},
			v: uint64(55),
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_JsonIetfVal{
				JsonIetfVal: []byte("{\"key\":\"value\"}"),
			}},
			v: map[string]interface{}{"key": "value"},
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_JsonVal{
				JsonVal: []byte("{\"key\":\"value\"}"),
			}},
			v: map[string]interface{}{"key": "value"},
		},
		{
			tv: &gpb.TypedValue{Value: &gpb.TypedValue_LeaflistVal{
				LeaflistVal: &gpb.ScalarArray{
					Element: []*gpb.TypedValue{{
						Value: &gpb.TypedValue_StringVal{
							StringVal: "gnmi",
						},
					}},
				},
			}},
			v: []interface{}([]interface{}{"gnmi"}),
		},
	}

	for _, row := range tt {
		v, err := GetValue(row.tv)
		assert.NoError(t, err)
		assert.Equal(t, row.v, v)
	}
}

func TestMergeLabel(t *testing.T) {
	kl := map[string]string{"a": "b", "c": "q"}
	pl := map[string]string{"a": "b"}
	exp := map[string]string{"/prefix/a": "b", "a": "b", "c": "q"}

	m := MergeLabels(kl, pl, "/prefix")

	assert.Equal(t, exp, m)
}
