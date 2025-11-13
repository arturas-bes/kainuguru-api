package extractionjob

import (
	"context"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Repository describes persistence operations for extraction jobs.
type Repository interface {
	// Basic CRUD operations
	GetByID(ctx context.Context, id int64) (*models.ExtractionJob, error)
	GetAll(ctx context.Context, filters *Filters) ([]*models.ExtractionJob, error)
	Create(ctx context.Context, job *models.ExtractionJob) error
	Update(ctx context.Context, job *models.ExtractionJob) error
	Delete(ctx context.Context, id int64) error

	// Job queue operations
	GetNextJob(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error)
	GetPendingJobs(ctx context.Context, jobTypes []string, limit int) ([]*models.ExtractionJob, error)
	GetProcessingJobs(ctx context.Context, workerID string) ([]*models.ExtractionJob, error)

	// Cleanup helpers
	DeleteCompletedJobs(ctx context.Context, olderThan time.Time) (int64, error)
	DeleteFailedJobs(ctx context.Context, olderThan time.Time) (int64, error)
	DeleteExpiredJobs(ctx context.Context) (int64, error)
}
