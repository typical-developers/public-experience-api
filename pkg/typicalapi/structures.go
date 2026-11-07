package typicalapi

import (
	"time"

	"github.com/typical-developers/public-experience-api/cmd/public/handlers/oaklands_v1"
)

type ResponseData[T any] struct {
	Data T `json:"data"`
}

// Changelogs
type (
	OaklandsChangelogVersion = oaklands_v1.ChangelogVersion
	OaklandsChangelog        = oaklands_v1.Changelog
)

// Economy
type (
	OaklandsMaterialValues       = oaklands_v1.MaterialValues
	OaklandsStockMarket          = oaklands_v1.StockMarket
	OaklandsMaterialBuffValue    = oaklands_v1.MaterialBuffValue
	OaklandsMaterialBuff         = oaklands_v1.MaterialBuff
	OaklandsMaterialBuffResponse = oaklands_v1.MaterialBuffResponse
)

// Newsletters
type (
	OaklandsNewsletterContent       = oaklands_v1.NewsLetterContent
	OaklandsNewsletterParagraph     = oaklands_v1.NewsletterParagraph
	OaklandsNewsletterImage         = oaklands_v1.NewsletterImage
	OaklandsNewsletterImageCarousel = oaklands_v1.NewsletterImageCarousel
	OaklandsNewsletterVideo         = oaklands_v1.NewsletterVideo
	OaklandsNewsletterSection       = oaklands_v1.NewsletterSection
	OaklandsNewsletter              = oaklands_v1.Newsletter
	OaklandsNewsletterEntry         = oaklands_v1.NewsletterEntry
)

// Sync
type (
	OaklandsSyncInfo          = oaklands_v1.SyncInfo
	OaklandsSyncInfoWithReset = oaklands_v1.SyncdInfoWithReset
	OaklandsSyncMeta          = oaklands_v1.SyncMeta
	OaklandsSync              = oaklands_v1.Sync
)

// Stores
type (
	OaklandsStoreItem     = oaklands_v1.StoreItem
	OaklandsStoreResponse = oaklands_v1.StoreResponse
)

// Items
type (
	OaklandsItemFormUnboxType = oaklands_v1.ItemFormUnboxType
	OaklandsItemForm          = oaklands_v1.ItemForm
	OaklandsItem              = oaklands_v1.Item
)

// Version
type OaklandsVersionInfo = oaklands_v1.VersionInfo

// World
type (
	OaklandsMinimapPosition = oaklands_v1.MinimapPosition
	OaklandsMinimapLayer    = oaklands_v1.MinimapLayer
	OaklandsMinimap         = oaklands_v1.Minimap
)

type StoreResponse struct {
	RestockAt *time.Time `json:"restock_at,omitempty"`
	ResponseData[[]OaklandsStoreItem]
}
