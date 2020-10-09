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
	"github.com/yahoo/panoptes-stream/telemetry"
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
