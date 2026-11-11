package typicalapi

import "time"

// ResponseData is the {"data": ...} envelope that the API wraps its responses in.
type ResponseData[T any] struct {
	Data T `json:"data"`
}

// OaklandsStoreResponse is the response from /oaklands/v1/stores/{store_name}.
type OaklandsStoreResponse struct {
	// The time that the store will run a restock of items.
	RestockAt *time.Time `json:"restock_at,omitempty"`
	ResponseData[[]OaklandsStoreItem]
}

// OaklandsMaterialBuffResponse is the response from /oaklands/v1/economy/daily-material-buffs.
type OaklandsMaterialBuffResponse struct {
	// The time that the daily material buffs will refresh.
	RefreshAt time.Time `json:"refresh_at"`
	ResponseData[map[string][]OaklandsMaterialBuff]
}

type OaklandsMedia struct {
	// The Roblox asset id, if it exists.
	AssetID *string `json:"asset_id,omitempty"`
	// The URL to the hosted media file, if it exists.
	URL *string `json:"url,omitempty"`
} //	@name	Oaklands.V1.Media

type OaklandsMaterialValues struct {
	// The type of material.
	Type string `json:"type"`
	// The base value for the material.
	BaseValue float32 `json:"base_value"`
	// The current value of the material.
	CurrentValue float32 `json:"current_value"`
} //	@name	Oaklands.V1.MaterialValues

type OaklandsStockMarket struct {
	// The identifier for the material.
	Name string `json:"name"`
	// The display name of the material.
	DisplayName string `json:"display_name"`
	// The currency of the material.
	CurrencyType string `json:"currency_type"`
	// The current multiplier.
	CurrentMultiplier float32 `json:"current_multiplier"`
	// The multipler from the last time the stock refreshed.
	LastMultiplier float32 `json:"last_multiplier"`
	// The values of the different types for the materials
	Values []OaklandsMaterialValues `json:"values"`
} //	@name	Oaklands.V1.StockMarket

type OaklandsMaterialBuffValue struct {
	// The type of material.
	Type string `json:"type"`
	// The base value of the material if it were sold un-buffed.
	BaseValue float32 `json:"base_value"`
	// The value of the material buffed.
	BuffedValue float32 `json:"buffed_value"`
} //	@name	Oaklands.V1.MaterialBuffValue

type OaklandsMaterialBuff struct {
	// The ID for the resource.
	ResourceID string `json:"resource_id"`
	// The name of the resource.
	DisplayName string `json:"display_name"`
	// The currency type that the resource sells for.
	CurrencyType string `json:"currency_type"`
	// The multiplier that the resource will sell for.
	Multiplier float32 `json:"multiplier"`
	// The value multiplier information for each resource type.
	Values []OaklandsMaterialBuffValue `json:"values"`
} //	@name	Oaklands.V1.MaterialBuff

type OaklandsChangelogVersion struct {
	ID      int32     `json:"id"`
	Version string    `json:"version"`
	Date    time.Time `json:"date"`
} //	@name	Oaklands.V1.ChangelogVersion

type OaklandsChangelog struct {
	OaklandsChangelogVersion

	Changed []string `json:"changed"`
	Added   []string `json:"added"`
	Fixed   []string `json:"fixed"`
} //	@name	Oaklands.V1.Changelog

type OaklandsNewsletterContent struct {
	// The content type.
	Type string `json:"type" enums:"paragraph,image,image_carousel,video"`
} //	@name	Oaklands.V1.NewsLetterContent

type OaklandsNewsletterParagraph struct {
	OaklandsNewsletterContent
	// The text contents.
	Text string `json:"text"`
} //	@name	Oaklands.V1.NewsletterParagraph

type OaklandsNewsletterImage struct {
	OaklandsNewsletterContent
	// The Roblox ID for the image in the section.
	//
	// Deprecated: Use `media` instead.
	ImageId string `json:"image_id" extensions:"x-deprecated=true"`
	// The image in the section.
	Media *OaklandsMedia `json:"media"`
} //	@name	Oaklands.V1.NewsletterImage

