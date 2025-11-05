Feature: Public Access to Flyers
  As a public user
  I want to access flyer information without requiring authentication
  So that I can browse grocery deals without creating an account

  Background:
    Given the system has the following stores:
      | name    | code   | enabled |
      | IKI     | iki    | true    |
      | Maxima  | maxima | true    |
      | Rimi    | rimi   | true    |
    And there are current flyers available for all stores
    And I am not logged in to the system

  Scenario: Access health check endpoint publicly
    When I request the health check endpoint
    Then I should receive a successful response
    And the response should indicate system health
    And no authentication should be required

  Scenario: Access GraphQL playground publicly
    When I request the GraphQL playground endpoint
    Then I should receive the playground interface
    And I should see information about available queries
    And no authentication should be required
    And the playground should show current implementation status

  Scenario: Browse stores without authentication
    When I request all stores via GraphQL without authentication
    Then I should see all enabled stores
    And each store should have complete information
    And the response should be successful
    And no authentication headers should be required

  Scenario: Browse current flyers without authentication
    When I request current flyers via GraphQL without authentication
    Then I should see flyers from all enabled stores
    And each flyer should have complete information
    And flyer details should be accessible
    And no authentication should be required

  Scenario: View flyer pages without authentication
    Given there is a current flyer with pages
    When I request flyer pages via GraphQL without authentication
    Then I should see all pages for the flyer
    And page images should be accessible
    And page details should be complete
    And no authentication should be required

  Scenario: Search products without authentication
    Given there are products in current flyers
    When I search for products via GraphQL without authentication
    Then I should see search results
    And product information should be complete
    And search functionality should work normally
    And no authentication should be required

  Scenario: View product details without authentication
    Given there are products in current flyers
    When I request product details via GraphQL without authentication
    Then I should see complete product information
    And product prices should be visible
    And store information should be included
    And no authentication should be required

  Scenario: Access product similarity without authentication
    Given there is a product with similar products available
    When I request similar products via GraphQL without authentication
    Then I should see similar product suggestions
    And similarity results should be complete
    And no authentication should be required

  Scenario: Get search suggestions without authentication
    When I request search suggestions via GraphQL without authentication
    Then I should see relevant search suggestions
    And suggestions should be based on current flyers
    And no authentication should be required

  Scenario: Rate limiting applies to public access
    When I make multiple rapid requests to the API without authentication
    Then rate limiting should be applied consistently
    And I should receive appropriate rate limit responses
    And the system should handle public traffic properly

  Scenario: CORS headers for public access
    When I make a cross-origin request to the GraphQL endpoint
    Then appropriate CORS headers should be returned
    And the request should be allowed from configured origins
    And preflight requests should be handled correctly

  Scenario: Error handling for public requests
    When I make an invalid request via GraphQL without authentication
    Then I should receive appropriate error messages
    And errors should not expose sensitive information
    And error responses should be user-friendly

  Scenario: Performance for public requests
    When I make requests to public endpoints without authentication
    Then responses should be returned within reasonable time limits
    And performance should not be degraded for unauthenticated users
    And caching should be effective for public data

  Scenario: GraphQL introspection for public schema
    When I perform GraphQL introspection queries without authentication
    Then I should see the complete public schema
    And all available queries should be documented
    And schema should show proper types and relationships
    And no authentication should be required