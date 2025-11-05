Feature: View Individual Flyer Pages
  As a Lithuanian grocery shopper
  I want to view individual pages of weekly flyers
  So that I can see product details and prices

  Background:
    Given there is a current flyer for "Maxima" with 8 pages
    And the flyer has products extracted on some pages

  Scenario: View flyer page with products
    When I request page 1 of the Maxima flyer
    Then I should see the page image
    And I should see extracted products for that page
    And each product should have name, price, and unit information

  Scenario: View flyer page without products extracted
    Given page 3 of the Maxima flyer has no extracted products
    When I request page 3 of the Maxima flyer
    Then I should see the page image
    And I should see a message that products are being processed
    And the page should still be accessible

  Scenario: View flyer page with failed extraction
    Given page 5 of the Maxima flyer has failed product extraction
    When I request page 5 of the Maxima flyer
    Then I should see the page image
    And I should not see any extracted products
    And the page should still be accessible for manual viewing

  Scenario: Navigate between flyer pages
    When I request page 2 of the Maxima flyer
    Then I should see navigation information
    And I should see that this is page 2 of 8
    And I should see options to view previous and next pages

  Scenario: Handle invalid page number
    When I request page 15 of the Maxima flyer
    Then I should receive an error
    And the error should indicate the valid page range is 1-8

  Scenario: View Lithuanian product names correctly
    Given page 1 has products with Lithuanian names containing "ą, č, ę, ė, į, š, ų, ū, ž"
    When I request page 1 of the Maxima flyer
    Then I should see Lithuanian characters displayed correctly
    And product names should be properly encoded in UTF-8

  Scenario: Access flyer pages without authentication
    Given I am not logged in
    When I request any page of the Maxima flyer
    Then I should successfully receive the page
    And no authentication should be required