Feature: Browse Weekly Grocery Flyers
  As a user
  I want to browse current weekly flyers from grocery stores
  So that I can see what products are available and plan my shopping

  Background:
    Given the system has the following stores:
      | name    | code   | enabled |
      | IKI     | iki    | true    |
      | Maxima  | maxima | true    |
      | Rimi    | rimi   | true    |
    And there are current flyers available for all stores

  Scenario: View list of all stores
    When I request all stores via GraphQL
    Then I should see 3 stores
    And each store should have id, name, code, and enabled fields
    And only enabled stores should be returned

  Scenario: View current flyers for all stores
    When I request current flyers via GraphQL
    Then I should see flyers from all enabled stores
    And each flyer should have id, store_id, valid_from, valid_to, and page_count
    And all flyers should be currently valid
    And flyers should be ordered by valid_from descending

  Scenario: View current flyers for specific store
    Given I want to see flyers only for "IKI"
    When I request current flyers for store_ids [1] via GraphQL
    Then I should see only flyers from IKI store
    And the flyers should be currently valid

  Scenario: View flyer pages for a specific flyer
    Given there is a current flyer for IKI with pages
    When I request pages for that flyer via GraphQL
    Then I should see all pages for the flyer
    And each page should have id, flyer_id, page_number, and image_url
    And pages should be ordered by page_number ascending

  Scenario: View flyer details by ID
    Given there is a flyer with ID 1
    When I request flyer details for ID 1 via GraphQL
    Then I should see the complete flyer information
    And it should include store details
    And it should include valid date range

  Scenario: Handle empty flyer results
    Given there are no current flyers for store_ids [999]
    When I request current flyers for store_ids [999] via GraphQL
    Then I should see an empty flyers list
    And the response should be successful

  Scenario: View store details by code
    When I request store by code "iki" via GraphQL
    Then I should see the IKI store details
    And the store should include id, name, code, and enabled status

  Scenario: Pagination of flyers
    Given there are more than 10 flyers available
    When I request the first 5 flyers via GraphQL
    Then I should see exactly 5 flyers
    And I should receive pagination information with hasNextPage
    When I request the next 5 flyers using pagination cursor
    Then I should see the next 5 flyers

  Scenario: Access flyers without authentication
    Given I am not logged in
    When I request current flyers via GraphQL
    Then I should successfully receive the flyers
    And no authentication should be required

  Scenario: System performance requirements
    Given there are flyers from all major stores
    When I request current flyers via GraphQL
    Then the response should be returned within 500ms
    And the GraphQL query should execute efficiently