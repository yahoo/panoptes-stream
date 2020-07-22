package influxdb

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb/pkg/escape"
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
	Database   string
	Token      string
	BatchSize  uint
	MaxRetries uint

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
	var (
		tagSet, fieldSet       []string
		measurement, timestamp string
		config                 = new(influxDBConfig)
	)

	i.logger.Info("influxdb set up", zap.String("name", i.cfg.Name),
		zap.String("server url", i.cfg.Config["server"].(string)))

	b, _ := json.Marshal(i.cfg.Config)
	json.Unmarshal(b, config)

	client, err := i.getClient(config)
	if err != nil {
		i.logger.Error("influxdb", zap.Error(err))
		os.Exit(1)
	}

	writeApi := client.WriteApi("", config.Database)

	for {
		select {
		case v, ok := <-i.ch:
			if !ok {
				break
			}

			out := strings.Split(v.Output, "::")
			if len(out) < 2 {
				i.logger.Error("influxdb.invalid", zap.String("output", v.Output))
				continue
			}

			measurement = out[1]

			for k, v := range v.DS {
				switch k {
				case "__prefix":
					tagSet = append(tagSet, fmt.Sprintf("prefix=%s", v.(string)))
				case "__labels":
					labels := v.(map[string]string)
					for k, v := range labels {
						tagSet = append(tagSet, fmt.Sprintf("%s=%s", escape.String(k), v))
					}
				case "__system_id":
					tagSet = append(tagSet, fmt.Sprintf("system_id=%s", v.(string)))
				case "__timestamp":
					timestamp = getValueString(v)
				default:
					fieldSet = append(fieldSet, fmt.Sprintf(" %s=%s", escape.String(k), getValueString(v)))
				}
			}

			line := fmt.Sprintf("%s,%s %s %s", measurement, strings.Join(tagSet, ","), strings.Join(fieldSet, ","), timestamp)
			writeApi.WriteRecord(line)

			tagSet = tagSet[:0]
			fieldSet = fieldSet[:0]

		case <-i.ctx.Done():
			i.logger.Info("database has been terminated", zap.String("name", i.cfg.Name))
			return
		}
	}

}

func (i *InfluxDB) getClient(config *influxDBConfig) (influxdb2.Client, error) {
	opts := influxdb2.DefaultOptions()

	if config.TLSConfig.CertFile != "" && !config.TLSConfig.Disabled {
		tls, err := secret.GetTLSConfig(&config.TLSConfig)
		if err != nil {
			return nil, err
		}
		opts = opts.SetTlsConfig(tls)
	}

	if config.BatchSize != 0 {
		opts.SetBatchSize(config.BatchSize)
	}

	if config.MaxRetries != 0 {
		opts.SetMaxRetries(config.MaxRetries)
	}

	client := influxdb2.NewClientWithOptions(config.Server, config.Token, opts)

	return client, nil
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
