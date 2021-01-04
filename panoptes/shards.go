//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"hash/fnv"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/discovery"
	"github.com/yahoo/panoptes-stream/telemetry"
)

// Shards represents sharding service.
// Panoptes shards devices for horizontal scaling and high availability.
type Shards struct {
	cfg                config.Config
	id                 string
	logger             *zap.Logger
	discovery          discovery.Discovery
	telemetry          *telemetry.Telemetry
	numberOfNodes      int
	initializingShards int
	minimumShards      int
	isSuspension       bool
	updateRequest      chan struct{}
}

// NewShards constructs a shard service.
func NewShards(cfg config.Config, telemetry *telemetry.Telemetry, discovery discovery.Discovery, updateRequest chan struct{}) *Shards {
	return &Shards{
		cfg:                cfg,
		discovery:          discovery,
		telemetry:          telemetry,
		updateRequest:      updateRequest,
		logger:             cfg.Logger(),
		numberOfNodes:      cfg.Global().Shards.NumberOfNodes,
		initializingShards: cfg.Global().Shards.InitializingShards,
		minimumShards:      cfg.Global().Shards.MinimumShards,
	}

}

// Start runs sharding service.
// it watches other nodes through service discovery and it creates
// proper device filters then refresh the devices.
func (s *Shards) Start() {
	s.logger.Info("shards", zap.Int("configured.nodes", s.numberOfNodes))

	// discovery
	notifyChan := make(chan struct{}, 1)
	s.waitForDiscoveryRegister()
	go s.discovery.Watch(notifyChan)

	s.waitForInitialShards()

	// in case of recovery it has to wait until other nodes
	// disconnected from targets that this node is responsible.
	time.Sleep(35 * time.Second)

	s.telemetry.AddFilterOpt("mainShard", mainShard(s.id, s.numberOfNodes))
	s.updateRequest <- struct{}{}

	// takeover if all nodes are not available
	go func() {
		<-time.After(time.Second * 35)
		instances, err := s.discovery.GetInstances()
		if err != nil {
			s.logger.Error("discovery shards failed", zap.Error(err))
			return
		}

		if !isAllNodesRunning(s.numberOfNodes, instances) &&
			availableShards(instances) >= s.minimumShards {
			s.telemetry.AddFilterOpt("extraShard", extraShards(s.id, s.numberOfNodes, instances))
			s.updateRequest <- struct{}{}
		}
	}()

	serviceChanged := false
	for {
		select {
		case <-notifyChan:
			serviceChanged = true
		case <-time.After(time.Second * 30):
			if !serviceChanged {
				continue
			}

			serviceChanged = false
			instances, err := s.discovery.GetInstances()
			if err != nil {
				s.logger.Error("discovery shards failed", zap.Error(err))
				continue
			}

			if availableShards(instances) >= s.minimumShards {
				s.logger.Info("shards", zap.String("event", "shards has been changed"))
				s.telemetry.AddFilterOpt("extraShard", extraShards(s.id, s.numberOfNodes, instances))

				if s.isSuspension {
					s.unsuspend()
				}
			} else {
				s.suspend()
			}

			s.updateRequest <- struct{}{}
		}

	}

}

func mainShard(myID string, shardSize int) telemetry.DeviceFilterOpt {
	whoami, _ := strconv.Atoi(myID)
	return func(d config.Device) bool {
		return getGroupID(d)%shardSize == whoami
	}
}

func extraShards(myID string, shardSize int, instances []discovery.Instance) telemetry.DeviceFilterOpt {
	var id int
	failed := []int{}
	mapIndex := make(map[int]int)
	nodes := make([]*discovery.Instance, shardSize)

	for _, instance := range instances {
		instance := instance
		if _, ok := instance.Meta["shards_enabled"]; !ok {
			continue
		}

		id, _ = strconv.Atoi(instance.ID)
		nodes[id] = &instance
	}

	for i, instance := range nodes {
		// haven't started yet
		if instance == nil {
			failed = append(failed, i)
			continue
		}

		if instance.Status == "passing" {
			id, _ = strconv.Atoi(instance.ID)
			mapIndex[id] = id - len(failed)
			continue
		}

		failed = append(failed, i)
	}

	whoami, _ := strconv.Atoi(myID)

	return func(d config.Device) bool {
		for _, j := range failed {
			i := getGroupID(d)
			if i%shardSize == j {
				if i%(shardSize-len(failed)) == mapIndex[whoami] {
					return true
				}
			}
		}
		return false
	}
}

func isAllNodesRunning(shardSize int, instances []discovery.Instance) bool {
	var passing int
	for _, instance := range instances {
		if instance.Status == "passing" {
			passing++
		}
	}

	return passing == shardSize
}

func getGroupID(d config.Device) int {
	if d.GroupID != 0 {
		return d.GroupID
	}

	return getHash(d.Host)
}

func getHash(key string) int {
	hash := fnv.New32()
	hash.Write([]byte(key))
	hSum32 := hash.Sum32()
	return int(hSum32)
}

func (s *Shards) waitForDiscoveryRegister() {
	hostname, _ := os.Hostname()
	for i := 0; i < 15; i++ {
		instances, err := s.discovery.GetInstances()
		if err != nil {
			s.logger.Error("discovery shards failed", zap.Error(err))
			continue
		}

		for _, instance := range instances {
			if instance.Address == hostname && instance.Status == "passing" {
				s.id = instance.ID
				return
			}
		}

		time.Sleep(time.Second * 1)
	}

	panic("discovery registration failed")
}

// waitForInitialShards waits for the configured initial shards to appear.
func (s *Shards) waitForInitialShards() {
	for {
		time.Sleep(time.Second * 10)

		currentAvailableNodes := 0

		instances, err := s.discovery.GetInstances()
		if err != nil {
			s.logger.Error("shards", zap.String("event", "discovery shards failed"), zap.Error(err))
			continue
		}

		for _, instance := range instances {
			if _, ok := instance.Meta["shards_enabled"]; !ok {
				continue
			}
			if instance.Status == "passing" {
				currentAvailableNodes++
			}
		}

		if currentAvailableNodes >= s.initializingShards {
			s.logger.Info("shards", zap.String("event", "initialized"), zap.Int("available.nodes", currentAvailableNodes))
			break
		}

		s.logger.Info("shards", zap.String("event", "wait.initial"), zap.Int("available.nodes", currentAvailableNodes))
	}
}

func (s *Shards) suspend() {
	s.telemetry.DelFilterOpt("extraShard")
	s.telemetry.DelFilterOpt("mainShard")
	s.isSuspension = true

	s.logger.Warn("shards", zap.String("event", "node has been suspended"))
}

func (s *Shards) unsuspend() {
	s.telemetry.AddFilterOpt("mainShard", mainShard(s.id, s.numberOfNodes))
	s.isSuspension = false

	s.logger.Warn("shards", zap.String("event", "node has been unsuspended"))
}

func availableShards(instances []discovery.Instance) int {
	var available int

	for _, instance := range instances {
		if _, ok := instance.Meta["shards_enabled"]; !ok {
			continue
		}
		if instance.Status == "passing" {
			available++
		}
	}

	return available
}
