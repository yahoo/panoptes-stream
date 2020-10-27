//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package demux

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/database"
	"github.com/yahoo/panoptes-stream/producer"
	"github.com/yahoo/panoptes-stream/register"
	"github.com/yahoo/panoptes-stream/telemetry"
)

var cfg = config.NewMockConfig()

func TestStartErrors(t *testing.T) {
	var (
		outChan = make(telemetry.ExtDSChan, 2)
		inChan  = make(telemetry.ExtDSChan, 2)
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := New(ctx, cfg, nil, nil, inChan)
	d.chMap.add("test", outChan)
	d.Start()

	cfg.LogOutput.Reset()

	inChan <- telemetry.ExtDataStore{Output: "test1::test"}

	e := ""
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Microsecond)
		if len(cfg.LogOutput.String()) > 0 {
			e = cfg.LogOutput.Unmarshal()["error"]
			break
		}
	}

	assert.Equal(t, "channel not found", e)

	inChan <- telemetry.ExtDataStore{Output: "test1"}

	e = ""
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Microsecond)
		if len(cfg.LogOutput.String()) > 0 {
			e = cfg.LogOutput.Unmarshal()["error"]
			break
		}
	}

	assert.Equal(t, "output not found", e)

}

func TestRegisteration(t *testing.T) {
	inChan := make(telemetry.ExtDSChan, 2)

	ctx := context.Background()
	cfg.LogOutput.Reset()

	// producer
	producerRegistrar := producer.NewRegistrar(cfg.Logger())
	register.Producer(producerRegistrar)

	// database
	databaseRegistrar := database.NewRegistrar(cfg.Logger())
	register.Database(databaseRegistrar)

	d := New(ctx, cfg, producerRegistrar, databaseRegistrar, inChan)

	// not exist
	p := config.Producer{
		Name:    "notexist",
		Service: "notexist",
	}
	err := d.subscribeProducer(p)
	assert.Error(t, err)

	// subscribe
	p = config.Producer{
		Name:    "console",
		Service: "console",
	}
	err = d.subscribeProducer(p)
	assert.NoError(t, err)
	assert.Equal(t, p, d.producers["console"])
	err = d.subscribeProducer(p)
	assert.Error(t, err)

	// duplicate subscription

	// unsubscribe
	d.unsubscribeProducer(p)

	// already unsubscribe
	d.unsubscribeProducer(p)

	// not exist
	db := config.Database{
		Name:    "notexist",
		Service: "notexist",
	}
	err = d.subscribeDatabase(db)
	assert.Error(t, err)

	// subscribe
	db = config.Database{
		Name:    "influx1",
		Service: "influxdb",
	}
	err = d.subscribeDatabase(db)
	assert.NoError(t, err)
	// duplicate subscription
	err = d.subscribeDatabase(db)
	assert.Error(t, err)

	// unsubscribe
	d.unsubscribeDatabase(db)

	// already unsubscribe
	d.unsubscribeDatabase(db)
}

func TestUpdateDatabase(t *testing.T) {
	cfg.MDatabases = []config.Database{}
	d := &Demux{
		cfg: cfg,
		databases: map[string]config.Database{
			"influx01": {
				Name:    "influx01",
				Service: "influxdb",
			},
		},
	}
	_, cancel := context.WithCancel(context.Background())
	d.register = map[string]context.CancelFunc{
		"influx01": cancel,
	}
	ch := make(telemetry.ExtDSChan)
	d.chMap = &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)}
	d.chMap.add("influx01", ch)

	d.updateDatabase()
	assert.Equal(t, 0, len(d.databases))
}

func TestUpdateProducer(t *testing.T) {
	cfg.MProducers = []config.Producer{}
	d := &Demux{
		cfg: cfg,
		producers: map[string]config.Producer{
			"kafka01": {
				Name:    "kafka01",
				Service: "kafka",
			},
		},
	}
	_, cancel := context.WithCancel(context.Background())
	d.register = map[string]context.CancelFunc{
		"kafka01": cancel,
	}
	ch := make(telemetry.ExtDSChan)
	d.chMap = &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)}
	d.chMap.add("kafka01", ch)

	d.updateProducer()
	assert.Equal(t, 0, len(d.producers))
}

func BenchmarkDemux(b *testing.B) {
	var (
		outChan = make(telemetry.ExtDSChan, 1)
		inChan  = make(telemetry.ExtDSChan, 1)
	)

	ctx := context.Background()
	cfg := config.NewMockConfig()
	d := New(ctx, cfg, nil, nil, inChan)
	d.chMap.add("test", outChan)
	go d.Start()

	for i := 0; i < b.N; i++ {
		inChan <- telemetry.ExtDataStore{Output: "test::test"}
		<-outChan
	}
}
