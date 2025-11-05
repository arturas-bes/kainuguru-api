package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type JobType string

const (
	JobTypeScrapeFlyer     JobType = "scrape_flyer"
	JobTypeExtractProducts JobType = "extract_products"
	JobTypeUpdatePrices    JobType = "update_prices"
	JobTypeArchiveData     JobType = "archive_data"
	JobTypeCleanupData     JobType = "cleanup_data"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
)

type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	Priority    int                    `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	RetryDelay  time.Duration          `json:"retry_delay"`
}

type JobQueue struct {
	redis         *redis.Client
	queueName     string
	processingSet string
	deadLetterSet string
}

func NewJobQueue(redisClient *redis.Client, queueName string) *JobQueue {
	return &JobQueue{
		redis:         redisClient,
		queueName:     queueName,
		processingSet: queueName + ":processing",
		deadLetterSet: queueName + ":dead_letter",
	}
}

func (q *JobQueue) Enqueue(ctx context.Context, job *Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	job.Status = JobStatusPending

	if job.MaxAttempts == 0 {
		job.MaxAttempts = 3
	}

	if job.RetryDelay == 0 {
		job.RetryDelay = 30 * time.Second
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// If job is scheduled, add to scheduled set
	if job.ScheduledAt != nil && job.ScheduledAt.After(time.Now()) {
		score := float64(job.ScheduledAt.Unix())
		return q.redis.ZAdd(ctx, q.queueName+":scheduled", &redis.Z{
			Score:  score,
			Member: string(jobData),
		}).Err()
	}

	// Add to priority queue (higher priority = lower score for ZPOPMIN)
	score := float64(-job.Priority) // Negative for reverse ordering
	return q.redis.ZAdd(ctx, q.queueName, &redis.Z{
		Score:  score,
		Member: string(jobData),
	}).Err()
}

func (q *JobQueue) Dequeue(ctx context.Context, timeout time.Duration) (*Job, error) {
	// First check for scheduled jobs that are ready
	err := q.moveScheduledJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to move scheduled jobs: %w", err)
	}

	// Pop job with highest priority
	result, err := q.redis.BZPopMin(ctx, timeout, q.queueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	var job Job
	err = json.Unmarshal([]byte(result.Member.(string)), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Move to processing set
	job.Status = JobStatusProcessing
	job.StartedAt = &time.Time{}
	*job.StartedAt = time.Now()
	job.UpdatedAt = time.Now()

	jobData, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal processing job: %w", err)
	}

	err = q.redis.SAdd(ctx, q.processingSet, string(jobData)).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to add job to processing set: %w", err)
	}

	return &job, nil
}

func (q *JobQueue) Complete(ctx context.Context, job *Job) error {
	job.Status = JobStatusCompleted
	job.CompletedAt = &time.Time{}
	*job.CompletedAt = time.Now()
	job.UpdatedAt = time.Now()

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal completed job: %w", err)
	}

	// Remove from processing set
	return q.redis.SRem(ctx, q.processingSet, string(jobData)).Err()
}

func (q *JobQueue) Fail(ctx context.Context, job *Job, errorMsg string) error {
	job.Error = errorMsg
	job.Attempts++
	job.UpdatedAt = time.Now()

	// Check if we should retry
	if job.Attempts < job.MaxAttempts {
		job.Status = JobStatusRetrying

		// Schedule retry with exponential backoff
		delay := time.Duration(job.Attempts) * job.RetryDelay
		retryAt := time.Now().Add(delay)
		job.ScheduledAt = &retryAt

		jobData, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal retry job: %w", err)
		}

		// Remove from processing set
		oldJobData, _ := json.Marshal(&Job{
			ID:     job.ID,
			Status: JobStatusProcessing,
		})
		q.redis.SRem(ctx, q.processingSet, string(oldJobData))

		// Add to scheduled set for retry
		score := float64(job.ScheduledAt.Unix())
		return q.redis.ZAdd(ctx, q.queueName+":scheduled", &redis.Z{
			Score:  score,
			Member: string(jobData),
		}).Err()
	}

	// Move to dead letter queue
	job.Status = JobStatusFailed
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal failed job: %w", err)
	}

	// Remove from processing set
	oldJobData, _ := json.Marshal(&Job{
		ID:     job.ID,
		Status: JobStatusProcessing,
	})
	q.redis.SRem(ctx, q.processingSet, string(oldJobData))

	// Add to dead letter set
	return q.redis.SAdd(ctx, q.deadLetterSet, string(jobData)).Err()
}

func (q *JobQueue) GetQueueStats(ctx context.Context) (map[string]int64, error) {
	pending, err := q.redis.ZCard(ctx, q.queueName).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}

	processing, err := q.redis.SCard(ctx, q.processingSet).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get processing count: %w", err)
	}

	scheduled, err := q.redis.ZCard(ctx, q.queueName+":scheduled").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled count: %w", err)
	}

	deadLetter, err := q.redis.SCard(ctx, q.deadLetterSet).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get dead letter count: %w", err)
	}

	return map[string]int64{
		"pending":     pending,
		"processing":  processing,
		"scheduled":   scheduled,
		"dead_letter": deadLetter,
	}, nil
}

func (q *JobQueue) moveScheduledJobs(ctx context.Context) error {
	now := time.Now().Unix()

	// Get jobs that are ready to be processed
	jobs, err := q.redis.ZRangeByScore(ctx, q.queueName+":scheduled", &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%d", now),
	}).Result()
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return nil
	}

	// Move jobs to main queue
	pipe := q.redis.Pipeline()
	for _, jobStr := range jobs {
		var job Job
		if err := json.Unmarshal([]byte(jobStr), &job); err != nil {
			continue
		}

		// Update job status
		job.Status = JobStatusPending
		job.ScheduledAt = nil
		job.UpdatedAt = time.Now()

		jobData, err := json.Marshal(job)
		if err != nil {
			continue
		}

		// Add to main queue with priority
		score := float64(-job.Priority)
		pipe.ZAdd(ctx, q.queueName, &redis.Z{
			Score:  score,
			Member: string(jobData),
		})

		// Remove from scheduled set
		pipe.ZRem(ctx, q.queueName+":scheduled", jobStr)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (q *JobQueue) CleanupStaleJobs(ctx context.Context, timeout time.Duration) error {
	// Find jobs that have been processing for too long
	cutoff := time.Now().Add(-timeout)

	members, err := q.redis.SMembers(ctx, q.processingSet).Result()
	if err != nil {
		return fmt.Errorf("failed to get processing jobs: %w", err)
	}

	for _, member := range members {
		var job Job
		if err := json.Unmarshal([]byte(member), &job); err != nil {
			continue
		}

		if job.StartedAt != nil && job.StartedAt.Before(cutoff) {
			// Job has been processing too long, move back to queue for retry
			err := q.Fail(ctx, &job, "job processing timeout")
			if err != nil {
				return fmt.Errorf("failed to fail stale job %s: %w", job.ID, err)
			}
		}
	}

	return nil
}
