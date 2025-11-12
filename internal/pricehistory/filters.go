package pricehistory

import "time"

// Filters define the available options when querying price history entries.
type Filters struct {
	IsOnSale    *bool
	IsAvailable *bool
	IsActive    *bool
	MinPrice    *float64
	MaxPrice    *float64
	Source      *string
	DateFrom    *time.Time
	DateTo      *time.Time
	Limit       int
	Offset      int
	OrderBy     string
	OrderDir    string
}
