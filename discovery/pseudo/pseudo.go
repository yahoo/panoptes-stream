//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package pseudo

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"hash/fnv"
	"net"
	"net/http"
	"os"
	"path"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

// pseudo represents pseudo service discovery.
// it checks the configured nodes through http/https probe.
type pseudo struct {
	cfg       config.Config
	logger    *zap.Logger
	instances []*instance

	probe     string
	path      string
	timeout   time.Duration
	interval  time.Duration
	maxRetry  int
	tlsConfig *config.TLSConfig
}

type pseudoConfig struct {
	Instances []string
	Probe     string
	Path      string
	Timeout   int
	Interval  int
	MaxRetry  int
	TLSConfig *config.TLSConfig
}

type instance struct {
	hostname string
	address  string
	id       string
	status   string
}

// New constructs pseudo service discovery.
func New(cfg config.Config) (discovery.Discovery, error) {
	conf, err := getConfig(cfg)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_discovery_pseudo"
	err = envconfig.Process(prefix, conf)
	if err != nil {
		return nil, err
	}

	config.SetDefault(&conf.Interval, 15)
	config.SetDefault(&conf.Timeout, 1)
	config.SetDefault(&conf.MaxRetry, 3)

	p := &pseudo{
		cfg:      cfg,
		logger:   cfg.Logger(),
		probe:    conf.Probe,
		path:     conf.Path,
		maxRetry: conf.MaxRetry,

		interval: time.Duration(conf.Interval) * time.Second,
		timeout:  time.Duration(conf.Timeout) * time.Second,

		tlsConfig: conf.TLSConfig,
	}

	localIPs, err := getLocalIPaddrs()
	if err != nil {
		return nil, err
	}

	for id, addr := range consensusOrdinal(conf.Instances) {
		hostname, err := getHostname(localIPs, addr)
		if err != nil {
			return nil, err
		}

		p.instances = append(p.instances, &instance{
			id:       strconv.Itoa(id),
			address:  addr,
			hostname: hostname,
			status:   "failure",
		})

		if hostname != "" {
			p.logger.Info("pseudo", zap.String("event", "register"), zap.String("id", strconv.Itoa(id)))
		}
	}

	go p.check()

	return p, nil
}

// Register satisfies discovery interface.
// pseudo doesn't have registry service.
func (*pseudo) Register() error {
	// not available
	return nil
}

// Deregister satisfies discovery interface.
// pseudo doesn't have deregistry service.
func (*pseudo) Deregister() error {
	// not available
	return nil
}

// GetInstances returns all instances.
func (p *pseudo) GetInstances() ([]discovery.Instance, error) {
	var instances []discovery.Instance

	meta := map[string]string{
		"shards_enabled": "true",
		"shards_nodes":   strconv.Itoa(p.cfg.Global().Shards.NumberOfNodes),
		"version":        p.cfg.Global().Version,
	}

	for _, instance := range p.instances {
		instances = append(instances, discovery.Instance{
			ID:      instance.id,
			Meta:    meta,
			Address: instance.hostname,
			Status:  instance.status,
		})
	}
	return instances, nil
}

// Watch monitors status of instances and notify through the channel.
func (p *pseudo) Watch(ch chan<- struct{}) {
	var instances = make([]*instance, len(p.instances))

	ticker := time.NewTicker(2 * time.Second)

	deepCopy(instances, p.instances)

	for {
		<-ticker.C
		if !reflect.DeepEqual(instances, p.instances) {
			p.logger.Debug("pseudo", zap.String("event", "watch.trigger"))

			select {
			case ch <- struct{}{}:
			default:
				p.logger.Debug("pseudo", zap.String("event", "watch.response.drop"))
			}

			deepCopy(instances, p.instances)
		}
	}
}

func (p *pseudo) check() {
	// check insterval
	ticker := time.NewTicker(p.interval)
	// warm-up time (status server)
	time.Sleep(2 * time.Second)

	for {
		for _, instance := range p.instances {
			if p.probe == "http" || p.probe == "https" {
				go p.checkHTTP(instance)
			} else {
				p.logger.Fatal("pseudo", zap.String("msg", "probe doesn't support"))
			}
		}

		<-ticker.C
	}
}

func getConfig(cfg config.Config) (*pseudoConfig, error) {
	panoptesConfig := new(pseudoConfig)
	b, err := json.Marshal(cfg.Global().Discovery.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, panoptesConfig)

	return panoptesConfig, err
}

func (p *pseudo) checkHTTP(instance *instance) {
	var (
		tlsConfig *tls.Config
		resp      *http.Response
		err       error
		retry     int
	)

	if p.tlsConfig.Enabled {
		tlsConfig, err = secret.GetTLSConfig(p.tlsConfig)
		p.logger.Fatal("discovery.panoptes", zap.Error(err))
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsConfig,
			DisableKeepAlives: true,
		},
		Timeout: p.timeout,
	}

	for {
		retry++
		resp, err = client.Get(p.probe + "://" + path.Join(instance.address, p.path))
		if err == nil {
			break
		} else if retry == p.maxRetry {
			instance.status = "failure"
			return
		}

		time.Sleep(time.Second)
	}

	if resp.StatusCode < 400 && resp.StatusCode >= 200 {
		instance.status = "passing"
	} else {
		instance.status = "failure"
	}

	resp.Body.Close()
}

func consensusOrdinal(slice []string) []string {
	mapHash := map[int]string{}
	hashSlice := []int{}

	for _, s := range slice {
		hash := getHash(s)
		mapHash[hash] = s
		hashSlice = append(hashSlice, hash)
	}

	sort.Ints(hashSlice)

	newSlice := []string{}
	for _, kk := range hashSlice {
		newSlice = append(newSlice, mapHash[kk])
	}

	return newSlice
}

func getHash(key string) int {
	hash := fnv.New32()
	hash.Write([]byte(key))
	hSum32 := hash.Sum32()
	return int(hSum32)
}

func deepCopy(dst, src []*instance) {
	for i := range src {
		a := *src[i]
		dst[i] = &a
	}
}

func getHostname(localIPs map[string]struct{}, instance string) (string, error) {
	host, _, err := net.SplitHostPort(instance)
	if err != nil {
		return "", err
	}
	resolver := net.DefaultResolver
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	addrs, err := resolver.LookupHost(ctx, host)
	cancel()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if _, ok := localIPs[addr]; ok {
			hostname, _ := os.Hostname()
			return hostname, nil
		}
	}

	return "", nil
}

func getLocalIPaddrs() (map[string]struct{}, error) {
	ips := map[string]struct{}{}

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ips[v.IP.String()] = struct{}{}
			case *net.IPAddr:
				ips[v.IP.String()] = struct{}{}
			}
		}
	}

	return ips, nil
}
