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

func TestExtractionJobService_GetByIDDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &mockExtractionJobRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ExtractionJob, error) {
			called = true
			if id != 42 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return &models.ExtractionJob{ID: 42, JobType: "test"}, nil
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	result, err := service.GetByID(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 42 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestExtractionJobService_GetByIDPropagatesError(t *testing.T) {
	ctx := context.Background()
	want := errors.New("not found")
	repo := &mockExtractionJobRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ExtractionJob, error) {
			return nil, want
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	_, err := service.GetByID(ctx, 999)
	if !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

func TestExtractionJobService_GetAllDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &mockExtractionJobRepository{
		getAllFunc: func(ctx context.Context, filters *extractionjob.Filters) ([]*models.ExtractionJob, error) {
			called = true
			if filters == nil || filters.Limit != 10 {
				t.Fatalf("filters not forwarded")
			}
			return []*models.ExtractionJob{{ID: 1}, {ID: 2}}, nil
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	result, err := service.GetAll(ctx, ExtractionJobFilters{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestExtractionJobService_CreateSetsDefaults(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Unix(1000, 0)
	repo := &mockExtractionJobRepository{
		createFunc: func(ctx context.Context, job *models.ExtractionJob) error {
			if job.Priority != 5 {
				t.Fatalf("expected default priority 5, got %d", job.Priority)
			}
			if job.MaxAttempts != 3 {
				t.Fatalf("expected default max attempts 3, got %d", job.MaxAttempts)
			}
			if job.Status != string(models.JobStatusPending) {
				t.Fatalf("expected pending status, got %s", job.Status)
			}
			if job.ScheduledFor != fixedTime {
				t.Fatalf("expected scheduled for to be set")
			}
			if job.CreatedAt != fixedTime || job.UpdatedAt != fixedTime {
				t.Fatalf("timestamps not set correctly")
			}
			return nil
		},
	}
	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return fixedTime },
	}

	job := &models.ExtractionJob{JobType: "test", Priority: 0}
	if err := service.Create(ctx, job); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractionJobService_CreateRejectsNilJob(t *testing.T) {
	service := &extractionJobService{repo: &mockExtractionJobRepository{}, logger: slog.Default(), now: time.Now}
	err := service.Create(context.Background(), nil)
	if err == nil || err.Error() != "job cannot be nil" {
		t.Fatalf("expected nil job error, got %v", err)
	}
}

func TestExtractionJobService_CreateRejectsEmptyJobType(t *testing.T) {
	service := &extractionJobService{repo: &mockExtractionJobRepository{}, logger: slog.Default(), now: time.Now}
	err := service.Create(context.Background(), &models.ExtractionJob{})
	if err == nil || err.Error() != "job type is required" {
		t.Fatalf("expected job type error, got %v", err)
	}
}

func TestExtractionJobService_UpdateSetsTimestamp(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Unix(2000, 0)
	repo := &mockExtractionJobRepository{
		updateFunc: func(ctx context.Context, job *models.ExtractionJob) error {
			if job.UpdatedAt != fixedTime {
				t.Fatalf("UpdatedAt not set correctly")
			}
			return nil
		},
	}
	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return fixedTime },
	}

	job := &models.ExtractionJob{ID: 10, JobType: "test"}
	if err := service.Update(ctx, job); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractionJobService_DeleteDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &mockExtractionJobRepository{
		deleteFunc: func(ctx context.Context, id int64) error {
			called = true
			if id != 99 {
				t.Fatalf("unexpected ID: %d", id)
			}
			return nil
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	if err := service.Delete(ctx, 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected delegation to repository")
	}
}

func TestExtractionJobService_GetNextJobDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &mockExtractionJobRepository{
		getNextFunc: func(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error) {
			called = true
			if len(jobTypes) != 2 || jobTypes[0] != "typeA" || workerID != "worker1" {
				t.Fatalf("unexpected args: %v %s", jobTypes, workerID)
			}
			return &models.ExtractionJob{ID: 50}, nil
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	result, err := service.GetNextJob(ctx, []string{"typeA", "typeB"}, "worker1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || result.ID != 50 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestExtractionJobService_GetPendingJobsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &mockExtractionJobRepository{
		getPendingFunc: func(ctx context.Context, jobTypes []string, limit int) ([]*models.ExtractionJob, error) {
			called = true
			if len(jobTypes) != 1 || limit != 0 {
				t.Fatalf("unexpected args: %v %d", jobTypes, limit)
			}
			return []*models.ExtractionJob{{ID: 1}, {ID: 2}}, nil
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	result, err := service.GetPendingJobs(ctx, []string{"typeX"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 2 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestExtractionJobService_GetProcessingJobsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &mockExtractionJobRepository{
		getProcessingFunc: func(ctx context.Context, workerID string) ([]*models.ExtractionJob, error) {
			called = true
			if workerID != "worker2" {
				t.Fatalf("unexpected workerID: %s", workerID)
			}
			return []*models.ExtractionJob{{ID: 3}}, nil
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	result, err := service.GetProcessingJobs(ctx, "worker2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || len(result) != 1 {
		t.Fatalf("expected delegation to repository")
	}
}

func TestExtractionJobService_CompleteJobSetsStatus(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Unix(3000, 0)
	job := &models.ExtractionJob{ID: 20, Status: string(models.JobStatusProcessing)}
	repo := &mockExtractionJobRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ExtractionJob, error) {
			return job, nil
		},
		updateFunc: func(ctx context.Context, j *models.ExtractionJob) error {
			if j.Status != string(models.JobStatusCompleted) {
				t.Fatalf("expected completed status, got %s", j.Status)
			}
			if j.CompletedAt == nil || *j.CompletedAt != fixedTime {
				t.Fatalf("CompletedAt not set correctly")
			}
			return nil
		},
	}
	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return fixedTime },
	}

	if err := service.CompleteJob(ctx, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractionJobService_FailJobSetsStatusAndError(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Unix(4000, 0)
	job := &models.ExtractionJob{ID: 25, Status: string(models.JobStatusProcessing), ErrorCount: 1}
	repo := &mockExtractionJobRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ExtractionJob, error) {
			return job, nil
		},
		updateFunc: func(ctx context.Context, j *models.ExtractionJob) error {
			if j.Status != string(models.JobStatusFailed) {
				t.Fatalf("expected failed status, got %s", j.Status)
			}
			if j.ErrorMessage == nil || *j.ErrorMessage != "test error" {
				t.Fatalf("ErrorMessage not set correctly")
			}
			if j.ErrorCount != 2 {
				t.Fatalf("ErrorCount not incremented, got %d", j.ErrorCount)
			}
			if j.CompletedAt == nil || *j.CompletedAt != fixedTime {
				t.Fatalf("CompletedAt not set correctly")
			}
			return nil
		},
	}
	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return fixedTime },
	}

	if err := service.FailJob(ctx, 25, "test error"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractionJobService_CancelJobSetsStatusAndClearsWorker(t *testing.T) {
	ctx := context.Background()
	fixedTime := time.Unix(5000, 0)
	workerID := "worker-cancel"
	job := &models.ExtractionJob{ID: 30, Status: string(models.JobStatusProcessing), WorkerID: &workerID}
	repo := &mockExtractionJobRepository{
		getByIDFunc: func(ctx context.Context, id int64) (*models.ExtractionJob, error) {
			return job, nil
		},
		updateFunc: func(ctx context.Context, j *models.ExtractionJob) error {
			if j.Status != string(models.JobStatusCancelled) {
				t.Fatalf("expected cancelled status, got %s", j.Status)
			}
			if j.WorkerID != nil {
				t.Fatalf("WorkerID should be cleared")
			}
			if j.CompletedAt == nil || *j.CompletedAt != fixedTime {
				t.Fatalf("CompletedAt not set correctly")
			}
			return nil
		},
	}
	service := &extractionJobService{
		repo:   repo,
		logger: slog.Default(),
		now:    func() time.Time { return fixedTime },
	}

	if err := service.CancelJob(ctx, 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractionJobService_CreateExtractPageJobSetsPayload(t *testing.T) {
	ctx := context.Background()
	repo := &mockExtractionJobRepository{
		createFunc: func(ctx context.Context, job *models.ExtractionJob) error {
			if job.JobType != string(models.JobTypeExtractPage) {
				t.Fatalf("expected extract_page job, got %s", job.JobType)
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

	if err := service.CreateExtractPageJob(ctx, 15, 8); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractionJobService_CreateMatchProductsJobSetsPayload(t *testing.T) {
	ctx := context.Background()
	repo := &mockExtractionJobRepository{
		createFunc: func(ctx context.Context, job *models.ExtractionJob) error {
			if job.JobType != string(models.JobTypeMatchProducts) {
				t.Fatalf("expected match_products job, got %s", job.JobType)
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

	if err := service.CreateMatchProductsJob(ctx, 25, 6); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractionJobService_CleanupExpiredJobsDelegates(t *testing.T) {
	ctx := context.Background()
	called := false
	repo := &mockExtractionJobRepository{
		deleteExpiredFunc: func(ctx context.Context) (int64, error) {
			called = true
			return 10, nil
		},
	}
	service := &extractionJobService{repo: repo, logger: slog.Default(), now: time.Now}

	count, err := service.CleanupExpiredJobs(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called || count != 10 {
		t.Fatalf("expected delegation to repository, got count %d", count)
	}
}

type mockExtractionJobRepository struct{
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
