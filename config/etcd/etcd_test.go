package etcd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/integration"
)

func TestNewEtcd(t *testing.T) {
	cluster := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer cluster.Terminate(t)

	client := cluster.RandClient()
	os.Setenv("PANOPTES_CONFIG_ETCD_ENDPOINTS", client.Endpoints()[0])
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	kv := clientv3.NewKV(client)
	kv.Put(ctx, "config/devices/core1.lax", `{"host": "core1.bur","port": 50051,"sensors" : ["sensor1"]}`)
	kv.Put(ctx, "config/sensors/sensor1", `{"service": "arista.gnmi","output":"console::stdout", "path": "/interfaces/", "mode": "sample"}`)
	kv.Put(ctx, "config/databases/db1", `{"service": "influxdb", "config": {"server": "https://localhost:8086"}}`)
	kv.Put(ctx, "config/producers/kafka1", `{"service": "kafka", "config" : {"brokers": ["127.0.0.1:9092"], "topics":["bgp"]}}`)
	kv.Put(ctx, "config/global", `{"logger": {"level":"info", "encoding": "console", "outputPaths": ["stdout"], "errorOutputPaths":["stderr"]}, "status": {"addr":"127.0.0.2:8081"}}`)

	cfg, err := New("")
	assert.Equal(t, nil, err)

	devices := cfg.Devices()
	databases := cfg.Databases()
	producers := cfg.Producers()

	assert.Len(t, devices, 1)
	assert.Len(t, databases, 1)
	assert.Len(t, producers, 2)

	assert.Equal(t, "core1.bur", devices[0].Host)
	assert.Equal(t, "influxdb", databases[0].Service)
	assert.Equal(t, "kafka", producers[0].Service)
	assert.Equal(t, "127.0.0.2:8081", cfg.Global().Status.Addr)
	assert.NotEqual(t, nil, cfg.Logger())

	// make sure watch is ready
	time.Sleep(time.Second)

	kv.Put(ctx, "config/databases/db2", `{"service": "influxdb", "config": {"server": "https://localhost:8086"}}`)

	select {
	case <-cfg.Informer():
	case <-ctx.Done():
		assert.Fail(t, "context deadline exceeded")
	}

	cfg.Update()
	databases = cfg.Databases()
	assert.Len(t, databases, 2)

	// invalid json data
	kv.Put(ctx, "config/devices/core1.lax", `"host": "core1.bur","port": 50051,"sensors" : ["sensor1"]}`)
	err = cfg.Update()
	assert.NotEqual(t, nil, err)
}

func TestEmptyConfig(t *testing.T) {
	cluster := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer cluster.Terminate(t)

	client := cluster.RandClient()
	os.Setenv("PANOPTES_CONFIG_ETCD_ENDPOINTS", client.Endpoints()[0])
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	kv := clientv3.NewKV(client)
	kv.Put(ctx, "config/", "")

	cfg, err := New("")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, cfg.Logger())
}
