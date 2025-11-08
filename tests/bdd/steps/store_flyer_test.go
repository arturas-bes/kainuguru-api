//go:build integration
// +build integration

package steps

import (
	"os"
	"testing"

	"github.com/kainuguru/kainuguru-api/tests/bdd/helpers"
)

// TestStoreAndFlyerAPIIntegration verifies store and flyer operations hit real GraphQL endpoints
func TestStoreAndFlyerAPIIntegration(t *testing.T) {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	testCtx := helpers.NewTestContext(apiURL)

	t.Run("Can query all stores via real API", func(t *testing.T) {
		query := `
			query GetStores {
				stores {
					edges {
						node {
							id
							code
							name
							logoURL
							websiteURL
							isActive
							createdAt
							updatedAt
						}
					}
					pageInfo {
						hasNextPage
					}
				}
			}
		`

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, nil)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify response structure
		if resp.Data == nil {
			t.Fatal("No data returned")
		}

		dataMap := resp.Data.(map[string]interface{})
		storesData := dataMap["stores"].(map[string]interface{})
		edges := storesData["edges"].([]interface{})

		if len(edges) == 0 {
			t.Log("⚠️  No stores found in database")
		} else {
			t.Logf("✅ Successfully retrieved %d stores via real API", len(edges))
		}
	})

	t.Run("Can query store by ID via real API", func(t *testing.T) {
		query := `
			query GetStore($id: Int!) {
				store(id: $id) {
					id
					code
					name
					logoURL
					websiteURL
					isActive
					createdAt
					updatedAt
				}
			}
		`

		variables := map[string]interface{}{
			"id": 1,
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		if resp.Data == nil {
			t.Fatal("No data returned")
		}

		dataMap := resp.Data.(map[string]interface{})
		storeData := dataMap["store"]

		if storeData == nil {
			t.Log("⚠️  Store with ID 1 not found")
		} else {
			store := storeData.(map[string]interface{})
			t.Logf("✅ Successfully retrieved store '%s' via real API", store["name"])
		}
	})

	t.Run("Can query all flyers via real API", func(t *testing.T) {
		query := `
			query GetFlyers {
				flyers {
					edges {
						node {
							id
							storeID
							title
							validFrom
							validTo
							status
							createdAt
							updatedAt
						}
					}
					pageInfo {
						hasNextPage
					}
				}
			}
		`

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, nil)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify response structure
		if resp.Data == nil {
			t.Fatal("No data returned")
		}

		dataMap := resp.Data.(map[string]interface{})
		flyersData := dataMap["flyers"].(map[string]interface{})
		edges := flyersData["edges"].([]interface{})

		if len(edges) == 0 {
			t.Log("⚠️  No flyers found in database")
		} else {
			t.Logf("✅ Successfully retrieved %d flyers via real API", len(edges))
		}
	})

	t.Run("Can query current flyers via real API", func(t *testing.T) {
		query := `
			query GetCurrentFlyers {
				currentFlyers {
					edges {
						node {
							id
							storeID
							title
							validFrom
							validTo
							status
							daysRemaining
							validityPeriod
						}
					}
					pageInfo {
						hasNextPage
					}
				}
			}
		`

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, nil)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			// Log error details
			for i, gqlErr := range resp.Errors {
				t.Logf("Error %d: %s", i+1, gqlErr.Message)
				if gqlErr.Extensions != nil {
					t.Logf("  Extensions: %v", gqlErr.Extensions)
				}
			}
			t.Fatalf("GraphQL errors encountered")
		}

		// Verify response structure
		if resp.Data == nil {
			t.Fatal("No data returned")
		}

		dataMap := resp.Data.(map[string]interface{})
		flyersData := dataMap["currentFlyers"].(map[string]interface{})
		edges := flyersData["edges"].([]interface{})

		if len(edges) == 0 {
			t.Log("⚠️  No current flyers found in database")
		} else {
			// Get first flyer to check computed fields
			firstEdge := edges[0].(map[string]interface{})
			node := firstEdge["node"].(map[string]interface{})

			t.Logf("✅ Successfully retrieved %d current flyers via real API", len(edges))
			t.Logf("  First flyer: %s", node["title"])
			t.Logf("  Days remaining: %.0f", node["daysRemaining"])
			t.Logf("  Validity period: %s", node["validityPeriod"])
		}
	})

	t.Run("Can query flyer with store relation via real API", func(t *testing.T) {
		// First, get a valid flyer ID
		listQuery := `
			query GetFlyers {
				flyers(first: 1) {
					edges {
						node {
							id
						}
					}
				}
			}
		`

		listResp, _, err := testCtx.Client.QueryWithMeasurement(listQuery, nil)
		if err != nil {
			t.Fatalf("Failed to get flyer list: %v", err)
		}

		if len(listResp.Errors) > 0 {
			t.Fatalf("Failed to get flyer list: %v", listResp.Errors)
		}

		listData := listResp.Data.(map[string]interface{})
		flyersData := listData["flyers"].(map[string]interface{})
		edges := flyersData["edges"].([]interface{})

		if len(edges) == 0 {
			t.Skip("No flyers in database to test")
		}

		firstEdge := edges[0].(map[string]interface{})
		node := firstEdge["node"].(map[string]interface{})
		flyerID := int(node["id"].(float64))

		// Now query that flyer with store relation
		query := `
			query GetFlyerWithStore($id: Int!) {
				flyer(id: $id) {
					id
					title
					validFrom
					validTo
					status
					store {
						id
						code
						name
					}
				}
			}
		`

		variables := map[string]interface{}{
			"id": flyerID,
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(query, variables)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			for i, gqlErr := range resp.Errors {
				t.Logf("Error %d: %s", i+1, gqlErr.Message)
			}
			t.Fatalf("GraphQL errors encountered")
		}

		if resp.Data == nil {
			t.Fatal("No data returned")
		}

		dataMap := resp.Data.(map[string]interface{})
		flyerData := dataMap["flyer"]

		if flyerData == nil {
			t.Fatalf("Flyer with ID %d not found", flyerID)
		}

		flyer := flyerData.(map[string]interface{})
		store := flyer["store"].(map[string]interface{})
		t.Logf("✅ Successfully retrieved flyer '%s' from store '%s' via real API",
			flyer["title"], store["name"])
	})
}
