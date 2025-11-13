package shoppinglistitem

// ItemOrder represents a desired ordering for a shopping list item.
type ItemOrder struct {
	ItemID    int64 `json:"item_id"`
	SortOrder int   `json:"sort_order"`
}
