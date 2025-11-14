package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

type JobHandler func(ctx context.Context, job *Job) error

type WorkerProcessor struct {
	queue         *JobQueue
	handlers      map[JobType]JobHandler
	concurrency   int
	shutdownCh    chan struct{}
	wg            sync.WaitGroup
	mu            sync.RWMutex
	running       bool
	cleanupTicker *time.Ticker
	workerID      string
	redis         *redis.Client
	lockKeyPrefix string
}

type ProcessorConfig struct {
	Concurrency     int
	CleanupInterval time.Duration
	JobTimeout      time.Duration
	WorkerID        string
	LockKeyPrefix   string
}

func NewWorkerProcessor(queue *JobQueue, redis *redis.Client, config ProcessorConfig) *WorkerProcessor {
	if config.Concurrency == 0 {
		config.Concurrency = 5
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}
	if config.JobTimeout == 0 {
		config.JobTimeout = 30 * time.Minute
	}
	if config.WorkerID == "" {
		config.WorkerID = fmt.Sprintf("worker-%d", time.Now().UnixNano())
	}
	if config.LockKeyPrefix == "" {
		config.LockKeyPrefix = "job_lock:"
	}

	return &WorkerProcessor{
		queue:         queue,
		handlers:      make(map[JobType]JobHandler),
		concurrency:   config.Concurrency,
		shutdownCh:    make(chan struct{}),
		redis:         redis,
		workerID:      config.WorkerID,
		lockKeyPrefix: config.LockKeyPrefix,
		cleanupTicker: time.NewTicker(config.CleanupInterval),
	}
}

func (wp *WorkerProcessor) RegisterHandler(jobType JobType, handler JobHandler) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.handlers[jobType] = handler
}

func (wp *WorkerProcessor) Start(ctx context.Context) error {
	wp.mu.Lock()
	if wp.running {
		wp.mu.Unlock()
		return apperrors.Conflict("worker processor is already running")
	}
	wp.running = true
	wp.mu.Unlock()

	log.Printf("Starting worker processor %s with concurrency %d", wp.workerID, wp.concurrency)

	// Start worker goroutines
	for i := 0; i < wp.concurrency; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}

	// Start cleanup goroutine
	wp.wg.Add(1)
	go wp.cleanupWorker(ctx)

	return nil
}

func (wp *WorkerProcessor) Stop() error {
	wp.mu.Lock()
	if !wp.running {
		wp.mu.Unlock()
		return apperrors.Conflict("worker processor is not running")
	}
	wp.running = false
	wp.mu.Unlock()

	log.Printf("Stopping worker processor %s", wp.workerID)

	close(wp.shutdownCh)
	wp.cleanupTicker.Stop()
	wp.wg.Wait()

	log.Printf("Worker processor %s stopped", wp.workerID)
	return nil
}

func (wp *WorkerProcessor) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-wp.shutdownCh:
			log.Printf("Worker %d shutting down", workerID)
			return
		case <-ctx.Done():
			log.Printf("Worker %d context cancelled", workerID)
			return
		default:
			// Try to dequeue a job
			job, err := wp.queue.Dequeue(ctx, 5*time.Second)
			if err != nil {
				log.Printf("Worker %d: Failed to dequeue job: %v", workerID, err)
				continue
			}

			if job == nil {
				// No jobs available, continue polling
				continue
			}

			// Process the job with distributed locking
			err = wp.processJobWithLock(ctx, job, workerID)
			if err != nil {
				log.Printf("Worker %d: Failed to process job %s: %v", workerID, job.ID, err)
			}
		}
	}
}

func (wp *WorkerProcessor) processJobWithLock(ctx context.Context, job *Job, workerID int) error {
	lockKey := wp.lockKeyPrefix + job.ID
	lockValue := fmt.Sprintf("%s-%d", wp.workerID, workerID)
	lockTTL := 30 * time.Minute

	// Try to acquire distributed lock
	acquired, err := wp.acquireLock(ctx, lockKey, lockValue, lockTTL)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to acquire lock")
	}

	if !acquired {
		// Another worker is processing this job
		log.Printf("Worker %d: Job %s is already being processed by another worker", workerID, job.ID)
		return nil
	}

	defer func() {
		// Release lock
		err := wp.releaseLock(ctx, lockKey, lockValue)
		if err != nil {
			log.Printf("Worker %d: Failed to release lock for job %s: %v", workerID, job.ID, err)
		}
	}()

	return wp.processJob(ctx, job, workerID)
}

func (wp *WorkerProcessor) processJob(ctx context.Context, job *Job, workerID int) error {
	log.Printf("Worker %d: Processing job %s of type %s", workerID, job.ID, job.Type)

	wp.mu.RLock()
	handler, exists := wp.handlers[job.Type]
	wp.mu.RUnlock()

	if !exists {
		errorMsg := fmt.Sprintf("no handler registered for job type %s", job.Type)
		err := wp.queue.Fail(ctx, job, errorMsg)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to mark job as failed")
		}
		return apperrors.Validation(errorMsg)
	}

	// Create context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// Process the job
	err := handler(jobCtx, job)
	if err != nil {
		log.Printf("Worker %d: Job %s failed: %v", workerID, job.ID, err)
		failErr := wp.queue.Fail(ctx, job, err.Error())
		if failErr != nil {
			return apperrors.Wrap(failErr, apperrors.ErrorTypeInternal, "failed to mark job as failed")
		}
		return err
	}

	// Mark job as completed
	err = wp.queue.Complete(ctx, job)
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to mark job as completed")
	}

	log.Printf("Worker %d: Job %s completed successfully", workerID, job.ID)
	return nil
}

func (wp *WorkerProcessor) cleanupWorker(ctx context.Context) {
	defer wp.wg.Done()

	log.Printf("Cleanup worker started")

	for {
		select {
		case <-wp.shutdownCh:
			log.Printf("Cleanup worker shutting down")
			return
		case <-ctx.Done():
			log.Printf("Cleanup worker context cancelled")
			return
		case <-wp.cleanupTicker.C:
			err := wp.queue.CleanupStaleJobs(ctx, 30*time.Minute)
			if err != nil {
				log.Printf("Failed to cleanup stale jobs: %v", err)
			}
		}
	}
}

func (wp *WorkerProcessor) acquireLock(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	result, err := wp.redis.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}

func (wp *WorkerProcessor) releaseLock(ctx context.Context, key, value string) error {
	// Use Lua script to ensure we only delete the lock if we own it
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	_, err := wp.redis.Eval(ctx, script, []string{key}, value).Result()
	return err
}

func (wp *WorkerProcessor) IsRunning() bool {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return wp.running
}

func (wp *WorkerProcessor) GetStats() map[string]interface{} {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	return map[string]interface{}{
		"worker_id":   wp.workerID,
		"running":     wp.running,
		"concurrency": wp.concurrency,
		"handlers":    len(wp.handlers),
	}
}
