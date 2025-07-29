package experiences

import (
	"strings"
	"time"
)

// --- newsletters ---

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

// --- changelogs ---

type OaklandsChangelog struct {
	ID      int64    `json:"_id"`
	Date    string   `json:"date"`
	Changed []string `json:"changed"`
	Added   []string `json:"added"`
	Fixed   []string `json:"fixed"`
}

// --- items ---

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

// --- ore rarity ---

type OaklandsOreRarity_v1 map[string]map[string]map[string]int64

// --- stock market ---

type OaklandStockMarketValue struct {
	BaseValue    float64 `json:"base_value"`
	Type         string  `json:"type"`
	CurrentValue float64 `json:"current_value"`
}

type OaklandsStockMarketEntry struct {
	Name              string                    `json:"name"`
	Values            []OaklandStockMarketValue `json:"values"`
	CurrentDifference float64                   `json:"current_difference"`
	LastDifference    float64                   `json:"last_difference"`
	CurrencyType      string                    `json:"currency_type"`
}

type OaklandsStockMarketV1 struct {
	Trees map[string]OaklandsStockMarketEntry `json:"trees"`
	Rocks map[string]OaklandsStockMarketEntry `json:"rocks"`
	Ores  map[string]OaklandsStockMarketEntry `json:"ores"`
}

func (s *OaklandsStockMarketV1) NextReset() string {
	now := time.Now().UTC()
	hour := now.Hour()

	var nextReset int
	for _, h := range []int{4, 10, 16, 22} {
		if hour > h {
			continue
		}

		nextReset = h
		break
	}

	return time.Date(
		now.Year(), now.Month(), now.Day(),
		nextReset, 0, 0, 0, time.UTC,
	).UTC().Format(time.RFC3339)
}

// ---
