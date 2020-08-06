package demux

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

func TestStartErrors(t *testing.T) {
	var (
		outChan = make(telemetry.ExtDSChan, 2)
		inChan  = make(telemetry.ExtDSChan, 2)
	)

	ctx := context.Background()
	cfg := &config.MockConfig{}
	d := New(ctx, cfg, nil, nil, inChan)
	d.chMap.add("test", outChan)
	go d.Start()

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

func BenchmarkDemux(b *testing.B) {
	var (
		outChan = make(telemetry.ExtDSChan, 1)
		inChan  = make(telemetry.ExtDSChan, 1)
	)

	ctx := context.Background()
	cfg := &config.MockConfig{}
	d := New(ctx, cfg, nil, nil, inChan)
	d.chMap.add("test", outChan)
	go d.Start()

	for i := 0; i < b.N; i++ {
		inChan <- telemetry.ExtDataStore{Output: "test::test"}
		<-outChan
	}
}
