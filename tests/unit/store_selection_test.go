package wizard_test

import (
	"testing"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services/wizard"
	"github.com/stretchr/testify/assert"
)

// T033: Verify maxStores constraint is never violated
func TestSelectOptimalStores_MaxStoresConstraint(t *testing.T) {
	tests := []struct {
		name            string
		suggestions     map[int][]*models.Suggestion
		maxStores       int
		expectedStores  int
		expectedCoverage float64
	}{
		{
			name: "single store covers all items",
			suggestions: map[int][]*models.Suggestion{
				1: {
					{FlyerProductID: 101, StoreID: 1, ProductName: "Coca-Cola", Price: 1.99},
				},
				2: {
					{FlyerProductID: 102, StoreID: 1, ProductName: "Pepsi", Price: 1.79},
				},
				3: {
					{FlyerProductID: 103, StoreID: 1, ProductName: "Sprite", Price: 1.89},
				},
			},
			maxStores:        2,
			expectedStores:   1,
			expectedCoverage: 100.0,
		},
		{
			name: "two stores required for full coverage",
			suggestions: map[int][]*models.Suggestion{
				1: {
					{FlyerProductID: 101, StoreID: 1, ProductName: "Coca-Cola", Price: 1.99},
				},
				2: {
					{FlyerProductID: 102, StoreID: 2, ProductName: "Pepsi", Price: 1.79},
				},
				3: {
					{FlyerProductID: 103, StoreID: 1, ProductName: "Sprite", Price: 1.89},
					{FlyerProductID: 104, StoreID: 2, ProductName: "Sprite", Price: 1.95},
				},
			},
			maxStores:        2,
			expectedStores:   2,
			expectedCoverage: 100.0,
		},
		{
			name: "five stores available but max 2 enforced",
			suggestions: map[int][]*models.Suggestion{
				1: {{FlyerProductID: 101, StoreID: 1, ProductName: "Item1", Price: 1.99}},
				2: {{FlyerProductID: 102, StoreID: 2, ProductName: "Item2", Price: 2.49}},
				3: {{FlyerProductID: 103, StoreID: 3, ProductName: "Item3", Price: 3.99}},
				4: {{FlyerProductID: 104, StoreID: 4, ProductName: "Item4", Price: 4.99}},
				5: {{FlyerProductID: 105, StoreID: 5, ProductName: "Item5", Price: 5.99}},
			},
			maxStores:        2,
			expectedStores:   2,
			expectedCoverage: 40.0,
		},
		{
			name: "maxStores=0 defaults to 1",
			suggestions: map[int][]*models.Suggestion{
				1: {
					{FlyerProductID: 101, StoreID: 1, ProductName: "Item1", Price: 1.99},
					{FlyerProductID: 201, StoreID: 2, ProductName: "Item1", Price: 2.49},
				},
			},
			maxStores:        1,
			expectedStores:   1,
			expectedCoverage: 100.0,
		},
		{
			name: "maxStores>2 capped to 2 (constitution)",
			suggestions: map[int][]*models.Suggestion{
				1: {{FlyerProductID: 101, StoreID: 1, ProductName: "Item1", Price: 1.99}},
				2: {{FlyerProductID: 102, StoreID: 2, ProductName: "Item2", Price: 2.49}},
				3: {{FlyerProductID: 103, StoreID: 3, ProductName: "Item3", Price: 3.99}},
			},
			maxStores:        5,
			expectedStores:   2,
			expectedCoverage: 66.67,
		},
		{
			name: "greedy picks best coverage first",
			suggestions: map[int][]*models.Suggestion{
				1: {
					{FlyerProductID: 101, StoreID: 1, ProductName: "Item1", Price: 1.99},
				},
				2: {
					{FlyerProductID: 102, StoreID: 1, ProductName: "Item2", Price: 2.49},
				},
				3: {
					{FlyerProductID: 103, StoreID: 1, ProductName: "Item3", Price: 3.99},
				},
				4: {
					{FlyerProductID: 104, StoreID: 2, ProductName: "Item4", Price: 4.99},
				},
			},
			maxStores:        2,
			expectedStores:   2,
			expectedCoverage: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wizard.SelectOptimalStores(tt.suggestions, tt.maxStores)

			assert.LessOrEqual(t, len(result.SelectedStores), tt.maxStores, "store count exceeds maxStores")
			assert.LessOrEqual(t, len(result.SelectedStores), 2, "store count exceeds constitution limit of 2")
			assert.Equal(t, tt.expectedStores, len(result.SelectedStores), "unexpected number of stores selected")
			assert.InDelta(t, tt.expectedCoverage, result.CoveragePercent, 1.0, "unexpected coverage percentage")

			result2 := wizard.SelectOptimalStores(tt.suggestions, tt.maxStores)
			assert.Equal(t, result.SelectedStores, result2.SelectedStores, "non-deterministic store selection")
		})
	}
}

