package hashmap

import (
	"hash/fnv"
	"sync"
)

type Map struct {
	Data     []*MapShard
	shardNum int
}

type MapShard struct {
	Data map[string]interface{}
	sync.RWMutex
}

func New(shardNum int) *Map {
	m := new(Map)
	m.Data = make([]*MapShard, shardNum)
	for i := 0; i < shardNum; i++ {
		m.Data[i] = &MapShard{Data: make(map[string]interface{})}
	}

	return m
}

func (m *Map) getShard(key string) uint {
	hash := fnv.New32()
	hash.Write([]byte(key))
	hSum32 := hash.Sum32()
	return uint(hSum32) % uint(m.shardNum)
}

func (m *Map) Set(key string, value interface{}) {
	shard := m.getShard(key)
	m.Data[shard].Lock()
	defer m.Data[shard].Unlock()
	m.Data[shard].Data[key] = value
}

func (m *Map) Get(key string) interface{} {
	shard := m.getShard(key)
	m.Data[shard].RLock()
	defer m.Data[shard].RUnlock()
	return m.Data[shard].Data
}
