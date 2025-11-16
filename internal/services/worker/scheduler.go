package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/robfig/cron/v3"
)

type ScheduledJob struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Schedule string                 `json:"schedule"` // Cron expression
	JobType  JobType                `json:"job_type"`
	Payload  map[string]interface{} `json:"payload"`
	Enabled  bool                   `json:"enabled"`
	LastRun  *time.Time             `json:"last_run,omitempty"`
	NextRun  *time.Time             `json:"next_run,omitempty"`
}

type JobScheduler struct {
	cron          *cron.Cron
	queue         *JobQueue
	redis         *redis.Client
	lockManager   *LockManager
	scheduledJobs map[string]*ScheduledJob
	running       bool
}

func NewJobScheduler(queue *JobQueue, redis *redis.Client) *JobScheduler {
	return &JobScheduler{
		cron:          cron.New(cron.WithSeconds()),
		queue:         queue,
		redis:         redis,
		lockManager:   NewLockManager(redis, "scheduler_lock:"),
		scheduledJobs: make(map[string]*ScheduledJob),
		running:       false,
	}
}

func (js *JobScheduler) Start() error {
	if js.running {
		return apperrors.Conflict("scheduler is already running")
	}

	js.cron.Start()
	js.running = true

	log.Println("Job scheduler started")
	return nil
}

func (js *JobScheduler) Stop() {
	if !js.running {
		return
	}

	js.cron.Stop()
	js.running = false

	log.Println("Job scheduler stopped")
}

func (js *JobScheduler) AddJob(scheduledJob *ScheduledJob) error {
	if !js.running {
		return apperrors.Conflict("scheduler is not running")
	}

	if !scheduledJob.Enabled {
		return nil
	}

	entryID, err := js.cron.AddFunc(scheduledJob.Schedule, func() {
		js.executeScheduledJob(scheduledJob)
	})
	if err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to add scheduled job")
	}

	// Store the cron entry ID for later removal
	scheduledJob.ID = fmt.Sprintf("%d", entryID)
	js.scheduledJobs[scheduledJob.ID] = scheduledJob

	// Update next run time
	entries := js.cron.Entries()
	for _, entry := range entries {
		if fmt.Sprintf("%d", entry.ID) == scheduledJob.ID {
			scheduledJob.NextRun = &entry.Next
			break
		}
	}

	log.Printf("Added scheduled job: %s (%s)", scheduledJob.Name, scheduledJob.Schedule)
	return nil
}

func (js *JobScheduler) RemoveJob(jobID string) error {
	scheduledJob, exists := js.scheduledJobs[jobID]
	if !exists {
		return apperrors.NotFound(fmt.Sprintf("scheduled job %s not found", jobID))
	}

	// Parse entry ID and remove from cron
	var entryID cron.EntryID
	fmt.Sscanf(jobID, "%d", &entryID)
	js.cron.Remove(entryID)

	delete(js.scheduledJobs, jobID)

	log.Printf("Removed scheduled job: %s", scheduledJob.Name)
	return nil
}

func (js *JobScheduler) executeScheduledJob(scheduledJob *ScheduledJob) {
	ctx := context.Background()

	// Use distributed lock to ensure only one instance executes the job
	lockResource := fmt.Sprintf("scheduled_job_%s", scheduledJob.ID)

	err := js.lockManager.WithLock(ctx, lockResource, 5*time.Minute, func() error {
		log.Printf("Executing scheduled job: %s", scheduledJob.Name)

		// Create and enqueue the job
		job := &Job{
			Type:        scheduledJob.JobType,
			Priority:    5, // Medium priority for scheduled jobs
			Payload:     scheduledJob.Payload,
			MaxAttempts: 3,
			RetryDelay:  time.Minute,
		}

		err := js.queue.Enqueue(ctx, job)
		if err != nil {
			log.Printf("Failed to enqueue scheduled job %s: %v", scheduledJob.Name, err)
			return err
		}

		// Update last run time
		now := time.Now()
		scheduledJob.LastRun = &now

		log.Printf("Successfully enqueued scheduled job: %s (job ID: %s)", scheduledJob.Name, job.ID)
		return nil
	})

	if err != nil {
		log.Printf("Failed to execute scheduled job %s: %v", scheduledJob.Name, err)
	}
}

