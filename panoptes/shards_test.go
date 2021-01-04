//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/discovery"
	"github.com/yahoo/panoptes-stream/telemetry"
)

func TestShardThreeNodes(t *testing.T) {
	devices := []config.Device{
		{DeviceConfig: config.DeviceConfig{Host: "core1.lax"}}, // even
		{DeviceConfig: config.DeviceConfig{Host: "core1.bur"}}, // even
		{DeviceConfig: config.DeviceConfig{Host: "core1.cdg"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core2.cdg"}}, // even
		{DeviceConfig: config.DeviceConfig{Host: "core3.cdg"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core4.cdg"}}, // even
	}
	instances := []discovery.Instance{}
	instances = append(instances, discovery.Instance{
		ID:     "0",
		Status: "passing",
		Meta:   map[string]string{"shards_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "1",
		Status: "critical",
		Meta:   map[string]string{"shards_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "2",
		Status: "passing",
		Meta:   map[string]string{"shards_enabled": "true"},
	})

	shardSize := 3

	r := []bool{false, false, true, false, true, false}
	f := mainShard("0", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard 0 failed")
	}

	r = []bool{true, false, false, true, false, false}
	f = extraShards("0", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShards 0 failed")
	}

	r = []bool{false, true, false, false, false, true}
	f = mainShard("2", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard 2 failed")
	}

	r = []bool{false, false, false, false, false, false}
	f = extraShards("2", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShards 2 failed")
	}

}

func TestShard2(t *testing.T) {
	// one unavailable node (nil instance)
	devices := []config.Device{
		{DeviceConfig: config.DeviceConfig{Host: "core2.lax"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core1.bur"}}, // even
		{DeviceConfig: config.DeviceConfig{Host: "core1.cdg"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core1.dca"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core1.sea"}}, // even
		{DeviceConfig: config.DeviceConfig{Host: "core1.sjc"}}, // odd
	}

	instances := []discovery.Instance{}
	instances = append(instances, discovery.Instance{
		ID:     "0",
		Status: "passing",
		Meta:   map[string]string{"shards_enabled": "true"},
	})

	shardSize := 2
	r := []bool{false, true, false, false, true, false}
	f := mainShard("0", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard failed")
	}

	r = []bool{true, false, true, true, false, true}
	f = extraShards("0", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShards failed")
	}
}

func TestShard3(t *testing.T) {
	// two critical nodes of 3 (first and last)
	devices := []config.Device{
		{DeviceConfig: config.DeviceConfig{Host: "core2.lax"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core1.bur"}}, // even
		{DeviceConfig: config.DeviceConfig{Host: "core1.cdg"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core1.dca"}}, // odd
		{DeviceConfig: config.DeviceConfig{Host: "core1.sea"}}, // even
		{DeviceConfig: config.DeviceConfig{Host: "core1.sjc"}}, // odd
	}

	instances := []discovery.Instance{}
	instances = append(instances, discovery.Instance{
		ID:     "0",
		Status: "critical",
		Meta:   map[string]string{"shards_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "1",
		Status: "passing",
		Meta:   map[string]string{"shards_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "2",
		Status: "critical",
		Meta:   map[string]string{"shards_enabled": "true"},
	})

	shardSize := 3
	r := []bool{false, false, false, false, false, false}
	f := mainShard("1", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard failed")
	}

	r = []bool{true, true, true, true, true, true}
	f = extraShards("1", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShards failed")
	}

}

func TestAvailableShards(t *testing.T) {
	instances := []discovery.Instance{}
	instances = append(instances, discovery.Instance{
		ID:     "0",
		Status: "passing",
		Meta:   map[string]string{"shards_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "1",
		Status: "critical",
		Meta:   map[string]string{"shards_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "2",
		Status: "passing",
		Meta:   map[string]string{"shards_enabled": "true"},
	})

	n := availableShards(instances)
	assert.Equal(t, 2, n)
}

func TestSuspendUnSuspend(t *testing.T) {
	//cfg := config.NewMockConfig()
	devices := []config.Device{
		{
			DeviceConfig: config.DeviceConfig{
				Host: "foo02.bar",
			},
		},
	}
	cfg := &config.MockConfig{MDevices: devices}

	tm := telemetry.New(context.Background(), cfg, nil, nil)

	s := Shards{
		cfg:           cfg,
		id:            "0",
		logger:        cfg.Logger(),
		telemetry:     tm,
		numberOfNodes: 2,
		updateRequest: make(chan struct{}),
	}

	s.unsuspend()
	d := s.telemetry.GetDevices()
	assert.Len(t, d, 1)
	s.suspend()
	d = s.telemetry.GetDevices()
	assert.Len(t, d, 0)
}
