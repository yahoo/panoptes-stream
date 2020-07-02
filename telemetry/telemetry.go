package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Telemetry represents telemetry
type Telemetry struct {
	register           map[string]context.CancelFunc
	cfg                config.Config
	ctx                context.Context
	lg                 *zap.Logger
	outChan            ExtDSChan
	telemetryRegistrar *Registrar
	informer           chan struct{}
}

type deltaDevices struct {
	add []config.Device
	del []config.Device
	mod []config.Device
}

// New creates a new telemetry
func New(ctx context.Context, cfg config.Config, lg *zap.Logger, tr *Registrar, outChan ExtDSChan) *Telemetry {
	return &Telemetry{
		ctx:                ctx,
		cfg:                cfg,
		lg:                 lg,
		register:           make(map[string]context.CancelFunc),
		informer:           make(chan struct{}, 1),
		outChan:            outChan,
		telemetryRegistrar: tr,
	}
}

func (p *Telemetry) subscribe(parentCtx context.Context, device config.Device) {
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
					ctx = metadata.AppendToOutgoingContext(ctx,
						"username", device.Username, "password", device.Password)
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

				select {
				case <-time.After(time.Second * 10):
				case <-ctx.Done():
					p.lg.Info("unsubscribed", zap.String("host", device.Host),
						zap.String("service", sName))
					return
				}
			}
		}(sName, sensors)
	}
}

func (p *Telemetry) unsubscribe(device config.Device) {
	cancel := p.register[device.Host]
	cancel()
}

func (p *Telemetry) Start(ctx context.Context) {
	p.ctx = ctx
	for _, device := range p.cfg.Devices() {
		p.subscribe(ctx, device)
	}
}

func (p *Telemetry) deltaDevices() *deltaDevices {
	oldDevices := make(map[string]config.Device)
	devices := make(map[string]config.Device)
	delta := new(deltaDevices)

	for _, device := range p.cfg.Devices() {
		oldDevices[device.Host] = device
	}

	p.cfg.Update()

	for _, device := range p.cfg.Devices() {
		devices[device.Host] = device

		if _, ok := oldDevices[device.Host]; !ok {
			delta.add = append(delta.add, device)
		} else {
			if ok := reflect.DeepEqual(oldDevices[device.Host], device); !ok {
				delta.mod = append(delta.mod, device)
			}
		}
	}

	for host, device := range oldDevices {
		if _, ok := devices[host]; !ok {
			delta.del = append(delta.del, device)
		}
	}

	return delta
}

func (p *Telemetry) Watcher() {
	ticker := time.NewTicker(time.Second * 60)
	go func() {
		for {
			<-ticker.C
			d := p.deltaDevices()
			for _, device := range d.add {
				p.subscribe(p.ctx, device)
			}

			for _, device := range d.del {
				p.unsubscribe(device)
			}

			for _, device := range d.mod {
				p.unsubscribe(device)
				p.subscribe(p.ctx, device)
			}
		}
	}()
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
