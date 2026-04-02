package cmd

import (
	"fmt"
	"github.com/lmxdawn/wallet/server"
	"github.com/urfave/cli"
	"os"
)

func Run(isSwag bool) {
	app := cli.NewApp()
	app.Name = "wallet"
	app.Usage = "wallet -c config/config.yml"
	printVersion := false
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "conf, c",
			Value: "config/dev.yml",
			Usage: "配置文件路径，如 config/dev.yml（本地开发）或 config/config-example.yml（含 BTC 等示例）",
		},
		cli.BoolFlag{
			Name:        "version, v",
			Required:    false,
			Usage:       "-v",
			Destination: &printVersion,
		},
	}

	app.Action = func(c *cli.Context) error {

		if printVersion {
			fmt.Printf("{%#v}", GetVersion())
			return nil
		}

		conf := c.String("conf")
		server.Start(isSwag, conf)

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
