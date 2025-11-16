package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

type ProductData struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	NormalizedName   string                 `json:"normalized_name"`
	Brand            string                 `json:"brand"`
	NormalizedBrand  string                 `json:"normalized_brand"`
	Category         string                 `json:"category"`
	Price            float64                `json:"price"`
	OriginalPrice    *float64               `json:"original_price,omitempty"`
	DiscountPercent  *float64               `json:"discount_percent,omitempty"`
	Unit             string                 `json:"unit"`
	NormalizedUnit   string                 `json:"normalized_unit"`
	UnitSize         *float64               `json:"unit_size,omitempty"`
	Description      string                 `json:"description"`
	ImageURL         string                 `json:"image_url"`
	Position         *Position              `json:"position,omitempty"`
	Confidence       float64                `json:"confidence"`
	ExtractionSource string                 `json:"extraction_source"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	ExtractedAt      time.Time              `json:"extracted_at"`
}

type Position struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ExtractionResult struct {
	FlyerID         string         `json:"flyer_id"`
	PageNum         int            `json:"page_num"`
	Products        []*ProductData `json:"products"`
	ExtractionModel string         `json:"extraction_model"`
	ProcessingTime  time.Duration  `json:"processing_time"`
	TokensUsed      int            `json:"tokens_used"`
	Cost            float64        `json:"cost"`
	Status          string         `json:"status"`
	Error           string         `json:"error,omitempty"`
	ExtractedAt     time.Time      `json:"extracted_at"`
	ValidatedAt     *time.Time     `json:"validated_at,omitempty"`
}

type ExtractionCache struct {
	redis      *redis.Client
	keyPrefix  string
	defaultTTL time.Duration
}

func NewExtractionCache(redis *redis.Client) *ExtractionCache {
	return &ExtractionCache{
		redis:      redis,
		keyPrefix:  "extraction:",
		defaultTTL: 7 * 24 * time.Hour, // Cache extractions for 1 week
	}
}

func (ec *ExtractionCache) SetExtractionResult(ctx context.Context, result *ExtractionResult) error {
	key := ec.extractionKey(result.FlyerID, result.PageNum)

	data, err := json.Marshal(result)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to marshal extraction result")
	}

	err = ec.redis.Set(ctx, key, data, ec.defaultTTL).Err()
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to cache extraction result")
	}

	// Add to flyer index
	flyerKey := ec.flyerIndexKey(result.FlyerID)
	err = ec.redis.SAdd(ctx, flyerKey, result.PageNum).Err()
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update flyer index")
	}

	// Set TTL on flyer index
	ec.redis.Expire(ctx, flyerKey, ec.defaultTTL)

	// Cache individual products for search
	for _, product := range result.Products {
		err := ec.setProductData(ctx, product, result.FlyerID, result.PageNum)
		if err != nil {
			// Log error but don't fail the entire operation
			continue
		}
	}

	return nil
}

func (ec *ExtractionCache) GetExtractionResult(ctx context.Context, flyerID string, pageNum int) (*ExtractionResult, error) {
	key := ec.extractionKey(flyerID, pageNum)

	data, err := ec.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Not found
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get extraction result from cache")
	}

	var result ExtractionResult
	err = json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to unmarshal extraction result")
	}

	return &result, nil
}

func (ec *ExtractionCache) GetFlyerExtractions(ctx context.Context, flyerID string) ([]*ExtractionResult, error) {
	flyerKey := ec.flyerIndexKey(flyerID)

	pageNums, err := ec.redis.SMembers(ctx, flyerKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []*ExtractionResult{}, nil
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get page numbers for flyer")
	}

	results := make([]*ExtractionResult, 0, len(pageNums))
	for _, pageNumStr := range pageNums {
		var pageNum int
		fmt.Sscanf(pageNumStr, "%d", &pageNum)

		result, err := ec.GetExtractionResult(ctx, flyerID, pageNum)
		if err != nil {
			continue // Skip invalid entries
		}
		if result != nil {
			results = append(results, result)
		}
	}

	return results, nil
}

func (ec *ExtractionCache) setProductData(ctx context.Context, product *ProductData, flyerID string, pageNum int) error {
	productKey := ec.productKey(product.ID)

	// Add flyer and page info to product metadata
	if product.Metadata == nil {
		product.Metadata = make(map[string]interface{})
	}
	product.Metadata["flyer_id"] = flyerID
	product.Metadata["page_num"] = pageNum

	data, err := json.Marshal(product)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to marshal product data")
	}

	err = ec.redis.Set(ctx, productKey, data, ec.defaultTTL).Err()
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to cache product")
	}

	// Add to search indices
	err = ec.addToSearchIndices(ctx, product)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to update search indices")
	}

	return nil
}

func (ec *ExtractionCache) GetProduct(ctx context.Context, productID string) (*ProductData, error) {
	key := ec.productKey(productID)

	data, err := ec.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Not found
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get product from cache")
	}

	var product ProductData
	err = json.Unmarshal([]byte(data), &product)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to unmarshal product data")
	}

	return &product, nil
}

func (ec *ExtractionCache) SearchProductsByName(ctx context.Context, searchTerm string, limit int) ([]*ProductData, error) {
	// Search in normalized name index
	nameKey := ec.nameIndexKey(searchTerm)

	productIDs, err := ec.redis.SMembers(ctx, nameKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to search products by name")
	}

	// If no exact match, try fuzzy search with pattern matching
	if len(productIDs) == 0 {
		pattern := ec.keyPrefix + "name:*" + searchTerm + "*"
		keys, err := ec.redis.Keys(ctx, pattern).Result()
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to perform fuzzy search")
		}

		// Collect product IDs from all matching keys
		for _, key := range keys {
			ids, err := ec.redis.SMembers(ctx, key).Result()
			if err != nil {
				continue
			}
			productIDs = append(productIDs, ids...)
		}
	}

	// Limit results
	if limit > 0 && len(productIDs) > limit {
		productIDs = productIDs[:limit]
	}

	// Fetch product data
	products := make([]*ProductData, 0, len(productIDs))
	for _, productID := range productIDs {
		product, err := ec.GetProduct(ctx, productID)
		if err != nil {
			continue // Skip invalid entries
		}
		if product != nil {
			products = append(products, product)
		}
	}

	return products, nil
}

func (ec *ExtractionCache) SearchProductsByBrand(ctx context.Context, brand string, limit int) ([]*ProductData, error) {
	brandKey := ec.brandIndexKey(brand)

	productIDs, err := ec.redis.SMembers(ctx, brandKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return []*ProductData{}, nil
		}
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to search products by brand")
	}

	// Limit results
	if limit > 0 && len(productIDs) > limit {
		productIDs = productIDs[:limit]
	}

	// Fetch product data
	products := make([]*ProductData, 0, len(productIDs))
	for _, productID := range productIDs {
		product, err := ec.GetProduct(ctx, productID)
		if err != nil {
			continue // Skip invalid entries
		}
		if product != nil {
			products = append(products, product)
		}
	}

	return products, nil
}

func (ec *ExtractionCache) addToSearchIndices(ctx context.Context, product *ProductData) error {
	pipe := ec.redis.Pipeline()

	// Add to name index (normalized name)
	if product.NormalizedName != "" {
		nameKey := ec.nameIndexKey(product.NormalizedName)
		pipe.SAdd(ctx, nameKey, product.ID)
		pipe.Expire(ctx, nameKey, ec.defaultTTL)
	}

	// Add to brand index (normalized brand)
	if product.NormalizedBrand != "" {
		brandKey := ec.brandIndexKey(product.NormalizedBrand)
		pipe.SAdd(ctx, brandKey, product.ID)
		pipe.Expire(ctx, brandKey, ec.defaultTTL)
	}

	// Add to category index
	if product.Category != "" {
		categoryKey := ec.categoryIndexKey(product.Category)
		pipe.SAdd(ctx, categoryKey, product.ID)
		pipe.Expire(ctx, categoryKey, ec.defaultTTL)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (ec *ExtractionCache) DeleteExtraction(ctx context.Context, flyerID string, pageNum int) error {
	// Get extraction result to find products
	result, err := ec.GetExtractionResult(ctx, flyerID, pageNum)
	if err != nil {
		return err
	}
	if result == nil {
		return nil // Already deleted
	}

	pipe := ec.redis.Pipeline()

	// Delete extraction result
	extractionKey := ec.extractionKey(flyerID, pageNum)
	pipe.Del(ctx, extractionKey)

	// Remove from flyer index
	flyerKey := ec.flyerIndexKey(flyerID)
	pipe.SRem(ctx, flyerKey, pageNum)

	// Delete products and remove from search indices
	for _, product := range result.Products {
		productKey := ec.productKey(product.ID)
		pipe.Del(ctx, productKey)

		// Remove from search indices
		if product.NormalizedName != "" {
			nameKey := ec.nameIndexKey(product.NormalizedName)
			pipe.SRem(ctx, nameKey, product.ID)
		}
		if product.NormalizedBrand != "" {
			brandKey := ec.brandIndexKey(product.NormalizedBrand)
			pipe.SRem(ctx, brandKey, product.ID)
		}
		if product.Category != "" {
			categoryKey := ec.categoryIndexKey(product.Category)
			pipe.SRem(ctx, categoryKey, product.ID)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to delete extraction from cache")
	}

	return nil
}

func (ec *ExtractionCache) InvalidateFlyer(ctx context.Context, flyerID string) error {
	// Get all extractions for the flyer
	results, err := ec.GetFlyerExtractions(ctx, flyerID)
	if err != nil {
		return err
	}

	// Delete each extraction
	for _, result := range results {
		err := ec.DeleteExtraction(ctx, flyerID, result.PageNum)
		if err != nil {
			return err
		}
	}

	// Delete flyer index
	flyerKey := ec.flyerIndexKey(flyerID)
	err = ec.redis.Del(ctx, flyerKey).Err()
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to delete flyer index")
	}

	return nil
}

func (ec *ExtractionCache) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	// Count total extractions
	extractionPattern := ec.keyPrefix + "result:*"
	extractionKeys, err := ec.redis.Keys(ctx, extractionPattern).Result()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count extractions")
	}

	// Count total products
	productPattern := ec.keyPrefix + "product:*"
	productKeys, err := ec.redis.Keys(ctx, productPattern).Result()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count products")
	}

	// Count search indices
	namePattern := ec.keyPrefix + "name:*"
	nameKeys, err := ec.redis.Keys(ctx, namePattern).Result()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count name indices")
	}

	brandPattern := ec.keyPrefix + "brand:*"
	brandKeys, err := ec.redis.Keys(ctx, brandPattern).Result()
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to count brand indices")
	}

	return map[string]interface{}{
		"total_extractions": len(extractionKeys),
		"total_products":    len(productKeys),
		"name_indices":      len(nameKeys),
		"brand_indices":     len(brandKeys),
		"cache_ttl":         ec.defaultTTL.String(),
	}, nil
}

func (ec *ExtractionCache) SetTTL(ttl time.Duration) {
	ec.defaultTTL = ttl
}

// Helper methods for key generation
func (ec *ExtractionCache) extractionKey(flyerID string, pageNum int) string {
	return fmt.Sprintf("%sresult:%s:%d", ec.keyPrefix, flyerID, pageNum)
}

func (ec *ExtractionCache) flyerIndexKey(flyerID string) string {
	return fmt.Sprintf("%sflyer:%s", ec.keyPrefix, flyerID)
}

func (ec *ExtractionCache) productKey(productID string) string {
	return fmt.Sprintf("%sproduct:%s", ec.keyPrefix, productID)
}

func (ec *ExtractionCache) nameIndexKey(normalizedName string) string {
	return fmt.Sprintf("%sname:%s", ec.keyPrefix, normalizedName)
}

func (ec *ExtractionCache) brandIndexKey(normalizedBrand string) string {
	return fmt.Sprintf("%sbrand:%s", ec.keyPrefix, normalizedBrand)
}

func (ec *ExtractionCache) categoryIndexKey(category string) string {
	return fmt.Sprintf("%scategory:%s", ec.keyPrefix, category)
}
