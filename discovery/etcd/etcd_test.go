package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/integration"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestNewEtcdAndRegistration(t *testing.T) {
	cluster := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	defer cluster.Terminate(t)

	time.Sleep(time.Second * 2)

	client := cluster.RandClient()

	cfg := &config.MockConfig{
		MGlobal: &config.Global{
			Discovery: config.Discovery{
				Config: map[string]interface{}{
					"endpoints": []string{cluster.RandClient().Endpoints()[0]},
					"prefix":    "/panoptes/",
				},
			},
		},
	}

	disc, err := New(cfg)
	assert.Equal(t, nil, err)

	err = disc.Register()
	assert.Equal(t, nil, err)

	// check the registered id
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	kv := clientv3.NewKV(client)
	resp, err := kv.Get(ctx, "/panoptes/services/", clientv3.WithPrefix())
	assert.Len(t, resp.Kvs, 1)
	assert.Equal(t, "/panoptes/services/0", string(resp.Kvs[0].Key))

	// register second panoptes
	disc2 := &Etcd{
		cfg:    cfg,
		prefix: "/panoptes/",
		logger: cfg.Logger(),
	}
	disc2.client, err = clientv3.New(clientv3.Config{Endpoints: []string{cluster.RandClient().Endpoints()[0]}})
	assert.Equal(t, nil, err)
	err = disc2.register("1", "127.0.0.2", nil)
	assert.Equal(t, nil, err)

	insts, err := disc.GetInstances()
	assert.Equal(t, nil, err)
	assert.Len(t, insts, 2)

	// reregister
	err = disc.Register()
	assert.Equal(t, nil, err)

	// check again the number of instances
	insts, err = disc.GetInstances()
	assert.Equal(t, nil, err)
	assert.Len(t, insts, 2)
}

func TestGetID(t *testing.T) {
	assert.Equal(t, "1", getID([]int{0, 3}))
	assert.Equal(t, "2", getID([]int{0, 1}))
}
