package search

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
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
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid search request")
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
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid search request")
	}

	req.Query = SanitizeQuery(req.Query)
	startTime := time.Now()

	query := `
		SELECT
			product_id, name, brand, category, current_price,
			store_id, flyer_id, name_similarity, brand_similarity, combined_similarity
		FROM fuzzy_search_products($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	s.logger.Info("executing fuzzy search", "query", req.Query, "limit", req.Limit, "offset", req.Offset, "tags", req.Tags, "flyer_ids", req.FlyerIDs)

	rows, err := s.db.DB.QueryContext(ctx, query,
		req.Query,
		0.15, // similarity_threshold - lowered to 0.15 to catch typos (combined_similarity uses weighted formula)
		req.Limit,
		req.Offset,
		pq.Array(req.StoreIDs),
		req.Category,
		req.MinPrice,
		req.MaxPrice,
		req.OnSaleOnly,
		pq.Array(req.Tags),
		pq.Array(req.FlyerIDs),
	)
	if err != nil {
		s.logger.Error("fuzzy search failed", "error", err, "query", req.Query)
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "fuzzy search failed")
	}
	defer rows.Close()

	var productIDs []int
	var scoresMap = make(map[int]float64)
	rowCount := 0
	for rows.Next() {
		rowCount++
		var (
			productID                                    int64
			storeID, flyerID                             int
			name                                         string
			brand, category                              sql.NullString
			currentPrice, nameSim, brandSim, combinedSim float64
		)

		err := rows.Scan(&productID, &name, &brand, &category, &currentPrice,
			&storeID, &flyerID, &nameSim, &brandSim, &combinedSim)
		if err != nil {
			s.logger.Error("failed to scan fuzzy search result", "error", err)
			continue
		}

		s.logger.Info("found product from fuzzy search", "product_id", productID, "name", name, "combined_sim", combinedSim)
		productIDs = append(productIDs, int(productID))
		scoresMap[int(productID)] = combinedSim
	}

	s.logger.Info("fuzzy search DB scan complete", "row_count", rowCount, "product_ids", productIDs, "scores_map_size", len(scoresMap))

	// Load full products with relations
	var products []*models.Product
	if len(productIDs) > 0 {
		s.logger.Info("loading products with relations", "product_ids", productIDs)
		err = s.db.NewSelect().
			Model(&products).
			Where("p.id IN (?)", bun.In(productIDs)).
			Relation("Store").
			Relation("Flyer").
			Relation("FlyerPage").
			Scan(ctx)
		if err != nil {
			s.logger.Error("failed to load products with relations", "error", err)
			return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to load products with relations")
		}
		s.logger.Info("loaded products", "count", len(products))
		for _, p := range products {
			s.logger.Info("loaded product detail", "id", p.ID, "name", p.Name, "store_id", p.StoreID, "flyer_id", p.FlyerID)
		}

		// Set currency for all products (not stored in DB)
		for _, p := range products {
			p.Currency = "EUR"
		}
	} else {
		s.logger.Warn("no product IDs to load")
	}

	// Build results with loaded products
	var results []ProductSearchResult
	for _, product := range products {
		if score, ok := scoresMap[product.ID]; ok {
			s.logger.Info("adding product to results", "product_id", product.ID, "name", product.Name, "score", score)
			results = append(results, ProductSearchResult{
				Product:     product,
				SearchScore: score,
				MatchType:   "fuzzy",
				Similarity:  score,
			})
		} else {
			s.logger.Warn("product not in scores map", "product_id", product.ID, "name", product.Name)
		}
	}

	s.logger.Info("fuzzy search results built", "result_count", len(results))

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "error iterating fuzzy search results")
	}

	// Get total count for pagination
	var totalCount int
	countQuery := `
		SELECT COUNT(*) FROM fuzzy_search_products($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	err = s.db.DB.QueryRowContext(ctx, countQuery, req.Query, 0.15, 1000000, 0,
		pq.Array(req.StoreIDs), req.Category, req.MinPrice, req.MaxPrice, req.OnSaleOnly, pq.Array(req.Tags), pq.Array(req.FlyerIDs)).Scan(&totalCount)
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
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid search request")
	}

	req.Query = SanitizeQuery(req.Query)
	startTime := time.Now()

	query := `
		SELECT
			product_id, name, brand, current_price,
			store_id, flyer_id, search_score, match_type
		FROM hybrid_search_products($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	rows, err := s.db.DB.QueryContext(ctx, query,
		req.Query,
		req.Limit,
		req.Offset,
		pq.Array(req.StoreIDs),
		req.MinPrice,
		req.MaxPrice,
		req.PreferFuzzy,
		req.Category,
		req.OnSaleOnly,
		pq.Array(req.Tags),
		pq.Array(req.FlyerIDs),
	)
	if err != nil {
		s.logger.Error("hybrid search failed", "error", err, "query", req.Query)
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "hybrid search failed")
	}
	defer rows.Close()

	var productIDs []int
	var scoresMap = make(map[int]float64)
	var matchTypeMap = make(map[int]string)
	for rows.Next() {
		var (
			productID                 int64
			storeID, flyerID          int
			name, matchType           string
			brand                     sql.NullString
			currentPrice, searchScore float64
		)

		err := rows.Scan(&productID, &name, &brand, &currentPrice,
			&storeID, &flyerID, &searchScore, &matchType)
		if err != nil {
			s.logger.Error("failed to scan hybrid search result", "error", err)
			continue
		}

		productIDs = append(productIDs, int(productID))
		scoresMap[int(productID)] = searchScore
		matchTypeMap[int(productID)] = matchType
	}

	// Load full products with relations
	var products []*models.Product
	if len(productIDs) > 0 {
		err = s.db.NewSelect().
			Model(&products).
			Where("p.id IN (?)", bun.In(productIDs)).
			Relation("Store").
			Relation("Flyer").
			Relation("FlyerPage").
			Scan(ctx)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to load products with relations")
		}

		// Set currency for all products (not stored in DB)
		for _, p := range products {
			p.Currency = "EUR"
		}
	}

	// Build results with loaded products
	var results []ProductSearchResult
	for _, product := range products {
		if score, ok := scoresMap[product.ID]; ok {
			matchType := matchTypeMap[product.ID]
			results = append(results, ProductSearchResult{
				Product:     product,
				SearchScore: score,
				MatchType:   matchType,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "error iterating hybrid search results")
	}

	// Get total count
	var totalCount int
	countQuery := `
		SELECT COUNT(*) FROM hybrid_search_products($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	err = s.db.QueryRowContext(ctx, countQuery, req.Query, 1000000, 0,
		pq.Array(req.StoreIDs), req.MinPrice, req.MaxPrice, req.PreferFuzzy, req.Category, req.OnSaleOnly, pq.Array(req.Tags), pq.Array(req.FlyerIDs)).Scan(&totalCount)
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
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid suggestion request")
	}

	req.PartialQuery = SanitizeQuery(req.PartialQuery)

	query := `
		SELECT suggestion, frequency, min_price, max_price, store_count
		FROM get_search_suggestions($1, $2)
	`

	rows, err := s.db.QueryContext(ctx, query, req.PartialQuery, req.Limit)
	if err != nil {
		s.logger.Error("search suggestions failed", "error", err, "query", req.PartialQuery)
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "search suggestions failed")
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
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid similar products request")
	}

	query := `
		SELECT similar_product_id, name, brand, current_price, store_id, similarity_score
		FROM find_similar_products($1, $2)
	`

	rows, err := s.db.QueryContext(ctx, query, req.ProductID, req.Limit)
	if err != nil {
		s.logger.Error("similar products search failed", "error", err, "product_id", req.ProductID)
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "similar products search failed")
	}
	defer rows.Close()

	var productIDs []int
	var similarityMap = make(map[int]float64)
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

		productIDs = append(productIDs, productID)
		similarityMap[productID] = similarityScore
	}

	// Load full products with relations
	var loadedProducts []*models.Product
	if len(productIDs) > 0 {
		err = s.db.NewSelect().
			Model(&loadedProducts).
			Where("p.id IN (?)", bun.In(productIDs)).
			Relation("Store").
			Relation("Flyer").
			Relation("FlyerPage").
			Scan(ctx)
		if err != nil {
			return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to load products with relations")
		}
	}

	// Build results with loaded products
	var products []SimilarProduct
	for _, product := range loadedProducts {
		if similarity, ok := similarityMap[product.ID]; ok {
			products = append(products, SimilarProduct{
				Product:         product,
				SimilarityScore: similarity,
			})
		}
	}

	return &SimilarProductsResponse{
		Products: products,
	}, nil
}

func (s *searchService) SuggestQueryCorrections(ctx context.Context, req *CorrectionRequest) (*CorrectionResponse, error) {
	if err := ValidateCorrectionRequest(req); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid correction request")
	}

	req.Query = SanitizeQuery(req.Query)

	query := `
		SELECT suggestion, confidence
		FROM suggest_query_corrections($1, $2)
	`

	rows, err := s.db.QueryContext(ctx, query, req.Query, req.Limit)
	if err != nil {
		s.logger.Error("query corrections failed", "error", err, "query", req.Query)
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "query corrections failed")
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
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check index status")
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
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to refresh search suggestions")
	}

	s.logger.Info("search suggestions refreshed successfully")
	return nil
}
