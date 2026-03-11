package oaklands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/pkg/redisx"
	"golang.org/x/sync/singleflight"
)

type OaklandsRepository interface {
	// ContentSync will update various entries with updated data.
	ContentSync(ctx context.Context, data ContentSyncData) error

	// GetSyncTimes will get the last time content was updated.
	GetSyncTimes(ctx context.Context) (*SyncInfo, error)

	// SetStockMarket will update all of the stock market values.
	SetStockMarket(ctx context.Context, data map[string][]StockMarketMaterial) error
	// GetTreesStockMarket will get the tree stock market.
	GetTreesStockMarket(ctx context.Context) (*StockMarket, error)
	// GetRocksStockMarket will get the rock stock market
	GetRocksStockMarket(ctx context.Context) (*StockMarket, error)
	// GetOresStockMarket will get the ore stock market.
	GetOresStockMarket(ctx context.Context) (*StockMarket, error)
}

type OaklandsRepositoryImpl struct {
	sf    singleflight.Group
	redis *redis.Client
}

type OaklandsRepositoryOpts struct {
	RedisClient *redis.Client
}

const (
	redisKeyLastSync = "oaklands:last_sync"

	redisKeyNewsletterLatest = "oaklands:newsletter:latest"

	redisKeyStockMarketLastSync = "oaklands:stock_market:last_sync"
	redisKeyStockMarketNextSync = "oaklands:stock_market:next_sync"
)

func redisKeyChangelog(version string) string {
	return fmt.Sprintf("oaklands:changelog:%s", version)
}

func redisKeyNewsletter(version string) string {
	return fmt.Sprintf("oaklands:newsletter:%s", version)
}

func redisKeyStockMarket(marketType string) string {
	return fmt.Sprintf("oaklands:stock_market:%s", strings.ToLower(marketType))
}

func NewOaklandsRepository(opts *OaklandsRepositoryOpts) OaklandsRepository {
	return &OaklandsRepositoryImpl{
		redis: opts.RedisClient,
	}
}

// jsonGet will get json data from a redis value.
func (r *OaklandsRepositoryImpl) jsonGet(ctx context.Context, key, path string, v any) error {
	val, err, _ := r.sf.Do(key, func() (any, error) {
		var raw json.RawMessage
		if err := redisx.JSONUnwrap(ctx, r.redis, key, path, &raw); err != nil {
			return nil, err
		}

		return raw, nil
	})

	if err != nil {
		return err
	}

	return json.Unmarshal(val.(json.RawMessage), v)
}

// stockMarketReset wil get the timestamp for the next time the stock market resets.
func (r *OaklandsRepositoryImpl) stockMarketReset(now time.Time) time.Time {
	utc := now.UTC()
	year, month, day := utc.Date()
	location := utc.Location()
	hour := utc.Hour()

	switch {
	case hour >= 22:
		return time.Date(year, month, day, 22, 0, 0, 0, location)
	case hour >= 16:
		return time.Date(year, month, day, 16, 0, 0, 0, location)
	case hour >= 10:
		return time.Date(year, month, day, 10, 0, 0, 0, location)
	case hour >= 4:
		return time.Date(year, month, day, 4, 0, 0, 0, location)
	default:
		prev := utc.AddDate(0, 0, -1)
		y, m, d := prev.Date()
		return time.Date(y, m, d, 22, 0, 0, 0, location)
	}
}

