package main

import (
	"hash/fnv"
	"os"
	"strconv"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

// Shard represents sharding service.
// Panoptes shards devices for horizontal scaling and high availability.
type Shard struct {
	cfg               config.Config
	id                string
	logger            *zap.Logger
	discovery         discovery.Discovery
	telemetry         *telemetry.Telemetry
	numberOfNodes     int
	initializingShard int
	updateRequest     chan struct{}
}

// NewShard constructs a shard service.
func NewShard(cfg config.Config, telemetry *telemetry.Telemetry, discovery discovery.Discovery, updateRequest chan struct{}) *Shard {
	return &Shard{
		cfg:               cfg,
		discovery:         discovery,
		telemetry:         telemetry,
		updateRequest:     updateRequest,
		logger:            cfg.Logger(),
		numberOfNodes:     cfg.Global().Shard.NumberOfNodes,
		initializingShard: cfg.Global().Shard.InitializingShard,
	}

}

// Start runs sharding service.
// it watches other nodes through service discovery and it creates
// proper device filters then refresh the devices.
func (s *Shard) Start() {
	s.logger.Info("shard", zap.Int("configured.nodes", s.numberOfNodes))

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
			s.logger.Error("discovery shard failed", zap.Error(err))
			return
		}

		if !isAllNodesRunning(s.numberOfNodes, instances) {
			s.telemetry.AddFilterOpt("extraShard", extraShard(s.id, s.numberOfNodes, instances))
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
				s.logger.Error("discovery shard failed", zap.Error(err))
				continue
			}
			s.telemetry.AddFilterOpt("extraShard", extraShard(s.id, s.numberOfNodes, instances))
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

func extraShard(myID string, shardSize int, instances []discovery.Instance) telemetry.DeviceFilterOpt {
	var id int
	failed := []int{}
	mapIndex := make(map[int]int)
	nodes := make([]*discovery.Instance, shardSize)

	for _, instance := range instances {
		instance := instance
		if _, ok := instance.Meta["shard_enabled"]; !ok {
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

func (s *Shard) waitForDiscoveryRegister() {
	hostname, _ := os.Hostname()
	for i := 0; i < 15; i++ {
		instances, err := s.discovery.GetInstances()
		if err != nil {
			s.logger.Error("discovery shard failed", zap.Error(err))
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
func (s *Shard) waitForInitialShards() {
	for {
		time.Sleep(time.Second * 10)

		currentAvailableNodes := 0

		instances, err := s.discovery.GetInstances()
		if err != nil {
			s.logger.Error("shard", zap.String("event", "discovery shard failed"), zap.Error(err))
			continue
		}

		for _, instance := range instances {
			if _, ok := instance.Meta["shard_enabled"]; !ok {
				continue
			}
			if instance.Status == "passing" {
				currentAvailableNodes++
			}
		}

		if currentAvailableNodes >= s.initializingShard {
			s.logger.Info("shard", zap.String("event", "initialized"), zap.Int("available.nodes", currentAvailableNodes))
			break
		}
	}
}
