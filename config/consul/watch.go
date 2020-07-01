package consul

import (
	"log"

	"github.com/hashicorp/consul/api"
)

func (c *consul) watch(prefix string) {
	client := c.client.KV()
	opts := api.QueryOptions{}

	_, meta, _ := client.Get("config/global", &opts)

	for {
		opts.WaitIndex = meta.LastIndex
		_, meta, _ = client.Get("config/global", &opts)

		select {
		case c.informer <- struct{}{}:
		default:
		}

		log.Println("modified connfig/global")

	}
}
