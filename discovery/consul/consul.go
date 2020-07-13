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
func (c *Consul) Register() error {
	_, err := c.lock("panoptes_global_lock", nil)
	if err != nil {
		return err
	}
	defer c.ulock()

	meta := make(map[string]string)
	meta["shard_enabled"] = strconv.FormatBool(c.cfg.Global().Shard.Enabled)
	meta["shard_nodes"] = strconv.Itoa(c.cfg.Global().Shard.NumberOfNodes)
	meta["version"] = c.cfg.Global().Version

	ids := []int{}
	instances, err := c.GetInstances()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		id, err := strconv.Atoi(instance.ID)
		if err != nil {
			c.logger.Warn("consul.register", zap.Error(err))
			continue
		}
		ids = append(ids, id)
		// recover node
		if instance.Address == hostname() {
			if err := c.register(instance.ID, meta); err != nil {
				return err
			}

			c.logger.Info("consul service registery recovered", zap.String("id", instance.ID))

			c.id = instance.ID

			return nil
		}
	}

	// new register node
	// TODO: if id > numbber_of_nodes then needs clean up!
	c.id = getID(ids)
	if err := c.register(c.id, meta); err != nil {
		return err
	}

	c.logger.Info("consul service registered", zap.String("id", c.id))

	return nil
}

func (c *Consul) register(id string, meta map[string]string) error {
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

	return c.client.Agent().ServiceRegister(reg)
}

// GetInstances returns all registered instances
func (c *Consul) GetInstances() ([]discovery.Instance, error) {
	var instances []discovery.Instance
	_, checksInfo, err := c.client.Agent().AgentHealthServiceByName("panoptes")
	if err != nil {
		return nil, err
	}

	for _, info := range checksInfo {
		instances = append(instances, discovery.Instance{
			ID:      info.Service.ID,
			Meta:    info.Service.Meta,
			Address: info.Service.Address,
			Status:  info.Checks.AggregatedStatus(),
		})
	}
	return instances, nil
}

// Deregister deregisters the panoptes at consul
func (c *Consul) Deregister() error {
	return c.client.Agent().ServiceDeregister(c.id)
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
