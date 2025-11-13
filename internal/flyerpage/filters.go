package flyerpage

// Filters define the available options when querying flyer pages.
type Filters struct {
	FlyerIDs    []int
	Status      []string
	HasImage    *bool
	HasProducts *bool
	PageNumbers []int
	Limit       int
	Offset      int
	OrderBy     string
	OrderDir    string
}