type OaklandsNewsletterImageCarousel struct {
	OaklandsNewsletterContent
	// The Roblox IDs for the images in the section.
	//
	// Deprecated: Use `media` instead.
	ImageIds []string `json:"image_ids" extensions:"x-deprecated=true"`
	// The images in the section.
	Media []OaklandsMedia `json:"media"`
} //	@name	Oaklands.V1.NewsletterImageCarousel

type OaklandsNewsletterVideo struct {
	OaklandsNewsletterContent
	// The Roblox ID for the video in the section.
	//
	// Deprecated: Use `media` instead.
	VideoId string `json:"video_id" extensions:"x-deprecated=true"`
	// The video in the section.
	Media *OaklandsMedia `json:"media"`
} //	@name	Oaklands.V1.NewsletterVideo

type OaklandsNewsletterSection struct {
	// The header of the section.
	Header string `json:"header"`
	// The type of content in the section.
	Content []any `json:"content" oneOf:"OaklandsNewsletterParagraph,OaklandsNewsletterImage,OaklandsNewsletterImageCarousel,OaklandsNewsletterVideo"`
} //	@name	Oaklands.V1.NewsletterSection

type OaklandsNewsletter struct {
	// The primary header of the newsletter.
	Header string `json:"header"`
	// The secondary header of the newsletter.
	Subheader string `json:"subheader"`
	// The date that the newsletter was released.
	Date time.Time `json:"date"`
	// The Roblox ID for the banner of the newsletter.
	//
	// Deprecated: Use `banner` instead.
	BannerImageId string `json:"banner_image_id" extensions:"x-deprecated=true"`
	// The banner of the newsletter.
	Banner *OaklandsMedia `json:"banner"`
	// Sections of the newsletter.
	Sections []OaklandsNewsletterSection `json:"sections"`
} //	@name	Oaklands.V1.Newsletter

type OaklandsNewsletterEntry struct {
	// The ID for the newsletter.
	ID string `json:"id"`
	// The primary header of the newsletter.
	Header string `json:"header"`
	// The Roblox ID for the banner of the newsletter.
	//
	// Deprecated: Use `banner` instead.
	BannerImageId string `json:"banner_image_id" extensions:"x-deprecated=true"`
	// The banner of the newsletter.
	Banner *OaklandsMedia `json:"banner"`
	// The date that the newsletter was released.
	Date time.Time `json:"date"`
} //	@name	Oaklands.V1.NewsletterEntry

type OaklandsSyncInfo struct {
	//	The last time that the resource was updated.
	LastSync time.Time `json:"last_sync"`
} //	@name	Oaklands.V1.SyncInfo

type OaklandsSyncInfoWithReset struct {
	//	The last time that the resource was updated.
	LastSync time.Time `json:"last_sync"`
	// The next time that the resource will be updated
	NextSync time.Time `json:"next_sync"`
} //	@name	Oaklands.V1.SyncdInfoWithReset

type OaklandsSyncMeta struct {
	// The last time that data from Oaklands that can be updated from experience changes was synced and updated.
	LastUpdate time.Time `json:"last_update"`
	// The next time Oaklands will be checked for an update.
	NextCheck time.Time `json:"next_check"`
} //	@name	Oaklands.V1.SyncMeta

type OaklandsSync struct {
	// Information from the experience that is used for sync checks.
	Meta OaklandsSyncMeta `json:"meta"`
	// Sync information for the stock market.
	StockMarket OaklandsSyncInfoWithReset `json:"stock_market"`
	// Sync information for the classic store.
	ClassicStore OaklandsSyncInfoWithReset `json:"classic_store"`
} //	@name	Oaklands.V1.Sync

type OaklandsStoreItem struct {
	// The id  of the item.
	ID string `json:"id"`
	// The display name of the item.
	DisplayName string `json:"display_name"`
	// The image of the item.
	//
	// Deprecated: Use `media` instead.
	Image *string `json:"image" extensions:"x-deprecated=true"`
	// The image of the item.
	Media *OaklandsMedia `json:"media"`
	// The description of the item.
	Description string `json:"description"`
	// The currency that is used to purchase the item.
	Currency string `json:"currency"`
	// How much the item costs.
	Price float64 `json:"price"`
} //	@name	Oaklands.V1.StoreItem

