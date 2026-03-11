package oaklands

import (
	"strings"
	"time"
)

type NewsletterContentType string

var (
	NewsletterContentTypeParagraph     NewsletterContentType = "Paragraph"
	NewsletterContentTypeImage         NewsletterContentType = "Image"
	NewsletterContentTypeImageCarousel NewsletterContentType = "ImageCarousel"
	NewsletterContentTypeVideo         NewsletterContentType = "Video"
)

type NewsletterBaseContent struct {
	Type NewsletterContentType `json:"Type"`
}

type NewsletterParagraph struct {
	NewsletterBaseContent
	Text string `json:"Text"`
}

type NewsletterImage struct {
	NewsletterBaseContent
	ImageId string `json:"ImageId"`
}

type NewsletterImageCarousel struct {
	NewsletterBaseContent
	ImageIds []string `json:"ImageIds"`
}

type Video struct {
	NewsletterBaseContent
	VideoId string `json:"VideoId"`
}

type NewsletterSection struct {
	Header  string `json:"Header"`
	Content []any  `json:"Content"`
}

type Newsletter struct {
	Header        string              `json:"Header"`
	Subheader     string              `json:"Subheader"`
	BannerImageId string              `json:"BannerImageId"`
	Sections      []NewsletterSection `json:"Sections"`
}

type ChangelogVersion struct {
	ID      int32    `json:"_Id"`
	Version string   `json:"Version"`
	Date    string   `json:"Date"`
	Changed []string `json:"Changed"`
	Added   []string `json:"Added"`
	Fixed   []string `json:"Fixed"`
}

// DateToISO8601 will convert the date (i.e. January 1st, 2026 to an ISO8601 formatted date in UTC)
func (c *ChangelogVersion) DateToISO8601() string {
	dateStr := c.Date

	// Remove ordinal suffixes to match the "January 2, 2006" layout
	replacer := strings.NewReplacer("st,", ",", "nd,", ",", "rd,", ",", "th,", ",")
	dateStr = replacer.Replace(dateStr)

	t, err := time.Parse("January 2, 2006", dateStr)
	if err != nil {
		return c.Date
	}

	return t.UTC().Format(time.RFC3339)
}

type StockMarketValue struct {
	Type         string  `json:"Type"`
	BaseValue    float32 `json:"BaseValue"`
	CurrentValue float32 `json:"CurrentValue"`
}

type StockMarketMaterial struct {
	Name              string             `json:"Name"`
	DisplayName       string             `json:"DisplayName"`
	CurrencyType      string             `json:"CurrencyType"`
	CurrentMultiplier float32            `json:"CurrentMultiplier"`
	LastMultiplier    float32            `json:"LastMultiplier"`
	Values            []StockMarketValue `json:"Values"`
}
