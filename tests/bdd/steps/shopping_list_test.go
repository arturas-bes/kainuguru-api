//go:build integration
// +build integration

package steps

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/tests/bdd/helpers"
)

// TestShoppingListAPIIntegration verifies shopping list operations hit real GraphQL endpoints
func TestShoppingListAPIIntegration(t *testing.T) {
	apiURL := os.Getenv("API_URL")
	if apiURL == ""  {
		apiURL = "http://localhost:8080"
	}

	testCtx := helpers.NewTestContext(apiURL)

	// Helper to register and get auth token
	setupAuthUser := func(t *testing.T) string {
		email := fmt.Sprintf("shoplist_%d@example.com", time.Now().UnixNano())
		password := "SecurePass123!"

		registerMutation := `
			mutation Register($input: RegisterInput!) {
				register(input: $input) {
					accessToken
				}
			}
		`

		registerVars := map[string]interface{}{
			"input": map[string]interface{}{
				"email":    email,
				"password": password,
				"fullName": "Shopping List Test User",
			},
		}

		resp, _, err := testCtx.Client.QueryWithMeasurement(registerMutation, registerVars)
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		if len(resp.Errors) > 0 {
			t.Fatalf("Registration failed: %v", resp.Errors)
		}

		if resp.Data == nil {
			t.Fatalf("No response data received from registration")
		}

		dataMap := resp.Data.(map[string]interface{})
		registerData := dataMap["register"].(map[string]interface{})
		return registerData["accessToken"].(string)
	}

	t.Run("Can create shopping list via real API", func(t *testing.T) {
		token := setupAuthUser(t)
		testCtx.Client.SetAuthToken(token)

		mutation := `
			mutation CreateShoppingList($input: CreateShoppingListInput!) {
				createShoppingList(input: $input) {
					id
					name
					description
					isDefault
					isArchived
					createdAt
				}
			}
		`

		listName := "My Test List"
		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"name":        listName,
				"description": "Test shopping list",
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(mutation, variables)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify response
		dataMap := resp.Data.(map[string]interface{})
		listData := dataMap["createShoppingList"].(map[string]interface{})

		name := listData["name"].(string)
		if name != listName {
			t.Fatalf("Expected name '%s', got '%s'", listName, name)
		}

		t.Logf("✅ Successfully created shopping list '%s' via real API", name)
	})

	t.Run("Can query user shopping lists via real API", func(t *testing.T) {
		token := setupAuthUser(t)
		testCtx.Client.SetAuthToken(token)

		// First create a list
		createMutation := `
			mutation CreateShoppingList($input: CreateShoppingListInput!) {
				createShoppingList(input: $input) {
					id
				}
			}
		`

		createVars := map[string]interface{}{
			"input": map[string]interface{}{
				"name": "List to Query",
			},
		}

		testCtx.Client.QueryWithMeasurement(createMutation, createVars)

		// Now query lists
		query := `
			query GetShoppingLists {
				shoppingLists {
					edges {
						node {
							id
							name
							isDefault
							isArchived
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
		dataMap := resp.Data.(map[string]interface{})
		listsData := dataMap["shoppingLists"].(map[string]interface{})
		edges := listsData["edges"].([]interface{})

		if len(edges) == 0 {
			t.Fatal("Expected at least one shopping list")
		}

		t.Logf("✅ Successfully retrieved %d shopping lists via real API", len(edges))
	})

	t.Run("Can update shopping list via real API", func(t *testing.T) {
		token := setupAuthUser(t)
		testCtx.Client.SetAuthToken(token)

		// Create a list first
		createMutation := `
			mutation CreateShoppingList($input: CreateShoppingListInput!) {
				createShoppingList(input: $input) {
					id
					name
				}
			}
		`

		createVars := map[string]interface{}{
			"input": map[string]interface{}{
				"name": "Original Name",
			},
		}

		createResp, _, _ := testCtx.Client.QueryWithMeasurement(createMutation, createVars)
		if createResp.Data == nil {
			t.Fatalf("No data from create mutation")
		}
		createDataMap := createResp.Data.(map[string]interface{})
		listData := createDataMap["createShoppingList"].(map[string]interface{})
		listID := int(listData["id"].(float64))

		// Update the list
		updateMutation := `
			mutation UpdateShoppingList($id: Int!, $input: UpdateShoppingListInput!) {
				updateShoppingList(id: $id, input: $input) {
					id
					name
					description
				}
			}
		`

		newName := "Updated Name"
		updateVars := map[string]interface{}{
			"id": listID,
			"input": map[string]interface{}{
				"name": newName,
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(updateMutation, updateVars)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify update
		dataMap := resp.Data.(map[string]interface{})
		updatedList := dataMap["updateShoppingList"].(map[string]interface{})
		updatedName := updatedList["name"].(string)

		if updatedName != newName {
			t.Fatalf("Expected name '%s', got '%s'", newName, updatedName)
		}

		t.Logf("✅ Successfully updated shopping list name to '%s' via real API", updatedName)
	})

	t.Run("Can delete shopping list via real API", func(t *testing.T) {
		token := setupAuthUser(t)
		testCtx.Client.SetAuthToken(token)

		// Create a list first
		createMutation := `
			mutation CreateShoppingList($input: CreateShoppingListInput!) {
				createShoppingList(input: $input) {
					id
				}
			}
		`

		createVars := map[string]interface{}{
			"input": map[string]interface{}{
				"name": "List to Delete",
			},
		}

		createResp, _, _ := testCtx.Client.QueryWithMeasurement(createMutation, createVars)
		if createResp.Data == nil {
			t.Fatalf("No data from create mutation")
		}
		createDataMap := createResp.Data.(map[string]interface{})
		listData := createDataMap["createShoppingList"].(map[string]interface{})
		listID := int(listData["id"].(float64))

		// Delete the list
		deleteMutation := `
			mutation DeleteShoppingList($id: Int!) {
				deleteShoppingList(id: $id)
			}
		`

		deleteVars := map[string]interface{}{
			"id": listID,
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(deleteMutation, deleteVars)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		t.Logf("Response time: %v", duration)

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Verify deletion returned true
		if resp.Data == nil {
			t.Fatalf("No data from delete mutation")
		}
		dataMap := resp.Data.(map[string]interface{})
		success := dataMap["deleteShoppingList"].(bool)

		if !success {
			t.Fatal("Expected deletion to return true")
		}

		t.Logf("✅ Successfully deleted shopping list via real API")
	})

	t.Run("Cannot access other user's shopping list via real API", func(t *testing.T) {
		// User 1 creates a list
		token1 := setupAuthUser(t)
		testCtx.Client.SetAuthToken(token1)

		createMutation := `
			mutation CreateShoppingList($input: CreateShoppingListInput!) {
				createShoppingList(input: $input) {
					id
				}
			}
		`

		createVars := map[string]interface{}{
			"input": map[string]interface{}{
				"name": "User 1 List",
			},
		}

		createResp, _, _ := testCtx.Client.QueryWithMeasurement(createMutation, createVars)
		if createResp.Data == nil {
			t.Fatalf("No data from create mutation")
		}
		createDataMap := createResp.Data.(map[string]interface{})
		listData := createDataMap["createShoppingList"].(map[string]interface{})
		listID := int(listData["id"].(float64))

		// User 2 tries to access it
		token2 := setupAuthUser(t)
		testCtx.Client.SetAuthToken(token2)

		query := `
			query GetShoppingList($id: Int!) {
				shoppingList(id: $id) {
					id
					name
				}
			}
		`

		queryVars := map[string]interface{}{
			"id": listID,
		}

		resp, _, err := testCtx.Client.QueryWithMeasurement(query, queryVars)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		// Should have access denied error
		if len(resp.Errors) == 0 {
			t.Fatal("Expected access denied error, got none")
		}

		errorMsg := resp.Errors[0].Message
		t.Logf("Error message: %s", errorMsg)

		t.Logf("✅ Access correctly denied for other user's shopping list")
	})

	t.Run("Performance: Create shopping list should complete under 2 seconds", func(t *testing.T) {
		token := setupAuthUser(t)
		testCtx.Client.SetAuthToken(token)

		mutation := `
			mutation CreateShoppingList($input: CreateShoppingListInput!) {
				createShoppingList(input: $input) {
					id
					name
				}
			}
		`

		variables := map[string]interface{}{
			"input": map[string]interface{}{
				"name": "Performance Test List",
			},
		}

		resp, duration, err := testCtx.Client.QueryWithMeasurement(mutation, variables)
		if err != nil {
			t.Fatalf("Failed to call API: %v", err)
		}

		if len(resp.Errors) > 0 {
			t.Fatalf("GraphQL errors: %v", resp.Errors)
		}

		// Check performance requirement
		if duration > 2*time.Second {
			t.Errorf("Create shopping list took %v, exceeds 2 second threshold", duration)
		}

		t.Logf("✅ Create shopping list performance: %v (under 2s threshold)", duration)
	})
}
