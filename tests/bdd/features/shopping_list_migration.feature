Feature: Automatic Shopping List Migration
  As a user
  I want my shopping list items to automatically update when products expire
  So that I always see current prices and availability

  Background:
    Given I am logged in as user "jonas@example.com"
    And the following product masters exist:
      | id  | name              | brand      |
      | 100 | Pienas 2.5% 1L   | Rokiškio   |
      | 101 | Duona ruginė     | Vilniaus   |
      | 102 | Jogurtas braškių | Žemaitijos |

  Scenario: Migrate expired product to current product
    Given I have a shopping list "Weekly Shopping" with items:
      | product_id | master_id | name            | price | valid_until |
      | 1001       | 100       | Pienas 2.5% 1L | 1.99  | yesterday   |
    And a new product exists:
      | product_id | master_id | name            | price | valid_until | store  |
      | 1002       | 100       | Pienas 2.5% 1L | 2.15  | +7 days     | Maxima |
    When the shopping list migration worker runs
    Then the shopping list item should be updated:
      | field       | old_value | new_value |
      | product_id  | 1001      | 1002      |
      | price       | 1.99      | 2.15      |
      | master_id   | 100       | 100       |
    And I should receive a notification "1 item updated with new prices"

  Scenario: Migrate multiple expired items in one list
    Given I have a shopping list with expired items:
      | product_id | master_id | name            |
      | 1001       | 100       | Pienas 2.5% 1L |
      | 2001       | 101       | Duona ruginė   |
      | 3001       | 102       | Jogurtas       |
    And new products exist for all masters
    When the migration worker runs
    Then all 3 items should be migrated
    And the migration statistics should show:
      | items_processed | items_migrated | errors |
      | 3               | 3              | 0      |
    And the shopping list should remain valid

  Scenario: Skip migration if product still valid
    Given I have a shopping list with items:
      | product_id | master_id | name            | valid_until |
      | 1001       | 100       | Pienas 2.5% 1L | +5 days     |
      | 2001       | 101       | Duona ruginė   | yesterday   |
    When the migration worker runs
    Then only 1 item should be migrated
    And product 1001 should remain unchanged
    And product 2001 should be updated to new product

  Scenario: Handle no replacement product available
    Given I have a shopping list with an expired item:
      | product_id | master_id | name            | valid_until |
      | 1001       | 100       | Pienas 2.5% 1L | yesterday   |
    And no current products exist for master 100
    When the migration worker runs
    Then the item should remain in list with warning
    And the item should be marked as "needs_attention"
    And I should see notification "1 item needs manual update"

  Scenario: Prefer same store during migration
    Given I have a shopping list item:
      | product_id | master_id | name            | store  | expired |
      | 1001       | 100       | Pienas 2.5% 1L | Maxima | yes     |
    And replacement products exist:
      | product_id | master_id | name            | store  | price |
      | 1002       | 100       | Pienas 2.5% 1L | Maxima | 2.15  |
      | 1003       | 100       | Pienas 2.5% 1L | IKI    | 1.99  |
    When the migration worker runs with "prefer_same_store"
    Then product_id should be updated to 1002 (Maxima)
    And the preference for same store should be logged

  Scenario: Choose cheapest when same store not available
    Given I have a shopping list item:
      | product_id | master_id | name            | store  | expired |
      | 1001       | 100       | Pienas 2.5% 1L | Norfa  | yes     |
    And replacement products exist:
      | product_id | master_id | store  | price |
      | 1002       | 100       | Maxima | 2.15  |
      | 1003       | 100       | IKI    | 1.99  |
      | 1004       | 100       | RIMI   | 2.05  |
    When the migration worker runs
    Then the item should migrate to product 1003 (cheapest)
    And the reason should be "no_same_store_available"

  Scenario: Migration across multiple user lists
    Given the following users have expired items:
      | user                | list_name       | expired_items |
      | jonas@example.com   | Weekly Shopping | 3             |
      | petras@example.com  | Daily Needs     | 2             |
      | marija@example.com  | Family List     | 5             |
    When the migration worker runs
    Then 10 items should be processed
    And each user should receive individual notification
    And the statistics should track per-user results

  Scenario: Batch migration with error handling
    Given 100 shopping list items need migration
    And 5 products have no replacements
    And 2 products have database errors
    When the migration worker runs
    Then 93 items should be successfully migrated
    And 5 items should be marked "needs_attention"
    And 2 errors should be logged
    And the system should continue processing despite errors

  Scenario: Scheduled automatic migration
    Given the migration worker is scheduled to run daily at 02:00
    And it's currently 02:00
    When the scheduled job triggers
    Then the migration should run automatically
    And statistics should be logged
    And affected users should be notified
    And the next run should be scheduled for tomorrow

  Scenario: Migration preserves custom item properties
    Given I have a shopping list item with:
      | field       | value              |
      | quantity    | 3                  |
      | notes       | "For breakfast"    |
      | priority    | high               |
      | category_id | 5                  |
    And the product expires
    When migration occurs
    Then the new product should maintain:
      | field       | value              |
      | quantity    | 3                  |
      | notes       | "For breakfast"    |
      | priority    | high               |
      | category_id | 5                  |
    And only product_id and price should change

  Scenario: Migration notification batching
    Given I have 15 expired items migrated
    When the migration completes
    Then I should receive 1 consolidated notification
    And the notification should summarize:
      | field              | value |
      | items_updated      | 15    |
      | price_changes      | 5     |
      | store_changes      | 3     |
      | needs_attention    | 0     |

  Scenario: Manual migration trigger
    Given I view my shopping list
    And 2 items show "Product expired" warning
    When I click "Update expired items"
    Then manual migration should trigger immediately
    And I should see real-time progress
    And upon completion, the list should refresh

  Scenario: Migration with price change notification
    Given my shopping list item costs "1.99"
    When migrated to new product costing "2.45"
    Then the price change should be highlighted
    And I should see "+€0.46 (23% increase)"
    And the notification should warn about significant price changes

  Scenario: Rollback failed migration
    Given migration starts for 10 items
    And item 5 causes a database error
    When the transaction fails
    Then all 10 items should remain unchanged
    And the error should be logged
    And no partial updates should persist
    And the system should retry later

  Scenario: Migration metrics and monitoring
    Given the migration worker has run 30 times this month
    When I request migration metrics
    Then I should see:
      | metric                  | value |
      | total_migrations        | 1250  |
      | success_rate            | 98.5% |
      | avg_items_per_run       | 41    |
      | items_needing_attention | 18    |
      | avg_processing_time     | 2.3s  |
