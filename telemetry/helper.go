package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
)

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

func GetGNMISubscriptions(sensors []*config.Sensor) []*gpb.Subscription {
	var subscriptions []*gpb.Subscription

	for _, sensor := range sensors {
		path, _ := ygot.StringToPath(sensor.Path, ygot.StructuredPath, ygot.StringSlicePath)

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

func SanitizePath(path string) string {
	if !strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s/", path)
	}

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	return path
}

func MergeLabels(keyLabels, prefixLabels map[string]string, prefix string) map[string]string {
	if len(keyLabels) > 0 {
		for k, v := range prefixLabels {
			if _, ok := keyLabels[k]; ok {
				keyLabels[prefix+k] = v
			} else {
				keyLabels[k] = v
			}
		}
		return keyLabels
	}

	return prefixLabels
}
