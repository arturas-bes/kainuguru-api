Feature: Public Access to Flyers
  As any visitor to the Kainuguru website
  I want to access flyers without creating an account
  So that I can quickly browse deals without barriers

  Background:
    Given the system has current flyers available
    And I am not authenticated

  Scenario: Browse flyers without login
    When I visit the flyers page
    Then I should see all current flyers
    And I should not see any login prompts
    And I should be able to view all flyer content

  Scenario: View flyer pages without login
    Given there is a current flyer from "Rimi"
    When I view any page of the Rimi flyer
    Then I should see the full page content
    And I should see extracted product information
    And no authentication should be required

  Scenario: Search products without login
    Given there are products available in current flyers
    When I search for "pienas" (milk in Lithuanian)
    Then I should see relevant products
    And I should see price and store information
    And no authentication should be required

  Scenario: Rate limiting for anonymous users
    Given I am making requests as an anonymous user
    When I make more than 100 requests per minute
    Then I should be rate limited
    And I should receive appropriate error messages
    But previous requests should have worked normally

  Scenario: Access health check without authentication
    When I check the system health endpoint
    Then I should receive system status
    And no authentication should be required

  Scenario: Performance with anonymous traffic
    Given there are 50 concurrent anonymous users
    When they all browse flyers simultaneously
    Then all users should receive responses within 500ms
    And the system should remain stable
    And no user should be denied access due to authentication

  Scenario: CORS support for web browsers
    Given I am accessing from a web browser
    When I make requests to the API from a different domain
    Then CORS headers should be properly set
    And browser requests should succeed
    And preflight requests should be handled correctly