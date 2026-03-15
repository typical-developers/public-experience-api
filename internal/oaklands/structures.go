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
	ClassicStore    SyncInfoCategory
}

type Changelogs struct {
	ID      int32
	Version string
	Date    time.Time
}

type Newsletters struct {
	ID   string
	Date time.Time
}

type StoreItem struct {
	Name        string
	DisplayName string
	Description string
	Currency    string
	Price       float64
}