type OaklandsItemFormUnboxType struct {
	// The form type that the item can unbox as.
	Type string `json:"type"`
	// The image of the unboxed item type.
	//
	// Deprecated: Use `media` instead.
	Image *string `json:"image" extensions:"x-deprecated=true"`
	// The image of the unboxed item type.
	Media *OaklandsMedia `json:"media"`
	// The type that the unboxed form will convert to.
	ConvertType *string `json:"convert_type,omitempty"`
} //	@name	Oaklands.V1.ItemFormUnboxType

type OaklandsItemForm struct {
	// The type of form of the item.
	Type string `json:"type"`
	// The image of this form.
	//
	// Deprecated: Use `media` instead.
	Image *string `json:"image" extensions:"x-deprecated=true"`
	// The image of this form.
	Media *OaklandsMedia `json:"media"`
	// The id of the item that the object converts to.
	// ConvertID will be omitted on items that are not yet viewable,
	// e.g. gifts that have not been opened yet.
	ConvertID *string `json:"convert_id,omitempty"`
	// The type that the itme will convert to when unboxed.
	ConvertType *string `json:"convert_type,omitempty"`
	// The different, unique unbox types.
	UnboxTypes []OaklandsItemFormUnboxType `json:"unbox_types,omitempty"`
} //	@name	Oaklands.V1.ItemForm

type OaklandsItem struct {
	// The ID of the item
	ID string `json:"id"`
	// The name of the item.
	DisplayName string `json:"display_name"`
	// An image preview of the item.
	//
	// Deprecated: Use `media` instead.
	Image *string `json:"image" extensions:"x-deprecated=true"`
	// An image preview of the item.
	Media *OaklandsMedia `json:"media"`
	// The description of the item.
	Description string `json:"description"`

	// The price for the item if it's sold in a store.
	Price *float64 `json:"price,omitempty"`
	// The type of currency that the item sells for if it's sold in a store.
	Currency *string `json:"currency,omitempty"`

	// The time (e.g. for Gifts) that the item can be opened.
	UnboxTime *time.Time `json:"unbox_time,omitempty"`

	// The stats for an item, if it has any.
	Stats *map[string]any `json:"stats,omitempty"`

	// The forms that the item has / converts to.
	Forms []OaklandsItemForm `json:"forms"`
} //	@name	Oaklands.V1.Item

type OaklandsVersionInfo struct {
	// The current published version of the experience.
	Version string `json:"version"`
	// The last time the experience had an update published.
	LastUpdate time.Time `json:"last_update"`
	// The next time that the experience will be checked for updates.
	NextCheck time.Time `json:"next_check"`
} //	@name	Oaklands.V1.VersionInfo

type OaklandsMinimapPosition struct {
	// The minimum position.
	Min int64 `json:"min"`
	// The maximum position.
	Max int64 `json:"max"`
} //	@name	Oaklands.V1.MinimapPosition

type OaklandsMinimapLocationPosition struct {
	// The X position in the world.
	X float64 `json:"x"`
	// The Y position in the world.
	Y float64 `json:"y"`
	// The Z position in the world.
	Z float64 `json:"z"`
} //	@name	Oaklands.V1.MinimapLocationPosition

type OaklandsMinimapLocation struct {
	// The name of the location.
	Name string `json:"name"`
	// The position of the location in the world.
	Position OaklandsMinimapLocationPosition `json:"position"`
	// The location's marker.
	Marker *OaklandsMedia `json:"marker"`
	// The hex color of the location's marker.
	Color string `json:"color"`
	// The ID of the property, if the location is a property.
	PropertyID *int64 `json:"property_id,omitempty"`
	// Whether the location is only shown under certain conditions (e.g. events).
	Conditional bool `json:"conditional"`
} //	@name	Oaklands.V1.MinimapLocation

type OaklandsMinimapLayer struct {
	// The ID of the layer.
	LayerID string `json:"layer_id"`
	// The min/max X location in the world.
	X OaklandsMinimapPosition `json:"X"`
	// The min/max Y location in the world.
	Z OaklandsMinimapPosition `json:"Z"`
	// The image IDs of the minimap.
	// This is from left -> right.
	//
	// Deprecated: Use `media` instead.
	ImageIDs []string `json:"image_ids" extensions:"x-deprecated=true"`
	// The images of the minimap.
	// This is from left -> right.
	Media []OaklandsMedia `json:"media"`
	// The locations marked on the layer.
	Locations []OaklandsMinimapLocation `json:"locations"`
} //	@name	Oaklands.V1.MinimapLayer

