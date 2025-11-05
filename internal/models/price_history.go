package models

import (
	"time"

	"github.com/uptrace/bun"
)

// PriceHistory tracks historical price changes for products
type PriceHistory struct {
	bun.BaseModel `bun:"table:price_history,alias:ph"`

	ID        int64 `bun:"id,pk,autoincrement" json:"id"`
	ProductID int   `bun:"product_id,notnull" json:"product_id"`
	StoreID   int   `bun:"store_id,notnull" json:"store_id"`
	FlyerID   *int  `bun:"flyer_id" json:"flyer_id,omitempty"`

	// Price information
	Price         float64  `bun:"price,notnull" json:"price"`
	OriginalPrice *float64 `bun:"original_price" json:"original_price,omitempty"`
	Currency      string   `bun:"currency,default:'EUR'" json:"currency"`
	IsOnSale      bool     `bun:"is_on_sale,default:false" json:"is_on_sale"`

	// Timing information
	RecordedAt    time.Time  `bun:"recorded_at,notnull" json:"recorded_at"`
	ValidFrom     time.Time  `bun:"valid_from,notnull" json:"valid_from"`
	ValidTo       time.Time  `bun:"valid_to,notnull" json:"valid_to"`
	SaleStartDate *time.Time `bun:"sale_start_date" json:"sale_start_date,omitempty"`
	SaleEndDate   *time.Time `bun:"sale_end_date" json:"sale_end_date,omitempty"`

	// Source information
	Source           string  `bun:"source,default:'flyer'" json:"source"` // 'flyer', 'manual', 'api'
	ExtractionMethod string  `bun:"extraction_method,default:'ocr'" json:"extraction_method"`
	Confidence       float64 `bun:"confidence,default:1.0" json:"confidence"`

	// Availability and stock
	IsAvailable bool    `bun:"is_available,default:true" json:"is_available"`
	StockLevel  *string `bun:"stock_level" json:"stock_level,omitempty"`

	// Metadata
	Notes     *string   `bun:"notes" json:"notes,omitempty"`
	IsActive  bool      `bun:"is_active,default:true" json:"is_active"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`

	// Relations
	Product *Product `bun:"rel:belongs-to,join:product_id=id" json:"product,omitempty"`
	Store   *Store   `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
	Flyer   *Flyer   `bun:"rel:belongs-to,join:flyer_id=id" json:"flyer,omitempty"`
}

// PriceTrend represents price trend analysis data
type PriceTrend struct {
	bun.BaseModel `bun:"table:price_trends,alias:pt"`

	ID        int64 `bun:"id,pk,autoincrement" json:"id"`
	ProductID int   `bun:"product_id,notnull" json:"product_id"`
	StoreID   *int  `bun:"store_id" json:"store_id,omitempty"`

	// Period information
	Period       string    `bun:"period,notnull" json:"period"` // '7d', '30d', '90d', '1y'
	StartDate    time.Time `bun:"start_date,notnull" json:"start_date"`
	EndDate      time.Time `bun:"end_date,notnull" json:"end_date"`
	CalculatedAt time.Time `bun:"calculated_at,notnull" json:"calculated_at"`

	// Trend data
	Direction       string  `bun:"direction,notnull" json:"direction"` // 'RISING', 'FALLING', 'STABLE', 'VOLATILE'
	TrendPercent    float64 `bun:"trend_percent,notnull" json:"trend_percent"`
	Confidence      float64 `bun:"confidence,notnull" json:"confidence"`
	DataPoints      int     `bun:"data_points,notnull" json:"data_points"`
	VolatilityScore float64 `bun:"volatility_score,notnull" json:"volatility_score"`

	// Price statistics
	StartPrice  float64 `bun:"start_price,notnull" json:"start_price"`
	EndPrice    float64 `bun:"end_price,notnull" json:"end_price"`
	MinPrice    float64 `bun:"min_price,notnull" json:"min_price"`
	MaxPrice    float64 `bun:"max_price,notnull" json:"max_price"`
	AvgPrice    float64 `bun:"avg_price,notnull" json:"avg_price"`
	MedianPrice float64 `bun:"median_price,notnull" json:"median_price"`

	// Regression analysis
	Slope         float64 `bun:"slope" json:"slope"`
	Intercept     float64 `bun:"intercept" json:"intercept"`
	RSquared      float64 `bun:"r_squared" json:"r_squared"`
	IsSignificant bool    `bun:"is_significant,default:false" json:"is_significant"`

	// Moving averages
	MA7  float64 `bun:"ma_7" json:"ma_7"`
	MA14 float64 `bun:"ma_14" json:"ma_14"`
	MA30 float64 `bun:"ma_30" json:"ma_30"`

	// Metadata
	IsActive  bool      `bun:"is_active,default:true" json:"is_active"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Product *Product `bun:"rel:belongs-to,join:product_id=id" json:"product,omitempty"`
	Store   *Store   `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
}

// PriceAlert represents user-configured price alerts
type PriceAlert struct {
	bun.BaseModel `bun:"table:price_alerts,alias:pa"`

	ID        int64 `bun:"id,pk,autoincrement" json:"id"`
	UserID    int   `bun:"user_id,notnull" json:"user_id"`
	ProductID int   `bun:"product_id,notnull" json:"product_id"`
	StoreID   *int  `bun:"store_id" json:"store_id,omitempty"`

	// Alert configuration
	AlertType   string   `bun:"alert_type,notnull" json:"alert_type"` // 'PRICE_DROP', 'TARGET_PRICE', 'PERCENTAGE_DROP'
	TargetPrice float64  `bun:"target_price,notnull" json:"target_price"`
	DropPercent *float64 `bun:"drop_percent" json:"drop_percent,omitempty"`
	IsActive    bool     `bun:"is_active,default:true" json:"is_active"`
	NotifyEmail bool     `bun:"notify_email,default:true" json:"notify_email"`
	NotifyPush  bool     `bun:"notify_push,default:false" json:"notify_push"`

	// Trigger information
	LastTriggered *time.Time `bun:"last_triggered" json:"last_triggered,omitempty"`
	TriggerCount  int        `bun:"trigger_count,default:0" json:"trigger_count"`
	LastPrice     *float64   `bun:"last_price" json:"last_price,omitempty"`

	// Metadata
	Notes     *string    `bun:"notes" json:"notes,omitempty"`
	CreatedAt time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	ExpiresAt *time.Time `bun:"expires_at" json:"expires_at,omitempty"`

	// Relations
	User    *User    `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Product *Product `bun:"rel:belongs-to,join:product_id=id" json:"product,omitempty"`
	Store   *Store   `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
}

// Methods for PriceHistory

// IsCurrentlyValid checks if the price entry is currently valid
func (ph *PriceHistory) IsCurrentlyValid() bool {
	now := time.Now()
	return ph.ValidFrom.Before(now.Add(time.Hour)) && // Valid from now or earlier (with 1h grace)
		ph.ValidTo.After(now) && // Valid until after now
		ph.IsActive
}

// IsCurrentlySale checks if this price represents a current sale
func (ph *PriceHistory) IsCurrentlySale() bool {
	if !ph.IsOnSale {
		return false
	}

	now := time.Now()
	if ph.SaleStartDate != nil && ph.SaleStartDate.After(now) {
		return false
	}
	if ph.SaleEndDate != nil && ph.SaleEndDate.Before(now) {
		return false
	}

	return true
}

// GetDiscountAmount returns the discount amount if on sale
func (ph *PriceHistory) GetDiscountAmount() float64 {
	if !ph.IsOnSale || ph.OriginalPrice == nil {
		return 0.0
	}
	if ph.Price >= *ph.OriginalPrice {
		return 0.0
	}
	return *ph.OriginalPrice - ph.Price
}

// GetDiscountPercent returns the discount percentage if on sale
func (ph *PriceHistory) GetDiscountPercent() float64 {
	if !ph.IsOnSale || ph.OriginalPrice == nil || *ph.OriginalPrice <= 0 {
		return 0.0
	}
	if ph.Price >= *ph.OriginalPrice {
		return 0.0
	}
	return (((*ph.OriginalPrice - ph.Price) / *ph.OriginalPrice) * 100)
}

// IsExpired checks if this price entry has expired
func (ph *PriceHistory) IsExpired() bool {
	return ph.ValidTo.Before(time.Now())
}

// GetValidityDuration returns how long this price was/is valid
func (ph *PriceHistory) GetValidityDuration() time.Duration {
	return ph.ValidTo.Sub(ph.ValidFrom)
}

// Methods for PriceTrend

// IsUpTrend checks if the trend is upward
func (pt *PriceTrend) IsUpTrend() bool {
	return pt.Direction == "RISING" && pt.TrendPercent > 0
}

// IsDownTrend checks if the trend is downward
func (pt *PriceTrend) IsDownTrend() bool {
	return pt.Direction == "FALLING" && pt.TrendPercent < 0
}

// IsStable checks if the price is stable
func (pt *PriceTrend) IsStable() bool {
	return pt.Direction == "STABLE"
}

// IsVolatile checks if the price is volatile
func (pt *PriceTrend) IsVolatile() bool {
	return pt.Direction == "VOLATILE" || pt.VolatilityScore > 0.3
}

// GetTrendStrength returns the strength of the trend
func (pt *PriceTrend) GetTrendStrength() string {
	absPercent := pt.TrendPercent
	if absPercent < 0 {
		absPercent = -absPercent
	}

	if pt.Confidence < 0.3 || absPercent < 2 {
		return "WEAK"
	} else if pt.Confidence < 0.7 || absPercent < 10 {
		return "MODERATE"
	}
	return "STRONG"
}

// IsCurrentTrend checks if this trend analysis is recent
func (pt *PriceTrend) IsCurrentTrend() bool {
	// Consider trend current if calculated within last 24 hours
	return time.Since(pt.CalculatedAt) < 24*time.Hour
}

// Methods for PriceAlert

// ShouldTrigger checks if the alert should be triggered for the given price
func (pa *PriceAlert) ShouldTrigger(currentPrice float64) bool {
	if !pa.IsActive {
		return false
	}

	if pa.ExpiresAt != nil && pa.ExpiresAt.Before(time.Now()) {
		return false
	}

	switch pa.AlertType {
	case "TARGET_PRICE":
		return currentPrice <= pa.TargetPrice
	case "PRICE_DROP":
		if pa.LastPrice == nil {
			return false
		}
		return currentPrice < *pa.LastPrice
	case "PERCENTAGE_DROP":
		if pa.LastPrice == nil || pa.DropPercent == nil {
			return false
		}
		dropPercent := ((*pa.LastPrice - currentPrice) / *pa.LastPrice) * 100
		return dropPercent >= *pa.DropPercent
	default:
		return false
	}
}

// UpdateLastPrice updates the last known price for comparison
func (pa *PriceAlert) UpdateLastPrice(price float64) {
	pa.LastPrice = &price
	pa.UpdatedAt = time.Now()
}

// MarkTriggered marks the alert as triggered
func (pa *PriceAlert) MarkTriggered() {
	now := time.Now()
	pa.LastTriggered = &now
	pa.TriggerCount++
	pa.UpdatedAt = now
}

// IsActive checks if the alert is currently active
func (pa *PriceAlert) IsActiveAlert() bool {
	if !pa.IsActive {
		return false
	}
	if pa.ExpiresAt != nil && pa.ExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// TableName methods for Bun
func (ph *PriceHistory) TableName() string {
	return "price_history"
}

func (pt *PriceTrend) TableName() string {
	return "price_trends"
}

func (pa *PriceAlert) TableName() string {
	return "price_alerts"
}
