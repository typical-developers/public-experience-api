package oaklands

import "context"

type OaklandsUsecase interface {
	GetSyncTimes(ctx context.Context) (*SyncInfo, error)
	GetStockMarketTrees(ctx context.Context) (*StockMarket, error)
	GetStockMarketRocks(ctx context.Context) (*StockMarket, error)
	GetStockMarketOres(ctx context.Context) (*StockMarket, error)
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
