package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestRegister(t *testing.T) {
	var df DatabaseFactory

	cfg := config.NewMockConfig()
	r := NewRegistrar(cfg.Logger())
	r.Register("influxdb", "influxdata.com", df)
	_, ok := r.GetDatabaseFactory("influxdb")
	assert.True(t, ok)
}
