package store

// Filters defines the available options when querying stores.
type Filters struct {
	IsActive  *bool
	HasFlyers *bool
	Codes     []string
	Limit     int
	Offset    int
	OrderBy   string
	OrderDir  string
}
