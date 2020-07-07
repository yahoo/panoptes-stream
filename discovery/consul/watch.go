package consul

import (
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
)

func (c *Consul) Watch(service string, ch chan<- struct{}) {
	params := make(map[string]interface{})
	params["type"] = "service"
	params["service"] = service

	wp, err := watch.Parse(params)
	if err != nil {
		panic(err)
	}

	lastIdx := uint64(0)
	wp.Handler = func(idx uint64, data interface{}) {
		if lastIdx != 0 {
			c.logger.Info("watcher", zap.String("service", service))
			select {
			case ch <- struct{}{}:
			default:
			}
		}
		lastIdx = idx
	}

	if err := wp.Run("localhost:8500"); err != nil {
		panic(err)
	}
}
