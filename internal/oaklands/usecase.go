package oaklands

import (
	"context"
	"slices"
)

type OaklandsUsecase interface {
	GetSyncTimes(ctx context.Context) (*SyncInfo, error)

	GetStockMarketTrees(ctx context.Context) (*StockMarket, error)
	GetStockMarketRocks(ctx context.Context) (*StockMarket, error)
	GetStockMarketOres(ctx context.Context) (*StockMarket, error)

	GetChangelogs(ctx context.Context) ([]Changelogs, error)
	GetChangelogVersion(ctx context.Context, version string) (*ChangelogVersion, error)

	GetNewsletters(ctx context.Context) ([]Newsletters, error)
	GetNewsletter(ctx context.Context, id string) (*Newsletter, error)
}

type OaklandsUsecaseImpl struct {
	r OaklandsRepository
}

func NewOaklandsUsecase(r OaklandsRepository) OaklandsUsecase {
	return &OaklandsUsecaseImpl{r: r}
}

func (u *OaklandsUsecaseImpl) GetSyncTimes(ctx context.Context) (*SyncInfo, error) {
	return u.r.GetSyncTimes(ctx)
}

func (u *OaklandsUsecaseImpl) GetStockMarketTrees(ctx context.Context) (*StockMarket, error) {
	return u.r.GetTreesStockMarket(ctx)
}

func (u *OaklandsUsecaseImpl) GetStockMarketRocks(ctx context.Context) (*StockMarket, error) {
	return u.r.GetRocksStockMarket(ctx)
}

func (u *OaklandsUsecaseImpl) GetStockMarketOres(ctx context.Context) (*StockMarket, error) {
	return u.r.GetOresStockMarket(ctx)
}

func (u *OaklandsUsecaseImpl) GetChangelogs(ctx context.Context) ([]Changelogs, error) {
	return u.r.GetChangelogs(ctx)
}

func (u *OaklandsUsecaseImpl) GetChangelogVersion(ctx context.Context, version string) (*ChangelogVersion, error) {
	return u.r.GetChangelogVersion(ctx, version)
}

func (u *OaklandsUsecaseImpl) GetNewsletters(ctx context.Context) ([]Newsletters, error) {
	config, err := u.r.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	newsletters, err := u.r.GetNewsletters(ctx)
	if err != nil {
		return nil, err
	}

	if len(config.HiddenNewsletters) == 0 {
		return newsletters, nil
	}

	hidden := make(map[string]struct{}, len(config.HiddenNewsletters))
	for _, id := range config.HiddenNewsletters {
		hidden[id] = struct{}{}
	}

	return slices.DeleteFunc(newsletters, func(newsletter Newsletters) bool {
		_, ok := hidden[newsletter.ID]
		return ok
	}), nil
}

func (u *OaklandsUsecaseImpl) GetNewsletter(ctx context.Context, id string) (*Newsletter, error) {
	if id == "latest" {
		config, err := u.r.GetConfig(ctx)
		if err == nil && config.NewsletterOverride != "" && config.NewsletterOverride != "Default" {
			id = config.NewsletterOverride
		}
	}

	return u.r.GetNewsletter(ctx, id)
}
