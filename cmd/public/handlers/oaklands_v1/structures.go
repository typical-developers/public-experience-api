package oaklands_v1

import "time"

// @name	Oaklands.V1.MaterialValues
type MaterialValues struct {
	// The type of material.
	Type string `json:"type"`
	// The base value for the material.
	BaseValue float32 `json:"base_value"`
	// The current value of the material.
	CurrentValue float32 `json:"current_value"`
} //	@name	Oaklands.V1.MaterialValues

type StockMarket struct {
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
	Values []MaterialValues `json:"values"`
} //	@name	Oaklands.V1.StockMarket

type ChangelogVersion struct {
	ID      int32     `json:"id"`
	Version string    `json:"version"`
	Date    time.Time `json:"date"`
} //	@name	Oaklands.V1.ChangelogVersion

type Changelog struct {
	ChangelogVersion

	Changed []string `json:"changed"`
	Added   []string `json:"added"`
	Fixed   []string `json:"fixed"`
} //	@name	Oaklands.V1.Changelog

type NewsLetterContent struct {
	// The content type.
	Type string `json:"type" enums:"paragraph,image,image_carousel,video"`
} //	@name	Oaklands.V1.NewsLetterContent

type NewsletterParagraph struct {
	NewsLetterContent
	// The text contents.
	Text string `json:"text"`
} //	@name	Oaklands.V1.NewsletterParagraph

type NewsletterImage struct {
	NewsLetterContent
	// The Roblox ID for the image in the section.
	ImageId string `json:"image_id"`
} //	@name	Oaklands.V1.NewsletterImage

type NewsletterImageCarousel struct {
	NewsLetterContent
	// The Roblox IDs for the images in the section.
	ImageIds []string `json:"image_ids"`
} //	@name	Oaklands.V1.NewsletterImageCarousel

type NewsletterVideo struct {
	NewsLetterContent
	// The Roblox ID for the video in the section.
	VideoId string `json:"video_id"`
} //	@name	Oaklands.V1.NewsletterVideo

type NewsletterSection struct {
	// The header of the section.
	Header string `json:"header"`
	// The type of content in the section.
	Content []any `json:"content" oneOf:"NewsletterParagraph,NewsletterImage,NewsletterImageCarousel,NewsletterVideo"`
} //	@name	Oaklands.V1.NewsletterSection

type Newsletter struct {
	// The primary header of the newsletter.
	Header string `json:"header"`
	// The secondary header of the newsletter.
	Subheader string `json:"subheader"`
	// The date that the newsletter was released.
	Date time.Time `json:"date"`
	// The Roblox ID for the banner of the newsletter.
	BannerImageId string `json:"banner_image_id"`
	// Sections of the newsletter.
	Sections []NewsletterSection `json:"sections"`
} //	@name	Oaklands.V1.Newsletter

type NewsletterEntry struct {
	ID   string    `json:"id"`
	Date time.Time `json:"date"`
} //	@name	Oaklands.V1.NewsletterEntry

type SyncInfo struct {
	//	The last time that the resource was updated.
	LastSync time.Time `json:"last_sync"`
} //	@name	Oaklands.V1.SyncInfo

type SyncdInfoWithReset struct {
	//	The last time that the resource was updated.
	LastSync time.Time `json:"last_sync"`
	// The next time that the resource will be updated
	NextSync time.Time `json:"next_sync"`
} //	@name	Oaklands.V1.SyncdInfoWithReset

type SyncMeta struct {
	// The last time that data from Oaklands that can be updated from experience changes was synced and updated.
	LastUpdate time.Time `json:"last_update"`
	// The next time Oaklands will be checked for an update.
	NextCheck time.Time `json:"next_check"`
} //	@name	Oaklands.V1.SyncMeta

type Sync struct {
	// Information from the experience that is used for sync checks.
	Meta SyncMeta `json:"meta"`
	// Sync information for the stock market.
	StockMarket SyncdInfoWithReset `json:"stock_market"`
	// Sync information for the classic store.
	ClassicStore SyncdInfoWithReset `json:"classic_store"`
} //	@name	Oaklands.V1.Sync

type StoreItem struct {
	// The name (identifier) of the item.
	Name string `json:"name"`
	// The display name of the item.
	DisplayName string `json:"display_name"`
	// The image of the item.
	Image *string `json:"image"`
	// The description of the item.
	Description string `json:"description"`
	// The currency that is used to purchase the item.
	Currency string `json:"currency"`
	// How much the item costs.
	Price float64 `json:"price"`
} //	@name	Oaklands.V1.StoreItem
