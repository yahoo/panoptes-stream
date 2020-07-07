package consul

import (
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"
)

func (c *consul) watch(watchType, value string, ch chan<- struct{}) {
	params := make(map[string]interface{})
	params["type"] = watchType

	switch watchType {
	case "keyprefix":
		params["prefix"] = value
	case "key":
		params["key"] = value
	}

	wp, err := watch.Parse(params)
	if err != nil {
		panic(err)
	}

	lastIdx := uint64(0)
	wp.Handler = func(idx uint64, data interface{}) {
		if lastIdx != 0 {
			c.logger.Info("watcher", zap.String("name", value), zap.String("type", watchType))
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
