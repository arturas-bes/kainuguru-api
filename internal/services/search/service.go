package search

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/lib/pq"
	"github.com/uptrace/bun"
)

type searchService struct {
	db     *bun.DB
	logger *slog.Logger
}

func NewSearchService(db *bun.DB, logger *slog.Logger) Service {
	return &searchService{
		db:     db,
		logger: logger,
	}
}

func (s *searchService) SearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if err := ValidateSearchRequest(req); err != nil {
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	// Sanitize the query
	req.Query = SanitizeQuery(req.Query)

	if req.PreferFuzzy {
		return s.FuzzySearchProducts(ctx, req)
	}

	return s.HybridSearchProducts(ctx, req)
}

func (s *searchService) FuzzySearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if err := ValidateSearchRequest(req); err != nil {
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	req.Query = SanitizeQuery(req.Query)
	startTime := time.Now()

	query := `
		SELECT
			product_id, name, brand, category, current_price,
			store_id, flyer_id, name_similarity, brand_similarity, combined_similarity
		FROM fuzzy_search_products($1, $2, $3, $4, $5, $6, $7, $8)
	`

	rows, err := s.db.QueryContext(ctx, query,
		req.Query,
		0.3, // similarity_threshold
		req.Limit,
		req.Offset,
		pq.Array(req.StoreIDs),
		req.Category,
		req.MinPrice,
		req.MaxPrice,
	)
	if err != nil {
		s.logger.Error("fuzzy search failed", "error", err, "query", req.Query)
		return nil, fmt.Errorf("fuzzy search failed: %w", err)
	}
	defer rows.Close()

	var results []ProductSearchResult
	for rows.Next() {
		var (
			productID, storeID, flyerID                  int
			name, brand, category                        string
			currentPrice, nameSim, brandSim, combinedSim float64
		)

		err := rows.Scan(&productID, &name, &brand, &category, &currentPrice,
			&storeID, &flyerID, &nameSim, &brandSim, &combinedSim)
		if err != nil {
			s.logger.Error("failed to scan fuzzy search result", "error", err)
			continue
		}

		var brandPtr, categoryPtr *string
		if brand != "" {
			brandPtr = &brand
		}
		if category != "" {
			categoryPtr = &category
		}

		product := &models.Product{
			ID:           productID,
			Name:         name,
			Brand:        brandPtr,
			Category:     categoryPtr,
			CurrentPrice: currentPrice,
			StoreID:      storeID,
			FlyerID:      flyerID,
		}

		results = append(results, ProductSearchResult{
			Product:     product,
			SearchScore: combinedSim,
			MatchType:   "fuzzy",
			Similarity:  combinedSim,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating fuzzy search results: %w", err)
	}

	// Get total count for pagination
	var totalCount int
	countQuery := `
		SELECT COUNT(*) FROM fuzzy_search_products($1, $2, $3, $4, $5, $6, $7, $8)
	`
	err = s.db.QueryRowContext(ctx, countQuery, req.Query, 0.3, 1000000, 0,
		pq.Array(req.StoreIDs), req.Category, req.MinPrice, req.MaxPrice).Scan(&totalCount)
	if err != nil {
		s.logger.Warn("failed to get total count", "error", err)
		totalCount = len(results)
	}

	queryTime := time.Since(startTime)

	// Log analytics
	go s.LogSearchAnalytics(context.Background(), &SearchAnalytics{
		QueryText:     req.Query,
		ResultCount:   len(results),
		ExecutionTime: queryTime,
		MethodUsed:    "fuzzy",
		Timestamp:     startTime,
	})

	return &SearchResponse{
		Products:   results,
		TotalCount: totalCount,
		QueryTime:  queryTime,
		HasMore:    req.Offset+len(results) < totalCount,
	}, nil
}

func (s *searchService) HybridSearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if err := ValidateSearchRequest(req); err != nil {
		return nil, fmt.Errorf("invalid search request: %w", err)
	}

	req.Query = SanitizeQuery(req.Query)
	startTime := time.Now()

	query := `
		SELECT
			product_id, name, brand, current_price,
			store_id, flyer_id, search_score, match_type
		FROM hybrid_search_products($1, $2, $3, $4, $5, $6, $7)
	`

	rows, err := s.db.QueryContext(ctx, query,
		req.Query,
		req.Limit,
		req.Offset,
		pq.Array(req.StoreIDs),
		req.MinPrice,
		req.MaxPrice,
		req.PreferFuzzy,
	)
	if err != nil {
		s.logger.Error("hybrid search failed", "error", err, "query", req.Query)
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}
	defer rows.Close()

	var results []ProductSearchResult
	for rows.Next() {
		var (
			productID, storeID, flyerID int
			name, brand, matchType      string
			currentPrice, searchScore   float64
		)

		err := rows.Scan(&productID, &name, &brand, &currentPrice,
			&storeID, &flyerID, &searchScore, &matchType)
		if err != nil {
			s.logger.Error("failed to scan hybrid search result", "error", err)
			continue
		}

		var brandPtr *string
		if brand != "" {
			brandPtr = &brand
		}

		product := &models.Product{
			ID:           productID,
			Name:         name,
			Brand:        brandPtr,
			CurrentPrice: currentPrice,
			StoreID:      storeID,
			FlyerID:      flyerID,
		}

		results = append(results, ProductSearchResult{
			Product:     product,
			SearchScore: searchScore,
			MatchType:   matchType,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating hybrid search results: %w", err)
	}

	// Get total count
	var totalCount int
	countQuery := `
		SELECT COUNT(*) FROM hybrid_search_products($1, $2, $3, $4, $5, $6, $7)
	`
	err = s.db.QueryRowContext(ctx, countQuery, req.Query, 1000000, 0,
		pq.Array(req.StoreIDs), req.MinPrice, req.MaxPrice, req.PreferFuzzy).Scan(&totalCount)
	if err != nil {
		s.logger.Warn("failed to get total count", "error", err)
		totalCount = len(results)
	}

	queryTime := time.Since(startTime)

	// Log analytics
	go s.LogSearchAnalytics(context.Background(), &SearchAnalytics{
		QueryText:     req.Query,
		ResultCount:   len(results),
		ExecutionTime: queryTime,
		MethodUsed:    "hybrid",
		Timestamp:     startTime,
	})

	return &SearchResponse{
		Products:   results,
		TotalCount: totalCount,
		QueryTime:  queryTime,
		HasMore:    req.Offset+len(results) < totalCount,
	}, nil
}

func (s *searchService) GetSearchSuggestions(ctx context.Context, req *SuggestionRequest) (*SuggestionResponse, error) {
	if err := ValidateSuggestionRequest(req); err != nil {
		return nil, fmt.Errorf("invalid suggestion request: %w", err)
	}

	req.PartialQuery = SanitizeQuery(req.PartialQuery)

	query := `
		SELECT suggestion, frequency, min_price, max_price, store_count
		FROM get_search_suggestions($1, $2)
	`

	rows, err := s.db.QueryContext(ctx, query, req.PartialQuery, req.Limit)
	if err != nil {
		s.logger.Error("search suggestions failed", "error", err, "query", req.PartialQuery)
		return nil, fmt.Errorf("search suggestions failed: %w", err)
	}
	defer rows.Close()

	var suggestions []SearchSuggestion
	for rows.Next() {
		var suggestion SearchSuggestion
		err := rows.Scan(&suggestion.Text, &suggestion.Frequency,
			&suggestion.MinPrice, &suggestion.MaxPrice, &suggestion.StoreCount)
		if err != nil {
			s.logger.Error("failed to scan suggestion", "error", err)
			continue
		}
		suggestions = append(suggestions, suggestion)
	}

	return &SuggestionResponse{
		Suggestions: suggestions,
	}, nil
}

func (s *searchService) FindSimilarProducts(ctx context.Context, req *SimilarProductsRequest) (*SimilarProductsResponse, error) {
	if err := ValidateSimilarProductsRequest(req); err != nil {
		return nil, fmt.Errorf("invalid similar products request: %w", err)
	}

	query := `
		SELECT similar_product_id, name, brand, current_price, store_id, similarity_score
		FROM find_similar_products($1, $2)
	`

	rows, err := s.db.QueryContext(ctx, query, req.ProductID, req.Limit)
	if err != nil {
		s.logger.Error("similar products search failed", "error", err, "product_id", req.ProductID)
		return nil, fmt.Errorf("similar products search failed: %w", err)
	}
	defer rows.Close()

	var products []SimilarProduct
	for rows.Next() {
		var (
			productID, storeID            int
			name, brand                   string
			currentPrice, similarityScore float64
		)

		err := rows.Scan(&productID, &name, &brand, &currentPrice, &storeID, &similarityScore)
		if err != nil {
			s.logger.Error("failed to scan similar product", "error", err)
			continue
		}

		var brandPtr *string
		if brand != "" {
			brandPtr = &brand
		}

		product := &models.Product{
			ID:           productID,
			Name:         name,
			Brand:        brandPtr,
			CurrentPrice: currentPrice,
			StoreID:      storeID,
		}

		products = append(products, SimilarProduct{
			Product:         product,
			SimilarityScore: similarityScore,
		})
	}

	return &SimilarProductsResponse{
		Products: products,
	}, nil
}

func (s *searchService) SuggestQueryCorrections(ctx context.Context, req *CorrectionRequest) (*CorrectionResponse, error) {
	if err := ValidateCorrectionRequest(req); err != nil {
		return nil, fmt.Errorf("invalid correction request: %w", err)
	}

	req.Query = SanitizeQuery(req.Query)

	query := `
		SELECT suggestion, confidence
		FROM suggest_query_corrections($1, $2)
	`

	rows, err := s.db.QueryContext(ctx, query, req.Query, req.Limit)
	if err != nil {
		s.logger.Error("query corrections failed", "error", err, "query", req.Query)
		return nil, fmt.Errorf("query corrections failed: %w", err)
	}
	defer rows.Close()

	var corrections []QueryCorrection
	for rows.Next() {
		var correction QueryCorrection
		err := rows.Scan(&correction.Suggestion, &correction.Confidence)
		if err != nil {
			s.logger.Error("failed to scan correction", "error", err)
			continue
		}
		corrections = append(corrections, correction)
	}

	return &CorrectionResponse{
		Corrections: corrections,
	}, nil
}

func (s *searchService) LogSearchAnalytics(ctx context.Context, analytics *SearchAnalytics) error {
	s.logger.Info("search analytics",
		"query", analytics.QueryText,
		"result_count", analytics.ResultCount,
		"execution_time_ms", analytics.ExecutionTime.Milliseconds(),
		"method", analytics.MethodUsed,
	)
	return nil
}

func (s *searchService) GetSearchHealth(ctx context.Context) (*SearchHealthStatus, error) {
	health := &SearchHealthStatus{
		IndexStatus: make(map[string]bool),
	}

	// Check index status
	indexQuery := `
		SELECT indexname, idx_scan > 0 as is_used
		FROM pg_stat_user_indexes
		WHERE tablename = 'products' AND schemaname = current_schema()
	`

	rows, err := s.db.QueryContext(ctx, indexQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to check index status: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var indexName string
		var isUsed bool
		if err := rows.Scan(&indexName, &isUsed); err != nil {
			continue
		}
		health.IndexStatus[indexName] = isUsed
	}

	// Get suggestion count
	suggestionQuery := `SELECT COUNT(*) FROM popular_product_searches`
	err = s.db.QueryRowContext(ctx, suggestionQuery).Scan(&health.SuggestionCount)
	if err != nil {
		s.logger.Warn("failed to get suggestion count", "error", err)
	}

	health.AverageQueryTime = 50 * time.Millisecond // Placeholder
	health.TotalSearches = 0                        // Placeholder
	health.ErrorRate = 0.01                         // Placeholder

	return health, nil
}

func (s *searchService) RefreshSearchSuggestions(ctx context.Context) error {
	query := `SELECT refresh_search_suggestions()`
	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		s.logger.Error("failed to refresh search suggestions", "error", err)
		return fmt.Errorf("failed to refresh search suggestions: %w", err)
	}

	s.logger.Info("search suggestions refreshed successfully")
	return nil
}
