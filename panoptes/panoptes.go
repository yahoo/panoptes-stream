package main

import (
	"context"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
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

	dp := demux.New(ctx, cfg, producerRegistrar, databaseRegistrar, outChan)
	dp.Init()
	go dp.Start()

	t := telemetry.New(ctx, cfg, telemetryRegistrar, outChan)
	t.Start()

	s := status.New(cfg)
	s.Start()

	ticker := time.NewTicker(time.Second * 60)
	for {
		<-ticker.C
		devices := make(map[string]config.Device)
		producers := make(map[string]config.Producer)
		databases := make(map[string]config.Database)

		for _, device := range cfg.Devices() {
			devices[device.Host] = device
		}

		for _, producer := range cfg.Producers() {
			producers[producer.Name] = producer
		}

		for _, database := range cfg.Databases() {
			databases[database.Name] = database
		}

		cfg.Update()
		dp.Update(producers, databases)
		t.Update(devices)
	}

}
