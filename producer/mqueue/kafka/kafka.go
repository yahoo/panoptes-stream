package kafka

import (
	"context"
	"encoding/json"
	"strings"

	kafka "github.com/segmentio/kafka-go"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

type Kafka struct {
	cfg config.Producer
	ch  telemetry.ExtDSChan
	lg  *zap.Logger
}

func New(cfg config.Producer, lg *zap.Logger, inChan telemetry.ExtDSChan) producer.Producer {
	return &Kafka{cfg: cfg, lg: lg, ch: inChan}
}

func (k *Kafka) Start() {
	m := make(map[string]*kafka.Writer)
	brokers := convArray(k.cfg.Config["brokers"])
	topics := convArray(k.cfg.Config["topics"])

	for _, topic := range topics {
		m[topic] = kafka.NewWriter(kafka.WriterConfig{
			Brokers:  brokers,
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		})
	}

	for {
		v, ok := <-k.ch
		if !ok {
			break
		}

		out := strings.Split(v.Output, "::")
		if len(out) < 2 {
			k.lg.Error("wrong output", zap.String("output", v.Output))
			continue
		}
		msg, err := json.Marshal(v.DS)
		if err != nil {
			k.lg.Error("marshal ds", zap.Error(err))
			continue
		}

		err = m[out[1]].WriteMessages(context.Background(), kafka.Message{Value: msg})
		if err != nil {
			k.lg.Error("write message", zap.Error(err))
		}
	}
}

func convArray(i interface{}) []string {
	var strArray []string
	ifArray := i.([]interface{})
	for _, v := range ifArray {
		strArray = append(strArray, v.(string))
	}

	return strArray
}