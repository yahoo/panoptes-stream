package register

import (
	"strings"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper"
	"go.uber.org/zap"
)

func RegisterVendor(lg *zap.Logger) {
	logger := lg.Named("telemetry-register")
	list := juniper.Register()
	logger.Info(strings.Join(list, ","))
}
