package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/status"
)

// Telemetry represents telemetry
type Telemetry struct {
	register           map[string]context.CancelFunc
	devices            map[string]config.Device
	cfg                config.Config
	ctx                context.Context
	logger             *zap.Logger
	outChan            ExtDSChan
	telemetryRegistrar *Registrar
	informer           chan struct{}
	deviceFilterOpts   DeviceFilterOpts
}

type delta struct {
	add []config.Device
	del []config.Device
	mod []config.Device
}

type DeviceFilterOpts struct {
	sync.RWMutex
	filterOpts map[string]DeviceFilterOpt
}

type DeviceFilterOpt func(config.Device) bool

var (
	metricCurrentDevice  = status.NewGauge("current_device", "")
	metricCurrentGRPConn = status.NewGauge("grpc_connection", "")
	metricTotalReconnect = status.NewCounter("total_reconnect", "")
)

func init() {
	status.Register(
		metricCurrentDevice,
		metricCurrentGRPConn,
		metricTotalReconnect,
	)
}

// New creates a new telemetry
func New(ctx context.Context, cfg config.Config, tr *Registrar, outChan ExtDSChan) *Telemetry {
	return &Telemetry{
		ctx:                ctx,
		cfg:                cfg,
		logger:             cfg.Logger(),
		register:           make(map[string]context.CancelFunc),
		deviceFilterOpts:   DeviceFilterOpts{filterOpts: make(map[string]DeviceFilterOpt)},
		devices:            make(map[string]config.Device),
		informer:           make(chan struct{}, 1),
		outChan:            outChan,
		telemetryRegistrar: tr,
	}
}

func (t *Telemetry) subscribe(device config.Device) {
	var ctx context.Context

	if len(device.Sensors) < 1 {
		return
	}

	if _, ok := t.devices[device.Host]; ok {
		t.logger.Error("device already subscribed", zap.String("name", device.Host))
		return
	}

	t.devices[device.Host] = device

	ctx, t.register[device.Host] = context.WithCancel(t.ctx)
	metricCurrentDevice.Inc()

	for sName, sensors := range device.Sensors {
		go func(sName string, sensors []*config.Sensor) {

			addr := net.JoinHostPort(device.Host, strconv.Itoa(device.Port))
			ctx = setCredentials(ctx, &device)
			opts, err := dialOpts(&device, t.cfg.Global())
			if err != nil {
				t.logger.Error("dial options", zap.Error(err))
			}

			for {
				conn, err := grpc.DialContext(ctx, addr, opts...)
				if err != nil {
					t.logger.Error("connect to device", zap.Error(err))
				} else {
					metricCurrentGRPConn.Inc()

					new, _ := t.telemetryRegistrar.GetNMIFactory(sName)
					nmi := new(t.logger, conn, sensors, t.outChan)
					err = nmi.Start(ctx)

					conn.Close()

					if err != nil {
						metricCurrentGRPConn.Dec()
						t.logger.Warn("nmi.start", zap.Error(err), zap.String("host", device.Host))
					}
				}

				select {
				case <-time.After(time.Second * 30):
					metricTotalReconnect.Inc()

				case <-ctx.Done():
					metricCurrentGRPConn.Dec()
					t.logger.Info("unsubscribed", zap.String("host", device.Host), zap.String("service", sName))
					return
				}
			}
		}(sName, sensors)
	}
}

func (t *Telemetry) unsubscribe(device config.Device) {
	t.register[device.Host]()
	delete(t.register, device.Host)
	delete(t.devices, device.Host)
	metricCurrentDevice.Dec()
}

// Start subscribes configured devices
func (t *Telemetry) Start() {
	for _, device := range t.GetDevices() {
		t.subscribe(device)
	}
}

// Update updates device subscriptions
// subscribe/unsubscribe/resubscribe devices
func (t *Telemetry) Update() {
	if t.cfg.Global().Shard.Enabled && len(t.deviceFilterOpts.getOpts()) < 1 {
		return
	}

	newDevices := make(map[string]config.Device)
	delta := new(delta)

	for _, device := range t.GetDevices() {
		newDevices[device.Host] = device

		if _, ok := t.devices[device.Host]; !ok {
			delta.add = append(delta.add, device)
		} else {
			if ok := reflect.DeepEqual(t.devices[device.Host], device); !ok {
				delta.mod = append(delta.mod, device)
			}
		}

	}

	for host, device := range t.devices {
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

// GetDevices returns devices based on the filters (if exist)
func (t *Telemetry) GetDevices() []config.Device {
	var filteredDevcies []config.Device

	if len(t.deviceFilterOpts.getOpts()) < 1 {
		return t.cfg.Devices()
	}

	for _, device := range t.cfg.Devices() {
		for _, filter := range t.deviceFilterOpts.getOpts() {
			if filter(device) {
				filteredDevcies = append(filteredDevcies, device)
			}
		}
	}

	return filteredDevcies
}

// AddFilterOpt adds filter option
func (t *Telemetry) AddFilterOpt(index string, filterOpt DeviceFilterOpt) {
	t.deviceFilterOpts.set(index, filterOpt)
}

// DelFilterOpt deletes filter option
func (t *Telemetry) DelFilterOpt(index string) {
	t.deviceFilterOpts.del(index)
}

func (d *DeviceFilterOpts) set(key string, value DeviceFilterOpt) {
	d.Lock()
	defer d.Unlock()
	d.filterOpts[key] = value
}

func (d *DeviceFilterOpts) del(key string) {
	d.Lock()
	defer d.Unlock()
	delete(d.filterOpts, key)
}

func (d *DeviceFilterOpts) getOpts() []DeviceFilterOpt {
	var opts []DeviceFilterOpt
	d.Lock()
	for _, opt := range d.filterOpts {
		opts = append(opts, opt)
	}
	d.Unlock()

	return opts
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

func dialOpts(device *config.Device, gCfg *config.Global) ([]grpc.DialOption, error) {
	var (
		opts      []grpc.DialOption
		tlsConfig *config.TLSConfig
	)

	opts = append(opts, grpc.WithUserAgent("Panoptes"))

	if device.TLSConfig.CertFile != "" && device.TLSConfig.KeyFile != "" {
		tlsConfig = &device.TLSConfig
	} else if gCfg.TLSConfig.CertFile != "" && gCfg.TLSConfig.KeyFile != "" {
		tlsConfig = &gCfg.TLSConfig
	}

	if tlsConfig != nil {
		creds, err := transportClientCreds(
			tlsConfig.CertFile,
			tlsConfig.KeyFile,
			tlsConfig.CAFile,
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

func setCredentials(ctx context.Context, device *config.Device) context.Context {
	if len(device.Username) > 0 && len(device.Password) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx,
			"username", device.Username, "password", device.Password)
	}

	return ctx
}
