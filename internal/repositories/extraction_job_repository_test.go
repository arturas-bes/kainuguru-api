package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/extractionjob"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

func TestExtractionJobRepository_GetNextJobReservesHighestPriority(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupExtractionJobTestDB(t)
	defer cleanup()

	payload := json.RawMessage(`{"foo":"bar"}`)
	now := time.Unix(0, 0)

	low := insertTestJob(t, db, &models.ExtractionJob{
		JobType:      string(models.JobTypeExtractPage),
		Status:       string(models.JobStatusPending),
		Priority:     3,
		MaxAttempts:  3,
		Payload:      payload,
		ScheduledFor: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	_ = low
	high := insertTestJob(t, db, &models.ExtractionJob{
		JobType:      string(models.JobTypeScrapeFlyer),
		Status:       string(models.JobStatusPending),
		Priority:     9,
		MaxAttempts:  3,
		Payload:      payload,
		ScheduledFor: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	})

	job, err := repo.GetNextJob(ctx, nil, "worker-1")
	if err != nil {
		t.Fatalf("GetNextJob returned error: %v", err)
	}
	if job == nil || job.ID != high.ID {
		t.Fatalf("expected job %d, got %+v", high.ID, job)
	}
	if job.Status != string(models.JobStatusProcessing) {
		t.Fatalf("expected processing status, got %s", job.Status)
	}
	if job.WorkerID == nil || *job.WorkerID != "worker-1" {
		t.Fatalf("worker id not set correctly")
	}

	reloaded, err := repo.GetByID(ctx, high.ID)
	if err != nil {
		t.Fatalf("failed to reload job: %v", err)
	}
	if reloaded.Status != string(models.JobStatusProcessing) {
		t.Fatalf("status not persisted, got %s", reloaded.Status)
	}
}

func TestExtractionJobRepository_CleanupOperations(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupExtractionJobTestDB(t)
	defer cleanup()

	payload := json.RawMessage(`{}`)
	now := time.Now()
	old := now.Add(-48 * time.Hour)

	insertTestJob(t, db, &models.ExtractionJob{
		JobType:      string(models.JobTypeMatchProducts),
		Status:       string(models.JobStatusCompleted),
		Payload:      payload,
		MaxAttempts:  3,
		Priority:     5,
		CompletedAt:  &old,
		ScheduledFor: old,
		CreatedAt:    old,
		UpdatedAt:    old,
	})

	insertTestJob(t, db, &models.ExtractionJob{
		JobType:      string(models.JobTypeExtractPage),
		Status:       string(models.JobStatusFailed),
		Payload:      payload,
		MaxAttempts:  3,
		Priority:     5,
		CompletedAt:  &old,
		ScheduledFor: old,
		CreatedAt:    old,
		UpdatedAt:    old,
	})

	insertTestJob(t, db, &models.ExtractionJob{
		JobType:      string(models.JobTypeScrapeFlyer),
		Status:       string(models.JobStatusExpired),
		Payload:      payload,
		MaxAttempts:  3,
		Priority:     5,
		ExpiresAt:    &old,
		ScheduledFor: old,
		CreatedAt:    old,
		UpdatedAt:    old,
	})

	completed, err := repo.DeleteCompletedJobs(ctx, now.Add(-24*time.Hour))
	if err != nil || completed != 1 {
		t.Fatalf("DeleteCompletedJobs = %d, %v", completed, err)
	}

	failed, err := repo.DeleteFailedJobs(ctx, now.Add(-24*time.Hour))
	if err != nil || failed != 1 {
		t.Fatalf("DeleteFailedJobs = %d, %v", failed, err)
	}

	expired, err := repo.DeleteExpiredJobs(ctx)
	if err != nil || expired != 1 {
		t.Fatalf("DeleteExpiredJobs = %d, %v", expired, err)
	}
}

func TestExtractionJobRepository_GetAllOrdersAndFilters(t *testing.T) {
	ctx := context.Background()
	db, repo, cleanup := setupExtractionJobTestDB(t)
	defer cleanup()

	now := time.Unix(0, 0)
	insertTestJob(t, db, &models.ExtractionJob{
		JobType:      string(models.JobTypeExtractPage),
		Status:       string(models.JobStatusPending),
		Priority:     3,
		ScheduledFor: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	second := insertTestJob(t, db, &models.ExtractionJob{
		JobType:      string(models.JobTypeScrapeFlyer),
		Status:       string(models.JobStatusPending),
		Priority:     9,
		ScheduledFor: now,
		CreatedAt:    now.Add(time.Minute),
		UpdatedAt:    now.Add(time.Minute),
	})

	jobs, err := repo.GetAll(ctx, nil)
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(jobs) != 2 || jobs[0].ID != second.ID {
		t.Fatalf("expected default ordering by priority desc, got %+v", jobs)
	}

	filters := &extractionjob.Filters{
		JobTypes: []string{string(models.JobTypeScrapeFlyer)},
		OrderBy:  "created_at",
		OrderDir: "ASC",
		Limit:    1,
	}
	filtered, err := repo.GetAll(ctx, filters)
	if err != nil {
		t.Fatalf("filtered GetAll returned error: %v", err)
	}
	if len(filtered) != 1 || filtered[0].JobType != string(models.JobTypeScrapeFlyer) {
		t.Fatalf("expected filtered slice to include only scrape flyer job, got %+v", filtered)
	}
}

func setupExtractionJobTestDB(t *testing.T) (*bun.DB, extractionjob.Repository, func()) {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.DriverName(), "file:extraction_jobs_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	db := bun.NewDB(sqldb, sqlitedialect.New())

	ctx := context.Background()
	schema := `
CREATE TABLE extraction_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 5,
    payload TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    worker_id TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    error_message TEXT,
    error_count INTEGER NOT NULL DEFAULT 0,
    scheduled_for DATETIME NOT NULL,
    expires_at DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	repo := NewExtractionJobRepository(db)
	cleanup := func() {
		_ = db.Close()
	}
	return db, repo, cleanup
}

func insertTestJob(t *testing.T, db *bun.DB, job *models.ExtractionJob) *models.ExtractionJob {
	t.Helper()
	if job.Payload == nil {
		job.Payload = json.RawMessage(`{}`)
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = job.CreatedAt
	}
	if job.ScheduledFor.IsZero() {
		job.ScheduledFor = job.CreatedAt
	}

	if _, err := db.NewInsert().Model(job).Exec(context.Background()); err != nil {
		t.Fatalf("failed to insert job: %v", err)
	}
	return job
}
