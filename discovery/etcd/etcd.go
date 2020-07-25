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

type Etcd struct {
	id          string
	cfg         config.Config
	logger      *zap.Logger
	client      *clientv3.Client
	session     *concurrency.Session
	lockHandler *concurrency.Mutex
}

type etcdConfig struct {
	Endpoints []string
	Prefix    string

	TLSConfig config.TLSConfig
}

func New(cfg config.Config) (discovery.Discovery, error) {
	var tlsConfig *tls.Config

	config, err := getConfig(cfg)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_discovery_etcd"
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	if config.TLSConfig.CertFile != "" && !config.TLSConfig.Disabled {
		tlsConfig, err = secret.GetTLSConfig(&config.TLSConfig)
		if err != nil {
			return nil, err
		}
	}

	etcd := &Etcd{
		cfg:    cfg,
		logger: cfg.Logger(),
	}

	etcd.client, err = clientv3.New(clientv3.Config{
		Endpoints: config.Endpoints,
		TLS:       tlsConfig,
	})
	if err != nil {
		return nil, err
	}

	return etcd, nil
}

func (e *Etcd) Register() error {
	e.lock()
	defer e.unlock()

	meta := make(map[string]string)
	meta["shard_enabled"] = strconv.FormatBool(e.cfg.Global().Shard.Enabled)
	meta["shard_nodes"] = strconv.Itoa(e.cfg.Global().Shard.NumberOfNodes)
	meta["version"] = e.cfg.Global().Version

	ids := []int{}
	instances, err := e.GetInstances()
	if err != nil {
		return err
	}

	for _, instance := range instances {
		id, err := strconv.Atoi(instance.ID)
		if err != nil {
			e.logger.Warn("etcd.register", zap.Error(err))
			continue
		}
		ids = append(ids, id)

		if instance.Address == hostname() {
			if err := e.register(instance.ID, meta); err != nil {
				return err
			}

			e.logger.Info("consul service registery recovered", zap.String("id", instance.ID))

			e.id = instance.ID

			return nil
		}
	}

	e.id = getID(ids)
	e.register(e.id, meta)

	// TODO check lease id > 0

	go e.Watch(nil)

	return nil
}
func (e *Etcd) Deregister() error {
	return nil
}
func (e *Etcd) Watch(ch chan<- struct{}) {
	rch := e.client.Watch(context.Background(), e.cfg.Global().Discovery.Prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			e.logger.Info("etcd watcher triggered", zap.ByteString("key", ev.Kv.Key))
			select {
			case ch <- struct{}{}:
			default:
				e.logger.Info("etcd watcher response dropped")
			}
		}
	}
}

func (e *Etcd) hearthBeat(leaseID clientv3.LeaseID) {
	ch, err := e.client.KeepAlive(context.Background(), leaseID)
	if err != nil {
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

func (e *Etcd) register(id string, meta map[string]string) error {
	reg := discovery.Instance{
		ID:      id,
		Meta:    meta,
		Address: hostname(),
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
	prefix := path.Join(e.cfg.Global().Discovery.Prefix, e.id)
	_, err = e.client.Put(ctx, prefix, string(b), clientv3.WithLease(resp.ID))
	cancel()
	if err != nil {
		return err
	}

	e.hearthBeat(resp.ID)

	return err
}

func (e *Etcd) GetInstances() ([]discovery.Instance, error) {
	requestTimeout, _ := time.ParseDuration("5s")

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := e.client.Get(ctx, e.cfg.Global().Discovery.Prefix, clientv3.WithPrefix())
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

func (e *Etcd) lock() error {
	var err error
	e.session, err = concurrency.NewSession(e.client)
	if err != nil {
		return err
	}

	requestTimeout, _ := time.ParseDuration("5s")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	e.lockHandler = concurrency.NewMutex(e.session, "/panoptes_global_locki/")
	e.lockHandler.Lock(ctx)
	cancel()

	return nil
}

func (e *Etcd) unlock() error {
	defer e.session.Close()

	requestTimeout, _ := time.ParseDuration("5s")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
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
