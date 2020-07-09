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

func (s *Shard) Start() {
	s.logger.Info("sharding has been started", zap.Int("configured.nodes", s.numberOfNodes))

	// discovery
	notifyChan := make(chan struct{}, 1)
	s.waitForDiscoveryRegister()
	go s.discovery.Watch("panoptes", notifyChan)

	s.waitForInitialShards()

	s.telemetry.AddFilterOpt("mainShard", mainShard(s.id, s.numberOfNodes))
	s.updateRequest <- struct{}{}

	// takeover if all nodes are not available
	go func() {
		<-time.After(time.Second * 35)
		if !isAllNodesRunning(s.numberOfNodes, s.discovery.GetInstances()) {
			s.telemetry.AddFilterOpt("extraShard", extraShard(s.id, s.numberOfNodes, s.discovery.GetInstances()))
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
			s.telemetry.AddFilterOpt("extraShard", extraShard(s.id, s.numberOfNodes, s.discovery.GetInstances()))
			s.updateRequest <- struct{}{}
		}

	}

}

func mainShard(myID string, shardSize int) telemetry.DeviceFilterOpt {
	whoami, _ := strconv.Atoi(myID)
	return func(d config.Device) bool {
		if getGroupID(d)%shardSize == whoami {
			return true
		}

		return false
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

	if passing == shardSize {
		return true
	}

	return false
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
		for _, instance := range s.discovery.GetInstances() {
			if instance.Address == hostname && instance.Status == "passing" {
				s.id = instance.ID
				return
			}
		}

		time.Sleep(time.Second * 1)
	}

	panic("discovery registeration failed")
}

// waits for the initial shards to appear
func (s *Shard) waitForInitialShards() {
	for {
		currentAvailableNodes := 0
		for _, instance := range s.discovery.GetInstances() {
			if _, ok := instance.Meta["shard_enabled"]; !ok {
				continue
			}
			if instance.Status == "passing" {
				currentAvailableNodes++
			}
		}

		if currentAvailableNodes >= s.initializingShard {
			s.logger.Info("initializing shards", zap.Int("available.nodes", currentAvailableNodes))
			break
		}

		time.Sleep(time.Second * 10)
	}
}