func (js *JobScheduler) GetScheduledJobs() []*ScheduledJob {
	jobs := make([]*ScheduledJob, 0, len(js.scheduledJobs))
	for _, job := range js.scheduledJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (js *JobScheduler) UpdateJob(jobID string, updates *ScheduledJob) error {
	// Remove existing job
	err := js.RemoveJob(jobID)
	if err != nil {
		return err
	}

	// Add updated job
	updates.ID = jobID
	return js.AddJob(updates)
}

// SetupDefaultSchedules sets up the default scheduled jobs for the application
func (js *JobScheduler) SetupDefaultSchedules() error {
	defaultJobs := []*ScheduledJob{
		{
			Name:     "Weekly Flyer Scraping - All Stores",
			Schedule: "0 0 6 * * MON", // Every Monday at 6 AM
			JobType:  JobTypeScrapeFlyer,
			Payload: map[string]interface{}{
				"stores": []string{"iki", "maxima", "rimi"},
				"type":   "weekly_update",
			},
			Enabled: true,
		},
		{
			Name:     "Daily Price Update Check",
			Schedule: "0 0 8 * * *", // Every day at 8 AM
			JobType:  JobTypeUpdatePrices,
			Payload: map[string]interface{}{
				"type": "daily_check",
			},
			Enabled: true,
		},
		{
			Name:     "Weekly Data Archival",
			Schedule: "0 0 2 * * SUN", // Every Sunday at 2 AM
			JobType:  JobTypeArchiveData,
			Payload: map[string]interface{}{
				"archive_older_than_days": 90,
				"type":                    "weekly_archive",
			},
			Enabled: true,
		},
		{
			Name:     "Monthly Data Cleanup",
			Schedule: "0 0 3 1 * *", // First day of every month at 3 AM
			JobType:  JobTypeCleanupData,
			Payload: map[string]interface{}{
				"cleanup_older_than_days": 180,
				"type":                    "monthly_cleanup",
			},
			Enabled: true,
		},
		{
			Name:     "Hourly Product Extraction Queue Processing",
			Schedule: "0 0 * * * *", // Every hour
			JobType:  JobTypeExtractProducts,
			Payload: map[string]interface{}{
				"batch_size": 50,
				"type":       "hourly_batch",
			},
			Enabled: true,
		},
	}

	for _, job := range defaultJobs {
		err := js.AddJob(job)
		if err != nil {
			log.Printf("Failed to add default scheduled job %s: %v", job.Name, err)
			continue
		}
	}

	log.Printf("Set up %d default scheduled jobs", len(defaultJobs))
	return nil
}

func (js *JobScheduler) IsRunning() bool {
	return js.running
}

func (js *JobScheduler) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"running":        js.running,
		"scheduled_jobs": len(js.scheduledJobs),
		"cron_entries":   len(js.cron.Entries()),
	}

	// Add details about each scheduled job
	jobDetails := make([]map[string]interface{}, 0, len(js.scheduledJobs))
	for _, job := range js.scheduledJobs {
		detail := map[string]interface{}{
			"id":       job.ID,
			"name":     job.Name,
			"schedule": job.Schedule,
			"job_type": job.JobType,
			"enabled":  job.Enabled,
		}

		if job.LastRun != nil {
			detail["last_run"] = job.LastRun
		}
		if job.NextRun != nil {
			detail["next_run"] = job.NextRun
		}

		jobDetails = append(jobDetails, detail)
	}
	stats["job_details"] = jobDetails

	return stats
}
