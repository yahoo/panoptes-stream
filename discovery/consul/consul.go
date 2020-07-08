package consul

import (
	"os"
	"sort"
	"strconv"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
)

// Consul represents the consul
type Consul struct {
	id          string
	cfg         config.Config
	logger      *zap.Logger
	client      *api.Client
	lockHandler *api.Lock
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
	_, err := c.lock("panoptes_global_lock", nil)
	if err != nil {
		panic(err)
	}
	defer c.ulock()

	ids := []int{}
	for _, instance := range c.GetInstances() {
		id, err := strconv.Atoi(instance.ID)
		if err != nil {
			c.logger.Warn("consul.register", zap.Error(err))
			continue
		}
		ids = append(ids, id)
		// recover node
		if instance.Address == hostname() {
			c.logger.Info("consul service registery recovered", zap.String("id", instance.ID))
			c.register(instance.ID, instance.Meta)
			c.id = instance.ID
			return
		}
	}

	// new register node
	c.id = getID(ids)
	c.register(c.id, nil)
	c.logger.Info("consul service registered", zap.String("id", c.id))
}

func (c *Consul) register(id string, meta map[string]string) {
	reg := &api.AgentServiceRegistration{
		ID:      id,
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
	if err := c.client.Agent().ServiceDeregister(c.id); err != nil {
		c.logger.Error("deregister failed", zap.Error(err))
	}
}

func (c *Consul) lock(key string, stopChan chan struct{}) (<-chan struct{}, error) {
	var err error
	opts := &api.LockOptions{
		Key:        key,
		SessionTTL: "10s",
	}
	c.lockHandler, err = c.client.LockOpts(opts)
	if err != nil {
		return nil, err
	}

	return c.lockHandler.Lock(stopChan)
}

func (c *Consul) ulock() {
	if err := c.lockHandler.Unlock(); err != nil {
		c.logger.Error("consul.unlock", zap.Error(err))
	}
}

func hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	return name
}

func getID(ids []int) string {
	idStr := "0"

	if len(ids) < 1 {
		return idStr
	}

	sort.Ints(ids)
	for i, id := range ids {
		if i != id {
			idsStr := strconv.Itoa(i)
			return idsStr
		}
	}

	idStr = strconv.Itoa(len(ids))

	return idStr
}
