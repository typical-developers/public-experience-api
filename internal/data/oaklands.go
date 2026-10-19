package data

import (
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
)

type OaklandsConfig struct {
	NewsletterOverride string
	HiddenNewsletters  []string
}

type OaklandsVersion struct {
	Version    string
	LastUpdate time.Time
	LastCheck  time.Time
	NextCheck  time.Time
}

type OaklandsItemFormStore struct {
	Currency string
	Price    float64
}

type OaklandsItemForm struct {
	FormType    string
	ConvertType string
	Data        any
}

func (f *OaklandsItemForm) UnmarshalJSON(raw []byte) error {
	var form struct {
		FormType    string
		ConvertType string
		Data        json.RawMessage
	}

	if err := json.Unmarshal(raw, &form); err != nil {
		return err
	}

	var decoded any

	switch form.FormType {
	case "store":
		var store OaklandsItemFormStore
		if err := json.Unmarshal(form.Data, &store); err != nil {
			return err
		}
		decoded = store
	default:
		if len(form.Data) > 0 {
			if err := json.Unmarshal(form.Data, &decoded); err != nil {
				return err
			}
		}
	}

	*f = OaklandsItemForm{
		FormType:    form.FormType,
		ConvertType: form.ConvertType,
		Data:        decoded,
	}

	return nil
}

type OaklandsItem struct {
	Name        string
	DisplayName string
	Description string
	Forms       []OaklandsItemForm
}

type OaklandsChangelog struct {
	ID      int32
	Version string
	Date    time.Time
	Changed []string
	Added   []string
	Fixed   []string
}

type OaklandsNewsletterSection struct {
	Header  string
	Content []any
}

func (s *OaklandsNewsletterSection) UnmarshalJSON(raw []byte) error {
	var section oaklands.NewsletterSection
	if err := json.Unmarshal(raw, &section); err != nil {
		return err
	}

	s.Header = section.Header
	s.Content = section.Content

	return nil
}

type OaklandsNewsletter struct {
	Date          time.Time
	Header        string
	Subheader     string
	BannerImageId string
	Section       []OaklandsNewsletterSection
}

type OaklandsStore struct {
	Config       *RedisValue[OaklandsConfig]
	Version      *RedisValue[OaklandsVersion]
	DeadVersions *RedisHash[bool]

	StoreStock *RedisHash[[]string]
	Items      *RedisHash[OaklandsItem]

	Changelog        *RedisHash[OaklandsChangelog]
	Newsletter       *RedisHash[OaklandsNewsletter]
	NewsletterLatest *RedisValue[string]
	Translations     *RedisHash[map[string]string]
}

func NewOaklandsStore(c redis.UniversalClient) *OaklandsStore {
	return &OaklandsStore{
		Config:       NewRedisValue[OaklandsConfig](c, "oaklands:config"),
		Version:      NewRedisValue[OaklandsVersion](c, "oaklands:version"),
		DeadVersions: NewRedisHash[bool](c, "oaklands:dead_versions"),

		StoreStock: NewRedisHash[[]string](c, "oaklands:store_stock"),
		Items:      NewRedisHash[OaklandsItem](c, "oaklands:items"),

		Changelog:        NewRedisHash[OaklandsChangelog](c, "oaklands:changelog"),
		Newsletter:       NewRedisHash[OaklandsNewsletter](c, "oaklands:newsletter"),
		NewsletterLatest: NewRedisValue[string](c, "oaklands:newsletter_latest"),
		Translations:     NewRedisHash[map[string]string](c, "oaklands:translations"),
	}
}
