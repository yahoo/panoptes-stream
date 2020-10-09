//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package nsq

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	gonsq "github.com/nsqio/go-nsq"
	"go.uber.org/zap"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/producer"
	"github.com/yahoo/panoptes-stream/telemetry"
)

type nsqConfig struct {
	Addr         string
	Topics       []string
	BatchSize    int
	BatchTimeout int
}

type noLogger struct{}

// NSQ represents nsq producer
type NSQ struct {
	ctx    context.Context
	cfg    config.Producer
	ch     telemetry.ExtDSChan
	logger *zap.Logger
}

// New constructs an instance of NSQ producer.
func New(ctx context.Context, cfg config.Producer, lg *zap.Logger, inChan telemetry.ExtDSChan) producer.Producer {
	return &NSQ{
		ctx:    ctx,
		cfg:    cfg,
		ch:     inChan,
		logger: lg,
	}
}

// Start sends the data to the different topics (fan-out).
func (n *NSQ) Start() {
	chMap := make(map[string]chan telemetry.DataStore)
	config, err := n.getConfig()
	if err != nil {
		n.logger.Fatal("nsq", zap.Error(err))
	}

	for _, topic := range config.Topics {
		chMap[topic] = make(chan telemetry.DataStore, 1000)

		go func(topic string, ch chan telemetry.DataStore) {
			err := n.start(config, ch, topic)
			if err != nil {
				n.logger.Error("nsq", zap.Error(err))
			}
		}(topic, chMap[topic])
	}

L:
	for {
		select {
		case v, ok := <-n.ch:
			if !ok {
				break L
			}

			topic := strings.Split(v.Output, "::")
			if len(topic) < 2 {
				n.logger.Error("nsq", zap.String("msg", "topic not found"), zap.String("output", v.Output))
				continue
			}

			if _, ok := chMap[topic[1]]; ok {
				chMap[topic[1]] <- v.DS
			} else {
				n.logger.Error("nsq", zap.String("msg", "topic not found"), zap.String("name", topic[1]))
			}

		case <-n.ctx.Done():
			n.logger.Info("nsq", zap.String("event", "terminate"))
			return
		}
	}
}

func (n *NSQ) start(config *nsqConfig, ch chan telemetry.DataStore, topic string) error {
	var (
		batch = make([][]byte, 0)
		flush = false
	)

	pConfig := gonsq.NewConfig()
	pConfig.UserAgent = "panoptes"
	pConfig.DialTimeout = 2 * time.Second
	producer, _ := gonsq.NewProducer(config.Addr, pConfig)
	producer.SetLogger(&noLogger{}, 0)

	if err := producer.Ping(); err != nil {
		return err
	}

	flushTicker := time.NewTicker(time.Second * time.Duration(config.BatchTimeout))

	for {
		select {
		case v := <-ch:
			b, _ := json.Marshal(v)
			batch = append(batch, b)

		case <-flushTicker.C:
			if len(batch) > 0 {
				flush = true
			} else {
				continue
			}

		case <-n.ctx.Done():
			n.logger.Info("nsq", zap.String("event", "terminate"), zap.String("topic", topic))
			producer.MultiPublish(topic, batch)
			producer.Stop()
			return nil
		}

		if len(batch) == config.BatchSize || flush {
			for n.ctx.Err() == nil {
				err := producer.MultiPublish(topic, batch)
				if err != nil {
					n.logger.Error("nsq", zap.String("event", "publish"), zap.Error(err))

					// backoff
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

func (n *NSQ) getConfig() (*nsqConfig, error) {
	conf := new(nsqConfig)
	b, err := json.Marshal(n.cfg.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	prefix := "panoptes_producer_" + n.cfg.Name
	err = envconfig.Process(prefix, conf)
	if err != nil {
		return nil, err
	}

	config.SetDefault(&conf.BatchSize, 1000)
	config.SetDefault(&conf.BatchTimeout, 1)

	return conf, nil
}

func (*noLogger) Output(int, string) error {
	return nil
}
