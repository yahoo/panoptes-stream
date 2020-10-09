//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/yahoo/panoptes-stream/config"
)

// GetKey returns telemetry key and extracted labels.
func GetKey(buf *bytes.Buffer, path []*gpb.PathElem) (string, map[string]string) {
	labels := make(map[string]string)

	for _, elem := range path {
		if len(elem.Name) > 0 {
			buf.WriteRune('/')
			buf.WriteString(elem.Name)
		}

		for key, value := range elem.Key {
			if _, ok := labels[key]; ok {
				labels[buf.String()+"/"+key] = value
			} else {
				labels[key] = value
			}
		}
	}

	buf.ReadRune()

	return buf.String(), labels
}

// GetValue returns telemetry value.
func GetValue(tv *gpb.TypedValue) (interface{}, error) {
	var (
		jsondata []byte
		value    interface{}
		err      error
	)

	switch tv.Value.(type) {
	case *gpb.TypedValue_AsciiVal:
		value = tv.GetAsciiVal()
	case *gpb.TypedValue_BoolVal:
		value = tv.GetBoolVal()
	case *gpb.TypedValue_BytesVal:
		value = tv.GetBytesVal()
	case *gpb.TypedValue_DecimalVal:
		value = float64(tv.GetDecimalVal().Digits) / math.Pow(10, float64(tv.GetDecimalVal().Precision))
	case *gpb.TypedValue_FloatVal:
		value = tv.GetFloatVal()
	case *gpb.TypedValue_IntVal:
		value = tv.GetIntVal()
	case *gpb.TypedValue_StringVal:
		value = tv.GetStringVal()
	case *gpb.TypedValue_UintVal:
		value = tv.GetUintVal()
	case *gpb.TypedValue_JsonIetfVal:
		jsondata = tv.GetJsonIetfVal()
	case *gpb.TypedValue_JsonVal:
		jsondata = tv.GetJsonVal()
	case *gpb.TypedValue_LeaflistVal:
		elems := tv.GetLeaflistVal().GetElement()
		value, err = getLeafList(elems)
	default:
		err = fmt.Errorf("unknown value type %+v", tv.Value)
	}

	if jsondata != nil {
		err = json.Unmarshal(jsondata, &value)
	}

	return value, err
}

// GetGNMISubscriptions return gNMI subscription based on the sensors.
func GetGNMISubscriptions(sensors []*config.Sensor) []*gpb.Subscription {
	var subscriptions []*gpb.Subscription

	for _, sensor := range sensors {
		path, _ := ygot.StringToPath(sensor.Path, ygot.StructuredPath, ygot.StringSlicePath)
		path.Origin = sensor.Origin

		mode := gpb.SubscriptionMode_value[strings.ToUpper(sensor.Mode)]
		sampleInterval := time.Duration(sensor.SampleInterval) * time.Second
		heartbeatInterval := time.Duration(sensor.HeartbeatInterval) * time.Second
		subscriptions = append(subscriptions, &gpb.Subscription{
			Path:              path,
			Mode:              gpb.SubscriptionMode(mode),
			SampleInterval:    uint64(sampleInterval.Nanoseconds()),
			HeartbeatInterval: uint64(heartbeatInterval.Nanoseconds()),
			SuppressRedundant: sensor.SuppressRedundant,
		})

	}

	return subscriptions
}

// GetPathOutput returns path to output map.
func GetPathOutput(sensors []*config.Sensor) map[string]string {
	var pathOutput = make(map[string]string)

	for _, sensor := range sensors {
		if strings.HasSuffix(sensor.Path, "/") {
			pathOutput[sensor.Path] = sensor.Output
		} else {
			pathOutput[fmt.Sprintf("%s/", sensor.Path)] = sensor.Output
		}
	}

	return pathOutput
}

func getLeafList(elems []*gpb.TypedValue) (interface{}, error) {
	list := []interface{}{}
	for _, v := range elems {
		ev, err := GetValue(v)
		if err != nil {
			return nil, fmt.Errorf("leaflist error: %v", err)
		}
		list = append(list, ev)
	}

	return list, nil
}

// MergeLabels merges key labels with prefix labels.
func MergeLabels(keyLabels, prefixLabels map[string]string, prefix string) map[string]string {
	if len(keyLabels) > 0 {
		for k, v := range prefixLabels {
			if _, ok := keyLabels[k]; ok {
				keyLabels[prefix+"/"+k] = v
			} else {
				keyLabels[k] = v
			}
		}
		return keyLabels
	}

	return prefixLabels
}

// getPathWOKey returns path string without key/value.
func getPathWithoutKey(path string) string {
	var buf bytes.Buffer

	p, _ := ygot.StringToPath(path, ygot.StructuredPath, ygot.StringSlicePath)

	for _, elem := range p.Elem {
		if len(elem.Name) > 0 {
			buf.WriteRune('/')
			buf.WriteString(elem.Name)
		}

	}

	return buf.String()
}

// getSensors splits sensors if they have overlap with each other.
// arista.gnmi and cisco.gnmi can not distinguish between overlapped
// sensors once the metrics returned from devices (multi path use case)
// the only way to distinguish them is split them to different grpc connections.
func getSensors(deviceSensors map[string][]*config.Sensor) map[string][]*config.Sensor {
	var (
		i        = 0
		rSensors = map[string][]*config.Sensor{}
	)

	for service, sensors := range deviceSensors {
		paths := map[string]bool{}

		if service == "arista.gnmi" || service == "cisco.gnmi" {
			for _, sensor := range sensors {
				ps := getPathWithoutKey(sensor.Path)
				if _, ok := paths[ps]; ok {
					serviceName := fmt.Sprintf("%s::ext%d", service, i)
					rSensors[serviceName] = append(rSensors[serviceName], sensor)
					i++
				} else {
					rSensors[service] = append(rSensors[service], sensor)
				}

				paths[ps] = true
			}
		} else {
			rSensors[service] = append(rSensors[service], sensors...)
		}
	}

	return rSensors
}

// GetDefaultOutput returns default output if available.
func GetDefaultOutput(sensors []*config.Sensor) string {
	var output = make(map[string]bool)

	for _, sensor := range sensors {
		output[sensor.Output] = true
	}

	if len(output) == 1 {
		for k := range output {
			return k
		}
	}

	return ""
}
