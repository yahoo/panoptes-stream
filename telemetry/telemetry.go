package telemetry

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
	"git.vzbuilders.com/marshadrad/panoptes/status"
)

// Telemetry represents telemetry
type Telemetry struct {
	register           map[string]context.CancelFunc
	devices            map[string]config.Device
	cfg                config.Config
	ctx                context.Context
	group              singleflight.Group
	logger             *zap.Logger
	outChan            ExtDSChan
	telemetryRegistrar *Registrar
	informer           chan struct{}
	deviceFilterOpts   DeviceFilterOpts
	metrics            map[string]status.Metrics
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

// New creates a new telemetry
func New(ctx context.Context, cfg config.Config, tr *Registrar, outChan ExtDSChan) *Telemetry {
	var metrics = make(map[string]status.Metrics)

	metrics["devicesCurrent"] = status.NewGauge("subscribed_devices", "")
	metrics["gRPConnCurrent"] = status.NewGauge("active_grpc_connections", "")
	metrics["reconnectsTotal"] = status.NewCounter("grpc_reconnects_total", "")

	status.Register(nil, metrics)

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
		metrics:            metrics,
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
	t.metrics["devicesCurrent"].Inc()

	for sName, sensors := range device.Sensors {
		go func(sName string, sensors []*config.Sensor) {

			addr := net.JoinHostPort(device.Host, strconv.Itoa(device.Port))
			ctx, err := t.setCredentials(ctx, &device)
			if err != nil {
				t.logger.Error("subscribe.creds", zap.Error(err))
			}

			opts, err := t.getDialOpts(&device)
			if err != nil {
				t.logger.Error("subscribe.dialOpts", zap.Error(err))
			}

			for {
				conn, err := grpc.DialContext(ctx, addr, opts...)
				if err != nil {
					t.logger.Error("subscribe.grpc", zap.Error(err))
				} else {
					t.metrics["gRPConnCurrent"].Inc()
					t.logger.Info("subscribe.grpc", zap.String("host", device.Host), zap.String("service", sName))

					new, _ := t.telemetryRegistrar.GetNMIFactory(sName)
					nmi := new(t.logger, conn, sensors, t.outChan)
					err = nmi.Start(ctx)

					conn.Close()
					t.metrics["gRPConnCurrent"].Dec()

					if err != nil {
						t.logger.Warn("nmi.start", zap.Error(err), zap.String("host", device.Host), zap.String("service", sName))
					} else {
						t.logger.Warn("subscribe.grpc", zap.String("msg", "terminated"), zap.String("host", device.Host), zap.String("service", sName))
					}
				}

				select {
				case <-time.After(time.Second * 30):
					t.metrics["reconnectsTotal"].Inc()

				case <-ctx.Done():
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
	t.metrics["devicesCurrent"].Dec()
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
			continue
		}

		if ok := reflect.DeepEqual(t.devices[device.Host], device); !ok {
			delta.mod = append(delta.mod, device)
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

	d.RLock()
	defer d.RUnlock()

	for _, opt := range d.filterOpts {
		opts = append(opts, opt)
	}

	return opts
}

func (t *Telemetry) getDialOpts(device *config.Device) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithUserAgent("Panoptes"))

	if device.TLSConfig.Enabled {
		creds, err := t.getTransportCredentials(device)
		if err != nil {
			return opts, err
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))

	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	return opts, nil
}

func (t *Telemetry) getTransportCredentials(device *config.Device) (credentials.TransportCredentials, error) {
	var (
		tlsConfig *config.TLSConfig
		gCfg      = t.cfg.Global()
	)

	if device.TLSConfig.CertFile != "" {
		tlsConfig = &device.TLSConfig
	} else if gCfg.DeviceOptions.TLSConfig.Enabled {
		tlsConfig = &gCfg.DeviceOptions.TLSConfig
	} else {
		return nil, errors.New("TLS is not available")
	}

	tc, err, _ := t.group.Do(tlsConfig.CertFile, func() (interface{}, error) {
		return secret.GetTLSConfig(tlsConfig)
	})
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tc.(*tls.Config)), nil
}

func (t *Telemetry) setCredentials(ctx context.Context, device *config.Device) (context.Context, error) {
	var (
		dOptions           = t.cfg.Global().DeviceOptions
		username, password string
	)
	// no username and password
	if device.Username == "" && dOptions.Username == "" {
		return ctx, nil
	}

	if device.Username != "" {
		username, password = device.Username, device.Password
	} else {
		username, password = dOptions.Username, device.Password
	}

	// remote username and password
	sType, path, ok := secret.ParseRemoteSecretInfo(username)
	if ok {
		tc, err, _ := t.group.Do(username, func() (interface{}, error) {
			secrets, err := secret.GetCredentials(sType, path)
			return secrets, err
		})
		if err != nil {
			return ctx, err
		}

		for u, p := range tc.(map[string]string) {
			ctx = metadata.AppendToOutgoingContext(ctx, "username", u, "password", p)
			return ctx, nil
		}

		return ctx, errors.New("credentials are not available at remote host")
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "username", username, "password", password)

	return ctx, nil
}
