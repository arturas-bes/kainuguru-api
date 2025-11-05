package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type ExtractionJob struct {
	bun.BaseModel `bun:"table:extraction_jobs,alias:ej"`

	ID       int64  `bun:"id,pk,autoincrement" json:"id"`
	JobType  string `bun:"job_type,notnull" json:"job_type"`
	Status   string `bun:"status,default:'pending'" json:"status"`
	Priority int    `bun:"priority,default:5" json:"priority"`

	// Job payload
	Payload json.RawMessage `bun:"payload,type:jsonb,notnull" json:"payload"`

	// Processing metadata
	Attempts    int     `bun:"attempts,default:0" json:"attempts"`
	MaxAttempts int     `bun:"max_attempts,default:3" json:"max_attempts"`
	WorkerID    *string `bun:"worker_id" json:"worker_id,omitempty"`
	StartedAt   *time.Time `bun:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `bun:"completed_at" json:"completed_at,omitempty"`

	// Error tracking
	ErrorMessage *string `bun:"error_message" json:"error_message,omitempty"`
	ErrorCount   int     `bun:"error_count,default:0" json:"error_count"`

	// Scheduling
	ScheduledFor time.Time  `bun:"scheduled_for,default:current_timestamp" json:"scheduled_for"`
	ExpiresAt    *time.Time `bun:"expires_at" json:"expires_at,omitempty"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// ExtractionJobType represents different types of extraction jobs
type ExtractionJobType string

const (
	JobTypeScrapeFlyer   ExtractionJobType = "scrape_flyer"
	JobTypeExtractPage   ExtractionJobType = "extract_page"
	JobTypeMatchProducts ExtractionJobType = "match_products"
	JobTypeGenerateFlyer ExtractionJobType = "generate_flyer"
	JobTypeProcessImage  ExtractionJobType = "process_image"
	JobTypeValidateData  ExtractionJobType = "validate_data"
)

// ExtractionJobStatus represents job processing statuses
type ExtractionJobStatus string

const (
	JobStatusPending    ExtractionJobStatus = "pending"
	JobStatusProcessing ExtractionJobStatus = "processing"
	JobStatusCompleted  ExtractionJobStatus = "completed"
	JobStatusFailed     ExtractionJobStatus = "failed"
	JobStatusCancelled  ExtractionJobStatus = "cancelled"
	JobStatusExpired    ExtractionJobStatus = "expired"
)

// ScrapeFlyerPayload represents payload for scrape_flyer jobs
type ScrapeFlyerPayload struct {
	StoreID   int    `json:"store_id"`
	StoreCode string `json:"store_code"`
	SourceURL string `json:"source_url"`
	ForceRescrape bool `json:"force_rescrape,omitempty"`
}

// ExtractPagePayload represents payload for extract_page jobs
type ExtractPagePayload struct {
	FlyerID      int    `json:"flyer_id"`
	FlyerPageID  int    `json:"flyer_page_id"`
	ImageURL     string `json:"image_url"`
	UseGPTVision bool   `json:"use_gpt_vision,omitempty"`
}

// MatchProductsPayload represents payload for match_products jobs
type MatchProductsPayload struct {
	FlyerID    int   `json:"flyer_id"`
	ProductIDs []int `json:"product_ids,omitempty"`
	ForceRematch bool `json:"force_rematch,omitempty"`
}

// ProcessImagePayload represents payload for process_image jobs
type ProcessImagePayload struct {
	ImageURL        string  `json:"image_url"`
	FlyerPageID     int     `json:"flyer_page_id"`
	ProcessingType  string  `json:"processing_type"` // "ocr", "vision_ai", "layout_analysis"
	Confidence      float64 `json:"confidence,omitempty"`
}

// ValidateDataPayload represents payload for validate_data jobs
type ValidateDataPayload struct {
	EntityType string `json:"entity_type"` // "flyer", "product", "store"
	EntityID   int    `json:"entity_id"`
	Strict     bool   `json:"strict,omitempty"`
}

// GetPayload unmarshals the payload JSON into the provided interface
func (ej *ExtractionJob) GetPayload(v interface{}) error {
	if ej.Payload == nil {
		return errors.New("payload is nil")
	}
	return json.Unmarshal(ej.Payload, v)
}

// SetPayload marshals the provided interface into the payload JSON
func (ej *ExtractionJob) SetPayload(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	ej.Payload = data
	return nil
}

// GetScrapeFlyerPayload gets the payload as ScrapeFlyerPayload
func (ej *ExtractionJob) GetScrapeFlyerPayload() (ScrapeFlyerPayload, error) {
	var payload ScrapeFlyerPayload
	err := ej.GetPayload(&payload)
	return payload, err
}

// GetExtractPagePayload gets the payload as ExtractPagePayload
func (ej *ExtractionJob) GetExtractPagePayload() (ExtractPagePayload, error) {
	var payload ExtractPagePayload
	err := ej.GetPayload(&payload)
	return payload, err
}

// GetMatchProductsPayload gets the payload as MatchProductsPayload
func (ej *ExtractionJob) GetMatchProductsPayload() (MatchProductsPayload, error) {
	var payload MatchProductsPayload
	err := ej.GetPayload(&payload)
	return payload, err
}

// GetProcessImagePayload gets the payload as ProcessImagePayload
func (ej *ExtractionJob) GetProcessImagePayload() (ProcessImagePayload, error) {
	var payload ProcessImagePayload
	err := ej.GetPayload(&payload)
	return payload, err
}

// GetValidateDataPayload gets the payload as ValidateDataPayload
func (ej *ExtractionJob) GetValidateDataPayload() (ValidateDataPayload, error) {
	var payload ValidateDataPayload
	err := ej.GetPayload(&payload)
	return payload, err
}

// IsPending checks if the job is pending
func (ej *ExtractionJob) IsPending() bool {
	return ej.Status == string(JobStatusPending)
}

// IsProcessing checks if the job is currently being processed
func (ej *ExtractionJob) IsProcessing() bool {
	return ej.Status == string(JobStatusProcessing)
}

// IsCompleted checks if the job has completed successfully
func (ej *ExtractionJob) IsCompleted() bool {
	return ej.Status == string(JobStatusCompleted)
}

// IsFailed checks if the job has failed permanently
func (ej *ExtractionJob) IsFailed() bool {
	return ej.Status == string(JobStatusFailed)
}

// IsExpired checks if the job has expired
func (ej *ExtractionJob) IsExpired() bool {
	return ej.Status == string(JobStatusExpired) ||
		   (ej.ExpiresAt != nil && ej.ExpiresAt.Before(time.Now()))
}

// CanBeProcessed checks if the job can be picked up for processing
func (ej *ExtractionJob) CanBeProcessed() bool {
	return ej.IsPending() &&
		   ej.Attempts < ej.MaxAttempts &&
		   ej.ScheduledFor.Before(time.Now().Add(time.Minute)) &&
		   !ej.IsExpired()
}

// CanBeRetried checks if the job can be retried
func (ej *ExtractionJob) CanBeRetried() bool {
	return (ej.IsFailed() || ej.Status == string(JobStatusProcessing)) &&
		   ej.Attempts < ej.MaxAttempts &&
		   !ej.IsExpired()
}

// GetProcessingDuration returns how long the job has been/was processing
func (ej *ExtractionJob) GetProcessingDuration() *time.Duration {
	if ej.StartedAt == nil {
		return nil
	}

	endTime := ej.CompletedAt
	if endTime == nil && ej.IsProcessing() {
		now := time.Now()
		endTime = &now
	}

	if endTime == nil {
		return nil
	}

	duration := endTime.Sub(*ej.StartedAt)
	return &duration
}

// StartProcessing marks the job as being processed
func (ej *ExtractionJob) StartProcessing(workerID string) {
	now := time.Now()
	ej.Status = string(JobStatusProcessing)
	ej.WorkerID = &workerID
	ej.StartedAt = &now
	ej.Attempts++
	ej.UpdatedAt = now
}

// CompleteProcessing marks the job as successfully completed
func (ej *ExtractionJob) CompleteProcessing() {
	now := time.Now()
	ej.Status = string(JobStatusCompleted)
	ej.CompletedAt = &now
	ej.UpdatedAt = now
}

// FailProcessing marks the job as failed
func (ej *ExtractionJob) FailProcessing(errorMsg string) {
	now := time.Now()
	ej.ErrorMessage = &errorMsg
	ej.ErrorCount++

	if ej.Attempts >= ej.MaxAttempts {
		ej.Status = string(JobStatusFailed)
		ej.CompletedAt = &now
	} else {
		// Reset for retry
		ej.Status = string(JobStatusPending)
		ej.WorkerID = nil
		ej.StartedAt = nil
		ej.ScheduledFor = now.Add(5 * time.Minute) // Retry in 5 minutes
	}

	ej.UpdatedAt = now
}

// Cancel marks the job as cancelled
func (ej *ExtractionJob) Cancel() {
	now := time.Now()
	ej.Status = string(JobStatusCancelled)
	ej.CompletedAt = &now
	ej.UpdatedAt = now
}

// MarkExpired marks the job as expired
func (ej *ExtractionJob) MarkExpired() {
	now := time.Now()
	ej.Status = string(JobStatusExpired)
	ej.CompletedAt = &now
	ej.UpdatedAt = now
}

// ResetForRetry resets the job for retry processing
func (ej *ExtractionJob) ResetForRetry(delayMinutes int) {
	now := time.Now()
	ej.Status = string(JobStatusPending)
	ej.WorkerID = nil
	ej.StartedAt = nil
	ej.ScheduledFor = now.Add(time.Duration(delayMinutes) * time.Minute)
	ej.UpdatedAt = now
}

// SetExpiration sets when the job should expire
func (ej *ExtractionJob) SetExpiration(duration time.Duration) {
	expiresAt := time.Now().Add(duration)
	ej.ExpiresAt = &expiresAt
}

// GetPriorityDescription returns a human-readable priority description
func (ej *ExtractionJob) GetPriorityDescription() string {
	switch ej.Priority {
	case 1:
		return "Critical"
	case 2:
		return "High"
	case 3:
		return "Medium-High"
	case 4:
		return "Medium"
	case 5:
		return "Normal"
	case 6:
		return "Medium-Low"
	case 7:
		return "Low"
	case 8:
		return "Very Low"
	default:
		return "Unknown"
	}
}

// GetJobTypeDescription returns a human-readable job type description
func (ej *ExtractionJob) GetJobTypeDescription() string {
	switch ExtractionJobType(ej.JobType) {
	case JobTypeScrapeFlyer:
		return "Scrape Flyer"
	case JobTypeExtractPage:
		return "Extract Page"
	case JobTypeMatchProducts:
		return "Match Products"
	case JobTypeGenerateFlyer:
		return "Generate Flyer"
	case JobTypeProcessImage:
		return "Process Image"
	case JobTypeValidateData:
		return "Validate Data"
	default:
		return "Unknown Job Type"
	}
}

// TableName returns the table name for Bun
func (ej *ExtractionJob) TableName() string {
	return "extraction_jobs"
}