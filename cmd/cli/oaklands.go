package main

import (
	"context"

	"github.com/luckfire/goblox/opencloud"
)

type OaklandsHandler struct {
	oc *opencloud.Client
}

type OaklandsHandlerOpts struct {
	OpencloudClient *opencloud.Client
}

func NewOaklandsHandler(opts OaklandsHandlerOpts) *OaklandsHandler {
	return &OaklandsHandler{
		oc: opts.OpencloudClient,
	}
}

// Resync will fetch and force resync all current data.
func (h *OaklandsHandler) Resync(ctx context.Context) error {
	return nil
}
