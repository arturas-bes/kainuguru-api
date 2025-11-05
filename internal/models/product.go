package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type Product struct {
	bun.BaseModel `bun:"table:products,alias:p"`

	ID              int    `bun:"id,pk,autoincrement" json:"id"`
	FlyerID         int    `bun:"flyer_id,notnull" json:"flyer_id"`
	FlyerPageID     *int   `bun:"flyer_page_id" json:"flyer_page_id,omitempty"`
	StoreID         int    `bun:"store_id,notnull" json:"store_id"`
	ProductMasterID *int   `bun:"product_master_id" json:"product_master_id,omitempty"`

	// Basic product information
	Name            string  `bun:"name,notnull" json:"name"`
	NormalizedName  string  `bun:"normalized_name,notnull" json:"normalized_name"`
	Brand           *string `bun:"brand" json:"brand,omitempty"`
	Category        *string `bun:"category" json:"category,omitempty"`
	Subcategory     *string `bun:"subcategory" json:"subcategory,omitempty"`
	Description     *string `bun:"description" json:"description,omitempty"`

	// Pricing information
	CurrentPrice    float64  `bun:"current_price,notnull" json:"current_price"`
	OriginalPrice   *float64 `bun:"original_price" json:"original_price,omitempty"`
	DiscountPercent *float64 `bun:"discount_percent" json:"discount_percent,omitempty"`
	Currency        string   `bun:"currency,default:'EUR'" json:"currency"`

	// Product specifications
	UnitSize     *string `bun:"unit_size" json:"unit_size,omitempty"`
	UnitType     *string `bun:"unit_type" json:"unit_type,omitempty"`
	UnitPrice    *string `bun:"unit_price" json:"unit_price,omitempty"`
	PackageSize  *string `bun:"package_size" json:"package_size,omitempty"`
	Weight       *string `bun:"weight" json:"weight,omitempty"`
	Volume       *string `bun:"volume" json:"volume,omitempty"`

	// Image and location data
	ImageURL      *string           `bun:"image_url" json:"image_url,omitempty"`
	BoundingBox   *ProductBoundingBox `bun:"bounding_box,type:jsonb" json:"bounding_box,omitempty"`
	PagePosition  *ProductPosition  `bun:"page_position,type:jsonb" json:"page_position,omitempty"`

	// Availability and promotion
	IsOnSale      bool       `bun:"is_on_sale,default:false" json:"is_on_sale"`
	SaleStartDate *time.Time `bun:"sale_start_date" json:"sale_start_date,omitempty"`
	SaleEndDate   *time.Time `bun:"sale_end_date" json:"sale_end_date,omitempty"`
	IsAvailable   bool       `bun:"is_available,default:true" json:"is_available"`
	StockLevel    *string    `bun:"stock_level" json:"stock_level,omitempty"`

	// Extraction metadata
	ExtractionConfidence float64 `bun:"extraction_confidence,default:0.0" json:"extraction_confidence"`
	ExtractionMethod     string  `bun:"extraction_method,default:'ocr'" json:"extraction_method"`
	RequiresReview       bool    `bun:"requires_review,default:false" json:"requires_review"`

	// Search optimization
	SearchVector string `bun:"search_vector" json:"-"`

	// Validity period (partitioning key)
	ValidFrom time.Time `bun:"valid_from,notnull" json:"valid_from"`
	ValidTo   time.Time `bun:"valid_to,notnull" json:"valid_to"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Flyer           *Flyer          `bun:"rel:belongs-to,join:flyer_id=id" json:"flyer,omitempty"`
	FlyerPage       *FlyerPage      `bun:"rel:belongs-to,join:flyer_page_id=id" json:"flyer_page,omitempty"`
	Store           *Store          `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
	ProductMaster   *ProductMaster  `bun:"rel:belongs-to,join:product_master_id=id" json:"product_master,omitempty"`
}

// ProductBoundingBox represents the position of a product on a flyer page
type ProductBoundingBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ProductPosition represents additional positioning metadata
type ProductPosition struct {
	Row    int `json:"row"`
	Column int `json:"column"`
	Zone   string `json:"zone"` // e.g., "header", "main", "footer", "sidebar"
}

// Implement SQL driver interfaces for JSON fields
func (pb *ProductBoundingBox) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ProductBoundingBox")
	}

	return json.Unmarshal(bytes, pb)
}

func (pb ProductBoundingBox) Value() (driver.Value, error) {
	return json.Marshal(pb)
}

func (pp *ProductPosition) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan ProductPosition")
	}

	return json.Unmarshal(bytes, pp)
}

func (pp ProductPosition) Value() (driver.Value, error) {
	return json.Marshal(pp)
}

