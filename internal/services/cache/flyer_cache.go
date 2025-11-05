package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type FlyerData struct {
	ID          string    `json:"id"`
	StoreCode   string    `json:"store_code"`
	Title       string    `json:"title"`
	ValidFrom   time.Time `json:"valid_from"`
	ValidUntil  time.Time `json:"valid_until"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"image_url"`
	PageCount   int       `json:"page_count"`
	ScrapedAt   time.Time `json:"scraped_at"`
	ProcessedAt time.Time `json:"processed_at"`
}

type FlyerPageData struct {
	FlyerID   string `json:"flyer_id"`
	PageNum   int    `json:"page_num"`
	ImageURL  string `json:"image_url"`
	LocalPath string `json:"local_path"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	FileSize  int64  `json:"file_size"`
}

type FlyerCache struct {
	redis      *redis.Client
	keyPrefix  string
	defaultTTL time.Duration
}

func NewFlyerCache(redis *redis.Client) *FlyerCache {
	return &FlyerCache{
		redis:      redis,
		keyPrefix:  "flyer:",
		defaultTTL: 24 * time.Hour, // Cache flyers for 24 hours
	}
}

func (fc *FlyerCache) SetFlyer(ctx context.Context, flyer *FlyerData) error {
	key := fc.flyerKey(flyer.ID)

	data, err := json.Marshal(flyer)
	if err != nil {
		return fmt.Errorf("failed to marshal flyer data: %w", err)
	}

	err = fc.redis.Set(ctx, key, data, fc.defaultTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to cache flyer: %w", err)
	}

	// Add to store index
	storeKey := fc.storeIndexKey(flyer.StoreCode)
	err = fc.redis.SAdd(ctx, storeKey, flyer.ID).Err()
	if err != nil {
		return fmt.Errorf("failed to update store index: %w", err)
	}

	// Set TTL on store index
	fc.redis.Expire(ctx, storeKey, fc.defaultTTL)

	return nil
}

func (fc *FlyerCache) GetFlyer(ctx context.Context, flyerID string) (*FlyerData, error) {
	key := fc.flyerKey(flyerID)

	data, err := fc.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get flyer from cache: %w", err)
	}

	var flyer FlyerData
	err = json.Unmarshal([]byte(data), &flyer)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal flyer data: %w", err)
	}

	return &flyer, nil
}

func (fc *FlyerCache) GetFlyersByStore(ctx context.Context, storeCode string) ([]*FlyerData, error) {
	storeKey := fc.storeIndexKey(storeCode)

	flyerIDs, err := fc.redis.SMembers(ctx, storeKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []*FlyerData{}, nil
		}
		return nil, fmt.Errorf("failed to get flyer IDs for store: %w", err)
	}

	flyers := make([]*FlyerData, 0, len(flyerIDs))
	for _, flyerID := range flyerIDs {
		flyer, err := fc.GetFlyer(ctx, flyerID)
		if err != nil {
			continue // Skip invalid entries
		}
		if flyer != nil {
			flyers = append(flyers, flyer)
		}
	}

	return flyers, nil
}

func (fc *FlyerCache) SetFlyerPage(ctx context.Context, page *FlyerPageData) error {
	key := fc.pageKey(page.FlyerID, page.PageNum)

	data, err := json.Marshal(page)
	if err != nil {
		return fmt.Errorf("failed to marshal page data: %w", err)
	}

	err = fc.redis.Set(ctx, key, data, fc.defaultTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to cache page: %w", err)
	}

	// Add to flyer pages index
	pagesKey := fc.flyerPagesKey(page.FlyerID)
	err = fc.redis.SAdd(ctx, pagesKey, page.PageNum).Err()
	if err != nil {
		return fmt.Errorf("failed to update pages index: %w", err)
	}

	// Set TTL on pages index
	fc.redis.Expire(ctx, pagesKey, fc.defaultTTL)

	return nil
}

func (fc *FlyerCache) GetFlyerPage(ctx context.Context, flyerID string, pageNum int) (*FlyerPageData, error) {
	key := fc.pageKey(flyerID, pageNum)

	data, err := fc.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get page from cache: %w", err)
	}

	var page FlyerPageData
	err = json.Unmarshal([]byte(data), &page)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal page data: %w", err)
	}

	return &page, nil
}

func (fc *FlyerCache) GetFlyerPages(ctx context.Context, flyerID string) ([]*FlyerPageData, error) {
	pagesKey := fc.flyerPagesKey(flyerID)

	pageNums, err := fc.redis.SMembers(ctx, pagesKey).Result()
	if err != nil {
		if err == redis.Nil {
			return []*FlyerPageData{}, nil
		}
		return nil, fmt.Errorf("failed to get page numbers for flyer: %w", err)
	}

	pages := make([]*FlyerPageData, 0, len(pageNums))
	for _, pageNumStr := range pageNums {
		var pageNum int
		fmt.Sscanf(pageNumStr, "%d", &pageNum)

		page, err := fc.GetFlyerPage(ctx, flyerID, pageNum)
		if err != nil {
			continue // Skip invalid entries
		}
		if page != nil {
			pages = append(pages, page)
		}
	}

	return pages, nil
}

func (fc *FlyerCache) DeleteFlyer(ctx context.Context, flyerID string) error {
	// Get flyer data to find store code
	flyer, err := fc.GetFlyer(ctx, flyerID)
	if err != nil {
		return err
	}
	if flyer == nil {
		return nil // Already deleted
	}

	pipe := fc.redis.Pipeline()

	// Delete flyer data
	flyerKey := fc.flyerKey(flyerID)
	pipe.Del(ctx, flyerKey)

	// Remove from store index
	storeKey := fc.storeIndexKey(flyer.StoreCode)
	pipe.SRem(ctx, storeKey, flyerID)

	// Delete all pages
	pages, err := fc.GetFlyerPages(ctx, flyerID)
	if err == nil {
		for _, page := range pages {
			pageKey := fc.pageKey(flyerID, page.PageNum)
			pipe.Del(ctx, pageKey)
		}
	}

	// Delete pages index
	pagesKey := fc.flyerPagesKey(flyerID)
	pipe.Del(ctx, pagesKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete flyer from cache: %w", err)
	}

	return nil
}

func (fc *FlyerCache) InvalidateStore(ctx context.Context, storeCode string) error {
	// Get all flyers for the store
	flyers, err := fc.GetFlyersByStore(ctx, storeCode)
	if err != nil {
		return err
	}

	// Delete each flyer
	for _, flyer := range flyers {
		err := fc.DeleteFlyer(ctx, flyer.ID)
		if err != nil {
			return err
		}
	}

	// Delete store index
	storeKey := fc.storeIndexKey(storeCode)
	err = fc.redis.Del(ctx, storeKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete store index: %w", err)
	}

	return nil
}

func (fc *FlyerCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// Count total flyers
	flyerPattern := fc.keyPrefix + "flyer:*"
	flyerKeys, err := fc.redis.Keys(ctx, flyerPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count flyers: %w", err)
	}

	// Count total pages
	pagePattern := fc.keyPrefix + "page:*"
	pageKeys, err := fc.redis.Keys(ctx, pagePattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count pages: %w", err)
	}

	// Count stores
	storePattern := fc.keyPrefix + "store:*"
	storeKeys, err := fc.redis.Keys(ctx, storePattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count stores: %w", err)
	}

	return map[string]interface{}{
		"total_flyers": len(flyerKeys),
		"total_pages":  len(pageKeys),
		"total_stores": len(storeKeys),
		"cache_ttl":    fc.defaultTTL.String(),
	}, nil
}

func (fc *FlyerCache) SetTTL(ttl time.Duration) {
	fc.defaultTTL = ttl
}

// Helper methods for key generation
func (fc *FlyerCache) flyerKey(flyerID string) string {
	return fmt.Sprintf("%sflyer:%s", fc.keyPrefix, flyerID)
}

func (fc *FlyerCache) pageKey(flyerID string, pageNum int) string {
	return fmt.Sprintf("%spage:%s:%d", fc.keyPrefix, flyerID, pageNum)
}

func (fc *FlyerCache) storeIndexKey(storeCode string) string {
	return fmt.Sprintf("%sstore:%s", fc.keyPrefix, storeCode)
}

func (fc *FlyerCache) flyerPagesKey(flyerID string) string {
	return fmt.Sprintf("%spages:%s", fc.keyPrefix, flyerID)
}
