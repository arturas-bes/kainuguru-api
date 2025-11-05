Feature: Search Products Across All Flyers
  As a user
  I want to search for products across all current flyers
  So that I can find specific items and compare prices between stores

  Background:
    Given the system has the following stores:
      | name    | code   | enabled |
      | IKI     | iki    | true    |
      | Maxima  | maxima | true    |
      | Rimi    | rimi   | true    |
    And there are current flyers with products for all stores
    And the search indexes are properly configured for Lithuanian language

  Scenario: Basic product search
    Given there are products containing "duona" in current flyers
    When I search for "duona" via GraphQL
    Then I should see products matching "duona"
    And each product should have id, name, price, store_id, and flyer_id
    And results should be ranked by relevance
    And results should include store information

  Scenario: Lithuanian diacritics handling
    Given there are products with Lithuanian characters like "ąčęėįšųūž"
    When I search for "aceeisuuz" (without diacritics)
    Then I should see products matching the Lithuanian equivalents
    And the search should be case-insensitive
    And both "pienas" and "PIENAS" should match "pienas"

  Scenario: Fuzzy search with typos
    Given there are products containing "pienas" in current flyers
    When I search for "piens" (with typo) via GraphQL
    Then I should see products containing "pienas"
    And the fuzzy matching should handle minor typos
    And results should be ordered by similarity score

  Scenario: Search with filters
    Given there are products in multiple stores and categories
    When I search for "duona" with store filter for "IKI" via GraphQL
    Then I should see only products from IKI store
    And products should still match "duona"
    When I search for "duona" with price range 1.00-5.00 via GraphQL
    Then I should see only products in that price range

  Scenario: Empty search results
    When I search for "nonexistentproduct123" via GraphQL
    Then I should see an empty products list
    And the response should be successful
    And I should receive appropriate messaging

  Scenario: Search suggestions
    When I request search suggestions for "pien" via GraphQL
    Then I should see suggested terms like "pienas"
    And suggestions should be ordered by popularity
    And suggestions should be relevant to current flyers

  Scenario: Search with special characters
    Given there are products with special characters and numbers
    When I search for "100%" via GraphQL
    Then the search should handle special characters safely
    And I should see relevant products
    And no SQL injection should occur

  Scenario: Product similarity search
    Given there is a product with ID 1
    When I request similar products for product ID 1 via GraphQL
    Then I should see products similar to the original
    And similarity should be based on name and category
    And the original product should not be included in results

  Scenario: Search performance requirements
    Given there are thousands of products in current flyers
    When I search for a common term like "pienas" via GraphQL
    Then the search should complete within 200ms
    And results should be limited to a reasonable number (e.g., 50)
    And pagination should be available for more results

  Scenario: Search analytics
    When I perform multiple searches via GraphQL
    Then search queries should be logged for analytics
    And popular search terms should be tracked
    And search performance metrics should be recorded

  Scenario: Multi-word search
    Given there are products with multi-word names like "Baltojo šokolado plokštelė"
    When I search for "baltas šokoladas" via GraphQL
    Then I should see products matching both terms
    And partial word matches should be supported
    And word order should not strictly matter

  Scenario: Category-based search
    Given there are products categorized as "pieno produktai"
    When I search within category "dairy" via GraphQL
    Then I should see only dairy products
    And category filtering should work with text search
    And results should be relevant to both category and search term

  Scenario: Search result pagination
    Given there are more than 20 products matching "duona"
    When I search for "duona" with limit 10 via GraphQL
    Then I should see exactly 10 products
    And I should receive pagination information
    When I request the next 10 products using pagination cursor
    Then I should see the next 10 matching products
    And there should be no duplicate results

  Scenario: Hybrid search functionality
    Given there are products that match via full-text search and trigram similarity
    When I search for a term that triggers both search methods
    Then results should combine both approaches
    And hybrid ranking should prioritize full-text matches
    And trigram matches should supplement full-text results