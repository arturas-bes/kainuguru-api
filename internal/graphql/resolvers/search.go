package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/services/search"
)

// AdvancedSearchProducts resolves the advancedSearchProducts query
func (r *queryResolver) AdvancedSearchProducts(ctx context.Context, input model.AdvancedSearchInput) (*model.AdvancedSearchResponse, error) {
	limit := 50
	offset := 0
	category := ""

	if input.Limit != nil {
		limit = *input.Limit
	}
	if input.Offset != nil {
		offset = *input.Offset
	}
	if input.Category != nil {
		category = *input.Category
	}

	req := &search.SearchRequest{
		Query:       input.Query,
		StoreIDs:    input.StoreIDs,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		OnSaleOnly:  input.OnSaleOnly != nil && *input.OnSaleOnly,
		Category:    category,
		Limit:       limit,
		Offset:      offset,
		PreferFuzzy: input.PreferFuzzy != nil && *input.PreferFuzzy,
	}

	resp, err := r.Resolver.searchService.SearchProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("advanced search failed: %w", err)
	}

	var products []*model.ProductSearchResult
	for _, result := range resp.Products {
		graphqlProduct, err := MapProductToGraphQL(result.Product)
		if err != nil {
			continue
		}

		products = append(products, &model.ProductSearchResult{
			Product:     graphqlProduct,
			SearchScore: result.SearchScore,
			MatchType:   result.MatchType,
			Similarity:  &result.Similarity,
			Highlights:  result.Highlights,
		})
	}

	return &model.AdvancedSearchResponse{
		Products:    products,
		TotalCount:  resp.TotalCount,
		QueryTime:   resp.QueryTime.String(),
		Suggestions: resp.Suggestions,
		HasMore:     resp.HasMore,
	}, nil
}

// FuzzySearchProducts resolves the fuzzySearchProducts query
func (r *queryResolver) FuzzySearchProducts(ctx context.Context, input model.FuzzySearchInput) (*model.FuzzySearchResponse, error) {
	limit := 50
	offset := 0
	category := ""

	if input.Limit != nil {
		limit = *input.Limit
	}
	if input.Offset != nil {
		offset = *input.Offset
	}
	if input.Category != nil {
		category = *input.Category
	}

	req := &search.SearchRequest{
		Query:       input.Query,
		StoreIDs:    input.StoreIDs,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		Category:    category,
		Limit:       limit,
		Offset:      offset,
		PreferFuzzy: true,
	}

	resp, err := r.Resolver.searchService.FuzzySearchProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search failed: %w", err)
	}

	var products []*model.FuzzyProductResult
	for _, result := range resp.Products {
		graphqlProduct, err := MapProductToGraphQL(result.Product)
		if err != nil {
			continue
		}

		products = append(products, &model.FuzzyProductResult{
			Product:           graphqlProduct,
			SearchScore:       result.SearchScore,
			NameSimilarity:    result.SearchScore, // Using SearchScore as placeholder
			BrandSimilarity:   result.Similarity,
			CombinedSimilarity: result.SearchScore,
		})
	}

	return &model.FuzzySearchResponse{
		Products:   products,
		TotalCount: resp.TotalCount,
		QueryTime:  resp.QueryTime.String(),
		HasMore:    resp.HasMore,
	}, nil
}

// HybridSearchProducts resolves the hybridSearchProducts query
func (r *queryResolver) HybridSearchProducts(ctx context.Context, input model.HybridSearchInput) (*model.HybridSearchResponse, error) {
	limit := 50
	offset := 0

	if input.Limit != nil {
		limit = *input.Limit
	}
	if input.Offset != nil {
		offset = *input.Offset
	}

	req := &search.SearchRequest{
		Query:       input.Query,
		StoreIDs:    input.StoreIDs,
		MinPrice:    input.MinPrice,
		MaxPrice:    input.MaxPrice,
		Limit:       limit,
		Offset:      offset,
		PreferFuzzy: input.PreferFuzzy != nil && *input.PreferFuzzy,
	}

	resp, err := r.searchService.HybridSearchProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	var products []*model.HybridProductResult
	for _, result := range resp.Products {
		graphqlProduct, err := MapProductToGraphQL(result.Product)
		if err != nil {
			continue
		}

		products = append(products, &model.HybridProductResult{
			Product:     graphqlProduct,
			SearchScore: result.SearchScore,
			MatchType:   result.MatchType,
		})
	}

	return &model.HybridSearchResponse{
		Products:   products,
		TotalCount: resp.TotalCount,
		QueryTime:  resp.QueryTime.String(),
		HasMore:    resp.HasMore,
	}, nil
}

