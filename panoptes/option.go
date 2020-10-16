//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package main

import (
	"errors"
	"fmt"
	"io"

	cli "github.com/urfave/cli/v2"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/config/consul"
	"github.com/yahoo/panoptes-stream/config/etcd"
	"github.com/yahoo/panoptes-stream/config/yaml"
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

	cli.AppHelpTemplate = `Panoptes Streaming

-config filename       path to a file in yaml format to read configuration
-consul filename or -  enable consul configuration management
-etcd   filename or -  enable etcd configuration management
-help, -h      show help
-version, -v   show version

In case of consul or etcd, if you set dash as argument, Panoptes assumes
they available at localhost with default configuration.
for more information visit https://github.com/yahoo/panoptes-stream

`

	cli.VersionFlag = &cli.BoolFlag{
		Name: "version", Aliases: []string{"v"},
		Usage: "print only the version",
	}

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("Panoptes Streaming Version: %s\n\n", c.App.Version)
		cli.OsExiter(0)
	}

	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		fmt.Fprint(w, templ)
		cli.OsExiter(0)
	}

	app := &cli.App{
		Version: config.GetVersion(),
		Flags:   flags,
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
