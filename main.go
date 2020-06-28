package main

import (
	"context"
	"encoding/json"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/demux"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/register"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func main() {
	cfg := yaml.LoadConfig("etc/config.yaml")
	lg := GetLogger(cfg.Global().Logger)
	defer lg.Sync()

	lg.Info("starting ...")

	ctx := context.Background()

	register.RegisterVendor(lg)
	mqueue.Register()

	outChan := make(telemetry.ExtDSChan, 1)

	dp := demux.New(cfg, outChan)
	dp.Init()
	go dp.Start(ctx)

	p := NewPanoptes(ctx, lg, outChan)
	for _, device := range cfg.Devices() {
		p.subscribe(device)
	}

	<-ctx.Done()
}

type panoptes struct {
	register map[string]context.CancelFunc
	ctx      context.Context
	outChan  telemetry.ExtDSChan

	lg *zap.Logger
}

func NewPanoptes(ctx context.Context, lg *zap.Logger, outChan telemetry.ExtDSChan) *panoptes {
	return &panoptes{
		register: make(map[string]context.CancelFunc),
		ctx:      ctx,
		outChan:  outChan,
		lg:       lg,
	}
}

func (p *panoptes) subscribe(device config.Device) {
	var ctx context.Context
	ctx, p.register[device.Host] = context.WithCancel(p.ctx)

	for sName, sensors := range device.Sensors {
		go func(sName string, sensors []*config.Sensor) {
			for {
				conn, err := grpc.Dial(device.Host, grpc.WithInsecure(), grpc.WithUserAgent("Panoptes"))
				if err != nil {
					p.lg.Error("connect to device", zap.Error(err))
				} else {
					NewNMI := telemetry.GetNMIFactory(sName)
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

func GetLogger(lcfg map[string]interface{}) *zap.Logger {
	var cfg zap.Config
	b, err := json.Marshal(lcfg)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		panic(err)
	}

	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = nil

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
