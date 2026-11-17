package assets

import (
	"context"
	"fmt"
	"log"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type OaklandsItem struct {
	ID   string
	Form string
}

type OaklandsItemAsset struct {
	OaklandsItem
	Key string
}

type snapshot struct {
	// TODO:
	// 		This is design mainly around Oaklands APIs for the time being.
	// 		If the S3 bucket is used more than for this, this should be
	// 		refactored to allow for a more broad usecase, ideally some
	// 		usable helper with still allowing for snapshots.

	// id -> form -> full S3 key
	byID      map[string]map[string]string
	count     int
	refreshed time.Time
}

type Index struct {
	client *s3.Client
	bucket string
	prefix string

	snap      atomic.Pointer[snapshot]
	refreshMu sync.Mutex // serializes refreshes; never held by readers
}

func NewIndex(client *s3.Client, bucket, prefix string) *Index {
	ix := &Index{
		client: client,
		bucket: bucket,
		prefix: prefix,
	}

	ix.snap.Store(&snapshot{
		byID: map[string]map[string]string{},
	})

	return ix
}

// Start does one synchronous refresh (so the index is populated before you
// serve traffic), then refreshes every interval until ctx is cancelled.
// A failed background refresh keeps the previous snapshot.
func (ix *Index) Start(ctx context.Context, interval time.Duration) error {
	if err := ix.refreshWithTimeout(ctx); err != nil {
		return fmt.Errorf("initial asset index refresh: %w", err)
	}

	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := ix.refreshWithTimeout(ctx); err != nil {
					log.Printf("asset index refresh: %v", err)
				}
			}
		}
	}()

	return nil
}

func (ix *Index) refreshWithTimeout(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return ix.Refresh(ctx)
}

// Refresh lists the prefix and atomically swaps in a new snapshot.
// Safe to call after an upload to make the new asset visible immediately.
func (ix *Index) Refresh(ctx context.Context) error {
	ix.refreshMu.Lock()
	defer ix.refreshMu.Unlock()

	byID := map[string]map[string]string{}
	count := 0

	p := s3.NewListObjectsV2Paginator(ix.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(ix.bucket),
		Prefix: aws.String(ix.prefix),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list s3://%s/%s: %w", ix.bucket, ix.prefix, err)
		}
		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			id, form, ok := parseItemKey(key, ix.prefix)
			if !ok {
				continue
			}
			forms := byID[id]
			if forms == nil {
				forms = map[string]string{}
				byID[id] = forms
			}
			if _, dup := forms[form]; !dup { // first extension wins
				forms[form] = key
				count++
			}
		}
	}

	ix.snap.Store(&snapshot{byID: byID, count: count, refreshed: time.Now()})
	return nil
}

// Lookup reports whether an asset exists for id+form and returns its S3 key.
func (ix *Index) Lookup(id, form string) (string, bool) {
	key, ok := ix.snap.Load().byID[id][form]
	return key, ok
}

// URL returns the public "/assets/..." path for id+form, or nil if no image exists.
func (ix *Index) URL(id, form string) *string {
	key, ok := ix.Lookup(id, form)
	if !ok {
		return nil
	}
	url := "/assets/" + key
	return &url
}

// Forms returns the sorted forms that exist for an item.
func (ix *Index) Forms(id string) []string {
	forms := ix.snap.Load().byID[id]
	out := make([]string, 0, len(forms))
	for f := range forms {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// Assets returns every existing asset for the given items, in item order.
func (ix *Index) Assets(items []OaklandsItem) []OaklandsItemAsset {
	snap := ix.snap.Load() // one snapshot for the whole call, so results are consistent
	assets := []OaklandsItemAsset{}
	for _, it := range items {
		forms := snap.byID[it.ID]
		names := make([]string, 0, len(forms))
		for f := range forms {
			names = append(names, f)
		}
		sort.Strings(names)
		for _, f := range names {
			assets = append(assets, OaklandsItemAsset{
				OaklandsItem: OaklandsItem{
					ID:   it.ID,
					Form: f,
				},
				Key: forms[f],
			})
		}
	}
	return assets
}

// Stats returns how many assets are indexed and when the index was last refreshed.
func (ix *Index) Stats() (count int, refreshed time.Time) {
	s := ix.snap.Load()
	return s.count, s.refreshed
}

// parseItemKey turns "items/<id>/<form>.<ext>" into (id, form).
// Adjust this if your key layout differs.
func parseItemKey(key, prefix string) (id, form string, ok bool) {
	rest, found := strings.CutPrefix(key, prefix)
	if !found {
		return "", "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	ext := path.Ext(parts[1])
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif", ".avif", ".svg":
	default:
		return "", "", false
	}
	form = strings.TrimSuffix(parts[1], ext)
	if form == "" {
		return "", "", false
	}
	return parts[0], form, true
}
