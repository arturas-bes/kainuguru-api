package extractionjob

import "time"

// Filters defines the query options available when listing extraction jobs.
type Filters struct {
	JobTypes        []string
	Status          []string
	WorkerIDs       []string
	Priority        *int
	ScheduledBefore *time.Time
	ScheduledAfter  *time.Time
	CreatedBefore   *time.Time
	CreatedAfter    *time.Time
	Limit           int
	Offset          int
	OrderBy         string
	OrderDir        string
}
