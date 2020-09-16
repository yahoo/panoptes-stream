//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package console

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

// Console represents console
// It's just print pretty metrics on the stdout or stderr for testing purpose
type Console struct {
	ch     telemetry.ExtDSChan
	logger *zap.Logger
}

// New returns a new console instance
func New(ctx context.Context, cfg config.Producer, lg *zap.Logger, inChan telemetry.ExtDSChan) producer.Producer {
	return &Console{ch: inChan, logger: lg}
}

// Start starts printing available metric
func (c *Console) Start() {
	for {
		v, ok := <-c.ch
		if !ok {
			break
		}

		out := strings.Split(v.Output, "::")
		if len(out) < 2 {
			c.logger.Error("wrong output", zap.String("output", v.Output))
			continue
		}

		PrettyPrint(v.DS, out[1])
	}
}

// PrettyPrint prints metrics on the stdout or stderr in pretty format
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

// Register registers console as a producer at producer registrar
func Register(producerRegistrar *producer.Registrar) {
	producerRegistrar.Register("console", "-", New)
}
