package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/cmd/cli/config"
	"github.com/typical-developers/public-experience-api/internal/logger"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/urfave/cli/v3"
)

func main() {
	logger.Init(logger.Options{})

	if strings.ToLower(config.C.Environment) == "development" {
		oaklands.SetStaging()
	}

	if config.C.OaklandsPlaceVerisonOverride != nil {
		oaklands.SetPlaceVersionOverride(*config.C.OaklandsPlaceVerisonOverride)
	}

	oc := opencloud.NewClient().WithAPIKey(config.C.OpencloudKey)
	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.C.Redis.Host, config.C.Redis.Port),
		Password: config.C.Redis.Password,
		DB:       config.C.Redis.DB,
	})

	oaklandsRepository := oaklands.NewOaklandsRepository(&oaklands.OaklandsRepositoryOpts{RedisClient: redis})
	oaklandsCmdHandler := NewOaklandsHandler(OaklandsHandlerOpts{
		OpencloudClient: oc,
		Repository:      oaklandsRepository,
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
