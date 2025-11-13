package productmaster

import "time"

// Filters define the available options when querying product masters.
type Filters struct {
	Status        []string
	IsVerified    *bool
	IsActive      *bool
	Categories    []string
	Brands        []string
	MinMatches    *int
	MinConfidence *float64
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
}
