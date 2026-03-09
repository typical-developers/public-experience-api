package oaklands

import "context"

type OaklandsUsecase interface {
	GetStockMarketTrees(ctx context.Context) ([]StockMarketMaterial, error)
	GetStockMarketRocks(ctx context.Context) ([]StockMarketMaterial, error)
	GetStockMarketOres(ctx context.Context) ([]StockMarketMaterial, error)
}

type OaklandsUsecaseImpl struct {
	r OaklandsRepository
}

func NewOaklandsUsecase(r OaklandsRepository) OaklandsUsecase {
	return &OaklandsUsecaseImpl{r: r}
}

func (u *OaklandsUsecaseImpl) GetStockMarketTrees(ctx context.Context) ([]StockMarketMaterial, error) {
	return u.r.GetTreesStockMarket(ctx)
}

func (u *OaklandsUsecaseImpl) GetStockMarketRocks(ctx context.Context) ([]StockMarketMaterial, error) {
	return u.r.GetRocksStockMarket(ctx)
}

func (u *OaklandsUsecaseImpl) GetStockMarketOres(ctx context.Context) ([]StockMarketMaterial, error) {
	return u.r.GetOresStockMarket(ctx)
}
