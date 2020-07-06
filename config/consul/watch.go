package consul

import (
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
)

func (c *consul) watchKey(key string) {
	client := c.client.KV()
	opts := api.QueryOptions{}

	_, meta, _ := client.Get(key, &opts)

	for {
		opts.WaitIndex = meta.LastIndex
		_, meta, _ = client.Get(key, &opts)

		select {
		case c.informer <- struct{}{}:
		default:
		}

		if opts.WaitIndex != meta.LastIndex {
			c.logger.Info("modified " + key)
		}

	}
}

func (c *consul) watch(prefix string) {
	params := make(map[string]interface{})
	params["type"] = "keyprefix"
	params["prefix"] = prefix

	wp, err := watch.Parse(params)
	if err != nil {
		panic(err)
	}

	lastIdx := uint64(0)
	wp.Handler = func(idx uint64, data interface{}) {
		if lastIdx != 0 {
			c.logger.Info("watcher: keyprefix modified", zap.String("name", prefix))
		}
		lastIdx = idx
	}

	if err := wp.Run("localhost:8500"); err != nil {
		panic(err)
	}
}
