package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
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
		Meta:   map[string]string{"shard_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "1",
		Status: "critical",
		Meta:   map[string]string{"shard_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "2",
		Status: "passing",
		Meta:   map[string]string{"shard_enabled": "true"},
	})

	shardSize := 3

	r := []bool{false, false, true, false, true, false}
	f := mainShard("0", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard 0 failed")
	}

	r = []bool{true, false, false, true, false, false}
	f = extraShard("0", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShard 0 failed")
	}

	r = []bool{false, true, false, false, false, true}
	f = mainShard("2", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard 2 failed")
	}

	r = []bool{false, false, false, false, false, false}
	f = extraShard("2", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShard 2 failed")
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
		Meta:   map[string]string{"shard_enabled": "true"},
	})

	shardSize := 2
	r := []bool{false, true, false, false, true, false}
	f := mainShard("0", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard failed")
	}

	r = []bool{true, false, true, true, false, true}
	f = extraShard("0", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShard failed")
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
		Meta:   map[string]string{"shard_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "1",
		Status: "passing",
		Meta:   map[string]string{"shard_enabled": "true"},
	})
	instances = append(instances, discovery.Instance{
		ID:     "2",
		Status: "critical",
		Meta:   map[string]string{"shard_enabled": "true"},
	})

	shardSize := 3
	r := []bool{false, false, false, false, false, false}
	f := mainShard("1", shardSize)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "mainShard failed")
	}

	r = []bool{true, true, true, true, true, true}
	f = extraShard("1", shardSize, instances)
	for i, d := range devices {
		assert.Equal(t, r[i], f(d), "extraShard failed")
	}

}
