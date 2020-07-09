package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/demux"
	"git.vzbuilders.com/marshadrad/panoptes/discovery/consul"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/register"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

var (
	producerRegistrar  *producer.Registrar
	databaseRegistrar  *database.Registrar
	telemetryRegistrar *telemetry.Registrar
)

func main() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}

	lg := cfg.Logger()
	defer lg.Sync()

	lg.Info("starting ...")

	discovery, err := consul.New(cfg)
	if err != nil {
		panic(err)
	}
	discovery.Register()
	defer discovery.Deregister()

	ctx := context.Background()

	// producer
	producerRegistrar = producer.NewRegistrar(lg)
	register.Producer(producerRegistrar)

	// database
	databaseRegistrar = database.NewRegistrar(lg)
	register.Database(databaseRegistrar)

	// telemetry
	telemetryRegistrar = telemetry.NewRegistrar(lg)
	register.Telemetry(telemetryRegistrar)

	outChan := make(telemetry.ExtDSChan, 1)

	d := demux.New(ctx, cfg, producerRegistrar, databaseRegistrar, outChan)
	d.Init()
	go d.Start()

	t := telemetry.New(ctx, cfg, telemetryRegistrar, outChan)
	t.Start()

	s := status.New(cfg)
	s.Start()

	updateRequest := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-cfg.Informer():
			case <-updateRequest:
			}

			cfg.Update()
			d.Update()
			t.Update()
		}
	}()

	if cfg.Global().Shard.Enabled {
		shard := NewShard(cfg, t, discovery, updateRequest)
		go shard.Start()
	}

	<-signalCh
}