// SearchSuggestions resolves the searchSuggestions query
func (r *queryResolver) SearchSuggestions(ctx context.Context, input model.SearchSuggestionsInput) (*model.SearchSuggestionsResponse, error) {
	limit := 10
	if input.Limit != nil {
		limit = *input.Limit
	}

	req := &search.SuggestionRequest{
		PartialQuery: input.PartialQuery,
		Limit:        limit,
	}

	resp, err := r.searchService.GetSearchSuggestions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search suggestions failed: %w", err)
	}

	var suggestions []*model.SearchSuggestion
	for _, suggestion := range resp.Suggestions {
		suggestions = append(suggestions, &model.SearchSuggestion{
			Text:       suggestion.Text,
			Frequency:  int(suggestion.Frequency),
			MinPrice:   suggestion.MinPrice,
			MaxPrice:   suggestion.MaxPrice,
			StoreCount: suggestion.StoreCount,
		})
	}

	return &model.SearchSuggestionsResponse{
		Suggestions: suggestions,
	}, nil
}

// SimilarProducts resolves the similarProducts query
func (r *queryResolver) SimilarProducts(ctx context.Context, input model.SimilarProductsInput) (*model.SimilarProductsResponse, error) {
	limit := 10
	if input.Limit != nil {
		limit = *input.Limit
	}

	req := &search.SimilarProductsRequest{
		ProductID: input.ProductID,
		Limit:     limit,
	}

	resp, err := r.searchService.FindSimilarProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("similar products search failed: %w", err)
	}

	var products []*model.SimilarProduct
	for _, similarProduct := range resp.Products {
		graphqlProduct, err := MapProductToGraphQL(similarProduct.Product)
		if err != nil {
			continue
		}

		products = append(products, &model.SimilarProduct{
			Product:         graphqlProduct,
			SimilarityScore: similarProduct.SimilarityScore,
		})
	}

	return &model.SimilarProductsResponse{
		Products: products,
	}, nil
}

// QueryCorrections resolves the queryCorrections query
func (r *queryResolver) QueryCorrections(ctx context.Context, input model.QueryCorrectionsInput) (*model.QueryCorrectionsResponse, error) {
	limit := 5
	if input.Limit != nil {
		limit = *input.Limit
	}

	req := &search.CorrectionRequest{
		Query: input.Query,
		Limit: limit,
	}

	resp, err := r.searchService.SuggestQueryCorrections(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("query corrections failed: %w", err)
	}

	var corrections []*model.QueryCorrection
	for _, correction := range resp.Corrections {
		corrections = append(corrections, &model.QueryCorrection{
			Suggestion: correction.Suggestion,
			Confidence: correction.Confidence,
		})
	}

	return &model.QueryCorrectionsResponse{
		Corrections: corrections,
	}, nil
}

// SearchHealth resolves the searchHealth query
func (r *queryResolver) SearchHealth(ctx context.Context) (*model.SearchHealthResponse, error) {
	health, err := r.searchService.GetSearchHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf("search health check failed: %w", err)
	}

	var indexStatus []*model.SearchIndexStatus
	for indexName, isUsed := range health.IndexStatus {
		indexStatus = append(indexStatus, &model.SearchIndexStatus{
			IndexName: indexName,
			IsHealthy: true, // Assuming healthy if it exists
			IsUsed:    isUsed,
		})
	}

	return &model.SearchHealthResponse{
		IndexStatus:      indexStatus,
		LastRefreshTime:  health.LastRefreshTime,
		SuggestionCount:  int(health.SuggestionCount),
		AverageQueryTime: health.AverageQueryTime.String(),
		TotalSearches:    int(health.TotalSearches),
		ErrorRate:        health.ErrorRate,
	}, nil
}

// RefreshSearchSuggestions resolves the refreshSearchSuggestions mutation
func (r *mutationResolver) RefreshSearchSuggestions(ctx context.Context) (*model.RefreshSearchSuggestionsResponse, error) {
	err := r.searchService.RefreshSearchSuggestions(ctx)
	if err != nil {
		return &model.RefreshSearchSuggestionsResponse{
			Success:     false,
			RefreshedAt: time.Now(),
			SuggestionCount: 0,
		}, fmt.Errorf("failed to refresh search suggestions: %w", err)
	}

	// Get updated count
	health, err := r.searchService.GetSearchHealth(ctx)
	suggestionCount := 0
	if err == nil {
		suggestionCount = int(health.SuggestionCount)
	}

	return &model.RefreshSearchSuggestionsResponse{
		Success:         true,
		RefreshedAt:     time.Now(),
		SuggestionCount: suggestionCount,
	}, nil
}