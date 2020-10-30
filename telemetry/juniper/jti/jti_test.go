//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package jti

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/telemetry"
	jpb "github.com/yahoo/panoptes-stream/telemetry/juniper/proto/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/mock"
)

type Update struct{}

func (u *Update) Run(subReq *jpb.SubscriptionRequest, subServer jpb.OpenConfigTelemetry_TelemetrySubscribeServer) error {
	switch subReq.PathList[0].Path {
	case "/interfaces/interface[name='lo0']/state/counters/":
		return subServer.Send(mock.JuniperJTILo0InterfaceSample())
	case "/network-instances/network-instance/protocols/protocol/bgp/":
		return subServer.Send(mock.JuniperBGPSample())
	case "/mixes/mix[name='lo0']/state/":
		return subServer.Send(mock.JuniperJTIMix())
	}
	return nil
}

func TestWithJTIServer(t *testing.T) {
	var addr = "127.0.0.1:50500"

	ln, err := mock.StartJTIServer(addr, &Update{})
	assert.NoError(t, err)
	defer ln.Close()

	t.Run("JuniperLo0InterfaceSample", JuniperLo0InterfaceSample)
	t.Run("JuniperBGPSample", JuniperBGPSample)
	t.Run("JuniperMix", JuniperMix)

}

func JuniperLo0InterfaceSample(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 10)
		sensors []*config.Sensor
		labels  map[string]string
		prefix  string
	)

	cfg := config.NewMockConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "juniper.jti",
		Output:  "console::stdout",
		Path:    "/interfaces/interface[name='lo0']/state/counters/",
	})

	j := New(cfg.Logger(), conn, sensors, ch)
	j.Start(ctx)

	KV := mock.JuniperJTILo0InterfaceSample().Kv

	r := new(bytes.Buffer)
	w := new(bytes.Buffer)

	for _, metric := range KV {
		if metric.Key == "__prefix__" {
			labels, prefix = getLabels(r, w, getValue(metric).(string))
			if len(prefix) > 1 {
				prefix = prefix[:len(prefix)-1]
			}
			continue
		}

		if strings.HasPrefix(metric.Key, "__") {
			continue
		}

		select {
		case resp := <-ch:
			assert.Equal(t, metric.Key, resp.DS["key"])
			assert.Equal(t, getValue(metric), resp.DS["value"])
			assert.Equal(t, prefix, resp.DS["prefix"])
			assert.Equal(t, "/interfaces/interface", resp.DS["prefix"])
			assert.Equal(t, "core1.lax", resp.DS["system_id"])
			assert.Equal(t, uint64(1596067993610)*1000000, resp.DS["timestamp"])
			assert.Equal(t, labels, resp.DS["labels"])

		case <-ctx.Done():
			assert.Fail(t, "context deadline exceeded")
			return
		}
	}

	assert.Equal(t, "", cfg.LogOutput.String())
}

func JuniperMix(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 10)
		sensors []*config.Sensor
		labels  map[string]string
		prefix  string
	)

	cfg := config.NewMockConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "juniper.jti",
		Output:  "console::stdout",
		Path:    "/mixes/mix[name='lo0']/state/",
	})

	j := New(cfg.Logger(), conn, sensors, ch)
	j.Start(ctx)

	KV := mock.JuniperJTIMix().Kv

	r := new(bytes.Buffer)
	w := new(bytes.Buffer)

	for _, metric := range KV {
		if metric.Key == "__prefix__" {
			labels, prefix = getLabels(r, w, getValue(metric).(string))
			if len(prefix) > 1 {
				prefix = prefix[:len(prefix)-1]
			}
			continue
		}

		if strings.HasPrefix(metric.Key, "__") {
			continue
		}

		select {
		case resp := <-ch:
			keyLabels, key := getLabels(r, w, metric.Key)
			assert.Equal(t, key, resp.DS["key"])
			assert.Equal(t, prefix, resp.DS["prefix"])

			if len(keyLabels) > 0 {
				for k, v := range labels {
					keyLabels[k] = v
				}
			} else {
				keyLabels = labels
			}

			assert.Equal(t, keyLabels, resp.DS["labels"])

			if metric.Key == "state/counters/out-queue[queue-number=0]/pkts" {
				assert.Equal(t, 2, len(keyLabels))
			}

		case <-ctx.Done():
			assert.Fail(t, "context deadline exceeded")
			return
		}
	}

	assert.Equal(t, "", cfg.LogOutput.String())
}

