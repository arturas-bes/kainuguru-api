package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type ProductMaster struct {
	bun.BaseModel `bun:"table:product_masters,alias:pm"`

	ID               int64    `bun:"id,pk,autoincrement" json:"id"`
	Name             string   `bun:"name,notnull" json:"name"`
	NormalizedName   string   `bun:"normalized_name,notnull" json:"normalized_name"`
	Brand            *string  `bun:"brand" json:"brand"`
	Description      *string  `bun:"description" json:"description"`

	// Categorization
	Category         *string  `bun:"category" json:"category"`
	Subcategory      *string  `bun:"subcategory" json:"subcategory"`
	Tags             []string `bun:"tags,array" json:"tags"`

	// Standard units and packaging
	StandardUnit     *string  `bun:"standard_unit" json:"standard_unit"`
	UnitType         *string  `bun:"unit_type" json:"unit_type"`
	StandardSize     *float64 `bun:"standard_size" json:"standard_size"`
	PackagingVariants []string `bun:"packaging_variants,array" json:"packaging_variants"`

	// Product identifiers and matching
	Barcode           *string  `bun:"barcode" json:"barcode"`
	ManufacturerCode  *string  `bun:"manufacturer_code" json:"manufacturer_code"`
	AlternativeNames  []string `bun:"alternative_names,array" json:"alternative_names"`

	// Statistics and quality metrics
	MatchCount        int      `bun:"match_count,default:0" json:"match_count"`
	ConfidenceScore   float64  `bun:"confidence_score,default:0" json:"confidence_score"`
	LastSeenDate      *time.Time `bun:"last_seen_date" json:"last_seen_date"`

	// Price tracking
	AvgPrice          *float64   `bun:"avg_price" json:"avg_price"`
	MinPrice          *float64   `bun:"min_price" json:"min_price"`
	MaxPrice          *float64   `bun:"max_price" json:"max_price"`
	PriceTrend        string     `bun:"price_trend,default:'stable'" json:"price_trend"`
	LastPriceUpdate   *time.Time `bun:"last_price_update" json:"last_price_update"`

	// Availability and popularity
	AvailabilityScore float64 `bun:"availability_score,default:0" json:"availability_score"`
	PopularityScore   float64 `bun:"popularity_score,default:0" json:"popularity_score"`
	SeasonalAvailability json.RawMessage `bun:"seasonal_availability,type:jsonb" json:"seasonal_availability"`

	// Quality and user preferences
	UserRating        *float64 `bun:"user_rating" json:"user_rating"`
	ReviewCount       int      `bun:"review_count,default:0" json:"review_count"`
	NutritionalInfo   json.RawMessage `bun:"nutritional_info,type:jsonb" json:"nutritional_info"`
	Allergens         []string `bun:"allergens,array" json:"allergens"`

	// Search and matching
	SearchVector      string   `bun:"search_vector" json:"-"`
	MatchKeywords     []string `bun:"match_keywords,array" json:"match_keywords"`

	// Status and lifecycle
	Status            string   `bun:"status,default:'active'" json:"status"`
	MergedIntoID      *int64   `bun:"merged_into_id" json:"merged_into_id"`

	// Timestamps
	CreatedAt         time.Time `bun:"created_at,default:now()" json:"created_at"`
	UpdatedAt         time.Time `bun:"updated_at,default:now()" json:"updated_at"`

	// Relations
	Products          []*Product `bun:"rel:has-many,join:id=product_master_id" json:"products,omitempty"`
	MergedInto        *ProductMaster `bun:"rel:belongs-to,join:merged_into_id=id" json:"merged_into,omitempty"`
	ShoppingListItems []*ShoppingListItem `bun:"rel:has-many,join:id=product_master_id" json:"shopping_list_items,omitempty"`
}

// ProductMasterStatus represents possible product master statuses
type ProductMasterStatus string

const (
	ProductMasterStatusActive   ProductMasterStatus = "active"
	ProductMasterStatusInactive ProductMasterStatus = "inactive"
	ProductMasterStatusMerged   ProductMasterStatus = "merged"
	ProductMasterStatusDeleted  ProductMasterStatus = "deleted"
)

// PriceTrend represents price trend directions
type PriceTrend string

const (
	PriceTrendIncreasing PriceTrend = "increasing"
	PriceTrendDecreasing PriceTrend = "decreasing"
	PriceTrendStable     PriceTrend = "stable"
)

// MatchingKeywords represents keywords used for product matching
type MatchingKeywords struct {
	Primary   []string `json:"primary"`
	Secondary []string `json:"secondary"`
	Brand     []string `json:"brand"`
	Category  []string `json:"category"`
}

// AlternativeNames represents alternative names for the product
type AlternativeNames struct {
	Names     []string `json:"names"`
	Aliases   []string `json:"aliases"`
	Abbreviations []string `json:"abbreviations"`
}

