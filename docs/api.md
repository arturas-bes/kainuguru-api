# 🔌 Kainuguru API Documentation

Complete GraphQL API reference for the Kainuguru grocery shopping platform.

## 📋 Table of Contents

- [Authentication](#authentication)
- [Stores & Flyers](#stores--flyers)
- [Products & Search](#products--search)
- [Shopping Lists](#shopping-lists)
- [Price History & Analytics](#price-history--analytics)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Examples](#examples)

## 🔐 Authentication

### Registration

Register a new user account.

```graphql
mutation Register {
  register(input: {
    email: "user@example.com"
    password: "securePassword123"
    fullName: "Jonas Jonaitis"
    preferredLanguage: "lt"
  }) {
    user {
      id
      email
      fullName
      emailVerified
      createdAt
    }
    accessToken
    refreshToken
    expiresAt
    tokenType
  }
}
```

**Input Parameters:**
- `email` (String!): Valid email address
- `password` (String!): Minimum 8 characters
- `fullName` (String): Optional full name
- `preferredLanguage` (String): Default "lt"

### Login

Authenticate existing user.

```graphql
mutation Login {
  login(input: {
    email: "user@example.com"
    password: "securePassword123"
  }) {
    user {
      id
      email
      fullName
      lastLoginAt
    }
    accessToken
    refreshToken
    expiresAt
  }
}
```

### Token Refresh

Refresh access token using refresh token.

```graphql
mutation RefreshToken {
  refreshToken {
    user {
      id
      email
    }
    accessToken
    refreshToken
    expiresAt
  }
}
```

### Password Management

#### Request Password Reset
```graphql
mutation RequestPasswordReset {
  requestPasswordReset(input: {
    email: "user@example.com"
  })
}
```

#### Reset Password
```graphql
mutation ResetPassword {
  resetPassword(input: {
    token: "reset-token-from-email"
    newPassword: "newSecurePassword123"
  })
}
```

### Email Verification

#### Verify Email
```graphql
mutation VerifyEmail {
  verifyEmail(token: "verification-token-from-email")
}
```

#### Resend Verification
```graphql
mutation ResendVerification {
  resendEmailVerification
}
```

## 🏪 Stores & Flyers

### Get All Stores

```graphql
query GetStores {
  stores(first: 50) {
    edges {
      node {
        id
        code
        name
        logoURL
        websiteURL
        locations {
          city
          address
          lat
          lng
        }
        isActive
        lastScrapedAt
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

### Get Store by ID

```graphql
query GetStore {
  store(id: 1) {
    id
    name
    code
    logoURL
    flyers(first: 10) {
      edges {
        node {
          id
          title
          validFrom
          validTo
          isValid
          isCurrent
          pageCount
          productsExtracted
        }
      }
    }
  }
}
```

### Current Flyers

Get currently valid flyers across stores.

```graphql
query CurrentFlyers {
  currentFlyers(
    storeIDs: [1, 2, 3]
    first: 20
  ) {
    edges {
      node {
        id
        title
        validFrom
        validTo
        daysRemaining
        store {
          name
          code
        }
        products(first: 5) {
          edges {
            node {
              name
              currentPrice
              isOnSale
            }
          }
        }
      }
    }
  }
}
```

## 🔍 Products & Search

### Basic Product Search

```graphql
query SearchProducts {
  products(
    filters: {
      query: "pienas"
      storeIDs: [1, 2]
      minPrice: 0.5
      maxPrice: 5.0
      onSaleOnly: true
    }
    first: 20
  ) {
    edges {
      node {
        id
        name
        brand
        category
        currentPrice
        originalPrice
        discountPercent
        isOnSale
        unitSize
        unitType
        imageURL
        store {
          name
          code
        }
        isCurrentlyOnSale
        discountAmount
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

### Advanced Search

```graphql
query AdvancedSearch {
  advancedSearch(input: {
    query: "duona pilngrūdžių"
    storeIDs: [1, 2, 3]
    minPrice: 1.0
    maxPrice: 10.0
    onSaleOnly: false
    category: "Duona ir kepiniai"
    limit: 50
    offset: 0
    preferFuzzy: true
  }) {
    products {
      id
      name
      brand
      currentPrice
      isOnSale
      relevanceScore
      store {
        name
      }
    }
    totalCount
    searchTime
    suggestions
  }
}
```

### Fuzzy Search

For typo-tolerant search.

```graphql
query FuzzySearch {
  fuzzySearch(input: {
    query: "peinas" # typo for "pienas"
    storeIDs: [1, 2]
    limit: 20
    similarityThreshold: 0.3
  }) {
    products {
      id
      name
      currentPrice
      similarityScore
    }
    corrections {
      original
      suggested
      confidence
    }
  }
}
```

### Search Suggestions

Get autocomplete suggestions.

```graphql
query SearchSuggestions {
  searchSuggestions(input: {
    partialQuery: "pien"
    limit: 10
  }) {
    suggestions {
      text
      type
      category
      frequency
    }
  }
}
```

### Similar Products

```graphql
query SimilarProducts {
  similarProducts(input: {
    productID: 123
    limit: 10
  }) {
    products {
      id
      name
      currentPrice
      similarityScore
      store {
        name
      }
    }
  }
}
```

## 🛍️ Shopping Lists

### Create Shopping List

```graphql
mutation CreateShoppingList {
  createShoppingList(input: {
    name: "Weekly Shopping"
    description: "Groceries for the week"
    colorHex: "#FF6B6B"
    iconName: "shopping-cart"
  }) {
    id
    name
    description
    itemCount
    totalEstimatedPrice
    isDefault
    createdAt
  }
}
```

### Get Shopping Lists

```graphql
query GetShoppingLists {
  shoppingLists(
    filters: {
      includeArchived: false
    }
    first: 20
  ) {
    edges {
      node {
        id
        name
        description
        itemCount
        completedItemCount
        totalEstimatedPrice
        isDefault
        isShared
        shareCode
        lastModifiedAt
      }
    }
  }
}
```

### Add Items to List

```graphql
mutation AddShoppingListItem {
  createShoppingListItem(input: {
    listID: 1
    description: "Duona pilngrūdžių"
    quantity: 2
    unit: "vnt"
    categoryID: 5
    notes: "Preferably dark bread"
  }) {
    id
    description
    quantity
    unit
    isCompleted
    estimatedPrice
    matchedProduct {
      id
      name
      currentPrice
      store {
        name
      }
    }
    matchingConfidence
  }
}
```

### Update Shopping List Item

```graphql
mutation UpdateShoppingListItem {
  updateShoppingListItem(
    id: 123
    input: {
      description: "Duona juoda pilngrūdžių"
      quantity: 3
      isCompleted: true
      notes: "Bought at Maxima"
    }
  ) {
    id
    description
    quantity
    isCompleted
    completedAt
    actualPrice
  }
}
```

### Smart Item Matching

```graphql
mutation AutoMatchItem {
  autoMatchItem(itemID: 123) {
    item {
      id
      description
      matchedProduct {
        name
        currentPrice
      }
    }
    confidence
    alternatives {
      product {
        name
        currentPrice
      }
      confidence
    }
    success
  }
}
```

### List Sharing

```graphql
mutation GenerateShareCode {
  generateShareCode(id: 1) {
    id
    shareCode
    isShared
    sharedAt
  }
}

query GetSharedList {
  sharedShoppingList(shareCode: "ABC123DEF") {
    id
    name
    items {
      description
      quantity
      isCompleted
    }
    owner {
      fullName
    }
    isReadOnly
  }
}
```

## 📊 Price History & Analytics

### Price History

```graphql
query PriceHistory {
  priceHistory(
    productID: 123
    storeID: 1
    filters: {
      startDate: "2024-01-01T00:00:00Z"
      endDate: "2024-12-31T23:59:59Z"
      isOnSale: false
    }
    first: 100
  ) {
    edges {
      node {
        id
        price
        originalPrice
        isOnSale
        recordedAt
        validFrom
        validTo
        confidence
        store {
          name
        }
        isCurrentlyValid
        discountPercent
      }
    }
    totalCount
  }
}
```

### Current Prices

```graphql
query CurrentPrices {
  currentPrice(productID: 123, storeID: 1) {
    price
    isOnSale
    recordedAt
    store {
      name
    }
  }

  currentPrices(productID: 123, storeIDs: [1, 2, 3]) {
    price
    store {
      name
    }
    isOnSale
  }
}
```

### Price Trend Analysis

```graphql
query PriceTrendAnalysis {
  analyzeTrend(
    productID: 123
    period: "30_days"
    storeID: 1
  ) {
    productID
    direction  # RISING, FALLING, STABLE, VOLATILE
    trendPercentage
    confidence
    startPrice
    endPrice
    volatilityScore
    dataPoints
    linearRegression {
      slope
      intercept
      rSquared
      significant
    }
    movingAverages {
      ma7
      ma14
      ma30
      ma7AboveMa30
    }
    trendStrength  # WEAK, MODERATE, STRONG
  }
}
```

### Store Price Comparison

```graphql
query CompareStorePrices {
  compareStoreTrends(
    productID: 123
    storeIDs: [1, 2, 3]
    period: "90_days"
  ) {
    productID
    storeTrends {
      storeID
      storeName
      currentPrice
      trendAnalysis {
        direction
        trendPercentage
        confidence
      }
      priceStability
    }
    bestStore {
      storeName
      currentPrice
    }
    worstStore {
      storeName
      currentPrice
    }
    divergence
  }
}
```

### Price Predictions

```graphql
query PricePrediction {
  predictPrice(
    productID: 123
    daysAhead: 30
  ) {
    productID
    predictionDate
    predictedPrice
    confidenceInterval {
      lower
      upper
      confidence
    }
    methodology
    accuracy
    factors {
      name
      impact
      description
    }
  }
}
```

### Buying Recommendations

```graphql
query BuyingRecommendation {
  buyingRecommendation(productID: 123) {
    productID
    recommendation  # BUY_NOW, WAIT, MONITOR
    currentPrice
    recommendedAction
    confidence
    reasoning
    nextSaleDate
    potentialSavings
    waitDays
  }
}
```

### Price Alerts

```graphql
# Get alert suggestions
query PriceAlertSuggestions {
  priceAlertSuggestions(productID: 123) {
    productID
    currentPrice
    thresholds {
      level  # GOOD_PRICE, EXCELLENT_PRICE, PRICE_DROP
      price
      frequency
      description
    }
    historicalStats {
      averagePrice
      medianPrice
      percentile10
      percentile90
    }
  }
}

# Create price alert
mutation CreatePriceAlert {
  createPriceAlert(input: {
    productID: 123
    storeID: 1
    alertType: TARGET_PRICE
    targetPrice: 2.50
    notifyEmail: true
    notifyPush: false
    expiresAt: "2024-12-31T23:59:59Z"
  }) {
    id
    targetPrice
    alertType
    isActive
    createdAt
  }
}
```

### Seasonal Analysis

```graphql
query SeasonalTrends {
  seasonalTrends(productID: 123) {
    productID
    hasSeasonality
    seasonalIndex
    seasons {
      season
      startMonth
      endMonth
      avgPrice
      priceIndex
      nextStart
    }
    nextPeakSeason {
      season
      avgPrice
      nextStart
    }
    nextLowSeason {
      season
      avgPrice
      nextStart
    }
    recommendations {
      period
      advice
      potentialSavings
      timing
    }
  }
}
```

### Comprehensive Trend Summary

```graphql
query TrendSummary {
  trendSummary(productID: 123) {
    productID
    shortTermTrend {      # 7 days
      direction
      trendPercentage
      confidence
    }
    mediumTermTrend {     # 30 days
      direction
      trendPercentage
      confidence
    }
    longTermTrend {       # 90 days
      direction
      trendPercentage
      confidence
    }
    buyingAdvice {
      recommendation
      confidence
      reasoning
    }
    priceAlerts {
      thresholds {
        level
        price
      }
    }
    seasonalPattern {
      hasSeasonality
      nextPeakSeason {
        season
        nextStart
      }
    }
    keyInsights
  }
}
```

## ❌ Error Handling

### Error Response Format

```json
{
  "errors": [
    {
      "message": "Product not found",
      "locations": [
        {
          "line": 2,
          "column": 3
        }
      ],
      "path": ["product"],
      "extensions": {
        "code": "NOT_FOUND",
        "exception": {
          "id": "123"
        }
      }
    }
  ],
  "data": null
}
```

### Common Error Codes

- `UNAUTHENTICATED`: User not authenticated
- `UNAUTHORIZED`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `VALIDATION_ERROR`: Invalid input data
- `RATE_LIMITED`: Too many requests
- `INTERNAL_ERROR`: Server error

### Authentication Errors

```graphql
# Missing or invalid token
{
  "errors": [
    {
      "message": "Authentication required",
      "extensions": {
        "code": "UNAUTHENTICATED"
      }
    }
  ]
}

# Expired token
{
  "errors": [
    {
      "message": "Token expired",
      "extensions": {
        "code": "UNAUTHENTICATED",
        "expired": true
      }
    }
  ]
}
```

### Validation Errors

```graphql
{
  "errors": [
    {
      "message": "Validation failed",
      "extensions": {
        "code": "VALIDATION_ERROR",
        "validationErrors": [
          {
            "field": "email",
            "message": "Invalid email format"
          },
          {
            "field": "password",
            "message": "Password too short"
          }
        ]
      }
    }
  ]
}
```

## 🚦 Rate Limiting

### Limits

- **Anonymous users**: 100 requests per hour
- **Authenticated users**: 1000 requests per hour
- **Search queries**: 500 requests per hour
- **Authentication endpoints**: 10 requests per minute

### Rate Limit Headers

```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
X-RateLimit-Retry-After: 3600
```

### Rate Limit Exceeded

```json
{
  "errors": [
    {
      "message": "Rate limit exceeded",
      "extensions": {
        "code": "RATE_LIMITED",
        "retryAfter": 3600
      }
    }
  ]
}
```

## 💡 Examples

### Complete Shopping Workflow

```graphql
# 1. Search for products
query SearchForBread {
  advancedSearch(input: {
    query: "duona"
    storeIDs: [1, 2]
    onSaleOnly: true
    limit: 10
  }) {
    products {
      id
      name
      currentPrice
      isOnSale
      store {
        name
      }
    }
  }
}

# 2. Create shopping list
mutation CreateList {
  createShoppingList(input: {
    name: "Weekend Shopping"
  }) {
    id
    name
  }
}

# 3. Add items to list
mutation AddBread {
  createShoppingListItem(input: {
    listID: 1
    description: "Duona pilngrūdžių"
    quantity: 2
  }) {
    id
    matchedProduct {
      name
      currentPrice
    }
  }
}

# 4. Check price history
query CheckPriceHistory {
  priceHistory(productID: 456, first: 30) {
    edges {
      node {
        price
        recordedAt
        isOnSale
      }
    }
  }
}

# 5. Get buying recommendation
query GetRecommendation {
  buyingRecommendation(productID: 456) {
    recommendation
    reasoning
    potentialSavings
  }
}
```

### Price Monitoring Setup

```graphql
# 1. Check current prices across stores
query ComparePrices {
  currentPrices(productID: 789, storeIDs: [1, 2, 3]) {
    price
    store {
      name
    }
    isOnSale
  }
}

# 2. Analyze price trends
query AnalyzeTrend {
  analyzeTrend(productID: 789, period: "30_days") {
    direction
    trendPercentage
    confidence
  }
}

# 3. Set up price alert
mutation SetupAlert {
  createPriceAlert(input: {
    productID: 789
    alertType: PERCENTAGE_DROP
    dropPercent: 15
    notifyEmail: true
  }) {
    id
    targetPrice
    isActive
  }
}
```

---

## 📚 Additional Resources

- [GraphQL Playground](http://localhost:8080/graphql) - Interactive API explorer
- [GitHub Repository](https://github.com/kainuguru/kainuguru-api) - Source code and issues
- [BDD Test Scenarios](./tests/bdd/) - Behavioral test specifications

For more detailed examples and advanced use cases, visit the [GraphQL Playground](http://localhost:8080/graphql) where you can explore the complete schema interactively.