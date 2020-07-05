package main

import (
	"errors"
	"os"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/consul"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	cli "github.com/urfave/cli/v2"
)

type cmd struct {
	configFile string
	consul     bool
}

func getConfig() (config.Config, error) {
	var cfg config.Config

	cli, err := getCli()
	if err != nil {
		return nil, err
	}

	if cli.consul {
		cfg, err = consul.New("etc/consul.yaml")
		if err != nil {
			return nil, err
		}
	} else {
		cfg, err = yaml.New("etc/config.yaml")
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
		&cli.BoolFlag{
			Name:  "consul",
			Usage: "enable consul configuration management",
		},
	}

	app := &cli.App{
		Flags: flags,
		Action: func(c *cli.Context) error {
			cm = cmd{
				configFile: c.String("config"),
				consul:     c.Bool("consul"),
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
