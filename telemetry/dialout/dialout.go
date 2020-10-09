//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package dialout

import (
	"context"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/cisco/mdt"
)

// Dialout represents dial-out mode for all telemetries.
type Dialout struct {
	cfg     config.Config
	ctx     context.Context
	outChan telemetry.ExtDSChan

	mdtHandler *mdt.Dialout
}

// New creates a new dialout instance.
func New(ctx context.Context, cfg config.Config, outChan telemetry.ExtDSChan) *Dialout {
	return &Dialout{
		cfg:     cfg,
		ctx:     ctx,
		outChan: outChan,
	}
}

// Start starts available dialout telemetries.
func (d *Dialout) Start() {
	for service := range d.cfg.Global().Dialout.Services {
		if service == "cisco.mdt" {
			d.mdtHandler = mdt.NewDialout(d.ctx, d.cfg, d.outChan)
			d.mdtHandler.Start()
		}
	}
}

// Update updates dialout telemetries once configuration changed.
func (d *Dialout) Update() {
	for service := range d.cfg.Global().Dialout.Services {
		if service == "cisco.mdt" {
			d.mdtHandler.Update()
		}
	}
}