// ExclusionKeywords represents keywords that exclude this product from matching
type ExclusionKeywords struct {
	Exclusions []string `json:"exclusions"`
	Conflicts  []string `json:"conflicts"`
}

// GetMatchingKeywords returns the match keywords as strings
func (pm *ProductMaster) GetMatchingKeywords() []string {
	return pm.MatchKeywords
}

// SetMatchingKeywords sets the match keywords
func (pm *ProductMaster) SetMatchingKeywords(keywords []string) {
	pm.MatchKeywords = keywords
}

// GetAlternativeNames returns the alternative names
func (pm *ProductMaster) GetAlternativeNames() []string {
	return pm.AlternativeNames
}

// SetAlternativeNames sets the alternative names
func (pm *ProductMaster) SetAlternativeNames(names []string) {
	pm.AlternativeNames = names
}

// NormalizeName creates a normalized version of the name
func (pm *ProductMaster) NormalizeName() {
	normalized := strings.ToLower(pm.Name)

	// Lithuanian character normalization
	normalized = strings.ReplaceAll(normalized, "ą", "a")
	normalized = strings.ReplaceAll(normalized, "č", "c")
	normalized = strings.ReplaceAll(normalized, "ę", "e")
	normalized = strings.ReplaceAll(normalized, "ė", "e")
	normalized = strings.ReplaceAll(normalized, "į", "i")
	normalized = strings.ReplaceAll(normalized, "š", "s")
	normalized = strings.ReplaceAll(normalized, "ų", "u")
	normalized = strings.ReplaceAll(normalized, "ū", "u")
	normalized = strings.ReplaceAll(normalized, "ž", "z")

	// Remove punctuation and normalize spaces
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = strings.ReplaceAll(normalized, ".", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "/", " ")
	normalized = strings.ReplaceAll(normalized, "(", " ")
	normalized = strings.ReplaceAll(normalized, ")", " ")

	// Normalize multiple spaces
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	pm.NormalizedName = strings.TrimSpace(normalized)
}

// IsActive checks if the product master is active
func (pm *ProductMaster) IsActive() bool {
	return pm.Status == string(ProductMasterStatusActive)
}

// GetMatchSuccessRate returns the success rate of product matching
func (pm *ProductMaster) GetMatchSuccessRate() float64 {
	if pm.MatchCount == 0 {
		return 0.0
	}
	// For now, assume all matches are successful (can be enhanced later)
	return 100.0
}

// IncrementMatchCount increments the match counter
func (pm *ProductMaster) IncrementMatchCount() {
	pm.MatchCount++
	pm.LastSeenDate = &time.Time{}
	*pm.LastSeenDate = time.Now()
	pm.UpdatedAt = time.Now()
}

// Deactivate marks the product master as inactive
func (pm *ProductMaster) Deactivate() {
	pm.Status = string(ProductMasterStatusInactive)
	pm.UpdatedAt = time.Now()
}

// MarkAsMerged marks the product master as merged into another
func (pm *ProductMaster) MarkAsMerged(targetID int64) {
	pm.Status = string(ProductMasterStatusMerged)
	pm.MergedIntoID = &targetID
	pm.UpdatedAt = time.Now()
}

// CanBeMatched checks if the product master can be used for matching
func (pm *ProductMaster) CanBeMatched() bool {
	return pm.IsActive() && pm.ConfidenceScore >= 0.3
}

// AddMatchingKeyword adds a new keyword to the matching keywords
func (pm *ProductMaster) AddMatchingKeyword(keyword string) {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return
	}

	// Check if keyword already exists
	for _, existing := range pm.MatchKeywords {
		if existing == keyword {
			return
		}
	}

	pm.MatchKeywords = append(pm.MatchKeywords, keyword)
	pm.UpdatedAt = time.Now()
}

// AddAlternativeName adds a new alternative name
func (pm *ProductMaster) AddAlternativeName(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}

	// Check if name already exists
	for _, existing := range pm.AlternativeNames {
		if existing == name {
			return
		}
	}

	pm.AlternativeNames = append(pm.AlternativeNames, name)
	pm.UpdatedAt = time.Now()
}

// GetAllMatchableTerms returns all terms that can be used for matching
func (pm *ProductMaster) GetAllMatchableTerms() []string {
	var terms []string

	terms = append(terms, pm.Name)
	terms = append(terms, pm.NormalizedName)
	if pm.Brand != nil {
		terms = append(terms, *pm.Brand)
	}

	// Add keywords
	terms = append(terms, pm.GetMatchingKeywords()...)

	// Add alternative names
	terms = append(terms, pm.GetAlternativeNames()...)

	return terms
}

// TableName returns the table name for Bun
func (pm *ProductMaster) TableName() string {
	return "product_masters"
}