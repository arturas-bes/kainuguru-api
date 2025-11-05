Feature: Shopping List Management
  As a registered user
  I want to create and manage shopping lists
  So that I can organize my grocery shopping efficiently

  Background:
    Given I am a registered user with email "user@example.com"
    And I am logged in with valid credentials
    And there are stores with current flyers available

  Scenario: Create a new shopping list
    Given I am on the shopping lists page
    When I click "Create New List"
    And I enter list name "Weekly Groceries"
    And I enter description "My weekly grocery shopping list"
    And I click "Create List"
    Then I should see "List created successfully"
    And I should see "Weekly Groceries" in my shopping lists
    And the list should be empty with 0 items

  Scenario: Create default shopping list
    Given I have no shopping lists
    When I create a shopping list with name "My Default List"
    And I mark it as default
    Then it should be set as my default shopping list
    And when I navigate to shopping lists, it should be displayed first

  Scenario: View shopping list details
    Given I have a shopping list "Weekly Groceries" with 5 items
    When I click on "Weekly Groceries"
    Then I should see the list details page
    And I should see "Weekly Groceries" as the title
    And I should see "5 items" in the list summary
    And I should see all 5 items listed

  Scenario: Edit shopping list details
    Given I have a shopping list "Weekly Groceries"
    When I click "Edit List" for "Weekly Groceries"
    And I change the name to "Updated Weekly List"
    And I change the description to "Updated description"
    And I click "Save Changes"
    Then I should see "List updated successfully"
    And I should see "Updated Weekly List" in my shopping lists
    And the description should show "Updated description"

  Scenario: Delete empty shopping list
    Given I have an empty shopping list "Test List"
    When I click "Delete" for "Test List"
    And I confirm the deletion
    Then I should see "List deleted successfully"
    And "Test List" should not appear in my shopping lists

  Scenario: Delete shopping list with items
    Given I have a shopping list "Test List" with 3 items
    When I click "Delete" for "Test List"
    Then I should see a warning "This list contains 3 items. Are you sure?"
    When I confirm the deletion
    Then I should see "List deleted successfully"
    And "Test List" should not appear in my shopping lists
    And all 3 items should be removed from the database

  Scenario: Cannot delete default shopping list
    Given I have a default shopping list "My Default List"
    When I try to delete "My Default List"
    Then I should see an error "Cannot delete default shopping list"
    And "My Default List" should still appear in my shopping lists

  Scenario: View multiple shopping lists
    Given I have the following shopping lists:
      | Name               | Items | Default |
      | Weekly Groceries   | 5     | Yes     |
      | Party Supplies     | 8     | No      |
      | Emergency Items    | 2     | No      |
    When I view my shopping lists
    Then I should see all 3 shopping lists
    And "Weekly Groceries" should be marked as default
    And lists should be ordered with default first

  Scenario: Shopping list validation
    When I try to create a shopping list with empty name
    Then I should see "List name is required"
    And the list should not be created

    When I try to create a shopping list with name longer than 100 characters
    Then I should see "List name must be 100 characters or less"
    And the list should not be created

  Scenario: Shopping list privacy
    Given user "other@example.com" has a shopping list "Other's List"
    When I view my shopping lists
    Then I should not see "Other's List"
    And I should not be able to access "Other's List" directly

  Scenario: Maximum shopping lists limit
    Given I have 10 shopping lists (the maximum allowed)
    When I try to create another shopping list
    Then I should see "Maximum number of shopping lists reached (10)"
    And the list should not be created

  Scenario: Set new default shopping list
    Given I have shopping lists:
      | Name               | Default |
      | Weekly Groceries   | Yes     |
      | Party Supplies     | No      |
    When I set "Party Supplies" as default
    Then "Party Supplies" should be marked as default
    And "Weekly Groceries" should no longer be default
    And I should see "Default list updated successfully"