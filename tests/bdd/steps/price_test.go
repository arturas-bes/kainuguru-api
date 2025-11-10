//go:build integration
// +build integration

package steps

import (
	"os"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/tests/bdd/helpers"
)

// TestPriceHistoryAPIIntegration verifies that the BDD tests actually call real GraphQL endpoints
func TestPriceHistoryAPIIntegration(t *testing.T) {
	// Get API URL
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	// Create test context with real client
	testCtx := helpers.NewTestContext(apiURL)

	t.Run("Can query price history via real GraphQL API", func(t *testing.T) {
		query := `
			query GetPriceHistory($productMasterID: Int!, $first: Int) {
				priceHistory(productMasterID: $productMasterID, first: $first) {
					edges {
						node {
							id
							price
							currency
							isOnSale
							recordedAt
						}
					}
					pageInfo {
						hasNextPage
						hasPreviousPage
					}
				}
			}
		`

		variables := map[string]interface{}{
			"productMasterID": 1,
			"first":           10,
		}

		// Make real API call
		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		// Verify response
		if resp == nil {
			t.Fatal("Response is nil")
		}

		t.Logf("Response time: %v", duration)

		// Check for errors
		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify data structure
		if resp.Data == nil {
			t.Fatal("Response data is nil")
		}

		dataMap, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Invalid response data structure")
		}

		priceHistoryData, ok := dataMap["priceHistory"].(map[string]interface{})
		if !ok {
			t.Fatal("priceHistory not found in response")
		}

		edges, ok := priceHistoryData["edges"].([]interface{})
		if !ok {
			t.Fatal("edges not found in priceHistory")
		}

		if len(edges) == 0 {
			t.Fatal("Expected at least one price history entry")
		}

		// Verify edge structure
		firstEdge := edges[0].(map[string]interface{})
		node := firstEdge["node"].(map[string]interface{})

		// Check required fields
		if _, ok := node["price"]; !ok {
			t.Fatal("price field missing")
		}
		if _, ok := node["currency"]; !ok {
			t.Fatal("currency field missing")
		}
		if _, ok := node["recordedAt"]; !ok {
			t.Fatal("recordedAt field missing")
		}

		t.Logf("✅ Successfully retrieved %d price history entries from real API", len(edges))
	})

	t.Run("Can query current price via real GraphQL API", func(t *testing.T) {
		query := `
			query GetCurrentPrice($productMasterID: Int!, $storeID: Int) {
				currentPrice(productMasterID: $productMasterID, storeID: $storeID) {
					id
					price
					currency
					isOnSale
					recordedAt
				}
			}
		`

		variables := map[string]interface{}{
			"productMasterID": 1,
			"storeID":         1,
		}

		// Make real API call
		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		// Check response is under 2 seconds
		if duration > 2*time.Second {
			t.Errorf("Response time %v exceeds 2 second threshold", duration)
		}

		// Check for errors
		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify data
		dataMap := resp.Data.(map[string]interface{})
		currentPrice, ok := dataMap["currentPrice"].(map[string]interface{})
		if !ok || currentPrice == nil {
			t.Fatal("currentPrice not found in response")
		}

		// Verify price field exists
		price, ok := currentPrice["price"].(float64)
		if !ok {
			t.Fatal("price field missing or invalid type")
		}

		t.Logf("✅ Successfully retrieved current price: €%.2f from real API", price)
	})

	t.Run("Can filter price history by isOnSale via real API", func(t *testing.T) {
		query := `
			query GetPriceHistory($productMasterID: Int!, $filters: PriceHistoryFilters) {
				priceHistory(productMasterID: $productMasterID, filters: $filters) {
					edges {
						node {
							id
							price
							isOnSale
						}
					}
				}
			}
		`

		variables := map[string]interface{}{
			"productMasterID": 1,
			"filters": map[string]interface{}{
				"isOnSale": true,
			},
		}

		// Make real API call
		resp, _, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		// Check for errors
		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify filtered results
		dataMap := resp.Data.(map[string]interface{})
		priceHistoryData := dataMap["priceHistory"].(map[string]interface{})
		edges := priceHistoryData["edges"].([]interface{})

		// Verify all results have isOnSale = true
		for _, edge := range edges {
			edgeMap := edge.(map[string]interface{})
			node := edgeMap["node"].(map[string]interface{})
			isOnSale := node["isOnSale"].(bool)
			if !isOnSale {
				t.Errorf("Expected isOnSale=true, got false")
			}
		}

		t.Logf("✅ Successfully filtered %d sale prices from real API", len(edges))
	})
}
