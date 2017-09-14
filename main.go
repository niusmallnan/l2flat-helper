package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/l2flat-helper/macsync"
	"github.com/rancher/l2flat-helper/setting"
	"github.com/urfave/cli"
)

var VERSION = "v0.0.0-dev"

func main() {
	app := cli.NewApp()
	app.Name = "l2flat-helper"
	app.Version = VERSION
	app.Usage = "Support l2-flat networking"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "listen",
			Value: ":10110",
		},
		cli.BoolFlag{
			Name:   "debug, d",
			EnvVar: "RANCHER_DEBUG",
		},
		cli.StringFlag{
			Name:   "metadata-address",
			Value:  setting.DefaultMetadataAddress,
			EnvVar: "RANCHER_METADATA_ADDRESS",
		},
	}
	app.Action = func(c *cli.Context) {
		if err := appMain(c); err != nil {
			logrus.Fatal(err)
		}
	}

	app.Run(os.Args)
}

func appMain(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	done := make(chan error)

	mc, err := metadata.NewClientAndWait(fmt.Sprintf(setting.MetadataURL, ctx.String("metadata-address")))
	if err != nil {
		return errors.Wrap(err, "Failed to create metadata client")
	}
	dc, err := client.NewEnvClient()
	if err != nil {
		return errors.Wrap(err, "Failed to create docker client")
	}

	err = macsync.Watch(mc, dc)
	if err != nil {
		return err
	}

	listenPort := ctx.String("listen")
	logrus.Debugf("About to start server and listen on port: %v", listenPort)
	go func() {
		done <- ListenAndServe(listenPort)
	}()

	return <-done
}
