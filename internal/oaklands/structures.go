package oaklands

import "time"

type StockMarket struct {
	LastSync time.Time
	NextSync time.Time
	Stock    []StockMarketMaterial
}

type SyncInfoCategory struct {
	LastSync time.Time
	NextSync *time.Time
}

type SyncInfo struct {
	LastContentSync time.Time
	NextSyncCheck   time.Time
	StockMarket     SyncInfoCategory
}
