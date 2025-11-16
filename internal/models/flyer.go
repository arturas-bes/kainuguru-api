package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type Flyer struct {
	bun.BaseModel `bun:"table:flyers,alias:f"`

	ID      int `bun:"id,pk,autoincrement" json:"id"`
	StoreID int `bun:"store_id,notnull" json:"store_id"`

	// Basic information
	Title     *string   `bun:"title" json:"title,omitempty"`
	ValidFrom time.Time `bun:"valid_from,notnull" json:"valid_from"`
	ValidTo   time.Time `bun:"valid_to,notnull" json:"valid_to"`
	PageCount *int      `bun:"page_count" json:"page_count,omitempty"`
	SourceURL *string   `bun:"source_url" json:"source_url,omitempty"`

	// Archival status
	IsArchived bool       `bun:"is_archived,default:false" json:"is_archived"`
	ArchivedAt *time.Time `bun:"archived_at" json:"archived_at,omitempty"`

	// Processing metadata
	Status                string     `bun:"status,default:'pending'" json:"status"`
	ExtractionStartedAt   *time.Time `bun:"extraction_started_at" json:"extraction_started_at,omitempty"`
	ExtractionCompletedAt *time.Time `bun:"extraction_completed_at" json:"extraction_completed_at,omitempty"`
	ProductsExtracted     int        `bun:"products_extracted,default:0" json:"products_extracted"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Store      *Store       `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
	FlyerPages []*FlyerPage `bun:"rel:has-many,join:id=flyer_id" json:"flyer_pages,omitempty"`
	Products   []*Product   `bun:"rel:has-many,join:id=flyer_id" json:"products,omitempty"`
}

// FlyerStatus represents possible flyer processing statuses
type FlyerStatus string

const (
	FlyerStatusPending    FlyerStatus = "pending"
	FlyerStatusProcessing FlyerStatus = "processing"
	FlyerStatusCompleted  FlyerStatus = "completed"
	FlyerStatusFailed     FlyerStatus = "failed"
)

// IsValid checks if the flyer is currently valid
func (f *Flyer) IsValid() bool {
	now := time.Now()
	return !f.IsArchived &&
		f.ValidFrom.Before(now.Add(24*time.Hour)) && // Valid from today or earlier
		f.ValidTo.After(now) // Valid until after now
}

// IsCurrent checks if the flyer is for the current week
func (f *Flyer) IsCurrent() bool {
	now := time.Now()
	// Get start of current week (Monday)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()-time.Monday))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	weekEnd := weekStart.AddDate(0, 0, 7)

	return f.IsValid() &&
		f.ValidFrom.Before(weekEnd) &&
		f.ValidTo.After(weekStart)
}

// GetDaysRemaining returns the number of days until the flyer expires
func (f *Flyer) GetDaysRemaining() int {
	if f.IsArchived {
		return 0
	}

	days := int(time.Until(f.ValidTo).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// IsProcessingComplete checks if extraction is complete
func (f *Flyer) IsProcessingComplete() bool {
	return f.Status == string(FlyerStatusCompleted) && f.ExtractionCompletedAt != nil
}

// CanBeProcessed checks if the flyer can be processed
func (f *Flyer) CanBeProcessed() bool {
	return f.IsValid() &&
		f.Status != string(FlyerStatusCompleted) &&
		f.Status != string(FlyerStatusProcessing)
}

// GetProcessingDuration returns how long the processing took or is taking
func (f *Flyer) GetProcessingDuration() *time.Duration {
	if f.ExtractionStartedAt == nil {
		return nil
	}

	endTime := f.ExtractionCompletedAt
	if endTime == nil {
		now := time.Now()
		endTime = &now
	}

	duration := endTime.Sub(*f.ExtractionStartedAt)
	return &duration
}

// StartProcessing marks the flyer as being processed
func (f *Flyer) StartProcessing() {
	now := time.Now()
	f.Status = string(FlyerStatusProcessing)
	f.ExtractionStartedAt = &now
	f.UpdatedAt = now
}

// CompleteProcessing marks the flyer as processing complete
func (f *Flyer) CompleteProcessing(productsExtracted int) {
	now := time.Now()
	f.Status = string(FlyerStatusCompleted)
	f.ExtractionCompletedAt = &now
	f.ProductsExtracted = productsExtracted
	f.UpdatedAt = now
}

// FailProcessing marks the flyer processing as failed
func (f *Flyer) FailProcessing() {
	now := time.Now()
	f.Status = string(FlyerStatusFailed)
	f.UpdatedAt = now
}

// Archive marks the flyer as archived
func (f *Flyer) Archive() {
	now := time.Now()
	f.IsArchived = true
	f.ArchivedAt = &now
	f.UpdatedAt = now
}

// GetValidityPeriod returns a human-readable validity period
func (f *Flyer) GetValidityPeriod() string {
	layout := "2006-01-02"
	return f.ValidFrom.Format(layout) + " - " + f.ValidTo.Format(layout)
}

// TableName returns the table name for Bun
func (f *Flyer) TableName() string {
	return "flyers"
}

// GetFolderName returns the folder name for this flyer's images
// Format: YYYY-MM-DD-store-title-slug
func (f *Flyer) GetFolderName() string {
	if f.Store == nil {
		return ""
	}

	dateStr := f.ValidFrom.Format("2006-01-02")
	storeCode := strings.ToLower(f.Store.Code)

	title := "leidinys" // default
	if f.Title != nil && *f.Title != "" {
		title = slugify(*f.Title)
	}

	return fmt.Sprintf("%s-%s-%s", dateStr, storeCode, title)
}

// GetImageBasePath returns the base directory path for this flyer's images
// Format: flyers/{store}/{folder-name}
func (f *Flyer) GetImageBasePath() string {
	if f.Store == nil {
		return ""
	}

	storeCode := strings.ToLower(f.Store.Code)
	folderName := f.GetFolderName()

	return fmt.Sprintf("flyers/%s/%s", storeCode, folderName)
}

// slugify converts a title to URL-friendly slug
func slugify(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace Lithuanian characters
	replacements := map[string]string{
		"ą": "a", "č": "c", "ę": "e", "ė": "e",
		"į": "i", "š": "s", "ų": "u", "ū": "u", "ž": "z",
	}
	for lt, en := range replacements {
		slug = strings.ReplaceAll(slug, lt, en)
	}

	// Remove non-alphanumeric characters (keep spaces for now)
	reg := regexp.MustCompile("[^a-z0-9 -]")
	slug = reg.ReplaceAllString(slug, "")

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start/end
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 30 {
		slug = slug[:30]
		// Trim trailing hyphen if cut in middle of word
		slug = strings.TrimRight(slug, "-")
	}

	if slug == "" {
		slug = "leidinys"
	}

	return slug
}
