package consul

import (
	"os"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"github.com/hashicorp/consul/api"
)

// Consul represents the consul
type Consul struct {
	client *api.Client
	config *api.Config
}

func New(cfg config.Config) (discovery.Discovery, error) {
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Consul{client, config}, nil
}

// Register registers the panoptes at consul
func (c *Consul) Register() {
	reg := &api.AgentServiceRegistration{
		ID:      "panoptes",
		Name:    "panoptes",
		Address: hostname(),
	}

	reg.Check = &api.AgentServiceCheck{
		HTTP:     "localhost:8055/healthcheck",
		Interval: "10s",
		Timeout:  "2s",
	}

	c.client.Agent().ServiceRegister(reg)
}

// Deregister deregisters the panoptes at consul
func (c *Consul) Deregister() {
	c.client.Agent().ServiceDeregister("panoptes")
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	return name
}
