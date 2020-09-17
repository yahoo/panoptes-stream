//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"errors"

	cli "github.com/urfave/cli/v2"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/consul"
	"git.vzbuilders.com/marshadrad/panoptes/config/etcd"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
)

type cmd struct {
	configFile string
	consul     string
	etcd       string
}

func getConfig(args []string) (config.Config, error) {
	var cfg config.Config

	cli, err := getCli(args)
	if err != nil {
		return nil, err
	}

	if len(cli.consul) > 0 {
		cfg, err = consul.New(cli.consul)
		if err != nil {
			return nil, err
		}
	} else if len(cli.etcd) > 0 {
		cfg, err = etcd.New(cli.etcd)
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = yaml.New(cli.configFile)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func getCli(args []string) (*cmd, error) {
	cm := cmd{}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "path to a file in yaml format to read configuration",
		},
		&cli.StringFlag{
			Name:  "consul",
			Usage: "enable consul configuration management (path to a file in yaml format or -)",
		},
		&cli.StringFlag{
			Name:  "etcd",
			Usage: "enable etcd configuration management (path to a file in yaml format or -)",
		},
	}

	app := &cli.App{
		Name:  "Panoptes Streaming",
		Usage: "A cloud native distributed streaming network telemetry",
		Flags: flags,
		Action: func(c *cli.Context) error {
			cm = cmd{
				configFile: c.String("config"),
				consul:     c.String("consul"),
				etcd:       c.String("etcd"),
			}

			if c.NumFlags() < 1 {
				cli.ShowAppHelp(c)
				return errors.New("configuration not specified")
			}

			return nil
		},
	}

	err := app.Run(args)

	return &cm, err
}
