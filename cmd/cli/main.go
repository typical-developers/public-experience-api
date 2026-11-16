package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/luckfire/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/config"
	"github.com/typical-developers/public-experience-api/internal/logger"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/urfave/cli/v3"
)

func main() {
	logger.Init(logger.Options{})

	if strings.ToLower(config.C.Environment) == "development" {
		oaklands.SetStaging()
	}

	oc := opencloud.NewClient().WithAPIKey(config.C.OpencloudKey)
	oaklandsCmdHandler := NewOaklandsHandler(OaklandsHandlerOpts{
		OpencloudClient: oc,
	})

	oaklandsCmds := []*cli.Command{
		{
			Name:  "sync",
			Usage: "Resyncs all necessary content for Oaklands.",
			Action: func(ctx context.Context, c *cli.Command) error {
				return oaklandsCmdHandler.Resync(ctx)
			},
		},
	}

	cmd := &cli.Command{
		Name:        "ctl",
		Description: "Maintenance commands for the Typical Developers Public Experience API.",
		Commands: []*cli.Command{
			{
				Name:     "oaklands",
				Usage:    "Commands related to Oaklands.",
				Commands: oaklandsCmds,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
