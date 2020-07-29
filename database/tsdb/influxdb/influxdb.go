package influxdb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb/pkg/escape"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

type InfluxDB struct {
	ctx    context.Context
	ch     telemetry.ExtDSChan
	logger *zap.Logger
	cfg    config.Database
}

type influxDBConfig struct {
	Server     string
	Bucket     string
	Org        string
	Token      string
	BatchSize  uint
	MaxRetries uint
	Timeout    uint

	TLSConfig config.TLSConfig
}

func New(ctx context.Context, cfg config.Database, lg *zap.Logger, inChan telemetry.ExtDSChan) database.Database {
	return &InfluxDB{
		ctx:    ctx,
		cfg:    cfg,
		ch:     inChan,
		logger: lg,
	}
}

func (i *InfluxDB) Start() {
	config, err := i.getConfig()
	if err != nil {
		i.logger.Error("influxdb", zap.Error(err))
		os.Exit(1)
	}

	client, err := i.getClient(config)
	if err != nil {
		i.logger.Error("influxdb", zap.Error(err))
		os.Exit(1)
	}

	writeApi := client.WriteApi(config.Org, config.Bucket)

	i.logger.Info("influxdb", zap.String("name", i.cfg.Name), zap.String("server", config.Server), zap.String("bucket", config.Bucket))

	for {
		select {
		case v, ok := <-i.ch:
			if !ok {
				break
			}

			line, err := getLineProtocol(v)
			if err != nil {
				i.logger.Error("influxdb", zap.Error(err), zap.String("output", v.Output))
			}

			writeApi.WriteRecord(line)

		case <-i.ctx.Done():
			i.logger.Info("influxdb", zap.String("msg", "database has been terminated"), zap.String("name", i.cfg.Name))
			return
		}
	}

}

func getLineProtocol(v telemetry.ExtDataStore) (string, error) {
	out := strings.Split(v.Output, "::")
	if len(out) < 2 {
		return "", errors.New("invalid output")
	}

	buf := bytes.NewBufferString(out[1])
	buf.WriteRune(',')
	buf.WriteString("prefix=" + v.DS["prefix"].(string))
	buf.WriteRune(',')
	buf.WriteString("system_id=" + v.DS["system_id"].(string))
	for k, v := range v.DS["labels"].(map[string]string) {
		buf.WriteRune(',')
		buf.WriteString(escape.String(k) + "=" + v)
	}
	buf.WriteRune(' ')
	buf.WriteString(escape.String(v.DS["key"].(string)) + "=" + getValueString(v.DS["value"]))
	buf.WriteRune(' ')
	buf.WriteString(getValueString(v.DS["timestamp"]))

	return buf.String(), nil
}

func (i *InfluxDB) getClient(config *influxDBConfig) (influxdb2.Client, error) {
	opts := influxdb2.DefaultOptions()

	if config.TLSConfig.Enabled {
		tls, err := secret.GetTLSConfig(&config.TLSConfig)
		if err != nil {
			return nil, err
		}
		opts = opts.SetTlsConfig(tls)
	}

	token, err := getToken(config.Token)
	if err != nil {
		return nil, err
	}

	if config.BatchSize != 0 {
		opts.SetBatchSize(config.BatchSize)
	}

	if config.MaxRetries != 0 {
		opts.SetMaxRetries(config.MaxRetries)
	}

	if config.Timeout != 0 {
		opts.SetHttpRequestTimeout(config.Timeout)
	}

	client := influxdb2.NewClientWithOptions(config.Server, token, opts)

	return client, nil
}

func (i *InfluxDB) getConfig() (*influxDBConfig, error) {
	config := new(influxDBConfig)
	b, err := json.Marshal(i.cfg.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, config)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_database_" + i.cfg.Name
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func getToken(tokenConfig string) (string, error) {
	sType, path, ok := secret.ParseRemoteSecretInfo(tokenConfig)
	if !ok {
		return tokenConfig, nil
	}

	secrets, err := secret.GetCredentials(sType, path)
	if err != nil {
		return tokenConfig, err
	}

	if token, ok := secrets["token"]; ok {
		return token, nil
	}

	return tokenConfig, errors.New("token not found")
}

func getValueString(value interface{}) string {
	switch v := value.(type) {
	case uint64, uint32, uint16, uint8, uint,
		int64, int32, int16, int8, int:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case string:
		return fmt.Sprintf("\"%s\"", escape.String(v))
	}

	return ""
}