func JuniperBGPSample(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 100)
		sensors []*config.Sensor
		prefix  string
		labels  map[string]string
	)

	cfg := config.NewMockConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "juniper.jti",
		Output:  "console::stdout",
		Path:    "/network-instances/network-instance/protocols/protocol/bgp/",
	})

	j := New(cfg.Logger(), conn, sensors, ch)
	j.Start(ctx)

	r := new(bytes.Buffer)
	w := new(bytes.Buffer)

	KV := mock.JuniperBGPSample().Kv

	for _, metric := range KV {
		if metric.Key == "__prefix__" {
			labels, prefix = getLabels(r, w, getValue(metric).(string))
			if len(prefix) > 1 {
				prefix = prefix[:len(prefix)-1]
			}
			continue
		}

		if strings.HasPrefix(metric.Key, "__") {
			continue
		}

		select {
		case resp := <-ch:
			assert.Equal(t, labels, resp.DS["labels"])
			assert.Equal(t, prefix, resp.DS["prefix"])
			assert.Equal(t, metric.Key, resp.DS["key"])
			assert.Equal(t, getValue(metric), resp.DS["value"])
			assert.Equal(t, "core1.lax", resp.DS["system_id"])
			assert.Equal(t, uint64(1596087032354*1000000), resp.DS["timestamp"])

		case <-ctx.Done():
			assert.Fail(t, "context deadline exceeded")
			return
		}
	}
}

func TestGetLabelsSingleKey(t *testing.T) {
	expectedLabels := map[string]string{"name": "mem-util-kernel-bytes-allocated"}
	r := new(bytes.Buffer)
	w := new(bytes.Buffer)
	labels, key := getLabels(r, w, "property[name='mem-util-kernel-bytes-allocated']/state/value")

	assert.Equal(t, "property/state/value", key)
	assert.Equal(t, expectedLabels, labels)
}

func TestGetLabelsMultiKeys(t *testing.T) {
	expectedLabels := map[string]string{"afi-safi-name": "IPV4_UNICAST", "instance-name": "master", "peer-group-name": "BUR"}
	r := new(bytes.Buffer)
	w := new(bytes.Buffer)
	labels, prefix := getLabels(r, w, "/network-instances/network-instance[instance-name='master']/protocols/protocol/bgp/peer-groups/peer-group[peer-group-name='BUR']/afi-safis/afi-safi[afi-safi-name='IPV4_UNICAST']/")

	assert.Equal(t, "/network-instances/network-instance/protocols/protocol/bgp/peer-groups/peer-group/afi-safis/afi-safi/", prefix)
	assert.Equal(t, expectedLabels, labels)
}

func TestGetLabelsWOKeys(t *testing.T) {
	r := new(bytes.Buffer)
	w := new(bytes.Buffer)
	labels, prefix := getLabels(r, w, "/network-instances/network-instance/")
	assert.Equal(t, "/network-instances/network-instance/", prefix)
	assert.Equal(t, map[string]string{}, labels)
}

func TestGetLabelsNumber(t *testing.T) {
	expectedLabels := map[string]string{"queue-number": "0"}
	r := new(bytes.Buffer)
	w := new(bytes.Buffer)
	labels, key := getLabels(r, w, "interfaces/interface/state/counters/out-queue[queue-number=0]/pkts")

	assert.Equal(t, "interfaces/interface/state/counters/out-queue/pkts", key)
	assert.Equal(t, expectedLabels, labels)
}

func TestGetValue(t *testing.T) {
	kv := jpb.KeyValue{Value: &jpb.KeyValue_DoubleValue{DoubleValue: 5.5}}
	assert.Equal(t, 5.5, getValue(&kv))
	kv = jpb.KeyValue{Value: &jpb.KeyValue_IntValue{IntValue: 5}}
	assert.Equal(t, int64(5), getValue(&kv))
	kv = jpb.KeyValue{Value: &jpb.KeyValue_SintValue{SintValue: 5}}
	assert.Equal(t, int64(5), getValue(&kv))
	kv = jpb.KeyValue{Value: &jpb.KeyValue_BytesValue{BytesValue: []byte("test")}}
	assert.Equal(t, []byte("test"), getValue(&kv))
}

func TestVersion(t *testing.T) {
	assert.Equal(t, jtiVersion, Version())
}

func BenchmarkGetLabels(b *testing.B) {
	r := new(bytes.Buffer)
	w := new(bytes.Buffer)

	for i := 0; i < b.N; i++ {
		getLabels(r, w, "/network-instances/network-instance[instance-name='master']/protocols/protocol/bgp/peer-groups/peer-group[peer-group-name='BUR']/afi-safis/afi-safi[afi-safi-name='IPV4_UNICAST']/")
	}

}

func BenchmarkGetLabelsWithoutGlobalBuffers(b *testing.B) {

	for i := 0; i < b.N; i++ {
		r := new(bytes.Buffer)
		w := new(bytes.Buffer)
		getLabels(r, w, "/network-instances/network-instance[instance-name='master']/protocols/protocol/bgp/peer-groups/peer-group[peer-group-name='BUR']/afi-safis/afi-safi[afi-safi-name='IPV4_UNICAST']/")
	}

}

func BenchmarkSensorRegex(b *testing.B) {
	regxPath := regexp.MustCompile(`/:(/.*/):`)
	for i := 0; i < b.N; i++ {
		regxPath.FindStringSubmatch("sensor_1038_4_1:/interfaces/interface[name='lo0']/:/interfaces/interface[name='lo0']/:mib2d")
	}
}
