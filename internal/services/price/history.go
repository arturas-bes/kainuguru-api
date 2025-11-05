package price

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// HistoryService handles price history operations
type HistoryService interface {
	// GetPriceHistory returns historical prices for a product
	GetPriceHistory(ctx context.Context, productID int, limit int, offset int) ([]*models.PriceHistory, error)

	// GetPriceHistoryByDateRange returns price history within a date range
	GetPriceHistoryByDateRange(ctx context.Context, productID int, startDate, endDate time.Time) ([]*models.PriceHistory, error)

	// GetPriceHistoryByStore returns price history for a specific store
	GetPriceHistoryByStore(ctx context.Context, productID int, storeID int, limit int) ([]*models.PriceHistory, error)

	// GetCurrentPrice returns the most recent price for a product
	GetCurrentPrice(ctx context.Context, productID int) (*models.PriceHistory, error)

	// GetCurrentPriceByStore returns the most recent price for a product at a specific store
	GetCurrentPriceByStore(ctx context.Context, productID int, storeID int) (*models.PriceHistory, error)

	// AddPriceEntry records a new price entry
	AddPriceEntry(ctx context.Context, entry *models.PriceHistory) error

	// GetPriceStatistics returns basic statistics for a product's price history
	GetPriceStatistics(ctx context.Context, productID int) (*PriceStatistics, error)

	// GetMultiStorePriceHistory returns price history from multiple stores
	GetMultiStorePriceHistory(ctx context.Context, productID int, storeIDs []int, limit int) ([]*models.PriceHistory, error)
}

// PriceStatistics contains statistical information about a product's price history
type PriceStatistics struct {
	ProductID       int       `json:"product_id"`
	MinPrice        float64   `json:"min_price"`
	MaxPrice        float64   `json:"max_price"`
	AveragePrice    float64   `json:"average_price"`
	CurrentPrice    float64   `json:"current_price"`
	PriceEntries    int       `json:"price_entries"`
	FirstRecorded   time.Time `json:"first_recorded"`
	LastRecorded    time.Time `json:"last_recorded"`
	StoreCount      int       `json:"store_count"`
	VolatilityScore float64   `json:"volatility_score"`
}

// PriceComparisonResult contains comparison data between current and historical prices
type PriceComparisonResult struct {
	CurrentPrice       float64   `json:"current_price"`
	PreviousPrice      float64   `json:"previous_price"`
	PriceChange        float64   `json:"price_change"`
	PriceChangePercent float64   `json:"price_change_percent"`
	IsLowerThanAverage bool      `json:"is_lower_than_average"`
	AveragePrice       float64   `json:"average_price"`
	LastUpdated        time.Time `json:"last_updated"`
}

// DateRangeFilter represents date range filtering options
type DateRangeFilter struct {
	StartDate time.Time
	EndDate   time.Time
}

// PriceHistoryFilter represents filtering options for price history queries
type PriceHistoryFilter struct {
	ProductID    int
	StoreIDs     []int
	DateRange    *DateRangeFilter
	Limit        int
	Offset       int
	OrderBy      string // "date_desc", "date_asc", "price_desc", "price_asc"
	IncludeStats bool
}

// historyService implements HistoryService
type historyService struct {
	repo PriceHistoryRepository
}

// NewHistoryService creates a new price history service
func NewHistoryService(repo PriceHistoryRepository) HistoryService {
	return &historyService{
		repo: repo,
	}
}

func (s *historyService) GetPriceHistory(ctx context.Context, productID int, limit int, offset int) ([]*models.PriceHistory, error) {
	filter := &PriceHistoryFilter{
		ProductID: productID,
		Limit:     limit,
		Offset:    offset,
		OrderBy:   "date_desc",
	}
	return s.repo.GetPriceHistory(ctx, filter)
}

func (s *historyService) GetPriceHistoryByDateRange(ctx context.Context, productID int, startDate, endDate time.Time) ([]*models.PriceHistory, error) {
	filter := &PriceHistoryFilter{
		ProductID: productID,
		DateRange: &DateRangeFilter{
			StartDate: startDate,
			EndDate:   endDate,
		},
		OrderBy: "date_desc",
	}
	return s.repo.GetPriceHistory(ctx, filter)
}

func (s *historyService) GetPriceHistoryByStore(ctx context.Context, productID int, storeID int, limit int) ([]*models.PriceHistory, error) {
	filter := &PriceHistoryFilter{
		ProductID: productID,
		StoreIDs:  []int{storeID},
		Limit:     limit,
		OrderBy:   "date_desc",
	}
	return s.repo.GetPriceHistory(ctx, filter)
}

func (s *historyService) GetCurrentPrice(ctx context.Context, productID int) (*models.PriceHistory, error) {
	return s.repo.GetCurrentPrice(ctx, productID)
}

func (s *historyService) GetCurrentPriceByStore(ctx context.Context, productID int, storeID int) (*models.PriceHistory, error) {
	return s.repo.GetCurrentPriceByStore(ctx, productID, storeID)
}

func (s *historyService) AddPriceEntry(ctx context.Context, entry *models.PriceHistory) error {
	return s.repo.CreatePriceEntry(ctx, entry)
}

func (s *historyService) GetPriceStatistics(ctx context.Context, productID int) (*PriceStatistics, error) {
	return s.repo.GetPriceStatistics(ctx, productID)
}

func (s *historyService) GetMultiStorePriceHistory(ctx context.Context, productID int, storeIDs []int, limit int) ([]*models.PriceHistory, error) {
	filter := &PriceHistoryFilter{
		ProductID: productID,
		StoreIDs:  storeIDs,
		Limit:     limit,
		OrderBy:   "date_desc",
	}
	return s.repo.GetPriceHistory(ctx, filter)
}

// PriceHistoryRepository defines the interface for price history data access
type PriceHistoryRepository interface {
	GetPriceHistory(ctx context.Context, filter *PriceHistoryFilter) ([]*models.PriceHistory, error)
	GetCurrentPrice(ctx context.Context, productID int) (*models.PriceHistory, error)
	GetCurrentPriceByStore(ctx context.Context, productID int, storeID int) (*models.PriceHistory, error)
	CreatePriceEntry(ctx context.Context, entry *models.PriceHistory) error
	GetPriceStatistics(ctx context.Context, productID int) (*PriceStatistics, error)
	GetPriceCount(ctx context.Context, productID int) (int, error)
}
