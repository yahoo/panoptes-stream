package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type panoptes struct {
	register           map[string]context.CancelFunc
	cfg                config.Config
	lg                 *zap.Logger
	outChan            telemetry.ExtDSChan
	telemetryRegistrar *telemetry.Registrar
}

func NewPanoptes(cfg config.Config, lg *zap.Logger, tr *telemetry.Registrar, outChan telemetry.ExtDSChan) *panoptes {
	return &panoptes{
		register:           make(map[string]context.CancelFunc),
		cfg:                cfg,
		lg:                 lg,
		outChan:            outChan,
		telemetryRegistrar: tr,
	}
}

func (p *panoptes) subscribe(parentCtx context.Context, device config.Device) {
	var (
		addr string
		ctx  context.Context
		gCfg *config.Global
	)

	ctx, p.register[device.Host] = context.WithCancel(parentCtx)

	gCfg = p.cfg.Global()

	for sName, sensors := range device.Sensors {
		go func(sName string, sensors []*config.Sensor) {
			for {
				if device.Port > 0 {
					addr = net.JoinHostPort(device.Host, strconv.Itoa(device.Port))
				} else {
					addr = device.Host
				}

				opts, err := dialOpts(device, gCfg)
				if err != nil {
					p.lg.Error("diap options", zap.Error(err))
				}

				if len(device.Username) > 0 && len(device.Password) > 0 {
					ctx = metadata.AppendToOutgoingContext(ctx, "username", device.Username, "password", device.Password)
				}

				conn, err := grpc.Dial(addr, opts...)
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

func transportClientCreds(certFile, keyFile, caCertFile string) (credentials.TransportCredentials, error) {
	var caCertPool *x509.CertPool
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	if caCertFile != "" {
		caCert, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			return nil, err
		}

		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	tc := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	})

	return tc, nil
}

func dialOpts(device config.Device, gCfg *config.Global) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithUserAgent("Panoptes"))

	if gCfg.TLSCertFile != "" && gCfg.TLSKeyFile != "" {
		creds, err := transportClientCreds(
			gCfg.TLSCertFile,
			gCfg.TLSKeyFile,
			gCfg.CAFile,
		)

		if err != nil {
			return opts, err
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))

	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	return opts, nil
}
