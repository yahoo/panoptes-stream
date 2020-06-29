package main

import (
	"context"

	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/demux"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/register"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

var (
	producerRegistrar  *producer.Registrar
	telemetryRegistrar *telemetry.Registrar
)

func main() {
	cfg := yaml.LoadConfig("etc/config.yaml")
	lg := GetLogger(cfg.Global().Logger)
	defer lg.Sync()

	lg.Info("starting ...")

	ctx := context.Background()

	// producer
	producerRegistrar = producer.NewRegistrar(lg)
	register.Producer(producerRegistrar)

	// telemetry
	telemetryRegistrar = telemetry.NewRegistrar(lg)
	register.Telemetry(telemetryRegistrar)

	outChan := make(telemetry.ExtDSChan, 1)

	dp := demux.New(cfg, lg, producerRegistrar, outChan)
	dp.Init()
	go dp.Start(ctx)

	p := NewPanoptes(ctx, lg, telemetryRegistrar, outChan)
	for _, device := range cfg.Devices() {
		p.subscribe(device)
	}

	<-ctx.Done()
}
