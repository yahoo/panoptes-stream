package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"sync/atomic"
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
	logger             *zap.Logger
	outChan            ExtDSChan
	telemetryRegistrar *Registrar
	status             *Status
	informer           chan struct{}
}

type Status struct {
	TotalDevices     uint64
	ConnectedDevices uint64
	Reconnect        uint64
}

type delta struct {
	add []config.Device
	del []config.Device
	mod []config.Device
}

// New creates a new telemetry
func New(ctx context.Context, cfg config.Config, tr *Registrar, outChan ExtDSChan) *Telemetry {
	return &Telemetry{
		ctx:                ctx,
		cfg:                cfg,
		logger:             cfg.Logger(),
		register:           make(map[string]context.CancelFunc),
		informer:           make(chan struct{}, 1),
		outChan:            outChan,
		status:             &Status{},
		telemetryRegistrar: tr,
	}
}

func (t *Telemetry) subscribe(device config.Device) {
	var (
		addr string
		ctx  context.Context
		gCfg *config.Global
	)

	ctx, t.register[device.Host] = context.WithCancel(t.ctx)
	atomic.AddUint64(&t.status.TotalDevices, 1)

	gCfg = t.cfg.Global()

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
					t.logger.Error("diap options", zap.Error(err))
				}

				if len(device.Username) > 0 && len(device.Password) > 0 {
					ctx = metadata.AppendToOutgoingContext(ctx,
						"username", device.Username, "password", device.Password)
				}

				conn, err := grpc.Dial(addr, opts...)
				if err != nil {
					t.logger.Error("connect to device", zap.Error(err))
				} else {
					atomic.AddUint64(&t.status.ConnectedDevices, 1)

					new, _ := t.telemetryRegistrar.GetNMIFactory(sName)
					nmi := new(t.logger, conn, sensors, t.outChan)
					err = nmi.Start(ctx)

					if err != nil {
						atomic.AddUint64(&t.status.ConnectedDevices, ^uint64(0))
						t.logger.Warn("nmi start error", zap.Error(err))
					}
				}

				select {
				case <-time.After(time.Second * 30):
					atomic.AddUint64(&t.status.Reconnect, 1)
				case <-ctx.Done():
					t.logger.Info("unsubscribed", zap.String("host", device.Host),
						zap.String("service", sName))
					return
				}
			}
		}(sName, sensors)
	}
}

func (t *Telemetry) unsubscribe(device config.Device) {
	t.register[device.Host]()
	delete(t.register, device.Host)
	atomic.AddUint64(&t.status.TotalDevices, ^uint64(0))
}

// Start subscribe configured devices
func (t *Telemetry) Start() {
	for _, device := range t.cfg.Devices() {
		t.subscribe(device)
	}
}

func (t *Telemetry) Update(devices map[string]config.Device) {
	newDevices := make(map[string]config.Device)
	delta := new(delta)

	for _, device := range t.cfg.Devices() {
		newDevices[device.Host] = device

		if _, ok := devices[device.Host]; !ok {
			delta.add = append(delta.add, device)
		} else {
			if ok := reflect.DeepEqual(devices[device.Host], device); !ok {
				delta.mod = append(delta.mod, device)
			}
		}

	}

	for host, device := range devices {
		if _, ok := newDevices[host]; !ok {
			delta.del = append(delta.del, device)
		}
	}

	for _, device := range delta.add {
		t.subscribe(device)
	}

	for _, device := range delta.del {
		t.unsubscribe(device)
	}

	for _, device := range delta.mod {
		t.unsubscribe(device)
		t.subscribe(device)
	}
}

func (t *Telemetry) GetStatus() *Status {
	return t.status
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
