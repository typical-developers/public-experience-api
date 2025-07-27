package luau

import (
	"strings"
)

type NewsletterContentType string

const (
	NewsletterContentTypeImage         NewsletterContentType = "Image"
	NewsletterContentTypeVideo         NewsletterContentType = "Video"
	NewsletterContentTypeParagraph     NewsletterContentType = "Paragraph"
	NewsletterContentTypeImageCarousel NewsletterContentType = "ImageCarousel"
)

type OaklandsNewsletterContent struct {
	Type     NewsletterContentType `json:"type"`
	ImageID  *string               `json:"image_id,omitempty"`
	VideoID  *string               `json:"video_id,omitempty"`
	Text     *string               `json:"text,omitempty"`
	ImageIDs *[]string             `json:"image_ids,omitempty"`
}

func (n *OaklandsNewsletterContent) Value() string {
	switch n.Type {
	case NewsletterContentTypeImage:
		return *n.ImageID
	case NewsletterContentTypeVideo:
		return *n.VideoID
	case NewsletterContentTypeParagraph:
		return *n.Text
	case NewsletterContentTypeImageCarousel:
		return strings.Join(*n.ImageIDs, ",")
	default:
		return ""
	}
}

type OaklandsNewsletterSection struct {
	Header   string                      `json:"header"`
	Contents []OaklandsNewsletterContent `json:"contents"`
}

type OaklandsNewsletterPage struct {
	BannerImageID string                      `json:"banner_image_id"`
	HeaderText    string                      `json:"header_text"`
	SubheaderText string                      `json:"subheader_text"`
	Sections      []OaklandsNewsletterSection `json:"sections"`
}

type OaklandsNewsletters struct {
	LatestPage string                             `json:"latest_page"`
	Pages      *map[string]OaklandsNewsletterPage `json:"pages"`
}

// --

type OaklandsChangelog struct {
	ID      int64    `json:"_id"`
	Date    string   `json:"date"`
	Changed []string `json:"changed"`
	Added   []string `json:"added"`
	Fixed   []string `json:"fixed"`
}

// --

type OaklandsItemDetails struct {
	Identifier  string `json:"identifier"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type OaklandsItemStore struct {
	Price    int    `json:"price"`
	Currency string `json:"currency"`
	Type     string `json:"type"`
}

type OaklandsItemGift struct {
	UnboxEpoch int `json:"unbox_epoch"`
}

type OaklandsItemInfo struct {
	Details OaklandsItemDetails `json:"details"`
	Store   *OaklandsItemStore  `json:"store,omitempty"`
	Gift    *OaklandsItemGift   `json:"gift,omitempty"`
	Stats   any                 `json:"stats,omitempty"`
}

// --

type OaklandsStockMarketValue struct {
	BaseValue    float64 `json:"base_value"`
	Type         string  `json:"type"`
	CurrentValue float64 `json:"current_value"`
}

type OaklandsStockMarketItem struct {
	Name              string                     `json:"name"`
	Values            []OaklandsStockMarketValue `json:"values"`
	CurrentDifference float64                    `json:"current_difference"`
	LastDifference    float64                    `json:"last_difference"`
	CurrencyType      string                     `json:"currency_type"`
}

type OaklandsStockMarket struct {
	Trees map[string]OaklandsStockMarketItem `json:"Trees"`
	Rocks map[string]OaklandsStockMarketItem `json:"Rocks"`
	Ores  map[string]OaklandsStockMarketItem `json:"Ores"`
}

// --

type OaklandsUpdateData struct {
	Translations map[string]map[string]any `json:"translations"`
	Newsletters  OaklandsNewsletters       `json:"newsletters"`

	Changelogs  *map[string]OaklandsChangelog `json:"changelogs"`
	ItemDetails map[string]OaklandsItemInfo   `json:"item_details"`
	StockMarket OaklandsStockMarket           `json:"stock_market"`
	StoreItems  map[string][]string           `json:"store_items"`

	OreRarityV1 map[string]map[string]map[string]int64 `json:"ore_rarity_v1"`
}
