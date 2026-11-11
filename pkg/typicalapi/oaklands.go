package typicalapi

import (
	"context"
	"net/url"
)

type oaklands service

// Latest can be passed to GetChangelog or GetNewsletter to fetch the newest entry.
const Latest = "latest"

// Order is the direction results are ordered in.
// An empty Order uses the API's default, OrderDesc.
type Order string

const (
	OrderDesc Order = "desc"
	OrderAsc  Order = "asc"
)

// StoreSort is the field store items are sorted by.
// An empty StoreSort uses the API's default, StoreSortPrice.
type StoreSort string

const (
	StoreSortPrice StoreSort = "price"
	StoreSortName  StoreSort = "name"
)

var (
	OaklandsListChangelogs = NewEndpoint(
		"/oaklands/v1/changelogs",
	)

	OaklandsGetChangelog = NewEndpoint(
		"/oaklands/v1/changelogs/{version}",
	)

	OaklandsGetDailyMaterialBuffs = NewEndpoint(
		"/oaklands/v1/economy/daily-material-buffs",
	)

	OaklandsListIndexCategories = NewEndpoint(
		"/oaklands/v1/index",
	)

	OaklandsListIndex = NewEndpoint(
		"/oaklands/v1/index/{category}",
	)

	OaklandsGetIndexEntry = NewEndpoint(
		"/oaklands/v1/index/{category}/{id}",
	)

	OaklandsListItems = NewEndpoint(
		"/oaklands/v1/items",
	)

	OaklandsGetItem = NewEndpoint(
		"/oaklands/v1/items/{id}",
	)

	OaklandsListNewsletters = NewEndpoint(
		"/oaklands/v1/newsletters",
	)

	OaklandsGetNewsletter = NewEndpoint(
		"/oaklands/v1/newsletters/{id}",
	)

	OaklandsListStores = NewEndpoint(
		"/oaklands/v1/stores",
	)

	OaklandsGetStore = NewEndpoint(
		"/oaklands/v1/stores/{store_name}",
	)

	OaklandsGetTranslations = NewEndpoint(
		"/oaklands/v1/translations",
	)

	OaklandsGetVersionInfo = NewEndpoint(
		"/oaklands/v1/version-info",
	)

	OaklandsGetMinimap = NewEndpoint(
		"/oaklands/v1/world/minimap",
	)
)

// ListChangelogs fetches every changelog version, without the changelog contents.
func (o *oaklands) ListChangelogs(ctx context.Context, order Order) ([]OaklandsChangelogVersion, *Response, error) {
	query := url.Values{}
	if order != "" {
		query.Set("order_by", string(order))
	}

	changelogs := new(ResponseData[[]OaklandsChangelogVersion])
	resp, err := o.client.Call(ctx, OaklandsListChangelogs.Compile(query), nil, changelogs)
	if err != nil {
		return nil, resp, err
	}

	return changelogs.Data, resp, nil
}

// GetChangelog fetches a changelog by its version (e.g. "1.2.0"), or [Latest].
func (o *oaklands) GetChangelog(ctx context.Context, version string) (*OaklandsChangelog, *Response, error) {
	changelog := new(ResponseData[OaklandsChangelog])
	resp, err := o.client.Call(ctx, OaklandsGetChangelog.Compile(nil, version), nil, changelog)
	if err != nil {
		return nil, resp, err
	}

	return &changelog.Data, resp, nil
}

// GetChangelogByID fetches a changelog by its numeric ID.
func (o *oaklands) GetChangelogByID(ctx context.Context, id int32) (*OaklandsChangelog, *Response, error) {
	query := url.Values{"useId": {"true"}}

	changelog := new(ResponseData[OaklandsChangelog])
	resp, err := o.client.Call(ctx, OaklandsGetChangelog.Compile(query, id), nil, changelog)
	if err != nil {
		return nil, resp, err
	}

	return &changelog.Data, resp, nil
}

// GetDailyMaterialBuffs fetches today's material buffs and when they next refresh.
func (o *oaklands) GetDailyMaterialBuffs(ctx context.Context) (*OaklandsMaterialBuffResponse, *Response, error) {
	buffs := new(OaklandsMaterialBuffResponse)
	resp, err := o.client.Call(ctx, OaklandsGetDailyMaterialBuffs.Compile(nil), nil, buffs)
	if err != nil {
		return nil, resp, err
	}

	return buffs, resp, nil
}

// ListIndexCategories fetches the material index categories (e.g. "ores"). Use them with ListIndex.
func (o *oaklands) ListIndexCategories(ctx context.Context) ([]string, *Response, error) {
	categories := new(ResponseData[[]string])
	resp, err := o.client.Call(ctx, OaklandsListIndexCategories.Compile(nil), nil, categories)
	if err != nil {
		return nil, resp, err
	}

	return categories.Data, resp, nil
}

// ListIndex fetches the material index entries in a category (e.g. "ores"), without the entry details.
func (o *oaklands) ListIndex(ctx context.Context, category string) ([]OaklandsIndexEntryInfo, *Response, error) {
	entries := new(ResponseData[[]OaklandsIndexEntryInfo])
	resp, err := o.client.Call(ctx, OaklandsListIndex.Compile(nil, category), nil, entries)
	if err != nil {
		return nil, resp, err
	}

	return entries.Data, resp, nil
}

