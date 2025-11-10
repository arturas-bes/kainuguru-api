Feature: Product Master Matching and Deduplication
  As a system administrator
  I want products to be automatically matched to product masters
  So that duplicate products are deduplicated and price history is maintained

  Background:
    Given the database is clean
    And the following stores exist:
      | code   | name    |
      | maxima | Maxima  |
      | iki    | IKI     |
      | rimi   | RIMI    |

  Scenario: Exact name match with same brand
    Given the following product masters exist:
      | name                | brand      | category        | confidence |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | 0.9        |
    When a product is added:
      | name                | brand      | category        | store  |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | maxima |
    Then the product should be matched to master "Pienas 2.5% 1L"
    And the match confidence should be >= 0.9
    And the match method should be "auto"

  Scenario: Fuzzy name match with similar naming
    Given the following product masters exist:
      | name                | brand      | category        | confidence |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | 0.9        |
    When a product is added:
      | name                | brand      | category        | store  |
      | Pienas 2,5% 1 litras | Rokiškio   | Pieno produktai | iki    |
    Then the product should be matched to master "Pienas 2.5% 1L"
    And the match confidence should be >= 0.7
    And the match method should be "auto"

  Scenario: Brand and category match with word overlap
    Given the following product masters exist:
      | name                | brand      | category        | confidence |
      | Jogurtas braškių    | Žemaitijos | Pieno produktai | 0.8        |
    When a product is added:
      | name                   | brand      | category        | store  |
      | Braškių skonio jogurtas | Žemaitijos | Pieno produktai | rimi   |
    Then the product should be matched to master "Jogurtas braškių"
    And the match confidence should be >= 0.5

  Scenario: No match creates new product master
    Given no product masters exist for "Avižų dribsniai"
    When a product is added:
      | name              | brand    | category | store  |
      | Avižų dribsniai   | Vilniaus | Kruopos  | maxima |
    Then a new product master should be created
    And the new master should have name "Avižų dribsniai"
    And the master confidence should be 0.5
    And the match method should be "new_master"

  Scenario: Manual verification increases confidence
    Given the following product masters exist:
      | name                | brand      | category        | confidence |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | 0.7        |
    When an admin verifies the product master "Pienas 2.5% 1L"
    Then the master confidence should be 1.0
    And the master status should be "verified"

  Scenario: Duplicate master merging
    Given the following product masters exist:
      | name                | brand      | category        | confidence |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | 0.9        |
      | Pienas 2,5% 1L     | Rokiškio   | Pieno produktai | 0.6        |
    And the second master has 3 linked products
    When an admin marks "Pienas 2,5% 1L" as duplicate of "Pienas 2.5% 1L"
    Then all products should be linked to "Pienas 2.5% 1L"
    And the master "Pienas 2,5% 1L" should be marked as duplicate
    And the match count of "Pienas 2.5% 1L" should increase by 3

  Scenario: Background worker processes unmatched products
    Given the following unmatched products exist:
      | name                | brand      | category        | store  |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | maxima |
      | Duona ruginė       | Vilniaus   | Duona           | iki    |
      | Jogurtas braškių   | Žemaitijos | Pieno produktai | rimi   |
    When the product master worker runs
    Then all 3 products should be matched or have new masters created
    And the worker should log statistics

  Scenario: Confidence score increases with linked products
    Given the following product masters exist:
      | name                | brand      | category        | confidence |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | 0.5        |
    And the master has 1 linked product
    When 4 more products are linked to the master
    Then the master match count should be 5
    And the master confidence should be 0.7

  Scenario: Get masters requiring manual review
    Given the following product masters exist:
      | name                | brand      | category        | confidence |
      | Pienas 2.5% 1L     | Rokiškio   | Pieno produktai | 0.9        |
      | Duona ruginė       | Vilniaus   | Duona           | 0.6        |
      | Jogurtas braškių   | Žemaitijos | Pieno produktai | 0.4        |
      | Sviestas           | Kelmės     | Pieno produktai | 0.2        |
    When I request masters for review
    Then I should receive 2 masters
    And the list should include "Duona ruginė"
    And the list should include "Jogurtas braškių"
    And the list should not include "Pienas 2.5% 1L"
    And the list should not include "Sviestas"

  Scenario: Matching statistics tracking
    Given the following matching activity occurred today:
      | products_matched | auto_matches | manual_matches | new_masters |
      | 100              | 85           | 10             | 5           |
    When I request overall matching statistics
    Then the total products matched should be 100
    And the auto match rate should be 85%
    And the manual match rate should be 10%
    And the new master creation rate should be 5%
