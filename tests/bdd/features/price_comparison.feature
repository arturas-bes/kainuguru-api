Feature: Price Comparison Across Stores
  As a user
  I want to compare prices for the same product across different stores
  So that I can find the best deals and save money

  Background:
    Given I am logged in as user "jonas@example.com"
    And the following stores exist:
      | code   | name    |
      | maxima | Maxima  |
      | iki    | IKI     |
      | rimi   | RIMI    |
      | lidl   | Lidl    |
    And the following product masters exist:
      | name              | brand      | category        |
      | Pienas 2.5% 1L   | Rokiškio   | Pieno produktai |
      | Duona ruginė     | Vilniaus   | Duona           |
      | Jogurtas braškių | Žemaitijos | Pieno produktai |

  Scenario: Compare prices for a single product across all stores
    Given the product master "Pienas 2.5% 1L" has prices:
      | store  | price | valid_from | valid_until |
      | maxima | 1.99  | today      | +7 days     |
      | iki    | 2.15  | today      | +7 days     |
      | rimi   | 1.89  | today      | +7 days     |
      | lidl   | 1.95  | today      | +7 days     |
    When I request price comparison for "Pienas 2.5% 1L"
    Then I should see 4 store prices
    And the best price should be "1.89" at "RIMI"
    And the most expensive should be "2.15" at "IKI"
    And I should see potential savings of "0.26"

  Scenario: Compare prices for multiple products
    Given the following current prices exist:
      | product          | store  | price |
      | Pienas 2.5% 1L  | maxima | 1.99  |
      | Pienas 2.5% 1L  | iki    | 2.15  |
      | Duona ruginė    | maxima | 1.49  |
      | Duona ruginė    | rimi   | 1.39  |
    When I request price comparison for products:
      | Pienas 2.5% 1L |
      | Duona ruginė   |
    Then I should see comparison for 2 products
    And total savings potential should be "0.26"
    And the comparison should show:
      | product         | best_store | best_price | savings |
      | Pienas 2.5% 1L | iki        | 2.15       | 0.00    |
      | Duona ruginė   | rimi       | 1.39       | 0.10    |

  Scenario: Price comparison includes historical data
    Given the product "Pienas 2.5% 1L" at "Maxima" had prices:
      | price | date      |
      | 2.29  | -30 days  |
      | 2.15  | -14 days  |
      | 1.99  | today     |
    When I request price comparison with history for "Pienas 2.5% 1L"
    Then I should see the current price is "1.99"
    And I should see the price trend is "decreasing"
    And I should see the 30-day low was "1.99"
    And I should see the 30-day high was "2.29"

  Scenario: No prices available for product
    Given no stores have prices for "Avižų dribsniai"
    When I request price comparison for "Avižų dribsniai"
    Then I should receive an empty comparison
    And I should see a message "No prices found for this product"

  Scenario: Price comparison only shows current valid prices
    Given the product "Pienas 2.5% 1L" has prices:
      | store  | price | valid_from | valid_until |
      | maxima | 1.79  | -14 days   | -1 day      |
      | maxima | 1.99  | today      | +7 days     |
      | iki    | 2.15  | today      | +7 days     |
    When I request price comparison for "Pienas 2.5% 1L"
    Then I should see 2 store prices
    And the Maxima price should be "1.99"
    And the expired price "1.79" should not appear

  Scenario: Price comparison with savings calculation
    Given I have a shopping list with:
      | product         | quantity |
      | Pienas 2.5% 1L | 2        |
      | Duona ruginė   | 1        |
    And current prices are:
      | product         | store  | price |
      | Pienas 2.5% 1L | maxima | 1.99  |
      | Pienas 2.5% 1L | rimi   | 1.89  |
      | Duona ruginė   | maxima | 1.49  |
      | Duona ruginė   | rimi   | 1.59  |
    When I request optimized shopping comparison
    Then the best single-store option should be "Maxima" with total "5.47"
    And shopping at multiple stores could save "0.10"
    And the optimal split should suggest:
      | store  | products        | total |
      | RIMI   | Pienas (x2)     | 3.78  |
      | Maxima | Duona           | 1.49  |

  Scenario: Store-by-store total comparison
    Given I have a shopping list with 5 items
    When I request price comparison by store
    Then I should see totals for each store:
      | store  | total | available_items |
      | Maxima | 12.45 | 5               |
      | IKI    | 13.20 | 5               |
      | RIMI   | 11.89 | 4               |
      | Lidl   | 12.10 | 3               |
    And the recommendation should be "Shop at RIMI (€11.89)"
    And I should see "1 item not available at RIMI"

  Scenario: Compare unit prices for different package sizes
    Given the following products exist:
      | name              | store  | price | unit   |
      | Pienas 2.5% 1L   | maxima | 1.99  | 1L     |
      | Pienas 2.5% 2L   | maxima | 3.49  | 2L     |
      | Pienas 2.5% 500ml| iki    | 1.15  | 500ml  |
    When I compare unit prices for "Pienas 2.5%"
    Then the unit price comparison should show:
      | product          | store  | price | unit_price |
      | Pienas 2.5% 2L   | maxima | 3.49  | 1.75/L     |
      | Pienas 2.5% 1L   | maxima | 1.99  | 1.99/L     |
      | Pienas 2.5% 500ml| iki    | 1.15  | 2.30/L     |
    And the best value should be "Pienas 2.5% 2L" at Maxima
