package search

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

type SearchRequest struct {
	Query       string   `json:"query" validate:"required,min=1,max=255"`
	StoreIDs    []int    `json:"store_ids,omitempty"`
	MinPrice    *float64 `json:"min_price,omitempty" validate:"omitempty,gte=0"`
	MaxPrice    *float64 `json:"max_price,omitempty" validate:"omitempty,gte=0"`
	OnSaleOnly  bool     `json:"on_sale_only"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Limit       int      `json:"limit" validate:"min=1,max=100"`
	Offset      int      `json:"offset" validate:"min=0"`
	PreferFuzzy bool     `json:"prefer_fuzzy"`
}

type SearchResponse struct {
	Products    []ProductSearchResult `json:"products"`
	TotalCount  int                   `json:"total_count"`
	QueryTime   time.Duration         `json:"query_time"`
	Suggestions []string              `json:"suggestions,omitempty"`
	HasMore     bool                  `json:"has_more"`
}

type ProductSearchResult struct {
	Product     *models.Product `json:"product"`
	SearchScore float64         `json:"search_score"`
	MatchType   string          `json:"match_type"`
	Similarity  float64         `json:"similarity,omitempty"`
	Highlights  []string        `json:"highlights,omitempty"`
}

type SuggestionRequest struct {
	PartialQuery string `json:"partial_query" validate:"required,min=1,max=100"`
	Limit        int    `json:"limit" validate:"min=1,max=20"`
}

type SuggestionResponse struct {
	Suggestions []SearchSuggestion `json:"suggestions"`
}

type SearchSuggestion struct {
	Text       string  `json:"text"`
	Frequency  int64   `json:"frequency"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	StoreCount int     `json:"store_count"`
}

type SimilarProductsRequest struct {
	ProductID int `json:"product_id" validate:"required,gt=0"`
	Limit     int `json:"limit" validate:"min=1,max=50"`
}

type SimilarProductsResponse struct {
	Products []SimilarProduct `json:"products"`
}

type SimilarProduct struct {
	Product         *models.Product `json:"product"`
	SimilarityScore float64         `json:"similarity_score"`
}

type CorrectionRequest struct {
	Query string `json:"query" validate:"required,min=1,max=255"`
	Limit int    `json:"limit" validate:"min=1,max=10"`
}

type CorrectionResponse struct {
	Corrections []QueryCorrection `json:"corrections"`
}

type QueryCorrection struct {
	Suggestion string  `json:"suggestion"`
	Confidence float64 `json:"confidence"`
}

type SearchAnalytics struct {
	QueryText     string        `json:"query_text"`
	ResultCount   int           `json:"result_count"`
	ExecutionTime time.Duration `json:"execution_time"`
	MethodUsed    string        `json:"method_used"`
	UserAgent     string        `json:"user_agent,omitempty"`
	IPAddress     string        `json:"ip_address,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

type Service interface {
	// Core search functionality
	SearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// Advanced search methods
	FuzzySearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
	HybridSearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// Search suggestions and autocomplete
	GetSearchSuggestions(ctx context.Context, req *SuggestionRequest) (*SuggestionResponse, error)

	// Similar products (recommendations)
	FindSimilarProducts(ctx context.Context, req *SimilarProductsRequest) (*SimilarProductsResponse, error)

	// Query correction for typos
	SuggestQueryCorrections(ctx context.Context, req *CorrectionRequest) (*CorrectionResponse, error)

	// Search analytics and monitoring
	LogSearchAnalytics(ctx context.Context, analytics *SearchAnalytics) error

	// Search health and performance
	GetSearchHealth(ctx context.Context) (*SearchHealthStatus, error)
	RefreshSearchSuggestions(ctx context.Context) error
}

type SearchHealthStatus struct {
	IndexStatus      map[string]bool `json:"index_status"`
	LastRefreshTime  time.Time       `json:"last_refresh_time"`
	SuggestionCount  int64           `json:"suggestion_count"`
	AverageQueryTime time.Duration   `json:"average_query_time"`
	TotalSearches    int64           `json:"total_searches"`
	ErrorRate        float64         `json:"error_rate"`
}
