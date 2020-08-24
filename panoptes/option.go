package main

import (
	"errors"
	"os"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/consul"
	"git.vzbuilders.com/marshadrad/panoptes/config/etcd"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	cli "github.com/urfave/cli/v2"
)

type cmd struct {
	configFile string
	consul     string
	etcd       string
}

func getConfig() (config.Config, error) {
	var cfg config.Config

	cli, err := getCli()
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

func getCli() (*cmd, error) {
	cm := cmd{}

	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "path to a file in yaml format to read configuration",
		},
		&cli.StringFlag{
			Name:  "consul",
			Usage: "enable consul configuration management",
		},
		&cli.StringFlag{
			Name:  "etcd",
			Usage: "enable etcd configuration management",
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

	err := app.Run(os.Args)

	return &cm, err
}