func TestSelectOptimalStores_SecondStoreJustification(t *testing.T) {
	tests := []struct {
		name           string
		suggestions    map[int][]*models.Suggestion
		expectedStores int
		reason         string
	}{
		{
			name: "second store adds no new coverage - only 1 store selected",
			suggestions: map[int][]*models.Suggestion{
				1: {
					{FlyerProductID: 101, StoreID: 1, ProductName: "Item1", Price: 1.99},
					{FlyerProductID: 201, StoreID: 2, ProductName: "Item1", Price: 2.99},
				},
				2: {
					{FlyerProductID: 102, StoreID: 1, ProductName: "Item2", Price: 2.49},
					{FlyerProductID: 202, StoreID: 2, ProductName: "Item2", Price: 3.49},
				},
			},
			expectedStores: 1,
			reason:         "store 1 covers all items, store 2 adds no value",
		},
		{
			name: "second store adds 1 new item - justified",
			suggestions: map[int][]*models.Suggestion{
				1: {
					{FlyerProductID: 101, StoreID: 1, ProductName: "Item1", Price: 1.99},
				},
				2: {
					{FlyerProductID: 102, StoreID: 1, ProductName: "Item2", Price: 2.49},
				},
				3: {
					{FlyerProductID: 103, StoreID: 2, ProductName: "Item3", Price: 3.99},
				},
			},
			expectedStores: 2,
			reason:         "store 2 adds coverage for item 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wizard.SelectOptimalStores(tt.suggestions, 2)
			assert.Equal(t, tt.expectedStores, len(result.SelectedStores), tt.reason)
		})
	}
}

func TestSelectOptimalStores_PriceTieBreaking(t *testing.T) {
	suggestions := map[int][]*models.Suggestion{
		1: {
			{FlyerProductID: 101, StoreID: 1, ProductName: "Item1", Price: 5.00},
			{FlyerProductID: 201, StoreID: 2, ProductName: "Item1", Price: 3.00},
		},
		2: {
			{FlyerProductID: 102, StoreID: 1, ProductName: "Item2", Price: 5.00},
			{FlyerProductID: 202, StoreID: 2, ProductName: "Item2", Price: 3.00},
		},
	}

	result := wizard.SelectOptimalStores(suggestions, 2)

	assert.Equal(t, 1, len(result.SelectedStores), "should only pick one store when coverage is identical")
	assert.Contains(t, result.SelectedStores, 2, "should pick cheaper store when coverage is equal")
}

func TestSelectOptimalStores_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		suggestions map[int][]*models.Suggestion
		maxStores   int
	}{
		{
			name:        "empty suggestions map",
			suggestions: map[int][]*models.Suggestion{},
			maxStores:   2,
		},
		{
			name:        "nil suggestions map",
			suggestions: nil,
			maxStores:   2,
		},
		{
			name: "suggestions with empty suggestion lists",
			suggestions: map[int][]*models.Suggestion{
				1: {},
				2: {},
			},
			maxStores: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wizard.SelectOptimalStores(tt.suggestions, tt.maxStores)

			assert.NotNil(t, result)
			assert.Equal(t, 0, len(result.SelectedStores))
			assert.Equal(t, 0.0, result.CoveragePercent)
		})
	}
}
