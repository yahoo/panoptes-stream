//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCli(t *testing.T) {
	args := os.Args[0:1]
	args = append(args, "-config=/etc/panoptes/config.yaml")
	cli, err := getCli(args)
	assert.NoError(t, err)
	assert.Equal(t, "/etc/panoptes/config.yaml", cli.configFile)

	args = os.Args[0:1]
	args = append(args, "-consul=-")
	cli, err = getCli(args)
	assert.NoError(t, err)
	assert.Equal(t, "-", cli.consul)

	args = os.Args[0:1]
	args = append(args, "-etcd=-")
	cli, err = getCli(args)
	assert.NoError(t, err)
	assert.Equal(t, "-", cli.etcd)

	args = os.Args[0:1]
	_, err = getCli(args)
	assert.Error(t, err)
}

func TestGetConfig(t *testing.T) {
	args := os.Args[0:1]
	args = append(args, "-config=-")
	_, err := getConfig(args)
	assert.Error(t, err)

	args = os.Args[0:1]
	args = append(args, "-consul=-")
	_, err = getConfig(args)
	assert.Error(t, err)

	args = os.Args[0:1]
	args = append(args, "-etcd=-")
	_, err = getConfig(args)
	assert.Error(t, err)
}
