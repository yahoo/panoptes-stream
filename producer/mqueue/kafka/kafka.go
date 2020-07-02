package kafka

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

type kafkaConfig struct {
	Brokers []string
	Topics  []string

	BatchSize int
}

type Kafka struct {
	ctx context.Context
	cfg kafkaConfig
	ch  telemetry.ExtDSChan
	lg  *zap.Logger
}

func New(ctx context.Context, cfg config.Producer, lg *zap.Logger, inChan telemetry.ExtDSChan) producer.Producer {
	kConfig := convConfig(cfg.Config)
	return &Kafka{
		ctx: ctx,
		cfg: kConfig,
		lg:  lg,
		ch:  inChan,
	}
}

func (k *Kafka) Start() {
	chMap := make(map[string]chan telemetry.DataStore)

	for _, topic := range k.cfg.Topics {
		chMap[topic] = make(chan telemetry.DataStore, 1)

		go func(topic string) {
			k.start(chMap[topic], topic)
		}(topic)
	}

	for {
		select {
		case v, ok := <-k.ch:
			if !ok {
				break
			}

			topic := strings.Split(v.Output, "::")
			if len(topic) < 2 {
				k.lg.Error("topic not found", zap.String("output", v.Output))
				continue
			}

			chMap[topic[1]] <- v.DS

		case <-k.ctx.Done():
			k.lg.Info("kafka fanout has been terminated",
				zap.String("brokers", strings.Join(k.cfg.Brokers, ",")))
			return
		}
	}

}

func (k *Kafka) start(ch chan telemetry.DataStore, topic string) {
	batch := make([]kafka.Message, 0, k.cfg.BatchSize)
	flushTicker := time.NewTicker(time.Second * 10)
	flush := false

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  k.cfg.Brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})

	k.lg.Info("kafka producer set up", zap.String("brokers",
		strings.Join(k.cfg.Brokers, ",")), zap.String("topic", topic))

	for {
		select {
		case v, ok := <-ch:
			if !ok {
				break
			}

			b, err := json.Marshal(v)
			if err != nil {
				k.lg.Error("dataset marshal failed", zap.Error(err))
			}

			batch = append(batch, kafka.Message{Value: b})
		case <-flushTicker.C:
			flush = true
		case <-k.ctx.Done():
			k.lg.Info("kafka has been terminated", zap.String("topic", topic))
			return

		}

		if len(batch) == k.cfg.BatchSize || flush {
			err := w.WriteMessages(k.ctx, batch...)
			if err != nil {
				k.lg.Error("kafka write message failed", zap.Error(err))
			}

			flush = false
			batch = nil
		}
	}
}

func convConfig(c interface{}) kafkaConfig {
	kc := kafkaConfig{}
	b, _ := json.Marshal(c)
	json.Unmarshal(b, &kc)
	return kc
}
