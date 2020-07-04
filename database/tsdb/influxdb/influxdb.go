package influxdb

import (
	"context"
	"strings"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

type InfluxDB struct {
	ctx context.Context
	ch  telemetry.ExtDSChan
	lg  *zap.Logger
	cfg config.Database
}

func New(ctx context.Context, cfg config.Database, lg *zap.Logger, inChan telemetry.ExtDSChan) database.Database {
	return &InfluxDB{
		ctx: ctx,
		cfg: cfg,
		lg:  lg,
		ch:  inChan,
	}
}

func (i *InfluxDB) Start() {

	i.lg.Info("influxdb set up", zap.String("name", i.cfg.Name),
		zap.String("server url", i.cfg.Config["server"].(string)))

	for {
		select {
		case v, ok := <-i.ch:
			if !ok {
				break
			}

			out := strings.Split(v.Output, "::")
			if len(out) < 2 {
				i.lg.Error("wrong output", zap.String("output", v.Output))
				continue
			}

			v.DS.PrettyPrint(out[1])
		case <-i.ctx.Done():
			i.lg.Info("database has been terminated", zap.String("name", i.cfg.Name))
			return
		}
	}

}