type OaklandsMinimap struct {
	// The layers of the minimap.
	Layers []OaklandsMinimapLayer `json:"layers"`
} //	@name	Oaklands.V1.Minimap

type OaklandsIndexEntryInfo struct {
	// The category of the entry (trees, rocks, ores, creatures or forageables).
	Category string `json:"category"`
	// The ID of the entry.
	ID string `json:"id"`
	// The image of the entry.
	Media *OaklandsMedia `json:"media"`
} //	@name	Oaklands.V1.IndexEntryInfo

type OaklandsIndexRegion struct {
	// The ID of the region (e.g. FloweringMeadows).
	// Entries that aren't tied to a region use Special, Event or Unobtainable.
	ID string `json:"id"`
	// The display name of the region.
	DisplayName string `json:"display_name"`
} //	@name	Oaklands.V1.IndexRegion

type OaklandsIndexVariant struct {
	// The display name of the variant.
	DisplayName string `json:"display_name"`
	// The value of the variant, if it has one.
	Value *float64 `json:"value,omitempty"`
} //	@name	Oaklands.V1.IndexVariant

type OaklandsIndexDrop struct {
	// The ID of the material that can drop.
	ID string `json:"id"`
	// The display name of the material that can drop.
	DisplayName string `json:"display_name"`
	// The chance of the material dropping.
	Chance float64 `json:"chance"`
} //	@name	Oaklands.V1.IndexDrop

type OaklandsIndexEntry struct {
	OaklandsIndexEntryInfo

	// The display name of the entry.
	DisplayName string `json:"display_name"`
	// The description of the entry.
	Description string `json:"description"`
	// The regions that the entry can be found in.
	Regions []OaklandsIndexRegion `json:"regions"`
	// Whether the entry is an enemy.
	Enemy bool `json:"enemy"`
	// The tool tier needed to harvest the entry, if it has one.
	Tier *int64 `json:"tier,omitempty"`
	// The stats for the entry, if it has any.
	Stats map[string]any `json:"stats,omitempty"`
	// The variants of the entry (e.g. raw, refined, forged), if it has any.
	Variants map[string]OaklandsIndexVariant `json:"variants,omitempty"`
	// The materials that the entry can drop (enemies only).
	Drops []OaklandsIndexDrop `json:"drops,omitempty"`
} //	@name	Oaklands.V1.IndexEntry

type OaklandsRegionVector struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
} //	@name	Oaklands.V1.RegionVector

type OaklandsRegionBound struct {
	// The center position of the bound in the world.
	Position OaklandsRegionVector `json:"position"`
	// The size of the bound.
	Size OaklandsRegionVector `json:"size"`
	// The orientation of the bound, in degrees.
	Rotation OaklandsRegionVector `json:"rotation"`
} //	@name	Oaklands.V1.RegionBound

type OaklandsRegionIsland struct {
	// The ID of the island.
	ID string `json:"id"`
	// The display name of the island.
	DisplayName string `json:"display_name"`
	// The tool tier the island is meant for.
	Tier int64 `json:"tier"`
} //	@name	Oaklands.V1.RegionIsland

type OaklandsRegionResource struct {
	OaklandsIndexEntryInfo

	// The display name of the resource.
	DisplayName string `json:"display_name"`
} //	@name	Oaklands.V1.RegionResource

type OaklandsRegion struct {
	// The ID of the region.
	ID string `json:"id"`
	// The display name of the region.
	DisplayName string `json:"display_name"`
	// The description of the region.
	Description string `json:"description"`
	// The island that the region is on.
	Island *OaklandsRegionIsland `json:"island,omitempty"`
	// The tool tier the region is meant for.
	Tier int64 `json:"tier"`
	// The layer of the region. 0 is the surface, anything higher is a cave layer.
	Layer int64 `json:"layer"`
	// The amount of sprouts rewarded for discovering the region.
	DiscoveryReward int64 `json:"discovery_reward"`
	// The bounds that make up the region in the world.
	Bounds []OaklandsRegionBound `json:"bounds"`
	// The material index entries that can be found in the region.
	Resources []OaklandsRegionResource `json:"resources"`
} //	@name	Oaklands.V1.Region
