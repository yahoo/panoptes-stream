//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/config"
)

func TestDiscoveryRegister(t *testing.T) {
	cfg := config.NewMockConfig()
	_, err := discoveryRegister(cfg)
	assert.NoError(t, err)

	cfg.Global().Discovery.Service = "consul"
	_, err = discoveryRegister(cfg)
	assert.Error(t, err)

	cfg.Global().Discovery.Service = "etcd"
	_, err = discoveryRegister(cfg)
	assert.Error(t, err)

	cfg.Global().Discovery.Service = "k8s"
	_, err = discoveryRegister(cfg)
	assert.Error(t, err)

	cfg.Global().Discovery.Service = "pseudo"
	_, err = discoveryRegister(cfg)
	assert.NoError(t, err)
}