func (r *OaklandsRepositoryImpl) ContentSync(ctx context.Context, data ContentSyncData) error {
	now := time.Now().UTC()
	pipeline := r.redis.Pipeline()

	if len(data.Changelogs) > 0 {
		for version, changelog := range data.Changelogs {
			pipeline.JSONSet(ctx, redisKeyChangelog(version), "$", changelog)
		}
	}

	if data.Newsletters.Latest != "" {
		pipeline.Set(ctx, redisKeyNewsletterLatest, data.Newsletters.Latest, 0)
	}
	if len(data.Newsletters.Pages) > 0 {
		for version, newsletter := range data.Newsletters.Pages {
			pipeline.JSONSet(ctx, redisKeyNewsletter(version), "$", newsletter)
		}
	}

	if len(data.StockMarket) > 0 {
		nextSync := r.stockMarketReset(now)
		pipeline.Set(ctx, redisKeyStockMarketLastSync, now.Format(time.RFC3339), 0)
		pipeline.Set(ctx, redisKeyStockMarketNextSync, nextSync.Format(time.RFC3339), 0)

		for marketType, materials := range data.StockMarket {
			pipeline.JSONSet(ctx, redisKeyStockMarket(marketType), "$", materials)
		}
	}

	pipeline.Set(ctx, redisKeyLastSync, now.Format(time.RFC3339), 0)

	if _, err := pipeline.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *OaklandsRepositoryImpl) GetSyncTimes(ctx context.Context) (*SyncInfo, error) {
	v, err := r.redis.MGet(ctx,
		redisKeyLastSync,
		redisKeyStockMarketLastSync,
		redisKeyStockMarketNextSync,
	).Result()
	if err != nil {
		return nil, err
	}

	if len(v) != 3 {
		return nil, fmt.Errorf("unexpected sync time result size: %d", len(v))
	}

	lastContentSync, err := redisx.ParseTime(v[0])
	if err != nil {
		return nil, err
	}

	stockMarketLastSync, err := redisx.ParseTime(v[1])
	if err != nil {
		return nil, err
	}

	stockMarketNextSync, err := redisx.ParseTime(v[2])
	if err != nil {
		return nil, err
	}

	info := &SyncInfo{}
	if lastContentSync != nil {
		info.LastContentSync = *lastContentSync
	}
	if stockMarketLastSync != nil {
		info.StockMarket.LastSync = *stockMarketLastSync
	}
	info.StockMarket.NextSync = stockMarketNextSync

	return info, nil
}

func (r *OaklandsRepositoryImpl) SetStockMarket(ctx context.Context, data map[string][]StockMarketMaterial) error {
	now := time.Now().UTC()
	pipeline := r.redis.Pipeline()

	nextSync := r.stockMarketReset(now)
	pipeline.Set(ctx, redisKeyStockMarketLastSync, now.Format(time.RFC3339), 0)
	pipeline.Set(ctx, redisKeyStockMarketNextSync, nextSync.Format(time.RFC3339), 0)

	for marketType, materials := range data {
		pipeline.JSONSet(ctx, redisKeyStockMarket(marketType), "$", materials)
	}

	if _, err := pipeline.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *OaklandsRepositoryImpl) GetTreesStockMarket(ctx context.Context) (*StockMarket, error) {
	var trees []StockMarketMaterial
	if err := r.jsonGet(ctx, redisKeyStockMarket("trees"), "$", &trees); err != nil {
		return nil, err
	}

	lastSync, err := r.GetSyncTimes(ctx)
	if err != nil {
		return nil, err
	}

	return &StockMarket{
		LastSync: lastSync.StockMarket.LastSync,
		NextSync: *lastSync.StockMarket.NextSync,
		Stock:    trees,
	}, nil
}

func (r *OaklandsRepositoryImpl) GetRocksStockMarket(ctx context.Context) (*StockMarket, error) {
	var rocks []StockMarketMaterial
	if err := r.jsonGet(ctx, redisKeyStockMarket("rocks"), "$", &rocks); err != nil {
		return nil, err
	}

	lastSync, err := r.GetSyncTimes(ctx)
	if err != nil {
		return nil, err
	}

	return &StockMarket{
		LastSync: lastSync.StockMarket.LastSync,
		NextSync: *lastSync.StockMarket.NextSync,
		Stock:    rocks,
	}, nil
}

func (r *OaklandsRepositoryImpl) GetOresStockMarket(ctx context.Context) (*StockMarket, error) {
	var ores []StockMarketMaterial
	if err := r.jsonGet(ctx, redisKeyStockMarket("ores"), "$", &ores); err != nil {
		return nil, err
	}

	lastSync, err := r.GetSyncTimes(ctx)
	if err != nil {
		return nil, err
	}

	return &StockMarket{
		LastSync: lastSync.StockMarket.LastSync,
		NextSync: *lastSync.StockMarket.NextSync,
		Stock:    ores,
	}, nil
}
