package console

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

type Console struct {
	ch telemetry.ExtDSChan
	lg *zap.Logger
}

func New(ctx context.Context, cfg config.Producer, lg *zap.Logger, inChan telemetry.ExtDSChan) producer.Producer {
	return &Console{ch: inChan}
}

func (c *Console) Start() {
	for {
		v, ok := <-c.ch
		if !ok {
			break
		}

		out := strings.Split(v.Output, "::")
		if len(out) < 2 {
			c.lg.Error("wrong output", zap.String("output", v.Output))
			continue
		}

		PrettyPrint(v.DS, out[1])
	}
}

func PrettyPrint(ds telemetry.DataStore, fdType string) error {
	b, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}

	if fdType == "stdout" {
		os.Stdout.Write(b)
	} else {
		os.Stderr.Write(b)
	}
	return nil
}

func Register(producerRegistrar *producer.Registrar) {
	producerRegistrar.Register("console", "-", New)
}
