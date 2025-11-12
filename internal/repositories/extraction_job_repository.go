package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/extractionjob"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type extractionJobRepository struct {
	db                 *bun.DB
	supportsSkipLocked bool
}

// NewExtractionJobRepository returns a Bun-backed extraction job repository.
func NewExtractionJobRepository(db *bun.DB) extractionjob.Repository {
	return &extractionJobRepository{
		db:                 db,
		supportsSkipLocked: db.Dialect().Name() == dialect.PG,
	}
}

func (r *extractionJobRepository) GetByID(ctx context.Context, id int64) (*models.ExtractionJob, error) {
	job := new(models.ExtractionJob)
	if err := r.db.NewSelect().
		Model(job).
		Where("ej.id = ?", id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return job, nil
}

func (r *extractionJobRepository) GetAll(ctx context.Context, filters *extractionjob.Filters) ([]*models.ExtractionJob, error) {
	var jobs []*models.ExtractionJob
	query := r.db.NewSelect().Model(&jobs)
	query = applyExtractionJobFilters(query, filters)

	if filters != nil {
		orderBy := filters.OrderBy
		if orderBy == "" {
			orderBy = "priority"
		}
		orderDir := filters.OrderDir
		if orderDir == "" {
			orderDir = "DESC"
		}
		query = query.Order(fmt.Sprintf("ej.%s %s", orderBy, orderDir))

		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	} else {
		query = query.Order("ej.priority DESC").Order("ej.created_at ASC")
	}

	if err := query.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.ExtractionJob{}, nil
		}
		return nil, err
	}
	return jobs, nil
}

func (r *extractionJobRepository) Create(ctx context.Context, job *models.ExtractionJob) error {
	_, err := r.db.NewInsert().
		Model(job).
		Exec(ctx)
	return err
}

func (r *extractionJobRepository) Update(ctx context.Context, job *models.ExtractionJob) error {
	_, err := r.db.NewUpdate().
		Model(job).
		WherePK().
		Exec(ctx)
	return err
}

func (r *extractionJobRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*models.ExtractionJob)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *extractionJobRepository) GetNextJob(ctx context.Context, jobTypes []string, workerID string) (*models.ExtractionJob, error) {
	var reserved *models.ExtractionJob

	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		job := new(models.ExtractionJob)
		query := tx.NewSelect().
			Model(job).
			Where("ej.status = ?", models.JobStatusPending).
			Where("ej.scheduled_for <= ?", time.Now()).
			Where("ej.attempts < ej.max_attempts").
			Order("ej.priority DESC").
			Order("ej.created_at ASC").
			Limit(1)

		if len(jobTypes) > 0 {
			query = query.Where("ej.job_type IN (?)", bun.In(jobTypes))
		}
		if r.supportsSkipLocked {
			query = query.For("UPDATE SKIP LOCKED")
		}

		if err := query.Scan(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}

		now := time.Now()
		job.Status = string(models.JobStatusProcessing)
		job.StartedAt = &now
		job.UpdatedAt = now
		job.Attempts++

		if workerID != "" {
			job.WorkerID = stringPtr(workerID)
		} else {
			generated := fmt.Sprintf("worker_%d", now.UnixNano())
			job.WorkerID = &generated
		}

		if _, err := tx.NewUpdate().
			Model(job).
			Column("status", "worker_id", "started_at", "attempts", "updated_at").
			Where("id = ?", job.ID).
			Exec(ctx); err != nil {
			return err
		}

		reserved = job
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reserved, nil
}

func (r *extractionJobRepository) GetPendingJobs(ctx context.Context, jobTypes []string, limit int) ([]*models.ExtractionJob, error) {
	var jobs []*models.ExtractionJob
	query := r.db.NewSelect().
		Model(&jobs).
		Where("ej.status = ?", models.JobStatusPending).
		Where("ej.scheduled_for <= ?", time.Now()).
		Order("ej.priority DESC").
		Order("ej.created_at ASC")

	if len(jobTypes) > 0 {
		query = query.Where("ej.job_type IN (?)", bun.In(jobTypes))
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.ExtractionJob{}, nil
		}
		return nil, err
	}
	return jobs, nil
}

func (r *extractionJobRepository) GetProcessingJobs(ctx context.Context, workerID string) ([]*models.ExtractionJob, error) {
	var jobs []*models.ExtractionJob
	query := r.db.NewSelect().
		Model(&jobs).
		Where("ej.status = ?", models.JobStatusProcessing)

	if workerID != "" {
		query = query.Where("ej.worker_id = ?", workerID)
	}

	if err := query.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*models.ExtractionJob{}, nil
		}
		return nil, err
	}
	return jobs, nil
}

func (r *extractionJobRepository) DeleteCompletedJobs(ctx context.Context, olderThan time.Time) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*models.ExtractionJob)(nil)).
		Where("status = ?", models.JobStatusCompleted).
		Where("completed_at IS NOT NULL AND completed_at < ?", olderThan).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *extractionJobRepository) DeleteFailedJobs(ctx context.Context, olderThan time.Time) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*models.ExtractionJob)(nil)).
		Where("status = ?", models.JobStatusFailed).
		Where("completed_at IS NOT NULL AND completed_at < ?", olderThan).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *extractionJobRepository) DeleteExpiredJobs(ctx context.Context) (int64, error) {
	now := time.Now()
	res, err := r.db.NewDelete().
		Model((*models.ExtractionJob)(nil)).
		Where("(status = ? OR (expires_at IS NOT NULL AND expires_at <= ?))", models.JobStatusExpired, now).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func applyExtractionJobFilters(query *bun.SelectQuery, filters *extractionjob.Filters) *bun.SelectQuery {
	if filters == nil {
		return query
	}

	if len(filters.JobTypes) > 0 {
		query = query.Where("ej.job_type IN (?)", bun.In(filters.JobTypes))
	}
	if len(filters.Status) > 0 {
		query = query.Where("ej.status IN (?)", bun.In(filters.Status))
	}
	if len(filters.WorkerIDs) > 0 {
		query = query.Where("ej.worker_id IN (?)", bun.In(filters.WorkerIDs))
	}
	if filters.Priority != nil {
		query = query.Where("ej.priority = ?", *filters.Priority)
	}
	if filters.ScheduledBefore != nil {
		query = query.Where("ej.scheduled_for <= ?", *filters.ScheduledBefore)
	}
	if filters.ScheduledAfter != nil {
		query = query.Where("ej.scheduled_for >= ?", *filters.ScheduledAfter)
	}
	if filters.CreatedBefore != nil {
		query = query.Where("ej.created_at <= ?", *filters.CreatedBefore)
	}
	if filters.CreatedAfter != nil {
		query = query.Where("ej.created_at >= ?", *filters.CreatedAfter)
	}

	return query
}

func stringPtr(value string) *string {
	v := value
	return &v
}
