package models

import (
	"time"

	"github.com/uptrace/bun"
)

type FlyerPage struct {
	bun.BaseModel `bun:"table:flyer_pages,alias:fp"`

	ID      int `bun:"id,pk,autoincrement" json:"id"`
	FlyerID int `bun:"flyer_id,notnull" json:"flyer_id"`

	// Page information
	PageNumber  int     `bun:"page_number,notnull" json:"page_number"`
	ImageURL    *string `bun:"image_url" json:"image_url,omitempty"`
	ImageWidth  *int    `bun:"image_width" json:"image_width,omitempty"`
	ImageHeight *int    `bun:"image_height" json:"image_height,omitempty"`

	// Processing status
	Status                string     `bun:"status,default:'pending'" json:"status"`
	ExtractionStartedAt   *time.Time `bun:"extraction_started_at" json:"extraction_started_at,omitempty"`
	ExtractionCompletedAt *time.Time `bun:"extraction_completed_at" json:"extraction_completed_at,omitempty"`
	ProductsExtracted     int        `bun:"products_extracted,default:0" json:"products_extracted"`

	// Error tracking
	ExtractionErrors    int        `bun:"extraction_errors,default:0" json:"extraction_errors"`
	LastExtractionError *string    `bun:"last_extraction_error" json:"last_extraction_error,omitempty"`
	LastErrorAt         *time.Time `bun:"last_error_at" json:"last_error_at,omitempty"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Flyer    *Flyer     `bun:"rel:belongs-to,join:flyer_id=id" json:"flyer,omitempty"`
	Products []*Product `bun:"rel:has-many,join:id=flyer_page_id" json:"products,omitempty"`
}

// FlyerPageStatus represents possible flyer page processing statuses
type FlyerPageStatus string

const (
	FlyerPageStatusPending    FlyerPageStatus = "pending"
	FlyerPageStatusProcessing FlyerPageStatus = "processing"
	FlyerPageStatusCompleted  FlyerPageStatus = "completed"
	FlyerPageStatusFailed     FlyerPageStatus = "failed"
)

// IsProcessingComplete checks if extraction is complete
func (fp *FlyerPage) IsProcessingComplete() bool {
	return fp.Status == string(FlyerPageStatusCompleted) && fp.ExtractionCompletedAt != nil
}

// CanBeProcessed checks if the page can be processed
func (fp *FlyerPage) CanBeProcessed() bool {
	return fp.Status != string(FlyerPageStatusCompleted) &&
		   fp.Status != string(FlyerPageStatusProcessing) &&
		   fp.ExtractionErrors < 3 // Max 3 retry attempts
}

// HasImage checks if the page has a valid image URL
func (fp *FlyerPage) HasImage() bool {
	return fp.ImageURL != nil && *fp.ImageURL != ""
}

// GetImageDimensions returns image dimensions if available
func (fp *FlyerPage) GetImageDimensions() (width, height int, hasData bool) {
	if fp.ImageWidth != nil && fp.ImageHeight != nil {
		return *fp.ImageWidth, *fp.ImageHeight, true
	}
	return 0, 0, false
}

// GetProcessingDuration returns how long the processing took or is taking
func (fp *FlyerPage) GetProcessingDuration() *time.Duration {
	if fp.ExtractionStartedAt == nil {
		return nil
	}

	endTime := fp.ExtractionCompletedAt
	if endTime == nil {
		now := time.Now()
		endTime = &now
	}

	duration := endTime.Sub(*fp.ExtractionStartedAt)
	return &duration
}

// StartProcessing marks the page as being processed
func (fp *FlyerPage) StartProcessing() {
	now := time.Now()
	fp.Status = string(FlyerPageStatusProcessing)
	fp.ExtractionStartedAt = &now
	fp.UpdatedAt = now
}

// CompleteProcessing marks the page as processing complete
func (fp *FlyerPage) CompleteProcessing(productsExtracted int) {
	now := time.Now()
	fp.Status = string(FlyerPageStatusCompleted)
	fp.ExtractionCompletedAt = &now
	fp.ProductsExtracted = productsExtracted
	fp.UpdatedAt = now
}

// FailProcessing marks the page processing as failed
func (fp *FlyerPage) FailProcessing(errorMsg string) {
	now := time.Now()
	fp.Status = string(FlyerPageStatusFailed)
	fp.ExtractionErrors++
	fp.LastExtractionError = &errorMsg
	fp.LastErrorAt = &now
	fp.UpdatedAt = now
}

// ResetForRetry resets the page for retry processing
func (fp *FlyerPage) ResetForRetry() {
	now := time.Now()
	fp.Status = string(FlyerPageStatusPending)
	fp.ExtractionStartedAt = nil
	fp.ExtractionCompletedAt = nil
	fp.UpdatedAt = now
}

// SetImageDimensions sets the image dimensions
func (fp *FlyerPage) SetImageDimensions(width, height int) {
	fp.ImageWidth = &width
	fp.ImageHeight = &height
	fp.UpdatedAt = time.Now()
}

// IsRetryable checks if the page can be retried
func (fp *FlyerPage) IsRetryable() bool {
	return fp.Status == string(FlyerPageStatusFailed) && fp.ExtractionErrors < 3
}

// GetExtractionEfficiency returns the success rate as a percentage
func (fp *FlyerPage) GetExtractionEfficiency() float64 {
	if fp.ExtractionErrors == 0 && fp.ProductsExtracted > 0 {
		return 100.0
	}

	totalAttempts := fp.ExtractionErrors
	if fp.Status == string(FlyerPageStatusCompleted) || fp.ProductsExtracted > 0 {
		totalAttempts++ // Count successful attempt
	}

	if totalAttempts == 0 {
		return 0.0
	}

	successfulAttempts := 0
	if fp.ProductsExtracted > 0 {
		successfulAttempts = 1
	}

	return float64(successfulAttempts) / float64(totalAttempts) * 100.0
}

// TableName returns the table name for Bun
func (fp *FlyerPage) TableName() string {
	return "flyer_pages"
}