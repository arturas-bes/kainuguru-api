package flyer

import "time"

// Filters defines query parameters when listing flyers.
type Filters struct {
	StoreIDs   []int
	StoreCode  *string
	StoreCodes []string
	Status     []string
	IsArchived *bool
	ValidFrom  *time.Time
	ValidTo    *time.Time
	ValidOn    *string
	IsCurrent  *bool
	IsValid    *bool
	Limit      int
	Offset     int
	OrderBy    string
	OrderDir   string
}
