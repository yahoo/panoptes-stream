package consul

import (
	"os"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// Consul represents the consul
type Consul struct {
	client *api.Client
	cfg    config.Config
	logger *zap.Logger
}

func New(cfg config.Config) (discovery.Discovery, error) {
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &Consul{
		client: client,
		cfg:    cfg,
		logger: cfg.Logger(),
	}, nil
}

// Register registers the panoptes at consul
func (c *Consul) Register() {
	reg := &api.AgentServiceRegistration{
		ID:      "panoptes",
		Name:    "panoptes",
		Address: hostname(),
	}

	reg.Check = &api.AgentServiceCheck{
		HTTP:     "http://localhost:8081/healthcheck",
		Interval: "10s",
		Timeout:  "2s",
	}

	if err := c.client.Agent().ServiceRegister(reg); err != nil {
		c.logger.Error("register failed", zap.Error(err))
	}
}

// Deregister deregisters the panoptes at consul
func (c *Consul) Deregister() {
	if err := c.client.Agent().ServiceDeregister("panoptes"); err != nil {
		c.logger.Error("deregister failed", zap.Error(err))
	}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	return name
}