// GetIndexEntry fetches a material index entry by its category (e.g. "ores") and ID (e.g. "Rose Gold").
func (o *oaklands) GetIndexEntry(ctx context.Context, category, id string) (*OaklandsIndexEntry, *Response, error) {
	entry := new(ResponseData[OaklandsIndexEntry])
	resp, err := o.client.Call(ctx, OaklandsGetIndexEntry.Compile(nil, category, id), nil, entry)
	if err != nil {
		return nil, resp, err
	}

	return &entry.Data, resp, nil
}

// ListItems fetches the IDs of every available item.
func (o *oaklands) ListItems(ctx context.Context) ([]string, *Response, error) {
	items := new(ResponseData[[]string])
	resp, err := o.client.Call(ctx, OaklandsListItems.Compile(nil), nil, items)
	if err != nil {
		return nil, resp, err
	}

	return items.Data, resp, nil
}

// GetItem fetches an item by its ID.
func (o *oaklands) GetItem(ctx context.Context, id string) (*OaklandsItem, *Response, error) {
	item := new(ResponseData[OaklandsItem])
	resp, err := o.client.Call(ctx, OaklandsGetItem.Compile(nil, id), nil, item)
	if err != nil {
		return nil, resp, err
	}

	return &item.Data, resp, nil
}

// ListNewsletters fetches every newsletter, without the newsletter sections.
func (o *oaklands) ListNewsletters(ctx context.Context, order Order) ([]OaklandsNewsletterEntry, *Response, error) {
	query := url.Values{}
	if order != "" {
		query.Set("order_by", string(order))
	}

	newsletters := new(ResponseData[[]OaklandsNewsletterEntry])
	resp, err := o.client.Call(ctx, OaklandsListNewsletters.Compile(query), nil, newsletters)
	if err != nil {
		return nil, resp, err
	}

	return newsletters.Data, resp, nil
}

// GetNewsletter fetches a newsletter by its ID, or [Latest].
func (o *oaklands) GetNewsletter(ctx context.Context, id string) (*OaklandsNewsletter, *Response, error) {
	newsletter := new(ResponseData[OaklandsNewsletter])
	resp, err := o.client.Call(ctx, OaklandsGetNewsletter.Compile(nil, id), nil, newsletter)
	if err != nil {
		return nil, resp, err
	}

	return &newsletter.Data, resp, nil
}

// ListStores fetches the names of every store. Use the names with GetStore.
func (o *oaklands) ListStores(ctx context.Context) ([]string, *Response, error) {
	stores := new(ResponseData[[]string])
	resp, err := o.client.Call(ctx, OaklandsListStores.Compile(nil), nil, stores)
	if err != nil {
		return nil, resp, err
	}

	return stores.Data, resp, nil
}

// GetStore fetches the items for sale in a store and when it next restocks.
func (o *oaklands) GetStore(ctx context.Context, name string, sortBy StoreSort, order Order) (*OaklandsStoreResponse, *Response, error) {
	query := url.Values{}
	if sortBy != "" {
		query.Set("sort_by", string(sortBy))
	}
	if order != "" {
		query.Set("order_by", string(order))
	}

	store := new(OaklandsStoreResponse)
	resp, err := o.client.Call(ctx, OaklandsGetStore.Compile(query, name), nil, store)
	if err != nil {
		return nil, resp, err
	}

	return store, resp, nil
}

// GetTranslations fetches translation strings keyed by translation key.
// A non-empty prefix limits the result to keys starting with it
// (e.g. "ItemName" returns "ItemName" and "ItemName_description").
func (o *oaklands) GetTranslations(ctx context.Context, prefix string) (map[string]string, *Response, error) {
	query := url.Values{}
	if prefix != "" {
		query.Set("prefix", prefix)
	}

	translations := new(ResponseData[map[string]string])
	resp, err := o.client.Call(ctx, OaklandsGetTranslations.Compile(query), nil, translations)
	if err != nil {
		return nil, resp, err
	}

	return translations.Data, resp, nil
}

// GetVersionInfo fetches the experience's current published version.
func (o *oaklands) GetVersionInfo(ctx context.Context) (*OaklandsVersionInfo, *Response, error) {
	info := new(ResponseData[OaklandsVersionInfo])
	resp, err := o.client.Call(ctx, OaklandsGetVersionInfo.Compile(nil), nil, info)
	if err != nil {
		return nil, resp, err
	}

	return &info.Data, resp, nil
}

// GetMinimap fetches the world minimap layers.
func (o *oaklands) GetMinimap(ctx context.Context) (*OaklandsMinimap, *Response, error) {
	minimap := new(ResponseData[OaklandsMinimap])
	resp, err := o.client.Call(ctx, OaklandsGetMinimap.Compile(nil), nil, minimap)
	if err != nil {
		return nil, resp, err
	}

	return &minimap.Data, resp, nil
}
