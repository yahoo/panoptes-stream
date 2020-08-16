package dialout

import (
	"context"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/cisco/mdt"
)

type Dialout struct {
	cfg     config.Config
	ctx     context.Context
	outChan telemetry.ExtDSChan

	mdtHandler *mdt.MDTDialout
}

func New(ctx context.Context, cfg config.Config, outChan telemetry.ExtDSChan) *Dialout {
	return &Dialout{
		cfg:     cfg,
		ctx:     ctx,
		outChan: outChan,
	}
}

func (d *Dialout) Start() {
	for service := range d.cfg.Global().Dialout.Services {
		if service == "cisco.mdt" {
			d.mdtHandler = mdt.NewDialout(d.ctx, service, d.cfg, d.outChan)
			d.mdtHandler.Start()
		}
	}
}

func (d *Dialout) Update() {
	for service := range d.cfg.Global().Dialout.Services {
		if service == "cisco.mdt" {
			d.mdtHandler.Update()
		}
	}
}
