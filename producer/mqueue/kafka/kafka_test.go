//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/yahoo/panoptes-stream/config"
	pb "github.com/yahoo/panoptes-stream/producer/proto"
	"github.com/yahoo/panoptes-stream/telemetry"
	"google.golang.org/protobuf/proto"
)

var mockConfig = config.NewMockConfig()

func TestKafkaStart(t *testing.T) {
	seedBroker := sarama.NewMockBroker(t, 1)
	defer seedBroker.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := config.Producer{
		Name:    "kafka01",
		Service: "kafka",
		Config: map[string]interface{}{
			"Brokers":   []string{seedBroker.Addr()},
			"Topics":    []string{"topic1", "topic2"},
			"BatchSize": 1,
		},
	}

	ch := make(telemetry.ExtDSChan)

	producer := New(ctx, cfg, mockConfig.Logger(), ch)
	go producer.Start()

	time.Sleep(1 * time.Second)

	topics := seedBroker.History()[0].Request.(*sarama.MetadataRequest).Topics
	topics = append(topics, seedBroker.History()[1].Request.(*sarama.MetadataRequest).Topics...)

	assert.Len(t, seedBroker.History(), 2)
	assert.Contains(t, topics, "topic1")
	assert.Contains(t, topics, "topic2")
}

func TestBatchRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cfg := config.Producer{
		Name:    "kafka01",
		Service: "kafka",
		Config: map[string]interface{}{
			"Brokers":     []string{"127.0.0.1:9092"},
			"Topics":      []string{"topic1", "topic2"},
			"BatchSize":   1,
			"MaxAttempts": 1,
		},
	}

	ch := make(telemetry.ExtDSChan)
	mockConfig.LogOutput.Reset()

	producer := New(ctx, cfg, mockConfig.Logger(), ch)
	go producer.Start()

	ch <- telemetry.ExtDataStore{Output: "kafka01::topic1", DS: telemetry.DataStore{"labels": "test"}}

	time.Sleep(3 * time.Second)
	counter := 0
	for _, l := range mockConfig.LogOutput.UnmarshalSlice() {
		if v, ok := l["event"]; ok {
			if v == "write" {
				counter++
			}
		}
	}

	assert.Equal(t, 3, counter)
}

func TestPBMarshal(t *testing.T) {
	ds := telemetry.DataStore{
		"prefix":    "/foos/foo",
		"labels":    map[string]string{"l1": "v1"},
		"timestamp": int64(1610395484),
		"system_id": "core1.lax",
		"key":       "counter1",
		"value":     int(55),
	}

	// general error
	b, err := pbMarshal(ds)
	assert.NoError(t, err)
	assert.NotZero(t, len(b))

	// unmarshal
	m := pb.Panoptes{}
	proto.Unmarshal(b, &m)

	assert.Equal(t, "counter1", m.Key)
	assert.Equal(t, "core1.lax", m.SystemId)
	assert.Equal(t, "/foos/foo", m.Prefix)
	assert.Equal(t, int64(1610395484), m.Timestamp)
	assert.Equal(t, map[string]string{"l1": "v1"}, m.Labels)
	assert.Equal(t, "type.googleapis.com/google.protobuf.Int64Value", m.Value.TypeUrl)
	assert.Equal(t, m.Value.Value, []uint8{0x8, 0x37})

	// known types
	tt := []interface{}{"foo", int(5), int32(5), int64(5), uint(5), uint32(5), uint64(5), true, []byte{0x8}}

	for _, v := range tt {
		ds["value"] = v
		_, err = pbMarshal(ds)
		assert.NoError(t, err)
	}

	// unknown type
	ds["value"] = make(chan int)
	_, err = pbMarshal(ds)
	assert.Error(t, err)
}
