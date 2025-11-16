package wizard

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/monitoring"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
)

// TwoPassSearch performs brand-aware product search with two passes:
// Pass 1: Search with brand + name (prioritize same-brand alternatives)
// Pass 2: Search with name only (fallback for when brand match fails)
//
// This implements the constitution requirement: "Same-brand alternatives
// appear first regardless of store" (FR-002)
func (s *wizardService) TwoPassSearch(
	ctx context.Context,
	expiredItem *models.ShoppingListItem,
) ([]*models.Product, error) {
	if expiredItem == nil || expiredItem.LinkedProduct == nil {
		s.logger.Error("cannot perform two-pass search on nil item or product")
		return nil, nil
	}

	product := expiredItem.LinkedProduct
	productName := product.Name
	var brandName string
	if product.Brand != nil {
		brandName = *product.Brand
	}

	s.logger.Info("starting two-pass search",
		"item_id", expiredItem.ID,
		"product_name", productName,
		"brand", brandName)

	// Track which product IDs we've already seen to avoid duplicates
	seenProductIDs := make(map[int]bool)
	var allResults []*models.Product
	var pass1Count, pass2Count int

	// Pass 1: Brand + Name search (if brand is present)
	if brandName != "" {
		pass1Query := brandName + " " + productName
		pass1Request := &search.SearchRequest{
			Query:       pass1Query,
			StoreIDs:    nil, // Search across all stores for maximum coverage
			Limit:       20,  // Get top 20 brand matches
			Offset:      0,
			PreferFuzzy: true, // Use fuzzy matching for typo tolerance
		}

		s.logger.Debug("pass 1: brand+name search",
			"query", pass1Query,
			"limit", pass1Request.Limit)

		pass1Response, err := s.searchService.FuzzySearchProducts(ctx, pass1Request)
		if err != nil {
			s.logger.Error("pass 1 search failed",
				"query", pass1Query,
				"error", err)
			// Don't fail entirely - continue to pass 2
		} else {
			s.logger.Info("pass 1 results",
				"count", len(pass1Response.Products),
				"query_time_ms", pass1Response.QueryTime.Milliseconds())

			// Extract products from pass 1 results
			for _, result := range pass1Response.Products {
				if !seenProductIDs[result.Product.ID] {
					seenProductIDs[result.Product.ID] = true
					allResults = append(allResults, result.Product)
					pass1Count++
				}
			}
		}
	}

	// Pass 2: Name-only search (fallback or primary if no brand)
	pass2Request := &search.SearchRequest{
		Query:       productName,
		StoreIDs:    nil, // Search across all stores
		Limit:       30,  // Get more results for name-only (broader net)
		Offset:      0,
		PreferFuzzy: true,
	}

	s.logger.Debug("pass 2: name-only search",
		"query", productName,
		"limit", pass2Request.Limit)

	pass2Response, err := s.searchService.FuzzySearchProducts(ctx, pass2Request)
	if err != nil {
		s.logger.Error("pass 2 search failed",
			"query", productName,
			"error", err)
		// If both passes fail, return what we have (might be empty)
	} else {
		s.logger.Info("pass 2 results",
			"count", len(pass2Response.Products),
			"query_time_ms", pass2Response.QueryTime.Milliseconds())

		// Extract products from pass 2, excluding duplicates
		for _, result := range pass2Response.Products {
			if !seenProductIDs[result.Product.ID] {
				seenProductIDs[result.Product.ID] = true
				allResults = append(allResults, result.Product)
				pass2Count++
			}
		}
	}

	// Track metrics for suggestions returned
	hasSameBrand := "false"
	if brandName != "" {
		// Check if at least one result has same brand
		for _, p := range allResults {
			if p.Brand != nil && *p.Brand == brandName {
				hasSameBrand = "true"
				break
			}
		}
	}
	monitoring.WizardSuggestionsReturned.WithLabelValues(hasSameBrand).Observe(float64(len(allResults)))

	s.logger.Info("two-pass search completed",
		"item_id", expiredItem.ID,
		"total_results", len(allResults),
		"pass1_unique_results", pass1Count,
		"pass2_unique_results", pass2Count,
		"has_same_brand", hasSameBrand)

	return allResults, nil
}
