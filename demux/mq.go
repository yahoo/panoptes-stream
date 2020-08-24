package demux

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

// MQ represents MQ
type MQ struct {
	sync.Mutex

	ctx           context.Context
	logger        *zap.Logger
	chMap         *extDSChanMap
	producer      *nsq.Producer
	batchSize     int
	drainInterval time.Duration
	batch         map[string][][]byte
	consumers     map[string]context.CancelFunc
}

type messageHandler struct {
	ch  telemetry.ExtDSChan
	err error
}

type noLogger struct{}

// NewMQ constructs MQ
func NewMQ(ctx context.Context, lg *zap.Logger, chMap *extDSChanMap) (*MQ, error) {
	// producer
	config := nsq.NewConfig()
	producer, _ := nsq.NewProducer("127.0.0.1:4150", config)
	producer.SetLogger(&noLogger{}, 0)

	if err := producer.Ping(); err != nil {
		return nil, err
	}

	m := &MQ{
		ctx:           ctx,
		logger:        lg,
		chMap:         chMap,
		producer:      producer,
		batchSize:     100,
		drainInterval: time.Duration(30),
		batch:         make(map[string][][]byte),
		consumers:     make(map[string]context.CancelFunc),
	}

	// bach drainer
	m.drainer()

	// consumers
	for _, topic := range m.chMap.list() {
		m.register(topic)
	}

	return m, nil
}

// publish produces ds to specified topic
func (m *MQ) publish(ds telemetry.ExtDataStore, topic string) {
	b, _ := json.Marshal(ds)
	m.batch[topic] = append(m.batch[topic], b)

	m.Lock()
	if len(m.batch[topic]) > m.batchSize {
		err := m.producer.MultiPublish(topic, m.batch[topic])
		if err != nil {
			m.logger.Error("demux.mq", zap.Error(err))
		}

		m.batch[topic] = m.batch[topic][:0]
	}
	m.Unlock()
}

func (m *MQ) drainer() {
	go func() {
		for {
			<-time.After(m.drainInterval * time.Second)

			for _, topic := range m.chMap.list() {
				m.Lock()
				if len(m.batch[topic]) > 0 {
					err := m.producer.MultiPublish(topic, m.batch[topic])
					if err != nil {
						m.logger.Error("demux.mq", zap.Error(err))
					}

					m.batch[topic] = m.batch[topic][:0]
				}
				m.Unlock()
			}

		}
	}()
}

func (m *MQ) register(topic string) {
	var ctx context.Context

	m.Lock()
	defer m.Unlock()

	ctx, m.consumers[topic] = context.WithCancel(m.ctx)
	m.batch[topic] = [][]byte{}
	go m.consumer(ctx, topic)
}

func (m *MQ) unregister(topic string) {
	m.Lock()
	defer m.Unlock()

	m.consumers[topic]()
	delete(m.batch, topic)
}

// update updates consumers
func (m *MQ) update() {
	topics := map[string]bool{}

	for _, topic := range m.chMap.list() {
		if _, ok := m.consumers[topic]; !ok {
			// new topic
			m.register(topic)
		}

		topics[topic] = true
	}

	for topic := range m.consumers {
		if _, ok := topics[topic]; !ok {
			// removed topic
			m.unregister(topic)
		}
	}
}

func (m *MQ) consumer(ctx context.Context, topic string) {
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(topic, "channel", config)
	if err != nil {
		m.logger.Error("demux.mq", zap.Error(err))
		return
	}

	consumer.SetLogger(&noLogger{}, 0)

	ch, ok := m.chMap.get(topic)
	if !ok {
		m.logger.Error("demux.mq", zap.String("topic", topic))
	}

	handler := &messageHandler{
		ch:  ch,
		err: errors.New("failed"),
	}

	consumer.AddConcurrentHandlers(handler, 2)
	consumer.ConnectToNSQD("127.0.0.1:4150")

	<-ctx.Done()
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	var ds telemetry.ExtDataStore
	json.Unmarshal(m.Body, &ds)

	select {
	case h.ch <- ds:
	default:
		return h.err
	}

	return nil
}

func (*noLogger) Output(int, string) error {
	return nil
}
