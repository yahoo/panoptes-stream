package consul

import (
	"encoding/json"
	"path"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"github.com/hashicorp/consul/api"
)

type consul struct {
	client *api.Client

	filename  string
	devices   []config.Device
	producers []config.Producer
	global    config.Global

	informer chan struct{}
}

type yamlConfig struct {
	Address string
}

func New(filename string) (config.Config, error) {
	var (
		err    error
		cfg    = &yamlConfig{}
		consul = &consul{}
	)

	if err := yaml.Read(filename, cfg); err != nil {
		return nil, err
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.Address

	consul.client, err = api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	kv := consul.client.KV()

	pairs, _, err := kv.List("config/producers/", nil)
	if err != nil {
		return nil, err
	}

	consul.producers = configProducers(pairs)

	return consul, nil
}

func (e *consul) Devices() []config.Device {
	return e.devices
}

func (e *consul) Producers() []config.Producer {
	return e.producers
}

func (e *consul) Global() config.Global {
	return e.global
}

func (e *consul) Informer() chan struct{} {

	return e.informer
}

func configProducers(pairs api.KVPairs) []config.Producer {
	var producers []config.Producer

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		producer := config.Producer{}
		if err := json.Unmarshal(p.Value, &producer); err != nil {
			panic(err)
		}

		_, producer.Name = path.Split(p.Key)
		producers = append(producers, producer)
	}

	return producers
}
