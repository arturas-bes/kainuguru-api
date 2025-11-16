package wizard

import (
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// StoreSelectionResult represents the result of store selection optimization
type StoreSelectionResult struct {
	SelectedStores  []int         // Store IDs selected
	ItemCoverage    map[int][]int // Map of store_id -> list of item_ids covered
	TotalItems      int           // Total items covered across all stores
	UncoveredItems  []int         // Items not covered by any selected store
	CoveragePercent float64       // Percentage of items covered (0-100)
	TotalPrice      float64       // Total estimated price across all stores
	Explanation     string        // Human-readable explanation of selection
}

// SelectOptimalStores implements a greedy algorithm to select stores that maximize item coverage
// Constraints:
// - maxStores: Maximum number of stores to select (default 1, allow up to 2 per constitution)
// - Must maximize coverage (number of items covered)
// - Break ties by minimizing total price
// - Returns deterministic results for same inputs
func SelectOptimalStores(
	suggestions map[int][]*models.Suggestion, // Map of item_id -> list of suggestions
	maxStores int,
) *StoreSelectionResult {
	if maxStores <= 0 {
		maxStores = 1 // Default to single store
	}
	if maxStores > 2 {
		maxStores = 2 // Constitution: max 2 stores
	}

	// Build store coverage map: store_id -> list of items it can cover
	storeCoverage := make(map[int][]int)
	storePrice := make(map[int]float64)

	for itemID, itemSuggestions := range suggestions {
		if len(itemSuggestions) == 0 {
			continue
		}

		// For each suggestion, track which store can cover this item
		for _, suggestion := range itemSuggestions {
			storeID := suggestion.StoreID

			// Add item to store's coverage
			if _, exists := storeCoverage[storeID]; !exists {
				storeCoverage[storeID] = []int{}
			}
			storeCoverage[storeID] = append(storeCoverage[storeID], itemID)

			// Track minimum price per store (greedy: pick cheapest suggestion per item)
			storePrice[storeID] += suggestion.Price
		}
	}

	// Greedy selection: pick stores that maximize coverage
	selectedStores := []int{}
	coveredItems := make(map[int]bool)
	totalPrice := 0.0

	for len(selectedStores) < maxStores && len(storeCoverage) > 0 {
		// Find store that covers the most uncovered items
		bestStore := -1
		bestCoverage := 0
		bestPrice := 0.0

		for storeID, items := range storeCoverage {
			// Count how many NEW items this store covers
			newCoverage := 0
			for _, itemID := range items {
				if !coveredItems[itemID] {
					newCoverage++
				}
			}

			// Pick store with highest new coverage, break ties by lowest price
			if newCoverage > bestCoverage || (newCoverage == bestCoverage && storePrice[storeID] < bestPrice) {
				bestStore = storeID
				bestCoverage = newCoverage
				bestPrice = storePrice[storeID]
			}
		}

		if bestStore == -1 || bestCoverage == 0 {
			break // No more stores add value
		}

		// Add best store to selection
		selectedStores = append(selectedStores, bestStore)
		totalPrice += bestPrice

		// Mark items as covered
		for _, itemID := range storeCoverage[bestStore] {
			coveredItems[itemID] = true
		}

		// Remove selected store from pool
		delete(storeCoverage, bestStore)
	}

	// Build final result
	itemCoverage := make(map[int][]int)
	uncoveredItems := []int{}

	// Populate item coverage per store
	for itemID, itemSuggestions := range suggestions {
		covered := false
		for _, storeID := range selectedStores {
			// Check if this store has a suggestion for this item
			for _, suggestion := range itemSuggestions {
				if suggestion.StoreID == storeID {
					itemCoverage[storeID] = append(itemCoverage[storeID], itemID)
					covered = true
					break
				}
			}
			if covered {
				break
			}
		}
		if !covered {
			uncoveredItems = append(uncoveredItems, itemID)
		}
	}

	// Calculate coverage percentage
	totalItems := len(suggestions)
	coveredCount := len(coveredItems)
	coveragePercent := 0.0
	if totalItems > 0 {
		coveragePercent = (float64(coveredCount) / float64(totalItems)) * 100.0
	}

	// Generate explanation
	explanation := generateStoreSelectionExplanation(selectedStores, totalItems, coveredCount, totalPrice)

	return &StoreSelectionResult{
		SelectedStores:  selectedStores,
		ItemCoverage:    itemCoverage,
		TotalItems:      coveredCount,
		UncoveredItems:  uncoveredItems,
		CoveragePercent: coveragePercent,
		TotalPrice:      totalPrice,
		Explanation:     explanation,
	}
}

// generateStoreSelectionExplanation creates a human-readable explanation
func generateStoreSelectionExplanation(stores []int, totalItems, coveredItems int, totalPrice float64) string {
	if len(stores) == 0 {
		return "No stores selected - no available alternatives"
	}

	if len(stores) == 1 {
		return "Selected 1 store for optimal coverage"
	}

	return "Selected 2 stores to maximize item coverage"
}
