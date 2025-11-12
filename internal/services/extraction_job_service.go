package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/extractionjob"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
)

type extractionJobService struct {
	repo   extractionjob.Repository
	logger *slog.Logger
	now    func() time.Time
}

// NewExtractionJobService creates a new extraction job service instance using the registered repository factory.
func NewExtractionJobService(db *bun.DB) ExtractionJobService {
	return NewExtractionJobServiceWithRepository(newExtractionJobRepository(db))
}

// NewExtractionJobServiceWithRepository allows injecting a custom repository (useful for tests).
func NewExtractionJobServiceWithRepository(repo extractionjob.Repository) ExtractionJobService {
	return &extractionJobService{
		repo:   repo,
		logger: slog.Default().With("service", "extraction_job"),
		now:    time.Now,
	}
}

// Basic CRUD operations
func (s *extractionJobService) GetByID(ctx context.Context, id int64) (*models.ExtractionJob, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get extraction job %d: %w", id, err)
	}
	return job, nil
}

func (s *extractionJobService) GetAll(ctx context.Context, filters ExtractionJobFilters) ([]*models.ExtractionJob, error) {
	filterCopy := filters
	jobs, err := s.repo.GetAll(ctx, &filterCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to list extraction jobs: %w", err)
	}
	return jobs, nil
}

func (s *extractionJobService) Create(ctx context.Context, job *models.ExtractionJob) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}
	if job.JobType == "" {
		return fmt.Errorf("job type is required")
	}
	now := s.now()
	job.Priority = s.normalizePriority(job.Priority)
	if job.MaxAttempts == 0 {
		job.MaxAttempts = 3
	}
	if job.Status == "" {
		job.Status = string(models.JobStatusPending)
	}
	if job.ScheduledFor.IsZero() {
		job.ScheduledFor = now
	}
	job.CreatedAt = now
	job.UpdatedAt = now
	if err := s.repo.Create(ctx, job); err != nil {
		return fmt.Errorf("failed to create extraction job: %w", err)
	}
	s.logger.Info("extraction job created",
		slog.Int64("job_id", job.ID),
		slog.String("job_type", job.JobType),
		slog.Int("priority", job.Priority),
	)
	return nil
}

func (s *extractionJobService) Update(ctx context.Context, job *models.ExtractionJob) error {
	if job == nil {
		return fmt.Errorf("job cannot be nil")
	}
	job.UpdatedAt = s.now()
	if err := s.repo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update extraction job %d: %w", job.ID, err)
	}
	return nil
}

func (s *extractionJobService) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete extraction job %d: %w", id, err)
	}
	return nil
}

// Job queue operations
func (s *extractionJobService) GetNextJob(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error) {
	job, err := s.repo.GetNextJob(ctx, jobTypes, workerID)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve extraction job: %w", err)
	}
	return job, nil
}

func (s *extractionJobService) GetPendingJobs(ctx context.Context, jobTypes []string) ([]*models.ExtractionJob, error) {
	jobs, err := s.repo.GetPendingJobs(ctx, jobTypes, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending extraction jobs: %w", err)
	}
	return jobs, nil
}

func (s *extractionJobService) GetProcessingJobs(ctx context.Context, workerID string) ([]*models.ExtractionJob, error) {
	jobs, err := s.repo.GetProcessingJobs(ctx, workerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list processing extraction jobs: %w", err)
	}
	return jobs, nil
}

// Job lifecycle operations
func (s *extractionJobService) StartProcessing(ctx context.Context, jobID int64, workerID string) error {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get extraction job %d: %w", jobID, err)
	}
	now := s.now()
	job.Status = string(models.JobStatusProcessing)
	job.StartedAt = &now
	job.UpdatedAt = now
	job.WorkerID = stringPtr(workerID)
	job.Attempts++

	if err := s.repo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to mark extraction job %d as processing: %w", jobID, err)
	}
	return nil
}

