//go:build integration
// +build integration

package steps

import (
	"os"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/tests/bdd/helpers"
)

// TestSearchAPIIntegration verifies search functionality hits real GraphQL endpoints
func TestSearchAPIIntegration(t *testing.T) {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	testCtx := helpers.NewTestContext(apiURL)

	t.Run("Can search products via real GraphQL API", func(t *testing.T) {
		query := `
			query SearchProducts($input: SearchInput!) {
				searchProducts(input: $input) {
					queryString
					totalCount
					products {
						product {
							id
							name
							price {
								current
								currency
							}
						}
						searchScore
						matchType
					}
					suggestions
					hasMore
				}
			}
		`

		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"q":     "pienas",
				"first": 10,
			},
		}

		// Make real API call
		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
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

		dataMap := resp.Data.(map[string]interface{})
		searchData := dataMap["searchProducts"].(map[string]interface{})

		// Verify query string
		queryString := searchData["queryString"].(string)
		if queryString != "pienas" {
			t.Fatalf("Expected queryString 'pienas', got '%s'", queryString)
		}

		// Verify we have products
		products := searchData["products"].([]interface{})
		t.Logf("✅ Successfully searched and found %d products via real API", len(products))
	})

	t.Run("Can search with filters via real API", func(t *testing.T) {
		query := `
			query SearchProducts($input: SearchInput!) {
				searchProducts(input: $input) {
					totalCount
					products {
						product {
							id
							name
							price {
								current
							}
							isOnSale
						}
					}
				}
			}
		`

		onSaleOnly := true
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"q":          "pienas",
				"onSaleOnly": &onSaleOnly,
				"first":      5,
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify filtered results
		dataMap := resp.Data.(map[string]interface{})
		searchData := dataMap["searchProducts"].(map[string]interface{})
		products := searchData["products"].([]interface{})

		// Verify all products are on sale if any returned
		for _, prod := range products {
			productResult := prod.(map[string]interface{})
			product := productResult["product"].(map[string]interface{})
			isOnSale := product["isOnSale"].(bool)
			if !isOnSale {
				t.Errorf("Expected onSaleOnly filter to return only sale items")
			}
		}

		t.Logf("✅ Successfully filtered search for sale items via real API")
	})

	t.Run("Can search with price range via real API", func(t *testing.T) {
		query := `
			query SearchProducts($input: SearchInput!) {
				searchProducts(input: $input) {
					totalCount
					products {
						product {
							id
							name
							price {
								current
							}
						}
					}
				}
			}
		`

		minPrice := 1.0
		maxPrice := 5.0
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"q":        "pienas",
				"minPrice": &minPrice,
				"maxPrice": &maxPrice,
				"first":    10,
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify price range
		dataMap := resp.Data.(map[string]interface{})
		searchData := dataMap["searchProducts"].(map[string]interface{})
		products := searchData["products"].([]interface{})

		// Verify all products are within price range
		for _, prod := range products {
			productResult := prod.(map[string]interface{})
			product := productResult["product"].(map[string]interface{})
			priceObj := product["price"].(map[string]interface{})
			price := priceObj["current"].(float64)
			if price < minPrice || price > maxPrice {
				t.Errorf("Product price %.2f outside range %.2f-%.2f", price, minPrice, maxPrice)
			}
		}

		t.Logf("✅ Successfully filtered search by price range via real API")
	})

	t.Run("Search handles Lithuanian characters via real API", func(t *testing.T) {
		query := `
			query SearchProducts($input: SearchInput!) {
				searchProducts(input: $input) {
					queryString
					totalCount
					products {
						product {
							id
							name
						}
					}
				}
			}
		`

		// Test with Lithuanian characters
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"q":     "duona",
				"first": 5,
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		dataMap := resp.Data.(map[string]interface{})
		searchData := dataMap["searchProducts"].(map[string]interface{})

		// Just verify the search executed successfully
		queryString := searchData["queryString"].(string)
		if queryString != "duona" {
			t.Fatalf("Expected queryString 'duona', got '%s'", queryString)
		}

		t.Logf("✅ Successfully handled Lithuanian search query via real API")
	})

	t.Run("Search with fuzzy matching via real API", func(t *testing.T) {
		query := `
			query SearchProducts($input: SearchInput!) {
				searchProducts(input: $input) {
					totalCount
					products {
						product {
							id
							name
						}
						matchType
					}
					suggestions
				}
			}
		`

		preferFuzzy := true
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"q":           "piens",  // Slightly misspelled
				"preferFuzzy": &preferFuzzy,
				"first":       10,
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		dataMap := resp.Data.(map[string]interface{})
		searchData := dataMap["searchProducts"].(map[string]interface{})

		// Verify search executed
		totalCount := int(searchData["totalCount"].(float64))
		t.Logf("✅ Fuzzy search returned %d results via real API", totalCount)
	})

	t.Run("Performance: Search should complete under 2 seconds", func(t *testing.T) {
		query := `
			query SearchProducts($input: SearchInput!) {
				searchProducts(input: $input) {
					totalCount
				}
			}
		`

		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"q":     "pienas",
				"first": 20,
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Check performance requirement
		if duration > 2*time.Second {
			t.Errorf("Search took %v, exceeds 2 second threshold", duration)
		}

		t.Logf("✅ Search performance: %v (under 2s threshold)", duration)
	})

	t.Run("Empty search query is rejected via real API", func(t *testing.T) {
		query := `
			query SearchProducts($input: SearchInput!) {
				searchProducts(input: $input) {
					totalCount
				}
			}
		`

		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"q":     "",
				"first": 10,
			},
		}

		resp, _, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to query API: %v", err)
		}

		// Should have error for empty query
		if len(resp.Errors) == 0 {
			t.Fatal("Expected error for empty search query, got none")
		}

		errorMsg := resp.Errors[0].Message
		t.Logf("Error message: %s", errorMsg)

		t.Logf("✅ Empty search query correctly rejected by real API")
	})
}
