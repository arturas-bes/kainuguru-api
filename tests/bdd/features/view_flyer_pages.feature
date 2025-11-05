Feature: View Flyer Pages
  As a user
  I want to view individual pages of grocery flyers
  So that I can see detailed product information and images

  Background:
    Given the system has the following stores:
      | name    | code   | enabled |
      | IKI     | iki    | true    |
      | Maxima  | maxima | true    |
      | Rimi    | rimi   | true    |
    And there are current flyers with multiple pages for all stores

  Scenario: View all pages for a specific flyer
    Given there is a flyer with ID 1 that has 8 pages
    When I request all pages for flyer ID 1 via GraphQL
    Then I should see 8 flyer pages
    And each page should have id, flyer_id, page_number, and image_url
    And pages should be ordered by page_number from 1 to 8
    And all pages should belong to flyer ID 1

  Scenario: View specific flyer page by ID
    Given there is a flyer page with ID 1
    When I request flyer page details for ID 1 via GraphQL
    Then I should see the complete flyer page information
    And it should include flyer details
    And it should include the page image URL
    And the image URL should be accessible

  Scenario: View products on specific flyer page
    Given there is a flyer page with ID 1 that contains products
    When I request products for flyer page ID 1 via GraphQL
    Then I should see products associated with that page
    And each product should have name, price, and flyer_page_id
    And products should include complete product information

  Scenario: Navigate between flyer pages
    Given there is a flyer page with page_number 3 out of 8 total pages
    When I request flyer page details for that page via GraphQL
    Then I should see navigation information
    And I should know this is page 3 of 8
    And I should be able to determine next and previous pages

  Scenario: Handle missing flyer pages
    Given there is a flyer with ID 999 that has no pages
    When I request pages for flyer ID 999 via GraphQL
    Then I should see an empty pages list
    And the response should be successful
    And no error should be returned

  Scenario: View flyer pages with filters
    Given there are multiple flyers with pages
    When I request pages for flyer_id 1 via GraphQL
    Then I should see only pages belonging to flyer ID 1
    And pages should be filtered correctly
    When I request pages ordered by page_number DESC via GraphQL
    Then pages should be in reverse order

  Scenario: Pagination of flyer pages
    Given there is a flyer with more than 10 pages
    When I request the first 5 pages via GraphQL
    Then I should see exactly 5 pages
    And I should receive pagination information
    When I request the next 5 pages using pagination cursor
    Then I should see the next 5 pages
    And page numbering should continue correctly

  Scenario: View Lithuanian product names correctly
    Given there are pages with products containing Lithuanian characters "ąčęėįšųūž"
    When I request those pages via GraphQL
    Then I should see Lithuanian characters displayed correctly
    And product names should be properly encoded in UTF-8
    And search should work with both diacritics and without

  Scenario: Access flyer pages without authentication
    Given I am not logged in
    When I request flyer pages via GraphQL
    Then I should successfully receive the pages
    And no authentication should be required
    And all page details should be accessible

  Scenario: Flyer page performance requirements
    Given there are flyers with many pages
    When I request pages for a flyer via GraphQL
    Then the response should be returned within 300ms
    And image URLs should be properly formatted
    And the GraphQL query should be optimized

  Scenario: Invalid flyer page requests
    When I request flyer page details for ID -1 via GraphQL
    Then I should receive an appropriate error message
    And the error should be handled gracefully
    When I request pages for flyer_id "invalid" via GraphQL
    Then I should receive a validation error