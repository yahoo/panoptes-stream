//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package telemetry

import (
	"context"
	"crypto/tls"
	"errors"
	"math/rand"
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

// DeviceFilterOpts represents assigned filter options
type DeviceFilterOpts struct {
	sync.RWMutex
	filterOpts map[string]DeviceFilterOpt
}

// DeviceFilterOpt represents filter option
type DeviceFilterOpt func(config.Device) bool

type mdtCredentials struct {
	username string
	password string
}

type backoff struct {
	d    time.Duration
	last time.Time
}

func (b *backoff) reset() {
	rand.Seed(time.Now().UnixNano())

	b.last = time.Now()
	b.d = time.Duration(20 + rand.Intn(30))
}

func (b *backoff) next() time.Duration {
	// first call - bypass backoff
	if b.d == 0 {
		b.reset()
		return 0
	}

	// reset back off
	if time.Since(b.last).Seconds() > float64(30*60) {
		b.reset()
		return b.d
	}

	// limit backoff
	if time.Since(b.last).Seconds() < float64(5*60) {
		if b.d < time.Duration(5*60) {
			b.d += b.d * 50 / 100
			b.last = time.Now()
		}
	}

	return b.d
}

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

	if _, ok := t.devices[device.Host]; ok {
		t.logger.Error("subscribe", zap.String("msg", "device already has been subscribed"), zap.String("name", device.Host))
		return
	}

	t.devices[device.Host] = device
	ctx, t.register[device.Host] = context.WithCancel(t.ctx)
	t.metrics["devicesCurrent"].Inc()

	for service, sensors := range getSensors(device.Sensors) {
		go func(service string, sensors []*config.Sensor) {
			addr := net.JoinHostPort(device.Host, strconv.Itoa(device.Port))
			backoff := backoff{}

			for {
				select {
				case <-time.After(time.Second * backoff.next()):
					t.metrics["reconnectsTotal"].Inc()

				case <-ctx.Done():
					return
				}

				ctx, err := t.setCredentials(ctx, &device)
				if err != nil {
					t.logger.Error("subscribe", zap.String("event", "grpc.credentials"), zap.Error(err))
					continue
				}

				opts, err := t.getDialOpts(&device, service)
				if err != nil {
					t.logger.Error("subscribe", zap.String("event", "grpc.dialopts"), zap.Error(err))
					continue
				}

				conn, err := grpc.DialContext(ctx, addr, opts...)
				if err != nil {
					t.logger.Error("subscribe", zap.String("event", "grpc.dial"), zap.Error(err))
					continue
				}

				t.metrics["gRPConnCurrent"].Inc()
				t.logger.Info("subscribe", zap.String("event", "grpc.connect"), zap.String("host", device.Host), zap.String("service", service))

				new, _ := t.telemetryRegistrar.GetNMIFactory(service)
				nmi := new(t.logger, conn, sensors, t.outChan)
				err = nmi.Start(ctx)

				conn.Close()
				t.metrics["gRPConnCurrent"].Dec()

				if err != nil {
					t.logger.Warn("subscribe", zap.String("event", "nmi"), zap.Error(err), zap.String("host", device.Host), zap.String("service", service))
				} else {
					t.logger.Warn("subscribe", zap.String("event", "grpc.terminate"), zap.String("host", device.Host), zap.String("service", service))
				}

			}
		}(service, sensors)
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
	if t.cfg.Global().Shards.Enabled && len(t.deviceFilterOpts.getOpts()) < 1 {
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

func (t *Telemetry) getDialOpts(device *config.Device, service string) ([]grpc.DialOption, error) {
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

	if service == "cisco.mdt" {
		creds := mdtCredentials{username: device.Username, password: device.Password}
		opts = append(opts, grpc.WithPerRPCCredentials(creds))
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
		username, password = dOptions.Username, dOptions.Password
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

func (m mdtCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"username": m.username,
		"password": m.password,
	}, nil
}

func (mdtCredentials) RequireTransportSecurity() bool {
	return false
}
