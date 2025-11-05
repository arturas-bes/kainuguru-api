Feature: Price History Tracking
  As a user
  I want to view the price history of products
  So that I can make informed purchasing decisions based on price trends

  Background:
    Given I am a registered user
    And I am authenticated

  Scenario: View product price history
    Given there is a product "Pienas 1L" with historical prices:
      | date       | price | store     |
      | 2024-10-01 | 1.49  | Maxima    |
      | 2024-10-05 | 1.39  | Maxima    |
      | 2024-10-10 | 1.59  | Maxima    |
      | 2024-10-15 | 1.45  | Maxima    |
    When I request the price history for "Pienas 1L"
    Then I should see the price history with 4 entries
    And the prices should be ordered by date descending
    And each entry should contain date, price, and store information

  Scenario: View price history with multiple stores
    Given there is a product "Duona 500g" with historical prices:
      | date       | price | store     |
      | 2024-10-01 | 0.89  | Maxima    |
      | 2024-10-01 | 0.95  | Rimi      |
      | 2024-10-05 | 0.85  | Maxima    |
      | 2024-10-05 | 0.92  | Rimi      |
    When I request the price history for "Duona 500g"
    Then I should see the price history with 4 entries
    And the history should show prices from both stores
    And prices should be grouped by date with multiple store entries

  Scenario: View price history for specific date range
    Given there is a product "Kiaušiniai 10vnt" with historical prices:
      | date       | price | store     |
      | 2024-09-20 | 2.99  | Maxima    |
      | 2024-10-01 | 3.19  | Maxima    |
      | 2024-10-15 | 2.89  | Maxima    |
      | 2024-10-30 | 3.09  | Maxima    |
    When I request the price history for "Kiaušiniai 10vnt" from "2024-10-01" to "2024-10-30"
    Then I should see the price history with 3 entries
    And all entries should be within the specified date range

  Scenario: View price history for product with no history
    Given there is a product "Naujas produktas" with no price history
    When I request the price history for "Naujas produktas"
    Then I should see an empty price history
    And I should receive a message indicating no price data is available

  Scenario: View current vs historical price comparison
    Given there is a product "Aliejus 1L" with historical prices:
      | date       | price | store     |
      | 2024-10-01 | 2.99  | Maxima    |
      | 2024-10-15 | 3.19  | Maxima    |
    And the current price is 2.89 at "Maxima"
    When I request the price history for "Aliejus 1L"
    Then I should see the current price highlighted
    And I should see the price difference from the last recorded price
    And I should see if the current price is higher or lower than historical average

  Scenario: View price history with data aggregation by week
    Given there is a product "Cukrus 1kg" with daily price data over 4 weeks
    When I request the weekly aggregated price history for "Cukrus 1kg"
    Then I should see weekly average prices
    And each week should show min, max, and average prices
    And the data should be aggregated by calendar week

  Scenario: Access control for price history
    Given I am not authenticated
    When I try to request the price history for any product
    Then I should receive an authentication error
    And I should not see any price history data

  Scenario: Price history performance with large dataset
    Given there is a product with over 1000 historical price entries
    When I request the price history for this product
    Then the response should be returned within 2 seconds
    And the data should be paginated with 50 entries per page
    And I should be able to navigate through different pages