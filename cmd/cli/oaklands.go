package main

import (
	"context"

	"github.com/typical-developers/goblox/opencloud"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
)

type OaklandsHandler struct {
	oc *opencloud.Client
	r  oaklands.OaklandsRepository
}

type OaklandsHandlerOpts struct {
	OpencloudClient *opencloud.Client
	Repository      oaklands.OaklandsRepository
}

func NewOaklandsHandler(opts OaklandsHandlerOpts) *OaklandsHandler {
	return &OaklandsHandler{
		oc: opts.OpencloudClient,
		r:  opts.Repository,
	}
}

// Resync will fetch and force resync all current data.
func (h *OaklandsHandler) Resync(ctx context.Context) error {
	content, err := oaklands.GetContentSync(ctx, h.oc)
	if err != nil {
		return err
	}

	if err := h.r.ContentSync(ctx, *content); err != nil {
		return err
	}

	return nil
}
