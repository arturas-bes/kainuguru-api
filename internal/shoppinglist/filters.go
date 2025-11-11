package shoppinglist

import "time"

// Filters defines paging and filter parameters when listing shopping lists.
type Filters struct {
	IsDefault     *bool
	IsArchived    *bool
	IsPublic      *bool
	HasItems      *bool
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	OrderBy       string
	OrderDir      string
	Limit         int
	Offset        int
}
