package steps

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/tests/bdd/helpers"
)

// TestEndpointDeepValidation performs deep validation of all GraphQL endpoints
func TestEndpointDeepValidation(t *testing.T) {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	testCtx := helpers.NewTestContext(apiURL)

	// Helper to pretty print responses
	printResponse := func(name string, data interface{}) {
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		t.Logf("\n=== %s Response ===\n%s\n", name, string(jsonData))
	}

	// Helper to register and get auth token
	getAuthToken := func(t *testing.T) string {
		email := fmt.Sprintf("validate_%d@example.com", time.Now().UnixNano())
		mutation := `
			mutation Register($input: RegisterInput!) {
				register(input: $input) {
					user {
						id
						email
						fullName
						isActive
						emailVerified
						createdAt
					}
					accessToken
					refreshToken
					expiresAt
					tokenType
				}
			}
		`

		vars := map[string]interface{}{
			"input": map[string]interface{}{
				"email":    email,
				"password": "SecurePass123!",
				"fullName": "Validation Test User",
			},
		}

		resp, _, err := testCtx.Client.QueryWithMeasurement(mutation, vars)
		if err != nil || len(resp.Errors) > 0 || resp.Data == nil {
			t.Fatalf("Failed to get auth token: err=%v, errors=%v", err, resp.Errors)
		}

		dataMap := resp.Data.(map[string]interface{})
		registerData := dataMap["register"].(map[string]interface{})
		return registerData["accessToken"].(string)
	}

	t.Run("Validate Authentication Endpoints", func(t *testing.T) {
		t.Run("Register - Full Response Structure", func(t *testing.T) {
			email := fmt.Sprintf("register_test_%d@example.com", time.Now().UnixNano())

			mutation := `
				mutation Register($input: RegisterInput!) {
					register(input: $input) {
						user {
							id
							email
							fullName
							isActive
							emailVerified
							createdAt
							updatedAt
						}
						accessToken
						refreshToken
						expiresAt
						tokenType
					}
				}
			`

			vars := map[string]interface{}{
				"input": map[string]interface{}{
					"email":    email,
					"password": "SecurePass123!",
					"fullName": "Register Test User",
				},
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(mutation, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Register", resp.Data)

			// Deep validation
			dataMap := resp.Data.(map[string]interface{})
			registerData := dataMap["register"].(map[string]interface{})

			// Validate user object
			user := registerData["user"].(map[string]interface{})

			if user["email"].(string) != email {
				t.Errorf("Email mismatch: expected %s, got %s", email, user["email"])
			}

			if user["fullName"].(string) != "Register Test User" {
				t.Errorf("Full name mismatch: expected 'Register Test User', got %s", user["fullName"])
			}

			if !user["isActive"].(bool) {
				t.Error("User should be active after registration")
			}

			if user["emailVerified"].(bool) {
				t.Error("Email should not be verified immediately after registration")
			}

			// Validate tokens
			accessToken := registerData["accessToken"].(string)
			if len(accessToken) == 0 {
				t.Error("Access token is empty")
			}

			refreshToken := registerData["refreshToken"].(string)
			if len(refreshToken) == 0 {
				t.Error("Refresh token is empty")
			}

			tokenType := registerData["tokenType"].(string)
			if tokenType != "Bearer" {
				t.Errorf("Token type should be 'Bearer', got '%s'", tokenType)
			}

			// Validate expiresAt is in future
			expiresAt := registerData["expiresAt"].(string)
			expiresTime, err := time.Parse(time.RFC3339, expiresAt)
			if err != nil {
				t.Errorf("Failed to parse expiresAt: %v", err)
			}
			if !expiresTime.After(time.Now()) {
				t.Error("Token expiration should be in the future")
			}

			t.Logf("✅ Register endpoint validated successfully")
		})

		t.Run("Login - Full Response Structure", func(t *testing.T) {
			// First register a user
			email := fmt.Sprintf("login_test_%d@example.com", time.Now().UnixNano())
			password := "SecurePass123!"

			registerMutation := `
				mutation Register($input: RegisterInput!) {
					register(input: $input) {
						user { id }
					}
				}
			`

			registerVars := map[string]interface{}{
				"input": map[string]interface{}{
					"email":    email,
					"password": password,
					"fullName": "Login Test User",
				},
			}

			testCtx.Client.QueryWithMeasurement(registerMutation, registerVars)

			// Now test login
			loginMutation := `
				mutation Login($input: LoginInput!) {
					login(input: $input) {
						user {
							id
							email
							fullName
							isActive
							emailVerified
						}
						accessToken
						refreshToken
						expiresAt
						tokenType
					}
				}
			`

			loginVars := map[string]interface{}{
				"input": map[string]interface{}{
					"email":    email,
					"password": password,
				},
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(loginMutation, loginVars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Login", resp.Data)

			// Validate response structure
			dataMap := resp.Data.(map[string]interface{})
			loginData := dataMap["login"].(map[string]interface{})

			user := loginData["user"].(map[string]interface{})
			if user["email"].(string) != email {
				t.Errorf("Email mismatch after login")
			}

			if !user["isActive"].(bool) {
				t.Error("User should be active after login")
			}

			t.Logf("✅ Login endpoint validated successfully")
		})

		t.Run("Me - Protected Query", func(t *testing.T) {
			token := getAuthToken(t)
			testCtx.Client.SetAuthToken(token)

			query := `
				query {
					me {
						id
						email
						fullName
						isActive
						emailVerified
						createdAt
						updatedAt
					}
				}
			`

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Me", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			meData := dataMap["me"].(map[string]interface{})

			// Validate all fields are present and have correct types
			if _, ok := meData["id"].(string); !ok {
				t.Error("id should be a string")
			}

			if _, ok := meData["email"].(string); !ok {
				t.Error("email should be a string")
			}

			if _, ok := meData["fullName"].(string); !ok {
				t.Error("fullName should be a string")
			}

			if _, ok := meData["isActive"].(bool); !ok {
				t.Error("isActive should be a boolean")
			}

			if _, ok := meData["emailVerified"].(bool); !ok {
				t.Error("emailVerified should be a boolean")
			}

			t.Logf("✅ Me query validated successfully")
		})
	})

	t.Run("Validate Store Endpoints", func(t *testing.T) {
		t.Run("Stores - List with Pagination", func(t *testing.T) {
			query := `
				query GetStores($first: Int) {
					stores(first: $first) {
						edges {
							node {
								id
								code
								name
								description
								logoURL
								websiteURL
								isActive
								createdAt
								updatedAt
							}
							cursor
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
						totalCount
					}
				}
			`

			vars := map[string]interface{}{
				"first": 5,
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Stores", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			storesData := dataMap["stores"].(map[string]interface{})
			edges := storesData["edges"].([]interface{})

			if len(edges) == 0 {
				t.Error("Expected at least one store")
			}

			// Validate first store structure
			firstEdge := edges[0].(map[string]interface{})
			node := firstEdge["node"].(map[string]interface{})

			// Validate all required fields
			requiredFields := []string{"id", "code", "name", "isActive"}
			for _, field := range requiredFields {
				if _, ok := node[field]; !ok {
					t.Errorf("Store missing required field: %s", field)
				}
			}

			// Validate cursor
			if cursor, ok := firstEdge["cursor"].(string); !ok || cursor == "" {
				t.Error("Cursor should be a non-empty string")
			}

			// Validate pageInfo
			pageInfo := storesData["pageInfo"].(map[string]interface{})
			if _, ok := pageInfo["hasNextPage"].(bool); !ok {
				t.Error("pageInfo.hasNextPage should be a boolean")
			}

			t.Logf("✅ Stores query validated successfully - returned %d stores", len(edges))
		})

		t.Run("Store - Single by ID", func(t *testing.T) {
			query := `
				query GetStore($id: Int!) {
					store(id: $id) {
						id
						code
						name
						description
						logoURL
						websiteURL
						isActive
					}
				}
			`

			vars := map[string]interface{}{
				"id": 1,
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Store", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			store := dataMap["store"].(map[string]interface{})

			// Validate ID matches request
			storeID := int(store["id"].(float64))
			if storeID != 1 {
				t.Errorf("Expected store ID 1, got %d", storeID)
			}

			// Validate code is present
			if code, ok := store["code"].(string); !ok || code == "" {
				t.Error("Store code should be a non-empty string")
			}

			t.Logf("✅ Store query validated successfully")
		})
	})

	t.Run("Validate Product Endpoints", func(t *testing.T) {
		t.Run("Products - List with Full Details", func(t *testing.T) {
			query := `
				query GetProducts($first: Int) {
					products(first: $first) {
						edges {
							node {
								id
								name
								description
								sku
								slug
								price {
									current
									original
									currency
									discount
									discountPercent
									discountAmount
									isDiscounted
								}
								isOnSale
								isCurrentlyOnSale
								validFrom
								validTo
								isValid
								isExpired
								validityPeriod
								category
								brand
								imageURL
								thumbnailURL
							}
						}
						pageInfo {
							hasNextPage
							endCursor
						}
						totalCount
					}
				}
			`

			vars := map[string]interface{}{
				"first": 3,
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Products", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			productsData := dataMap["products"].(map[string]interface{})
			edges := productsData["edges"].([]interface{})

			if len(edges) == 0 {
				t.Error("Expected at least one product")
			}

			// Deep validate first product
			firstEdge := edges[0].(map[string]interface{})
			product := firstEdge["node"].(map[string]interface{})

			// Validate price object structure
			price := product["price"].(map[string]interface{})

			if current, ok := price["current"].(float64); !ok || current <= 0 {
				t.Error("Price current should be a positive number")
			}

			if currency, ok := price["currency"].(string); !ok || currency == "" {
				t.Error("Currency should be a non-empty string")
			}

			if _, ok := price["isDiscounted"].(bool); !ok {
				t.Error("isDiscounted should be a boolean")
			}

			// Validate computed fields
			if _, ok := product["isCurrentlyOnSale"].(bool); !ok {
				t.Error("isCurrentlyOnSale should be a boolean")
			}

			if _, ok := product["isValid"].(bool); !ok {
				t.Error("isValid should be a boolean")
			}

			if _, ok := product["isExpired"].(bool); !ok {
				t.Error("isExpired should be a boolean")
			}

			// Validate string fields
			if validityPeriod, ok := product["validityPeriod"].(string); !ok || validityPeriod == "" {
				t.Error("validityPeriod should be a non-empty string")
			}

			t.Logf("✅ Products query validated successfully - returned %d products", len(edges))
		})

		t.Run("Product - Nested Relations", func(t *testing.T) {
			query := `
				query GetProductWithRelations($id: Int!) {
					product(id: $id) {
						id
						name
						store {
							id
							code
							name
						}
						flyer {
							id
							title
							validFrom
							validTo
						}
						productMaster {
							id
							name
							category
							brand
						}
					}
				}
			`

			vars := map[string]interface{}{
				"id": 1,
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Product with Relations", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			product := dataMap["product"].(map[string]interface{})

			// Validate store relation
			if store, ok := product["store"].(map[string]interface{}); ok {
				if _, ok := store["id"].(float64); !ok {
					t.Error("Store ID should be a number")
				}
				if _, ok := store["code"].(string); !ok {
					t.Error("Store code should be a string")
				}
			} else {
				t.Error("Store relation should be present")
			}

			// Validate flyer relation
			if flyer, ok := product["flyer"].(map[string]interface{}); ok {
				if _, ok := flyer["id"].(float64); !ok {
					t.Error("Flyer ID should be a number")
				}
			} else {
				t.Error("Flyer relation should be present")
			}

			// productMaster is optional, but if present should be valid
			if productMaster := product["productMaster"]; productMaster != nil {
				pm := productMaster.(map[string]interface{})
				if _, ok := pm["id"].(float64); !ok {
					t.Error("ProductMaster ID should be a number")
				}
			}

			t.Logf("✅ Product with relations validated successfully")
		})
	})

	t.Run("Validate Price History Endpoints", func(t *testing.T) {
		t.Run("PriceHistory - Full Connection Structure", func(t *testing.T) {
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
								storeID
								productMasterID
							}
							cursor
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
						totalCount
					}
				}
			`

			vars := map[string]interface{}{
				"productMasterID": 1,
				"first":           5,
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Price History", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			priceHistoryData := dataMap["priceHistory"].(map[string]interface{})
			edges := priceHistoryData["edges"].([]interface{})

			if len(edges) == 0 {
				t.Error("Expected at least one price history entry")
			}

			// Validate first entry
			firstEdge := edges[0].(map[string]interface{})
			node := firstEdge["node"].(map[string]interface{})

			// Validate price is positive
			if price, ok := node["price"].(float64); !ok || price <= 0 {
				t.Error("Price should be a positive number")
			}

			// Validate currency
			if currency, ok := node["currency"].(string); !ok || currency == "" {
				t.Error("Currency should be a non-empty string")
			}

			// Validate recordedAt is valid timestamp
			if recordedAt, ok := node["recordedAt"].(string); ok {
				if _, err := time.Parse(time.RFC3339, recordedAt); err != nil {
					t.Errorf("recordedAt should be valid RFC3339 timestamp: %v", err)
				}
			} else {
				t.Error("recordedAt should be a string")
			}

			// Validate IDs
			if productMasterID, ok := node["productMasterID"].(float64); !ok || productMasterID != 1 {
				t.Error("productMasterID should match requested ID")
			}

			// Validate totalCount
			if totalCount, ok := priceHistoryData["totalCount"].(float64); !ok || totalCount <= 0 {
				t.Error("totalCount should be a positive number")
			}

			t.Logf("✅ Price history validated successfully - returned %d entries", len(edges))
		})

		t.Run("CurrentPrice - Latest Price Entry", func(t *testing.T) {
			query := `
				query GetCurrentPrice($productMasterID: Int!, $storeID: Int) {
					currentPrice(productMasterID: $productMasterID, storeID: $storeID) {
						id
						price
						currency
						isOnSale
						recordedAt
						storeID
						productMasterID
					}
				}
			`

			vars := map[string]interface{}{
				"productMasterID": 1,
				"storeID":         1,
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Current Price", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			currentPrice := dataMap["currentPrice"].(map[string]interface{})

			// Validate it returns the most recent entry
			if recordedAt, ok := currentPrice["recordedAt"].(string); ok {
				recordedTime, err := time.Parse(time.RFC3339, recordedAt)
				if err != nil {
					t.Errorf("Invalid recordedAt timestamp: %v", err)
				}

				// Should be recent (within last year for test data)
				if time.Since(recordedTime) > 365*24*time.Hour {
					t.Error("Current price recordedAt seems too old")
				}
			}

			t.Logf("✅ Current price validated successfully")
		})
	})

	t.Run("Validate Search Endpoint", func(t *testing.T) {
		t.Run("SearchProducts - Full Result Structure", func(t *testing.T) {
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
								isOnSale
							}
							searchScore
							matchType
						}
						suggestions
						hasMore
						facets {
							stores {
								name
								options {
									label
									value
									count
								}
							}
						}
						pagination {
							totalItems
							currentPage
							totalPages
							itemsPerPage
						}
					}
				}
			`

			vars := map[string]interface{}{
				"input": map[string]interface{}{
					"q":     "pienas",
					"first": 5,
				},
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Search Products", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			searchData := dataMap["searchProducts"].(map[string]interface{})

			// Validate query string echoed back
			if queryString, ok := searchData["queryString"].(string); !ok || queryString != "pienas" {
				t.Error("queryString should match input query")
			}

			// Validate totalCount
			if _, ok := searchData["totalCount"].(float64); !ok {
				t.Error("totalCount should be a number")
			}

			// Validate products array
			products := searchData["products"].([]interface{})
			if len(products) > 0 {
				firstResult := products[0].(map[string]interface{})

				// Validate search score
				if searchScore, ok := firstResult["searchScore"].(float64); !ok || searchScore < 0 {
					t.Error("searchScore should be a non-negative number")
				}

				// Validate match type
				if matchType, ok := firstResult["matchType"].(string); !ok || matchType == "" {
					t.Error("matchType should be a non-empty string")
				}
			}

			// Validate pagination
			pagination := searchData["pagination"].(map[string]interface{})
			if totalItems, ok := pagination["totalItems"].(float64); !ok || totalItems < 0 {
				t.Error("pagination.totalItems should be a non-negative number")
			}

			t.Logf("✅ Search products validated successfully")
		})
	})

	t.Run("Validate Shopping List Endpoints", func(t *testing.T) {
		token := getAuthToken(t)
		testCtx.Client.SetAuthToken(token)

		var listID int

		t.Run("CreateShoppingList - Full Response", func(t *testing.T) {
			mutation := `
				mutation CreateShoppingList($input: CreateShoppingListInput!) {
					createShoppingList(input: $input) {
						id
						name
						description
						isDefault
						isArchived
						isPublic
						createdAt
						updatedAt
						user {
							id
							email
						}
					}
				}
			`

			vars := map[string]interface{}{
				"input": map[string]interface{}{
					"name":        "Validation Test List",
					"description": "Testing full response structure",
				},
			}

			resp, duration, err := testCtx.Client.QueryWithMeasurement(mutation, vars)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Create Shopping List", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			list := dataMap["createShoppingList"].(map[string]interface{})

			// Store list ID for later tests
			listID = int(list["id"].(float64))

			// Validate name matches input
			if list["name"].(string) != "Validation Test List" {
				t.Error("Name should match input")
			}

			// Validate defaults
			if list["isDefault"].(bool) {
				t.Error("New list should not be default by default")
			}

			if list["isArchived"].(bool) {
				t.Error("New list should not be archived")
			}

			// Validate user relation
			user := list["user"].(map[string]interface{})
			if _, ok := user["id"].(string); !ok {
				t.Error("User ID should be present")
			}

			// Validate timestamps
			if createdAt, ok := list["createdAt"].(string); ok {
				if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
					t.Errorf("createdAt should be valid timestamp: %v", err)
				}
			}

			t.Logf("✅ Create shopping list validated - ID: %d", listID)
		})

		t.Run("ShoppingLists - Query All Lists", func(t *testing.T) {
			query := `
				query GetShoppingLists {
					shoppingLists {
						edges {
							node {
								id
								name
								description
								isDefault
								isArchived
								itemCount
								completedItemCount
								completionPercentage
								isCompleted
								canBeShared
							}
							cursor
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
						}
					}
				}
			`

			resp, duration, err := testCtx.Client.QueryWithMeasurement(query, nil)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			t.Logf("Response time: %v", duration)

			if len(resp.Errors) > 0 {
				t.Fatalf("GraphQL errors: %v", resp.Errors)
			}

			printResponse("Shopping Lists", resp.Data)

			dataMap := resp.Data.(map[string]interface{})
			listsData := dataMap["shoppingLists"].(map[string]interface{})
			edges := listsData["edges"].([]interface{})

			if len(edges) == 0 {
				t.Error("Expected at least one list (just created one)")
			}

			// Validate computed fields
			firstEdge := edges[0].(map[string]interface{})
			node := firstEdge["node"].(map[string]interface{})

			if _, ok := node["itemCount"].(float64); !ok {
				t.Error("itemCount should be a number")
			}

			if _, ok := node["completionPercentage"].(float64); !ok {
				t.Error("completionPercentage should be a number")
			}

			if _, ok := node["isCompleted"].(bool); !ok {
				t.Error("isCompleted should be a boolean")
			}

			if _, ok := node["canBeShared"].(bool); !ok {
				t.Error("canBeShared should be a boolean")
			}

			t.Logf("✅ Shopping lists query validated - returned %d lists", len(edges))
		})
	})
}
