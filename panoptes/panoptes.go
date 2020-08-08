package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/demux"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"git.vzbuilders.com/marshadrad/panoptes/discovery/consul"
	"git.vzbuilders.com/marshadrad/panoptes/discovery/etcd"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/register"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

var (
	producerRegistrar  *producer.Registrar
	databaseRegistrar  *database.Registrar
	telemetryRegistrar *telemetry.Registrar
)

func main() {
	var (
		discovery     discovery.Discovery
		signalCh      = make(chan os.Signal, 1)
		updateRequest = make(chan struct{}, 1)
		outChan       = make(telemetry.ExtDSChan, 10000)
		ctx           = context.Background()
	)

	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := getConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger := cfg.Logger()
	defer logger.Sync()

	logger.Info("starting ...")

	// discovery
	discovery, err = discoveryRegister(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if discovery != nil {
		defer discovery.Deregister()
	}

	// producer
	producerRegistrar = producer.NewRegistrar(logger)
	register.Producer(producerRegistrar)

	// database
	databaseRegistrar = database.NewRegistrar(logger)
	register.Database(databaseRegistrar)

	// telemetry
	telemetryRegistrar = telemetry.NewRegistrar(logger)
	register.Telemetry(telemetryRegistrar)

	// start demux
	d := demux.New(ctx, cfg, producerRegistrar, databaseRegistrar, outChan)
	d.Init()
	d.Start()

	// start telemetry
	t := telemetry.New(ctx, cfg, telemetryRegistrar, outChan)
	if !cfg.Global().Shard.Enabled {
		t.Start()
	}

	// status
	if !cfg.Global().Status.Disabled {
		s := status.New(cfg)
		s.Start()
	}

	go updateLoop(cfg, t, d, updateRequest)

	if cfg.Global().Shard.Enabled && discovery != nil {
		shard := NewShard(cfg, t, discovery, updateRequest)
		go shard.Start()
	}

	<-signalCh
}

func updateLoop(cfg config.Config, t *telemetry.Telemetry, d *demux.Demux, updateRequest chan struct{}) {
	var informed bool

	for {
		select {
		case <-cfg.Informer():
			informed = true
			continue

		case <-updateRequest:

		case <-time.After(time.Second * 10):
			if !informed {
				continue
			}
			informed = false
		}

		if err := cfg.Update(); err != nil {
			cfg.Logger().Error("update", zap.Error(err))
			continue
		}

		d.Update()
		t.Update()
	}
}

func discoveryRegister(cfg config.Config) (discovery.Discovery, error) {
	var (
		discovery discovery.Discovery
		err       error
	)

	switch cfg.Global().Discovery.Service {

	case "consul":
		discovery, err = consul.New(cfg)
		if err != nil {
			return nil, err
		}
	case "etcd":
		discovery, err = etcd.New(cfg)
		if err != nil {
			return nil, err
		}
	default:
		cfg.Logger().Info("discovery disabled")
		return nil, nil
	}

	return discovery, discovery.Register()
}
