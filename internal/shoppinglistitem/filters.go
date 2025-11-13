package shoppinglistitem

import "time"

// Filters defines paging and filter parameters when listing shopping list items.
type Filters struct {
	IsChecked     *bool
	Categories    []string
	Tags          []string
	HasPrice      *bool
	IsLinked      *bool
	StoreIDs      []int
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string
}
