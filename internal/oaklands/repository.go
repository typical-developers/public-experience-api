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

	// GetLastSyncTime will get the last time a ContentSync happened.
	GetLastSyncTime(ctx context.Context) (*time.Time, error)

	//
	GetTreesStockMarket(ctx context.Context) ([]StockMarketMaterial, error)
	//
	GetRocksStockMarket(ctx context.Context) ([]StockMarketMaterial, error)
	//
	GetOresStockMarket(ctx context.Context) ([]StockMarketMaterial, error)
}

type OaklandsRepositoryImpl struct {
	sf    singleflight.Group
	redis *redis.Client
}

type OaklandsRepositoryOpts struct {
	RedisClient *redis.Client
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

func (r *OaklandsRepositoryImpl) ContentSync(ctx context.Context, data ContentSyncData) error {
	pipeline := r.redis.Pipeline()

	for version, changelog := range data.Changelogs {
		key := fmt.Sprintf("oaklands:changelog:%s", version)
		pipeline.JSONSet(ctx, key, "$", changelog)
	}

	for version, newsletter := range data.Newsletters.Pages {
		key := fmt.Sprintf("oaklands:newsletter:%s", version)
		pipeline.JSONSet(ctx, key, "$", newsletter)
	}

	if data.Newsletters.Latest != "" {
		pipeline.Set(ctx, "oaklands:newsletter:latest", data.Newsletters.Latest, 0)
	}

	for marketType, materials := range data.StockMarket {
		key := fmt.Sprintf("oaklands:stock_market:%s", strings.ToLower(marketType))
		pipeline.JSONSet(ctx, key, "$", materials)
	}

	pipeline.Set(ctx, "oaklands:last_sync", time.Now().Format(time.RFC3339), 0)

	if _, err := pipeline.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (r *OaklandsRepositoryImpl) GetLastSyncTime(ctx context.Context) (*time.Time, error) {
	val, err := r.redis.Get(ctx, "oaklands:last_sync").Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *OaklandsRepositoryImpl) GetTreesStockMarket(ctx context.Context) ([]StockMarketMaterial, error) {
	var trees []StockMarketMaterial
	if err := r.jsonGet(ctx, "oaklands:stock_market:trees", "$", &trees); err != nil {
		return nil, err
	}

	return trees, nil
}

func (r *OaklandsRepositoryImpl) GetRocksStockMarket(ctx context.Context) ([]StockMarketMaterial, error) {
	var rocks []StockMarketMaterial
	if err := r.jsonGet(ctx, "oaklands:stock_market:rocks", "$", &rocks); err != nil {
		return nil, err
	}

	return rocks, nil
}

func (r *OaklandsRepositoryImpl) GetOresStockMarket(ctx context.Context) ([]StockMarketMaterial, error) {
	var ores []StockMarketMaterial
	if err := r.jsonGet(ctx, "oaklands:stock_market:ores", "$", &ores); err != nil {
		return nil, err
	}

	return ores, nil
}
