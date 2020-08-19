package consul

import (
	"encoding/json"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

// consul represents the Consul service discovery
type consul struct {
	id          string
	cfg         config.Config
	config      *consulConfig
	logger      *zap.Logger
	client      *api.Client
	lockHandler *api.Lock
}

type consulConfig struct {
	Address        string
	Prefix         string
	HealthcheckURL string

	TLSConfig config.TLSConfig
}

// New constructs consul service discovery
func New(cfg config.Config) (discovery.Discovery, error) {
	config, err := getConfig(cfg)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_discovery_consul"
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = config.Address

	if config.TLSConfig.Enabled {
		apiConfig.TLSConfig, err = getTLSConfig(config)
		if err != nil {
			return nil, err
		}
	}

	client, err := api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	if len(config.Prefix) < 1 {
		config.Prefix = "/panoptes/"
	}

	return &consul{
		cfg:    cfg,
		config: config,
		client: client,
		logger: cfg.Logger(),
	}, nil
}

// Register registers the panoptes at consul
func (c *consul) Register() error {
	key := path.Join(c.config.Prefix, "global_lock")[1:]
	_, err := c.lock(key, nil)
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
		// recovered node
		if instance.Address == hostname() {
			if err := c.register(instance.ID, hostname(), meta); err != nil {
				return err
			}

			c.logger.Info("consul service registery recovered", zap.String("id", instance.ID))

			c.id = instance.ID

			return nil
		}
	}

	// new register node
	// TODO: if id > numbber_of_nodes then it needs clean up!
	c.id = getID(ids)
	if err := c.register(c.id, hostname(), meta); err != nil {
		return err
	}

	c.logger.Info("consul service registered", zap.String("id", c.id))

	return nil
}

func (c *consul) register(id, hostname string, meta map[string]string) error {
	reg := &api.AgentServiceRegistration{
		ID:      id,
		Name:    "panoptes",
		Meta:    meta,
		Address: hostname,
	}

	reg.Check = &api.AgentServiceCheck{
		HTTP:     c.getHealthcheckURL(),
		Interval: "10s",
		Timeout:  "2s",
	}

	return c.client.Agent().ServiceRegister(reg)
}

// GetInstances returns all registered instances
func (c *consul) GetInstances() ([]discovery.Instance, error) {
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
func (c *consul) Deregister() error {
	return c.client.Agent().ServiceDeregister(c.id)
}

func (c *consul) lock(key string, stopChan chan struct{}) (<-chan struct{}, error) {
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

func (c *consul) ulock() {
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

// Watch monitors for updates at consul panoptes service
// and notify through the channel.
func (c *consul) Watch(ch chan<- struct{}) {
	params := make(map[string]interface{})
	params["type"] = "service"
	params["service"] = "panoptes"

	wp, err := watch.Parse(params)
	if err != nil {
		panic(err)
	}

	lastIdx := uint64(0)
	wp.Handler = func(idx uint64, data interface{}) {
		if lastIdx != 0 {
			c.logger.Info("consul watcher triggered")
			select {
			case ch <- struct{}{}:
			default:
			}
		}
		lastIdx = idx
	}

	if err := wp.Run(c.config.Address); err != nil {
		panic(err)
	}
}

func getConfig(cfg config.Config) (*consulConfig, error) {
	consulConfig := new(consulConfig)
	b, err := json.Marshal(cfg.Global().Discovery.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, consulConfig)

	return consulConfig, err
}

func getTLSConfig(cfg *consulConfig) (api.TLSConfig, error) {
	var CAPEM []byte

	sType, path, ok := secret.ParseRemoteSecretInfo(cfg.TLSConfig.CertFile)
	if ok {
		sec, err := secret.GetSecretEngine(sType)
		if err != nil {
			return api.TLSConfig{}, nil
		}

		secrets, err := sec.GetSecrets(path)
		if err != nil {
			return api.TLSConfig{}, nil
		}

		if v, ok := secrets["ca"]; ok {
			CAPEM = v
		}

		return api.TLSConfig{
			CertPEM:            secrets["cert"],
			KeyPEM:             secrets["key"],
			CAPem:              CAPEM,
			InsecureSkipVerify: cfg.TLSConfig.InsecureSkipVerify,
		}, nil

	}

	return api.TLSConfig{
		CertFile:           cfg.TLSConfig.CertFile,
		KeyFile:            cfg.TLSConfig.KeyFile,
		CAFile:             cfg.TLSConfig.CAFile,
		InsecureSkipVerify: cfg.TLSConfig.InsecureSkipVerify,
	}, nil
}

func (c *consul) getHealthcheckURL() string {
	if len(c.config.HealthcheckURL) > 0 {
		return c.config.HealthcheckURL
	}

	// TODO: if status binded to all and there is
	// one ip address, assign it to hc

	if len(c.cfg.Global().Status.Addr) > 0 {
		hc := path.Join(c.cfg.Global().Status.Addr, "healthcheck")
		if c.cfg.Global().Status.TLSConfig.Enabled {
			return "https://" + hc
		}

		return "http://" + hc
	}

	return "http://127.0.0.1:8081"
}
