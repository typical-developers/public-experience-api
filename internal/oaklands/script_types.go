package oaklands

import (
	"encoding/json"
	"strings"
	"time"
)

type Translations map[string]map[string]string

// DateToISO8601 will convert the date (i.e. January 1st, 2026 to an ISO8601 formatted date in UTC)
func dateToISO8601(dateStr string) string {
	// Remove ordinal suffixes to match the "January 2, 2006" layout
	replacer := strings.NewReplacer("st,", ",", "nd,", ",", "rd,", ",", "th,", ",")
	dateStr = replacer.Replace(dateStr)

	t, err := time.Parse("January 2, 2006", dateStr)
	if err != nil {
		return dateStr
	}

	return t.UTC().Format(time.RFC3339)
}

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

type NewsletterVideo struct {
	NewsletterBaseContent
	VideoId string `json:"VideoId"`
}

type NewsletterSection struct {
	Header  string `json:"Header"`
	Content []any  `json:"Content"`
}

type Newsletter struct {
	Date          string              `json:"Date"`
	Header        string              `json:"Header"`
	Subheader     string              `json:"Subheader"`
	BannerImageId string              `json:"BannerImageId"`
	Sections      []NewsletterSection `json:"Sections"`
}

func (n *Newsletter) DateToISO8601() string {
	return dateToISO8601(n.Date)
}

func (n *NewsletterSection) UnmarshalJSON(data []byte) error {
	var raw struct {
		Header  string            `json:"Header"`
		Content []json.RawMessage `json:"Content"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	n.Header = raw.Header
	n.Content = make([]any, 0, len(raw.Content))

	for _, item := range raw.Content {
		var base NewsletterBaseContent
		if err := json.Unmarshal(item, &base); err != nil {
			return err
		}

		switch base.Type {
		case NewsletterContentTypeParagraph:
			var paragraph NewsletterParagraph
			if err := json.Unmarshal(item, &paragraph); err != nil {
				return err
			}
			n.Content = append(n.Content, paragraph)
		case NewsletterContentTypeImage:
			var image NewsletterImage
			if err := json.Unmarshal(item, &image); err != nil {
				return err
			}
			n.Content = append(n.Content, image)
		case NewsletterContentTypeImageCarousel:
			var carousel NewsletterImageCarousel
			if err := json.Unmarshal(item, &carousel); err != nil {
				return err
			}
			n.Content = append(n.Content, carousel)
		case NewsletterContentTypeVideo:
			var video NewsletterVideo
			if err := json.Unmarshal(item, &video); err != nil {
				return err
			}
			n.Content = append(n.Content, video)
		default:
			n.Content = append(n.Content, base)
		}
	}

	return nil
}

type ChangelogVersion struct {
	ID      int32    `json:"_Id"`
	Version string   `json:"Version"`
	Date    string   `json:"Date"`
	Changed []string `json:"Changed"`
	Added   []string `json:"Added"`
	Fixed   []string `json:"Fixed"`
}

func (c *ChangelogVersion) DateToISO8601() string {
	return dateToISO8601(c.Date)
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

type ItemForm struct {
	FormType    string `json:"FormType"`
	ConvertType string `json:"ConvertType"`
	Data        any    `json:"Data"`
}

type ItemFormStoreData struct {
	Currency string  `json:"Currency"`
	Price    float64 `json:"Price"`
}

func (f *ItemForm) UnmarshalJSON(data []byte) error {
	var raw struct {
		FormType    string          `json:"FormType"`
		ConvertType string          `json:"ConvertType"`
		Data        json.RawMessage `json:"Data"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	f.FormType = raw.FormType
	f.ConvertType = raw.ConvertType

	switch raw.FormType {
	case "store":
		var storeData ItemFormStoreData
		if err := json.Unmarshal(raw.Data, &storeData); err != nil {
			return err
		}
		f.Data = storeData
	default:
		var data map[string]any
		if len(raw.Data) > 0 && string(raw.Data) != "null" {
			if err := json.Unmarshal(raw.Data, &data); err != nil {
				return err
			}

			f.Data = data
		}
	}

	return nil
}

type ItemDetails struct {
	Name        string     `json:"Name"`
	DisplayName string     `json:"DisplayName"`
	Description string     `json:"Description"`
	Forms       []ItemForm `json:"Forms"`
}
