//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/demux"
	"git.vzbuilders.com/marshadrad/panoptes/discovery"
	"git.vzbuilders.com/marshadrad/panoptes/discovery/consul"
	"git.vzbuilders.com/marshadrad/panoptes/discovery/etcd"
	"git.vzbuilders.com/marshadrad/panoptes/discovery/k8s"
	"git.vzbuilders.com/marshadrad/panoptes/discovery/pseudo"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/register"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/dialout"
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
		ctx           = context.Background()
	)

	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := getConfig(os.Args)
	if err != nil {
		fmt.Println("config", err)
		os.Exit(1)
	}

	logger := cfg.Logger()
	defer logger.Sync()

	outChan := make(telemetry.ExtDSChan, cfg.Global().BufferSize)

	// discovery
	discovery, err = discoveryRegister(cfg)
	if err != nil {
		logger.Fatal("discovery", zap.Error(err))
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
	d.Start()

	// start telemetry
	t := telemetry.New(ctx, cfg, telemetryRegistrar, outChan)
	if !cfg.Global().Shards.Enabled {
		t.Start()
	}

	// start telemetry dialout
	i := dialout.New(ctx, cfg, outChan)
	i.Start()

	// status
	if !cfg.Global().Status.Disabled {
		s := status.New(cfg)
		s.Start()
	}

	go updateLoop(cfg, t, d, i, updateRequest)

	if cfg.Global().Shards.Enabled && discovery != nil {
		shards := NewShards(cfg, t, discovery, updateRequest)
		go shards.Start()
	}

	<-signalCh
}

func updateLoop(cfg config.Config, t *telemetry.Telemetry, d *demux.Demux, i *dialout.Dialout, updateRequest chan struct{}) {
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
		i.Update()
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
	case "k8s":
		discovery, err = k8s.New(cfg)
		if err != nil {
			return nil, err
		}
	case "pseudo":
		discovery, err = pseudo.New(cfg)
		if err != nil {
			return nil, err
		}
	default:
		cfg.Logger().Info("discovery disabled")
		return nil, nil
	}

	return discovery, discovery.Register()
}
