package product

import "time"

// Filters define the available options when querying products.
type Filters struct {
	StoreIDs         []int
	FlyerIDs         []int
	FlyerPageIDs     []int
	ProductMasterIDs []int
	Categories       []string
	Brands           []string
	IsOnSale         *bool
	IsAvailable      *bool
	RequiresReview   *bool
	MinPrice         *float64
	MaxPrice         *float64
	Currency         string
	ValidFrom        *time.Time
	ValidTo          *time.Time
	Limit            int
	Offset           int
	OrderBy          string
	OrderDir         string
}
