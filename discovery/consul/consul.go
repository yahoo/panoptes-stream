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
	meta := make(map[string]string)
	reg := &api.AgentServiceRegistration{
		ID:      "panoptes",
		Name:    "panoptes",
		Meta:    meta,
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

// GetInstances returns all registered instances
func (c *Consul) GetInstances() []discovery.Instance {
	var instances []discovery.Instance
	_, checksInfo, err := c.client.Agent().AgentHealthServiceByName("panoptes")
	if err != nil {
		c.logger.Error("get health service failed", zap.Error(err))
	}

	for _, info := range checksInfo {
		instances = append(instances, discovery.Instance{
			ID:      info.Service.ID,
			Meta:    info.Service.Meta,
			Address: info.Service.Address,
			Status:  info.Checks.AggregatedStatus(),
		})
	}
	return instances
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
