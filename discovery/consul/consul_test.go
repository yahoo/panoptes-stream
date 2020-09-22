//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package consul

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

var srv *testutil.TestServer

func TestConsul(t *testing.T) {
	var err error
	srv, err = testutil.NewTestServerConfigT(t, nil)
	assert.Equal(t, nil, err)
	defer srv.Stop()

	time.Sleep(2 * time.Second)

	t.Run("testNewConsulAndRegistration", testNewConsulAndRegistration)
	t.Run("testSecondNodes", testSecondNodes)
	t.Run("testWatch", testWatch)
}

func testNewConsulAndRegistration(t *testing.T) {
	//t.Parallel()

	// srv, err := testutil.NewTestServerConfigT(t, nil)
	// assert.Equal(t, nil, err)
	// defer srv.Stop()

	// time.Sleep(time.Second * 2)

	cfg := config.NewMockConfig()
	cfg.MGlobal = &config.Global{
		Discovery: config.Discovery{
			Config: map[string]interface{}{
				"address": srv.HTTPAddr,
				"prefix":  "",
			},
		},
	}

	disc, err := New(cfg)
	assert.Equal(t, nil, err)

	err = disc.Register()
	assert.Equal(t, nil, err)
	instances, err := disc.GetInstances()
	assert.Equal(t, nil, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "0", instances[0].ID)
}

func testSecondNodes(t *testing.T) {
	//t.Parallel()

	// srv, err := testutil.NewTestServerConfigT(t, nil)
	// assert.Equal(t, nil, err)
	// defer srv.Stop()

	// time.Sleep(time.Second * 2)

	cfg := config.NewMockConfig()
	cfg.MGlobal = &config.Global{
		Shards: config.Shards{},
	}
	apiConfig := api.DefaultConfig()
	apiConfig.Address = srv.HTTPAddr
	client, err := api.NewClient(apiConfig)
	assert.Equal(t, nil, err)

	c := &consul{
		client: client,
		cfg:    cfg,
		config: &consulConfig{},
		logger: cfg.Logger(),
	}

	c.register("0", "1.0.0.1", map[string]string{})
	c.register("1", hostname(), map[string]string{})

	err = c.Register()
	assert.Equal(t, nil, err)

	instances, err := c.GetInstances()
	assert.Equal(t, nil, err)
	assert.Len(t, instances, 2)
}

func testWatch(t *testing.T) {
	//t.Parallel()

	// srv, err := testutil.NewTestServerConfigT(t, nil)
	// assert.Equal(t, nil, err)
	// defer srv.Stop()

	// time.Sleep(time.Second * 2)

	cfg := config.NewMockConfig()
	cfg.MGlobal = &config.Global{
		Discovery: config.Discovery{
			Config: map[string]interface{}{
				"address": srv.HTTPAddr,
				"prefix":  "",
			},
		},
	}

	disc, err := New(cfg)
	assert.Equal(t, nil, err)

	ch := make(chan struct{}, 1)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go disc.Watch(ch)

	time.Sleep(time.Microsecond * 300)
	srv.AddService(t, "panoptes", "passing", nil)

	select {
	case <-ch:
	case <-ctx.Done():
		assert.Fail(t, "context deadline exceeded")
	}

}

func TestGetID(t *testing.T) {
	assert.Equal(t, "1", getID([]int{0, 3}))
	assert.Equal(t, "2", getID([]int{0, 1}))
}
