package services

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/extractionjob"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

func TestExtractionJobService_CreateScrapeFlyerJob(t *testing.T) {
	ctx := context.Background()
	repo := &mockExtractionJobRepository{
		createFunc: func(ctx context.Context, job *models.ExtractionJob) error {
			if job.JobType != string(models.JobTypeScrapeFlyer) {
				t.Fatalf("expected scrape job, got %s", job.JobType)
			}
			if job.Priority != 7 {
				t.Fatalf("expected priority 7, got %d", job.Priority)
			}
			if job.Payload == nil {
				t.Fatalf("payload must be set")
			}
			return nil
		},
	}

	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return time.Unix(0, 0) },
	}

	if err := service.CreateScrapeFlyerJob(ctx, 5, 7); err != nil {
		t.Fatalf("CreateScrapeFlyerJob returned error: %v", err)
	}
}

func TestExtractionJobService_StartProcessing(t *testing.T) {
	ctx := context.Background()
	job := &models.ExtractionJob{
		ID:       100,
		Status:   string(models.JobStatusPending),
		Attempts: 1,
	}
	var updated *models.ExtractionJob

	repo := &mockExtractionJobRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ExtractionJob, error) {
			return job, nil
		},
		updateFunc: func(ctx context.Context, j *models.ExtractionJob) error {
			updated = j
			return nil
		},
	}

	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return time.Unix(100, 0) },
	}

	if err := service.StartProcessing(ctx, 100, "worker-A"); err != nil {
		t.Fatalf("StartProcessing returned error: %v", err)
	}
	if updated.Status != string(models.JobStatusProcessing) {
		t.Fatalf("expected processing status, got %s", updated.Status)
	}
	if updated.WorkerID == nil || *updated.WorkerID != "worker-A" {
		t.Fatalf("worker not set correctly")
	}
	if updated.Attempts != 2 {
		t.Fatalf("attempts not incremented, got %d", updated.Attempts)
	}
	if updated.StartedAt == nil {
		t.Fatalf("started_at not set")
	}
}

func TestExtractionJobService_RetryJob(t *testing.T) {
	ctx := context.Background()
	start := time.Unix(200, 0)
	job := &models.ExtractionJob{
		ID:           5,
		Status:       string(models.JobStatusFailed),
		WorkerID:     stringPtr("worker-x"),
		StartedAt:    &start,
		CompletedAt:  &start,
		ErrorMessage: stringPtr("bad"),
	}
	var scheduled time.Time

	repo := &mockExtractionJobRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ExtractionJob, error) {
			return job, nil
		},
		updateFunc: func(ctx context.Context, j *models.ExtractionJob) error {
			if j.Status != string(models.JobStatusPending) {
				t.Fatalf("expected pending status, got %s", j.Status)
			}
			if j.WorkerID != nil || j.StartedAt != nil || j.CompletedAt != nil || j.ErrorMessage != nil {
				t.Fatalf("job should be reset")
			}
			scheduled = j.ScheduledFor
			return nil
		},
	}

	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return time.Unix(300, 0) },
	}

	if err := service.RetryJob(ctx, 5, 15); err != nil {
		t.Fatalf("RetryJob returned error: %v", err)
	}
	expected := time.Unix(300, 0).Add(15 * time.Minute)
	if !scheduled.Equal(expected) {
		t.Fatalf("expected scheduled_for %v, got %v", expected, scheduled)
	}
}

func TestExtractionJobService_CleanupCompletedJobs(t *testing.T) {
	repo := &mockExtractionJobRepository{
		deleteCompletedFunc: func(ctx context.Context, olderThan time.Time) (int64, error) {
			return 3, nil
		},
		deleteFailedFunc: func(ctx context.Context, olderThan time.Time) (int64, error) {
			return 2, nil
		},
	}

	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return time.Unix(500, 0) },
	}

	deleted, err := service.CleanupCompletedJobs(context.Background(), time.Hour)
	if err != nil {
		t.Fatalf("CleanupCompletedJobs returned error: %v", err)
	}
	if deleted != 5 {
		t.Fatalf("expected 5 deletions, got %d", deleted)
	}
}

type mockExtractionJobRepository struct {
	getByIDFunc         func(ctx context.Context, id int64) (*models.ExtractionJob, error)
	getAllFunc          func(ctx context.Context, filters *extractionjob.Filters) ([]*models.ExtractionJob, error)
	createFunc          func(ctx context.Context, job *models.ExtractionJob) error
	updateFunc          func(ctx context.Context, job *models.ExtractionJob) error
	deleteFunc          func(ctx context.Context, id int64) error
	getNextFunc         func(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error)
	getPendingFunc      func(ctx context.Context, jobTypes []string, limit int) ([]*models.ExtractionJob, error)
	getProcessingFunc   func(ctx context.Context, workerID string) ([]*models.ExtractionJob, error)
	deleteCompletedFunc func(ctx context.Context, olderThan time.Time) (int64, error)
	deleteFailedFunc    func(ctx context.Context, olderThan time.Time) (int64, error)
	deleteExpiredFunc   func(ctx context.Context) (int64, error)
}

func (m *mockExtractionJobRepository) GetByID(ctx context.Context, id int64) (*models.ExtractionJob, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExtractionJobRepository) GetAll(ctx context.Context, filters *extractionjob.Filters) ([]*models.ExtractionJob, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc(ctx, filters)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExtractionJobRepository) Create(ctx context.Context, job *models.ExtractionJob) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, job)
	}
	return errors.New("not implemented")
}

func (m *mockExtractionJobRepository) Update(ctx context.Context, job *models.ExtractionJob) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, job)
	}
	return errors.New("not implemented")
}

func (m *mockExtractionJobRepository) Delete(ctx context.Context, id int64) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockExtractionJobRepository) GetNextJob(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error) {
	if m.getNextFunc != nil {
		return m.getNextFunc(ctx, jobTypes, workerID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExtractionJobRepository) GetPendingJobs(ctx context.Context, jobTypes []string, limit int) ([]*models.ExtractionJob, error) {
	if m.getPendingFunc != nil {
		return m.getPendingFunc(ctx, jobTypes, limit)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExtractionJobRepository) GetProcessingJobs(ctx context.Context, workerID string) ([]*models.ExtractionJob, error) {
	if m.getProcessingFunc != nil {
		return m.getProcessingFunc(ctx, workerID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockExtractionJobRepository) DeleteCompletedJobs(ctx context.Context, olderThan time.Time) (int64, error) {
	if m.deleteCompletedFunc != nil {
		return m.deleteCompletedFunc(ctx, olderThan)
	}
	return 0, errors.New("not implemented")
}

func (m *mockExtractionJobRepository) DeleteFailedJobs(ctx context.Context, olderThan time.Time) (int64, error) {
	if m.deleteFailedFunc != nil {
		return m.deleteFailedFunc(ctx, olderThan)
	}
	return 0, errors.New("not implemented")
}

func (m *mockExtractionJobRepository) DeleteExpiredJobs(ctx context.Context) (int64, error) {
	if m.deleteExpiredFunc != nil {
		return m.deleteExpiredFunc(ctx)
	}
	return 0, errors.New("not implemented")
}
