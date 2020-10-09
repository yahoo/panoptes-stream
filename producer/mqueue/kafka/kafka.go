//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/gzip"
	"github.com/segmentio/kafka-go/lz4"
	"github.com/segmentio/kafka-go/snappy"
	"go.uber.org/zap"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/producer"
	"github.com/yahoo/panoptes-stream/secret"
	"github.com/yahoo/panoptes-stream/telemetry"
)

type kafkaConfig struct {
	Brokers       []string
	Topics        []string
	BatchSize     int
	BatchTimeout  int
	MaxAttempts   int
	QueueCapacity int
	KeepAlive     int
	IOTimeout     int
	Compression   string

	TLSConfig config.TLSConfig
}

// Kafka represents Kafka Segment.io
type Kafka struct {
	ctx    context.Context
	cfg    config.Producer
	ch     telemetry.ExtDSChan
	logger *zap.Logger
}

// New constructs an instance of kafka producer.
func New(ctx context.Context, cfg config.Producer, lg *zap.Logger, inChan telemetry.ExtDSChan) producer.Producer {

	return &Kafka{
		ctx:    ctx,
		cfg:    cfg,
		ch:     inChan,
		logger: lg,
	}
}

// Start sends the data to the different topics (fan-out).
func (k *Kafka) Start() {
	chMap := make(map[string]chan telemetry.DataStore)
	config, err := k.getConfig()
	if err != nil {
		k.logger.Fatal("kafka", zap.Error(err))
	}

	for _, topic := range config.Topics {
		chMap[topic] = make(chan telemetry.DataStore, 1000)

		go func(topic string, ch chan telemetry.DataStore) {
			err := k.start(config, ch, topic)
			if err != nil {
				k.logger.Error("kafka", zap.Error(err))
			}
		}(topic, chMap[topic])
	}

L:
	for {
		select {
		case v, ok := <-k.ch:
			if !ok {
				break L
			}

			topic := strings.Split(v.Output, "::")
			if len(topic) < 2 {
				k.logger.Error("kafka", zap.String("msg", "topic not found"), zap.String("output", v.Output))
				continue
			}

			if _, ok := chMap[topic[1]]; ok {
				chMap[topic[1]] <- v.DS
			} else {
				k.logger.Error("kafka", zap.String("msg", "topic not found"), zap.String("name", topic[1]))
			}

		case <-k.ctx.Done():
			k.logger.Info("kafka", zap.String("event", "terminate"), zap.String("brokers", strings.Join(config.Brokers, ",")))
			return
		}
	}

}

func (k *Kafka) start(config *kafkaConfig, ch chan telemetry.DataStore, topic string) error {
	var (
		batch = make([]kafka.Message, 0, config.BatchSize)
		flush = false
	)

	flushTicker := time.NewTicker(time.Second * time.Duration(config.BatchTimeout))

	cfg, err := k.getWriterConfig(config, topic)
	if err != nil {
		return err
	}

	w := kafka.NewWriter(cfg)

	k.logger.Info("kafka", zap.String("name", k.cfg.Name), zap.String("brokers", strings.Join(config.Brokers, ",")), zap.String("topic", topic))

	for {
		select {
		case v := <-ch:
			b, err := json.Marshal(v)
			if err != nil {
				k.logger.Error("kafka", zap.Error(err))
				continue
			}

			batch = append(batch, kafka.Message{Value: b})

		case <-flushTicker.C:
			if len(batch) > 0 {
				flush = true
			} else {
				continue
			}

		case <-k.ctx.Done():
			k.logger.Info("kafka", zap.String("event", "terminate"), zap.String("topic", topic))
			w.WriteMessages(k.ctx, batch...)
			w.Close()
			return nil

		}

		if len(batch) == config.BatchSize || flush {
			for k.ctx.Err() == nil {
				err := w.WriteMessages(k.ctx, batch...)
				if err != nil {
					k.logger.Error("kafka", zap.String("event", "write"), zap.Error(err))

					// extra backoff
					time.Sleep(1 * time.Second)
					continue
				}

				break
			}

			flush = false
			batch = batch[:0]
		}
	}
}

func (k *Kafka) getConfig() (*kafkaConfig, error) {
	conf := new(kafkaConfig)
	b, err := json.Marshal(k.cfg.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_producer_" + k.cfg.Name
	err = envconfig.Process(prefix, conf)
	if err != nil {
		return nil, err
	}

	config.SetDefault(&conf.BatchSize, 1000)
	config.SetDefault(&conf.BatchTimeout, 1)

	return conf, nil
}

func (k *Kafka) getWriterConfig(config *kafkaConfig, topic string) (kafka.WriterConfig, error) {
	var err error

	cfg := kafka.WriterConfig{
		Brokers:       config.Brokers,
		Topic:         topic,
		Balancer:      &kafka.LeastBytes{},
		MaxAttempts:   config.MaxAttempts,
		QueueCapacity: config.QueueCapacity,
		ReadTimeout:   time.Duration(config.IOTimeout) * time.Second,
		WriteTimeout:  time.Duration(config.IOTimeout) * time.Second,
		Dialer: &kafka.Dialer{
			ClientID:  "panoptes",
			Timeout:   time.Duration(config.IOTimeout) * time.Second,
			KeepAlive: time.Duration(config.IOTimeout) * time.Second,
			DualStack: true,
		},
	}

	if config.TLSConfig.Enabled {
		cfg.Dialer.TLS, err = secret.GetTLSConfig(&config.TLSConfig)
		if err != nil {
			return cfg, err
		}
	}

	switch config.Compression {
	case "gzip":
		cfg.CompressionCodec = gzip.NewCompressionCodec()
	case "snappy":
		cfg.CompressionCodec = snappy.NewCompressionCodec()
	case "lz4":
		cfg.CompressionCodec = lz4.NewCompressionCodec()
	}

	err = cfg.Validate()

	return cfg, err
}
