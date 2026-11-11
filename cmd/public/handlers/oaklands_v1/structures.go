package oaklands_v1

import (
	"time"

	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/pkg/typicalapi"
)

// The response structures are defined in pkg/typicalapi so that the client
// can use them without importing the server. They are aliased here so the
// handlers can keep using the short names.

// Economy
type (
	MaterialValues    = typicalapi.OaklandsMaterialValues
	StockMarket       = typicalapi.OaklandsStockMarket
	MaterialBuffValue = typicalapi.OaklandsMaterialBuffValue
	MaterialBuff      = typicalapi.OaklandsMaterialBuff
)

// Changelogs
type (
	ChangelogVersion = typicalapi.OaklandsChangelogVersion
	Changelog        = typicalapi.OaklandsChangelog
)

// Newsletters
type (
	NewsLetterContent       = typicalapi.OaklandsNewsletterContent
	NewsletterParagraph     = typicalapi.OaklandsNewsletterParagraph
	NewsletterImage         = typicalapi.OaklandsNewsletterImage
	NewsletterImageCarousel = typicalapi.OaklandsNewsletterImageCarousel
	NewsletterVideo         = typicalapi.OaklandsNewsletterVideo
	NewsletterSection       = typicalapi.OaklandsNewsletterSection
	Newsletter              = typicalapi.OaklandsNewsletter
	NewsletterEntry         = typicalapi.OaklandsNewsletterEntry
)

// Sync
type (
	SyncInfo           = typicalapi.OaklandsSyncInfo
	SyncdInfoWithReset = typicalapi.OaklandsSyncInfoWithReset
	SyncMeta           = typicalapi.OaklandsSyncMeta
	Sync               = typicalapi.OaklandsSync
)

// Stores
type StoreItem = typicalapi.OaklandsStoreItem

type StoreResponse struct {
	// The time that the store will run a restock of items.
	RestockAt *time.Time `json:"restock_at,omitempty"`
	// The items in the store.
	rest.Response[[]StoreItem]
} //	@name	Oaklands.V1.StoreResponse

// Media
type Media = typicalapi.OaklandsMedia

// Items
type (
	ItemFormUnboxType = typicalapi.OaklandsItemFormUnboxType
	ItemForm          = typicalapi.OaklandsItemForm
	Item              = typicalapi.OaklandsItem
)

// Index
type (
	IndexEntryInfo = typicalapi.OaklandsIndexEntryInfo
	IndexSection   = typicalapi.OaklandsIndexSection
	IndexVariant   = typicalapi.OaklandsIndexVariant
	IndexDrop      = typicalapi.OaklandsIndexDrop
	IndexEntry     = typicalapi.OaklandsIndexEntry
)

// Economy responses
type MaterialBuffResponse struct {
	// The time that the daily material buffs will refresh.
	RefreshAt time.Time `json:"refresh_at"`
	// The items in the store.
	rest.Response[map[string][]MaterialBuff]
}

// Version
type VersionInfo = typicalapi.OaklandsVersionInfo

// World
type (
	MinimapPosition         = typicalapi.OaklandsMinimapPosition
	MinimapLocationPosition = typicalapi.OaklandsMinimapLocationPosition
	MinimapLocation         = typicalapi.OaklandsMinimapLocation
	MinimapLayer            = typicalapi.OaklandsMinimapLayer
	Minimap                 = typicalapi.OaklandsMinimap
	RegionVector            = typicalapi.OaklandsRegionVector
	RegionBound             = typicalapi.OaklandsRegionBound
	RegionIsland            = typicalapi.OaklandsRegionIsland
	RegionResource          = typicalapi.OaklandsRegionResource
	Region                  = typicalapi.OaklandsRegion
)
