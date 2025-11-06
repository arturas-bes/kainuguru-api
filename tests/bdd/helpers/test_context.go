package helpers

import (
	"time"
)

// TestContext holds shared state between BDD test steps
type TestContext struct {
	// GraphQL client
	Client *GraphQLClient

	// Response state
	LastResponse      *GraphQLResponse
	LastResponseTime  time.Duration
	LastError         error

	// Authentication state
	AuthToken         string
	CurrentUser       map[string]interface{}

	// Test data state
	TestProducts      map[string]int // product name -> product_master_id
	TestStores        map[string]int // store name -> store_id
	TestPriceHistory  []map[string]interface{}

	// Assertion state
	AssertionErrors   []string
}

// NewTestContext creates a new test context with GraphQL client
func NewTestContext(apiURL string) *TestContext {
	return &TestContext{
		Client:           NewGraphQLClient(apiURL),
		TestProducts:     make(map[string]int),
		TestStores:       make(map[string]int),
		TestPriceHistory: make([]map[string]interface{}, 0),
		AssertionErrors:  make([]string, 0),
	}
}

// Reset clears the context state for a new scenario
func (ctx *TestContext) Reset() {
	ctx.LastResponse = nil
	ctx.LastResponseTime = 0
	ctx.LastError = nil
	ctx.TestProducts = make(map[string]int)
	ctx.TestStores = make(map[string]int)
	ctx.TestPriceHistory = make([]map[string]interface{}, 0)
	ctx.AssertionErrors = make([]string, 0)
}

// AddAssertionError records an assertion failure
func (ctx *TestContext) AddAssertionError(message string) {
	ctx.AssertionErrors = append(ctx.AssertionErrors, message)
}

// HasAssertionErrors returns true if any assertions failed
func (ctx *TestContext) HasAssertionErrors() bool {
	return len(ctx.AssertionErrors) > 0
}
