package main

import (
	"context"
	"net"
	"strconv"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type panoptes struct {
	register           map[string]context.CancelFunc
	ctx                context.Context
	lg                 *zap.Logger
	outChan            telemetry.ExtDSChan
	telemetryRegistrar *telemetry.Registrar
}

func NewPanoptes(ctx context.Context, lg *zap.Logger, tr *telemetry.Registrar, outChan telemetry.ExtDSChan) *panoptes {
	return &panoptes{
		register:           make(map[string]context.CancelFunc),
		ctx:                ctx,
		lg:                 lg,
		outChan:            outChan,
		telemetryRegistrar: tr,
	}
}

func (p *panoptes) subscribe(device config.Device) {
	var (
		ctx  context.Context
		addr string
	)

	ctx, p.register[device.Host] = context.WithCancel(p.ctx)

	for sName, sensors := range device.Sensors {
		go func(sName string, sensors []*config.Sensor) {
			for {
				if device.Port > 0 {
					addr = net.JoinHostPort(device.Host, strconv.Itoa(device.Port))
				} else {
					addr = device.Host
				}

				conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithUserAgent("Panoptes"))
				if err != nil {
					p.lg.Error("connect to device", zap.Error(err))
				} else {
					NewNMI, _ := p.telemetryRegistrar.GetNMIFactory(sName)
					nmi := NewNMI(p.lg, conn, sensors, p.outChan)
					err = nmi.Start(ctx)
					if err != nil {
						p.lg.Warn("nmi start error", zap.Error(err))
					}
				}

				<-time.After(time.Second * 10)
			}
		}(sName, sensors)
	}
}

func (p *panoptes) unsubscribe(device config.Device) {
	cancel := p.register[device.Host]
	cancel()
}
