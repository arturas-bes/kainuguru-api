Feature: Smart Shopping List Optimization
  As a user
  I want the system to optimize my shopping list across stores
  So that I can minimize costs and save time

  Background:
    Given I am logged in as user "jonas@example.com"
    And the following stores exist:
      | code   | name    | location          |
      | maxima | Maxima  | Vilnius Center    |
      | iki    | IKI     | Vilnius North     |
      | rimi   | RIMI    | Vilnius South     |
      | lidl   | Lidl    | Vilnius West      |

  Scenario: Single store optimization (minimum hassle)
    Given I have a shopping list "Weekly Shopping" with items:
      | product         | quantity |
      | Pienas 2.5% 1L | 2        |
      | Duona ruginė   | 1        |
      | Jogurtas       | 4        |
      | Sviestas       | 1        |
      | Kiaušiniai     | 10       |
    And current prices exist at all stores
    When I request optimization with strategy "single_store"
    Then I should receive a single-store recommendation
    And the result should include:
      | field        | value   |
      | store        | Maxima  |
      | total_cost   | 15.45   |
      | item_count   | 5       |
      | strategy     | SINGLE  |
    And all items should be available at the chosen store

  Scenario: Multi-store optimization (maximum savings)
    Given I have a shopping list with items:
      | product         | quantity |
      | Pienas 2.5% 1L | 2        |
      | Duona ruginė   | 1        |
      | Jogurtas       | 4        |
    And the best prices are:
      | product         | best_store | price |
      | Pienas 2.5% 1L | RIMI       | 1.89  |
      | Duona ruginė   | Maxima     | 1.49  |
      | Jogurtas       | IKI        | 0.99  |
    When I request optimization with strategy "multi_store"
    Then I should receive a multi-store plan
    And the result should recommend shopping at 3 stores
    And total cost should be "8.23"
    And potential savings vs single store should be shown
    And I should see a shopping route

  Scenario: Balanced optimization (some savings, fewer stores)
    Given I have a shopping list with 10 items
    And single-store cost at Maxima would be "25.00"
    And shopping at 4 stores could save "2.50"
    When I request optimization with strategy "balanced"
    Then the system should recommend 2 stores maximum
    And total cost should be between "23.00" and "24.00"
    And I should see "Shopping at 2 stores saves €1.50"

  Scenario: Optimization with item availability constraints
    Given I have a shopping list with items:
      | product         | quantity |
      | Pienas 2.5% 1L | 2        |
      | Duona ruginė   | 1        |
      | Specialty item | 1        |
    And "Specialty item" is only available at "Lidl"
    When I request optimization with strategy "single_store"
    Then the recommendation should be "Lidl"
    And I should see a note "Required for specialty items"

  Scenario: Price alert integration with optimization
    Given I have price alerts for:
      | product         | target_price |
      | Pienas 2.5% 1L | 1.85         |
      | Jogurtas       | 0.95         |
    And current prices are:
      | product         | store  | price |
      | Pienas 2.5% 1L | RIMI   | 1.79  |
      | Jogurtas       | IKI    | 0.89  |
    When I request shopping optimization
    Then I should see alerts for 2 products
    And the optimization should prioritize these deals
    And I should see "2 items on sale at target price!"

  Scenario: Shopping list auto-migration on expired products
    Given I have a shopping list with items:
      | product         | current_master | linked_product_id |
      | Pienas 2.5% 1L | Master-123     | Product-456       |
    And product "Product-456" expired yesterday
    And a new product exists:
      | name            | store  | master     | valid_until |
      | Pienas 2.5% 1L | Maxima | Master-123 | +7 days     |
    When the shopping list migration worker runs
    Then my shopping list item should be updated to new product
    And the item should maintain same master ID
    And I should see a notification "1 item updated with current prices"

  Scenario: Optimization considers store distance
    Given I have set my location to "Vilnius Center"
    And store distances are:
      | store  | distance_km |
      | Maxima | 0.5         |
      | RIMI   | 3.2         |
      | IKI    | 5.8         |
    When I request optimization with "prefer_nearby" enabled
    Then closer stores should be weighted higher
    And if Maxima is only €0.50 more expensive, it should be recommended

  Scenario: Weekly shopping plan optimization
    Given I have 3 shopping lists:
      | name           | item_count | total_value |
      | Weekly basics  | 15         | 35.00       |
      | Fresh produce  | 8          | 12.00       |
      | Household      | 5          | 18.00       |
    When I request combined optimization
    Then I should receive a weekly shopping strategy
    And the plan should group items by store visit
    And I should see estimated time savings
    And total cost across all lists should be optimized

  Scenario: Optimization with promotional offers
    Given current promotions are:
      | store  | promotion               | products    | discount |
      | Maxima | Buy 2 Get 1 Free       | Jogurtas    | 33%      |
      | RIMI   | 20% off dairy          | All dairy   | 20%      |
    And I have "Jogurtas" (quantity: 3) in my list
    When I request optimization
    Then the system should detect the Maxima promotion
    And recommend buying at Maxima for "Buy 2 Get 1 Free"
    And show adjusted total with promotion applied

  Scenario: Alternative product suggestions
    Given I have "Premium Pienas" in my list
    And "Premium Pienas" costs "2.99" at all stores
    And similar product "Regular Pienas" costs "1.89"
    When I request optimization with suggestions enabled
    Then I should see "Consider similar product: Regular Pienas"
    And the potential saving should be "1.10 per item"
    And I can choose to swap the product

  Scenario: Optimization preserves user preferences
    Given I have marked "Maxima" as preferred store
    And I have disabled multi-store shopping
    When I request optimization
    Then the result should only recommend "Maxima"
    And I should not see multi-store options
    And the response should note "Based on your preferences"

  Scenario: Failed optimization graceful handling
    Given no stores have current prices for any items
    When I request optimization
    Then I should receive an error with helpful message
    And the system should suggest "Try adding items to your list"
    And fallback to showing all stores with item availability
