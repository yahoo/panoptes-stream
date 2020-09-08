//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"os"
	"path"
	"sort"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

// etcd represents the etcd as service discovery.
type etcd struct {
	id          string
	prefix      string
	cfg         config.Config
	logger      *zap.Logger
	client      *clientv3.Client
	session     *concurrency.Session
	lockHandler *concurrency.Mutex
}

type etcdConfig struct {
	Endpoints   []string
	Prefix      string
	DialTimeout int

	TLSConfig config.TLSConfig
}

// New constructs etcd service discovery.
func New(cfg config.Config) (discovery.Discovery, error) {
	var tlsConfig *tls.Config

	etcd := &etcd{
		cfg:    cfg,
		logger: cfg.Logger(),
	}

	config, err := getConfig(cfg)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_discovery_etcd"
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	if config.TLSConfig.Enabled {
		tlsConfig, err = secret.GetTLSConfig(&config.TLSConfig)
		if err != nil {
			return nil, err
		}
	}

	setDefaultConfig(config)

	etcd.prefix = config.Prefix

	etcd.client, err = clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: time.Duration(config.DialTimeout) * time.Second,
		TLS:         tlsConfig,
		LogConfig: &zap.Config{
			Development:      false,
			Level:            zap.NewAtomicLevelAt(zap.ErrorLevel),
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
	})
	if err != nil {
		return nil, err
	}

	return etcd, nil
}

// Register registers panoptes at etcd.
func (e *etcd) Register() error {
	if err := e.lock(); err != nil {
		return err
	}
	defer e.unlock()

	meta := make(map[string]string)
	meta["shards_enabled"] = strconv.FormatBool(e.cfg.Global().Shards.Enabled)
	meta["shards_nodes"] = strconv.Itoa(e.cfg.Global().Shards.NumberOfNodes)
	meta["version"] = e.cfg.Global().Version

	ids := []int{}
	instances, err := e.GetInstances()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		id, err := strconv.Atoi(instance.ID)
		if err != nil {
			e.logger.Warn("etcd", zap.String("event", "register"), zap.Error(err))
			continue
		}
		ids = append(ids, id)

		if instance.Address == hostname() {
			if err := e.register(instance.ID, hostname(), meta); err != nil {
				return err
			}

			e.logger.Info("etcd", zap.String("event", "register.recover"), zap.String("id", instance.ID))

			e.id = instance.ID

			return nil
		}
	}

	e.id = getID(ids)
	e.register(e.id, hostname(), meta)

	// TODO check lease id > 0

	go e.Watch(nil)

	e.logger.Info("etcd", zap.String("event", "register"), zap.String("id", e.id))

	return nil
}

// Deregister doesn't do anything as the panoptes
// deregister once the TTL is expired.
func (e *etcd) Deregister() error {
	return nil
}

// Watch monitors for updates at etcd panoptes service
// and notify through the channel.
func (e *etcd) Watch(ch chan<- struct{}) {
	prefix := path.Join(e.prefix, "services")
	rch := e.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			e.logger.Info("etcd", zap.String("event", "watcher.trigger"), zap.ByteString("key", ev.Kv.Key))
			select {
			case ch <- struct{}{}:
			default:
				e.logger.Debug("etcd", zap.String("event", "watcher.response.drop"))
			}
		}
	}
}

func (e *etcd) hearthBeat(leaseID clientv3.LeaseID) {
	ch, err := e.client.KeepAlive(context.Background(), leaseID)
	if err != nil {
		// TODO
		panic(err)
	}

	go func() {
		for {
			_, ok := <-ch
			if !ok {
				e.logger.Error("close channel")
				break
			}
		}
		// TODO etcd unreachable
	}()
}

func (e *etcd) register(id, hostname string, meta map[string]string) error {
	reg := discovery.Instance{
		ID:      id,
		Meta:    meta,
		Address: hostname,
		Status:  "passing",
	}

	e.id = id

	requestTimeout, _ := time.ParseDuration("5s")

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := e.client.Grant(ctx, 60)
	cancel()
	if err != nil {
		return err
	}

	b, _ := json.Marshal(&reg)

	ctx, cancel = context.WithTimeout(context.Background(), requestTimeout)
	prefix := path.Join(e.prefix, "services", e.id)
	_, err = e.client.Put(ctx, prefix, string(b), clientv3.WithLease(resp.ID))
	cancel()
	if err != nil {
		return err
	}

	e.hearthBeat(resp.ID)

	return err
}

// GetInstances returns all registered instances.
func (e *etcd) GetInstances() ([]discovery.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	prefix := path.Join(e.prefix, "services")
	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, err
	}

	var instances []discovery.Instance
	for _, ev := range resp.Kvs {
		instance := discovery.Instance{}
		if err := json.Unmarshal(ev.Value, &instance); err != nil {
			return nil, err
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

func (e *etcd) lock() error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	e.session, err = concurrency.NewSession(e.client, concurrency.WithContext(ctx))
	if err != nil {
		return err
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	prefix := path.Join(e.prefix, "global_lock")
	e.lockHandler = concurrency.NewMutex(e.session, prefix)

	return e.lockHandler.Lock(ctx)
}

func (e *etcd) unlock() error {
	defer e.session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.lockHandler.Unlock(ctx)
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

func getConfig(cfg config.Config) (*etcdConfig, error) {
	etcdConfig := new(etcdConfig)
	b, err := json.Marshal(cfg.Global().Discovery.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, etcdConfig)

	return etcdConfig, err
}

func setDefaultConfig(config *etcdConfig) {
	if len(config.Endpoints) < 1 {
		config.Endpoints = []string{"127.0.0.1:2379"}
	}

	if len(config.Prefix) < 1 {
		config.Prefix = "/panoptes/"
	}

	if config.DialTimeout < 1 {
		config.DialTimeout = 2
	}
}