// IsCurrentlyOnSale checks if the product is currently on sale
func (p *Product) IsCurrentlyOnSale() bool {
	if !p.IsOnSale {
		return false
	}

	now := time.Now()
	if p.SaleStartDate != nil && p.SaleStartDate.After(now) {
		return false
	}
	if p.SaleEndDate != nil && p.SaleEndDate.Before(now) {
		return false
	}

	return true
}

// GetDiscountAmount returns the discount amount in currency
func (p *Product) GetDiscountAmount() float64 {
	if p.OriginalPrice == nil || p.CurrentPrice >= *p.OriginalPrice {
		return 0.0
	}
	return *p.OriginalPrice - p.CurrentPrice
}

// CalculateDiscountPercent calculates and updates the discount percentage
func (p *Product) CalculateDiscountPercent() {
	if p.OriginalPrice == nil || *p.OriginalPrice <= 0 {
		p.DiscountPercent = nil
		return
	}

	if p.CurrentPrice >= *p.OriginalPrice {
		p.DiscountPercent = nil
		p.IsOnSale = false
		return
	}

	discount := ((*p.OriginalPrice - p.CurrentPrice) / *p.OriginalPrice) * 100
	p.DiscountPercent = &discount
	p.IsOnSale = true
}

// NormalizeName creates a normalized version of the product name for search
func (p *Product) NormalizeName() {
	// Basic Lithuanian text normalization
	normalized := strings.ToLower(p.Name)

	// Remove special characters and extra spaces
	normalized = strings.ReplaceAll(normalized, "ą", "a")
	normalized = strings.ReplaceAll(normalized, "č", "c")
	normalized = strings.ReplaceAll(normalized, "ę", "e")
	normalized = strings.ReplaceAll(normalized, "ė", "e")
	normalized = strings.ReplaceAll(normalized, "į", "i")
	normalized = strings.ReplaceAll(normalized, "š", "s")
	normalized = strings.ReplaceAll(normalized, "ų", "u")
	normalized = strings.ReplaceAll(normalized, "ū", "u")
	normalized = strings.ReplaceAll(normalized, "ž", "z")

	// Remove common punctuation and normalize spaces
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = strings.ReplaceAll(normalized, ".", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "/", " ")
	normalized = strings.ReplaceAll(normalized, "(", " ")
	normalized = strings.ReplaceAll(normalized, ")", " ")

	// Normalize multiple spaces to single space
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	p.NormalizedName = strings.TrimSpace(normalized)
}

// IsValid checks if the product is currently valid
func (p *Product) IsValid() bool {
	now := time.Now()
	return p.ValidFrom.Before(now.Add(24*time.Hour)) && // Valid from today or earlier
		   p.ValidTo.After(now) // Valid until after now
}

// IsExpired checks if the product has expired
func (p *Product) IsExpired() bool {
	return p.ValidTo.Before(time.Now())
}

// GetValidityPeriod returns a human-readable validity period
func (p *Product) GetValidityPeriod() string {
	layout := "2006-01-02"
	return p.ValidFrom.Format(layout) + " - " + p.ValidTo.Format(layout)
}

// SetBoundingBox sets the bounding box coordinates
func (p *Product) SetBoundingBox(x, y, width, height float64) {
	p.BoundingBox = &ProductBoundingBox{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

// SetPagePosition sets the page position metadata
func (p *Product) SetPagePosition(row, column int, zone string) {
	p.PagePosition = &ProductPosition{
		Row:    row,
		Column: column,
		Zone:   zone,
	}
}

// RequiresManualReview checks if the product needs manual review
func (p *Product) RequiresManualReview() bool {
	return p.RequiresReview ||
		   p.ExtractionConfidence < 0.8 ||
		   p.CurrentPrice <= 0 ||
		   strings.TrimSpace(p.Name) == ""
}

// GetPricePerUnit calculates price per unit if unit information is available
func (p *Product) GetPricePerUnit() *float64 {
	if p.UnitPrice != nil && *p.UnitPrice != "" {
		// Unit price already provided, could parse and return
		return nil // For now, return nil as parsing would require more logic
	}

	// Could implement unit price calculation based on UnitSize and CurrentPrice
	// This would require parsing various unit formats (kg, g, l, ml, etc.)
	return nil
}

// MarkForReview marks the product as requiring manual review
func (p *Product) MarkForReview(reason string) {
	p.RequiresReview = true
	if p.Description == nil {
		p.Description = &reason
	} else {
		combined := *p.Description + " | Review: " + reason
		p.Description = &combined
	}
	p.UpdatedAt = time.Now()
}

// UpdateExtractionMetadata updates extraction confidence and method
func (p *Product) UpdateExtractionMetadata(confidence float64, method string) {
	p.ExtractionConfidence = confidence
	p.ExtractionMethod = method
	p.RequiresReview = confidence < 0.8
	p.UpdatedAt = time.Now()
}

// TableName returns the table name for Bun
func (p *Product) TableName() string {
	return "products"
}