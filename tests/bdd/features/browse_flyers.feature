Feature: Browse Current Weekly Flyers
  As a Lithuanian grocery shopper
  I want to browse current weekly flyers from all major stores
  So that I can find deals and plan my shopping

  Background:
    Given the system has current flyers from major stores
    And today is within a valid flyer period

  Scenario: View list of current flyers
    When I request the current flyers list
    Then I should see flyers from active stores
    And each flyer should show store name, valid dates, and page count
    And flyers should be sorted by store popularity

  Scenario: View flyers only for current week
    Given there are flyers for current week and previous week
    When I request the current flyers list
    Then I should only see current week flyers
    And I should not see expired flyers

  Scenario: Handle no current flyers gracefully
    Given there are no current flyers available
    When I request the current flyers list
    Then I should receive an empty list
    And I should get a helpful message about checking back later

  Scenario: View flyer details
    Given there is a current flyer for "IKI" store
    When I request details for that flyer
    Then I should see the flyer information
    And I should see the valid date range
    And I should see the number of pages available

  Scenario: Access flyers without authentication
    Given I am not logged in
    When I request the current flyers list
    Then I should successfully receive the flyers
    And no authentication should be required

  Scenario: System performance requirements
    Given there are flyers from all major Lithuanian stores
    When I request the current flyers list
    Then the response should be returned within 500ms
    And the API should handle concurrent requests efficiently