func (s *extractionJobService) CompleteJob(ctx context.Context, jobID int64) error {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get extraction job %d: %w", jobID, err)
	}
	now := s.now()
	job.Status = string(models.JobStatusCompleted)
	job.CompletedAt = &now
	job.UpdatedAt = now
	if err := s.repo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to complete extraction job %d: %w", jobID, err)
	}
	return nil
}

func (s *extractionJobService) FailJob(ctx context.Context, jobID int64, errorMsg string) error {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get extraction job %d: %w", jobID, err)
	}
	now := s.now()
	job.Status = string(models.JobStatusFailed)
	job.CompletedAt = &now
	job.UpdatedAt = now
	job.ErrorMessage = stringPtr(errorMsg)
	job.ErrorCount++

	if err := s.repo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to mark extraction job %d as failed: %w", jobID, err)
	}
	return nil
}

func (s *extractionJobService) CancelJob(ctx context.Context, jobID int64) error {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get extraction job %d: %w", jobID, err)
	}
	now := s.now()
	job.Status = string(models.JobStatusCancelled)
	job.CompletedAt = &now
	job.UpdatedAt = now
	job.WorkerID = nil
	if err := s.repo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to cancel extraction job %d: %w", jobID, err)
	}
	return nil
}

func (s *extractionJobService) RetryJob(ctx context.Context, jobID int64, delayMinutes int) error {
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get extraction job %d: %w", jobID, err)
	}
	now := s.now()
	job.Status = string(models.JobStatusPending)
	job.WorkerID = nil
	job.StartedAt = nil
	job.CompletedAt = nil
	job.ErrorMessage = nil
	job.UpdatedAt = now
	job.ScheduledFor = now.Add(time.Duration(delayMinutes) * time.Minute)

	if err := s.repo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to reschedule extraction job %d: %w", jobID, err)
	}
	return nil
}

// Job creation helpers
func (s *extractionJobService) CreateScrapeFlyerJob(ctx context.Context, storeID int, priority int) error {
	payload := models.ScrapeFlyerPayload{
		StoreID: storeID,
	}
	return s.createJobWithPayload(ctx, models.JobTypeScrapeFlyer, priority, payload)
}

func (s *extractionJobService) CreateExtractPageJob(ctx context.Context, flyerPageID int, priority int) error {
	payload := models.ExtractPagePayload{
		FlyerPageID: flyerPageID,
	}
	return s.createJobWithPayload(ctx, models.JobTypeExtractPage, priority, payload)
}

func (s *extractionJobService) CreateMatchProductsJob(ctx context.Context, flyerID int, priority int) error {
	payload := models.MatchProductsPayload{
		FlyerID: flyerID,
	}
	return s.createJobWithPayload(ctx, models.JobTypeMatchProducts, priority, payload)
}

// Cleanup operations
func (s *extractionJobService) CleanupExpiredJobs(ctx context.Context) (int64, error) {
	count, err := s.repo.DeleteExpiredJobs(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired extraction jobs: %w", err)
	}
	return count, nil
}

func (s *extractionJobService) CleanupCompletedJobs(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := s.now().Add(-olderThan)
	completed, err := s.repo.DeleteCompletedJobs(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup completed jobs: %w", err)
	}
	failed, err := s.repo.DeleteFailedJobs(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup failed jobs: %w", err)
	}
	return completed + failed, nil
}

// Helpers
func (s *extractionJobService) createJobWithPayload(ctx context.Context, jobType models.ExtractionJobType, priority int, payload interface{}) error {
	job := &models.ExtractionJob{
		JobType:     string(jobType),
		Priority:    s.normalizePriority(priority),
		MaxAttempts: 3,
	}
	if err := job.SetPayload(payload); err != nil {
		return fmt.Errorf("failed to encode extraction job payload: %w", err)
	}
	return s.Create(ctx, job)
}

func (s *extractionJobService) normalizePriority(priority int) int {
	switch {
	case priority <= 0:
		return 5
	case priority > 10:
		return 10
	default:
		return priority
	}
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	v := value
	return &v
}
