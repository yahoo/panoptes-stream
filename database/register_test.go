//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestRegister(t *testing.T) {
	var df Factory

	cfg := config.NewMockConfig()
	r := NewRegistrar(cfg.Logger())
	r.Register("influxdb", "influxdata.com", df)
	_, ok := r.GetDatabaseFactory("influxdb")
	assert.True(t, ok)
}
