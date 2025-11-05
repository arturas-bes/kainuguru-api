Feature: Price Trend Analysis
  As a user
  I want to analyze price trends for products
  So that I can understand price patterns and make better shopping decisions

  Background:
    Given I am a registered user
    And I am authenticated

  Scenario: Calculate basic price trend for rising prices
    Given there is a product "Bananas 1kg" with historical prices:
      | date       | price | store     |
      | 2024-09-01 | 1.99  | Maxima    |
      | 2024-09-15 | 2.19  | Maxima    |
      | 2024-10-01 | 2.39  | Maxima    |
      | 2024-10-15 | 2.49  | Maxima    |
    When I request the price trend analysis for "Bananas 1kg"
    Then the trend should be "RISING"
    And the trend percentage should be approximately 25.1%
    And the average price change per week should be positive

  Scenario: Calculate basic price trend for falling prices
    Given there is a product "Bulvės 1kg" with historical prices:
      | date       | price | store     |
      | 2024-09-01 | 2.49  | Maxima    |
      | 2024-09-15 | 2.29  | Maxima    |
      | 2024-10-01 | 1.99  | Maxima    |
      | 2024-10-15 | 1.89  | Maxima    |
    When I request the price trend analysis for "Bulvės 1kg"
    Then the trend should be "FALLING"
    And the trend percentage should be approximately -24.1%
    And the average price change per week should be negative

  Scenario: Calculate stable price trend
    Given there is a product "Druska 1kg" with historical prices:
      | date       | price | store     |
      | 2024-09-01 | 0.99  | Maxima    |
      | 2024-09-15 | 1.01  | Maxima    |
      | 2024-10-01 | 0.98  | Maxima    |
      | 2024-10-15 | 1.00  | Maxima    |
    When I request the price trend analysis for "Druska 1kg"
    Then the trend should be "STABLE"
    And the trend percentage should be approximately 1.0%
    And the price volatility should be "LOW"

  Scenario: Analyze price volatility
    Given there is a product "Benzinas 1L" with highly variable prices:
      | date       | price | store     |
      | 2024-09-01 | 1.45  | Circle K  |
      | 2024-09-03 | 1.52  | Circle K  |
      | 2024-09-05 | 1.38  | Circle K  |
      | 2024-09-07 | 1.47  | Circle K  |
      | 2024-09-09 | 1.55  | Circle K  |
    When I request the price trend analysis for "Benzinas 1L"
    Then the price volatility should be "HIGH"
    And the standard deviation should be greater than 0.05
    And I should see volatility warnings

  Scenario: Compare trends across multiple stores
    Given there is a product "Pienas 1L" with prices in multiple stores:
      | date       | price | store     |
      | 2024-10-01 | 1.49  | Maxima    |
      | 2024-10-01 | 1.55  | Rimi      |
      | 2024-10-15 | 1.59  | Maxima    |
      | 2024-10-15 | 1.52  | Rimi      |
    When I request the comparative price trend analysis for "Pienas 1L"
    Then I should see trends for each store separately
    And "Maxima" trend should be "RISING" with 6.7% increase
    And "Rimi" trend should be "FALLING" with -1.9% decrease
    And I should see which store has better price stability

  Scenario: Seasonal trend detection
    Given there is a product "Šokoladas 100g" with seasonal price data:
      | date       | price | store     |
      | 2024-01-15 | 1.99  | Maxima    |
      | 2024-04-15 | 2.49  | Maxima    |  # Easter
      | 2024-07-15 | 1.89  | Maxima    |
      | 2024-10-15 | 1.95  | Maxima    |
      | 2024-12-15 | 2.89  | Maxima    |  # Christmas
    When I request the seasonal trend analysis for "Šokoladas 100g"
    Then I should see seasonal patterns identified
    And December should be marked as "HIGH_SEASON" with premium pricing
    And April should be marked as "SEASONAL_PEAK"
    And July should be marked as "LOW_SEASON"

  Scenario: Price prediction based on trends
    Given there is a product "Kava 500g" with consistent price trend:
      | date       | price | store     |
      | 2024-08-01 | 4.99  | Maxima    |
      | 2024-09-01 | 5.19  | Maxima    |
      | 2024-10-01 | 5.39  | Maxima    |
      | 2024-11-01 | 5.59  | Maxima    |
    When I request the price prediction for "Kava 500g"
    Then I should see a predicted price for next month
    And the prediction should be around 5.79 with confidence interval
    And I should see the prediction accuracy rating
    And I should see factors affecting the prediction

  Scenario: Best time to buy recommendation
    Given there is a product "Saldainiai 200g" with cyclical price pattern:
      | date       | price | store     |
      | 2024-09-01 | 2.99  | Maxima    |
      | 2024-09-15 | 1.99  | Maxima    |  # Sale period
      | 2024-10-01 | 2.99  | Maxima    |
      | 2024-10-15 | 1.99  | Maxima    |  # Sale period
    When I request the buying recommendation for "Saldainiai 200g"
    Then I should see the next predicted sale date
    And I should see the recommended action is "WAIT"
    And I should see the potential savings amount
    And I should see the confidence level of the recommendation

  Scenario: Price alert threshold suggestions
    Given there is a product "Aliejus 1L" with price history
    When I request price alert suggestions for "Aliejus 1L"
    Then I should see suggested alert thresholds based on historical data
    And the "GOOD_PRICE" threshold should be below historical average
    And the "EXCELLENT_PRICE" threshold should be in bottom 10% of prices
    And I should see the frequency of prices reaching each threshold

  Scenario: Trend analysis with insufficient data
    Given there is a product "Naujas produktas" with only 2 price points
    When I request the price trend analysis for "Naujas produktas"
    Then I should see a message indicating insufficient data for trend analysis
    And I should see the minimum data requirements
    And I should see the available data points

  Scenario: Macro trend analysis across product category
    Given there are multiple products in "Dairy" category with price trends
    When I request the category trend analysis for "Dairy"
    Then I should see the overall category trend direction
    And I should see the average price change across all dairy products
    And I should see which products are driving the category trend
    And I should see correlation with external factors if available