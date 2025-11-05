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
	PageNumber int     `bun:"page_number,notnull" json:"page_number"`
	ImageURL   *string `bun:"image_url" json:"image_url,omitempty"`

	// Processing status
	ExtractionStatus   string  `bun:"extraction_status,default:'pending'" json:"extraction_status"`
	ExtractionAttempts int     `bun:"extraction_attempts,default:0" json:"extraction_attempts"`
	ExtractionError    *string `bun:"extraction_error" json:"extraction_error,omitempty"`
	NeedsManualReview  bool    `bun:"needs_manual_review,default:false" json:"needs_manual_review"`

	// Raw extraction data as JSONB
	RawExtractionData map[string]interface{} `bun:"raw_extraction_data,type:jsonb" json:"raw_extraction_data,omitempty"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Flyer *Flyer `bun:"rel:belongs-to,join:flyer_id=id" json:"flyer,omitempty"`
}

// FlyerPageStatus represents the processing status of a flyer page
type FlyerPageStatus string

const (
	FlyerPageStatusPending    FlyerPageStatus = "pending"
	FlyerPageStatusProcessing FlyerPageStatus = "processing"
	FlyerPageStatusCompleted  FlyerPageStatus = "completed"
	FlyerPageStatusFailed     FlyerPageStatus = "failed"
)

// IsProcessingComplete checks if extraction is complete
func (fp *FlyerPage) IsProcessingComplete() bool {
	return fp.ExtractionStatus == string(FlyerPageStatusCompleted)
}

// CanBeProcessed checks if the page can be processed
func (fp *FlyerPage) CanBeProcessed() bool {
	return fp.ExtractionStatus != string(FlyerPageStatusCompleted) &&
		fp.ExtractionStatus != string(FlyerPageStatusProcessing) &&
		fp.ExtractionAttempts < 3 // Max 3 retry attempts
}

// HasImage checks if the page has a valid image URL
func (fp *FlyerPage) HasImage() bool {
	return fp.ImageURL != nil && *fp.ImageURL != ""
}

// StartProcessing marks the page as being processed
func (fp *FlyerPage) StartProcessing() {
	fp.ExtractionStatus = string(FlyerPageStatusProcessing)
	fp.UpdatedAt = time.Now()
}

// CompleteProcessing marks the page as completed
func (fp *FlyerPage) CompleteProcessing() {
	fp.ExtractionStatus = string(FlyerPageStatusCompleted)
	fp.UpdatedAt = time.Now()
}

// FailProcessing marks the page as failed
func (fp *FlyerPage) FailProcessing(errorMsg string) {
	fp.ExtractionStatus = string(FlyerPageStatusFailed)
	fp.ExtractionAttempts++
	fp.ExtractionError = &errorMsg
	fp.UpdatedAt = time.Now()

	// Mark for manual review if too many failures
	if fp.ExtractionAttempts >= 3 {
		fp.NeedsManualReview = true
	}
}

// ResetForRetry resets the page for another processing attempt
func (fp *FlyerPage) ResetForRetry() {
	fp.ExtractionStatus = string(FlyerPageStatusPending)
	fp.UpdatedAt = time.Now()
}