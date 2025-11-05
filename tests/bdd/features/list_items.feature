Feature: Shopping List Item Management
  As a registered user
  I want to add, edit, and manage items in my shopping lists
  So that I can track what I need to buy

  Background:
    Given I am a registered user with email "user@example.com"
    And I am logged in with valid credentials
    And I have a shopping list "Weekly Groceries"
    And there are stores with current flyers and products available

  Scenario: Add item by text description
    Given I am viewing "Weekly Groceries" list
    When I click "Add Item"
    And I enter text description "Organic milk 1L"
    And I set quantity to 2
    And I click "Add Item"
    Then I should see "Item added successfully"
    And I should see "Organic milk 1L" in the list
    And the quantity should show "2"
    And the item should be unchecked

  Scenario: Add item from product search
    Given I am viewing "Weekly Groceries" list
    When I click "Add from Products"
    And I search for "milk"
    And I select "Pienas 1L - Rimi" from search results
    And I set quantity to 1
    And I add notes "Low fat preferred"
    And I click "Add to List"
    Then I should see "Item added successfully"
    And I should see "Pienas 1L" in the list
    And the item should be linked to the Rimi product
    And notes should show "Low fat preferred"

  Scenario: Add item from current flyer
    Given I am viewing "Weekly Groceries" list
    And there is a current Maxima flyer with "Bread 2€"
    When I click "Add from Flyers"
    And I browse current flyers
    And I click "Add to List" on "Bread 2€" from Maxima flyer
    Then I should see "Item added successfully"
    And I should see "Bread" in the list
    And the item should show current price "2€"
    And the item should be linked to Maxima store

  Scenario: Edit item details
    Given I have an item "Milk 1L" with quantity 1 in my list
    When I click "Edit" on "Milk 1L"
    And I change the description to "Organic Milk 1L"
    And I change quantity to 2
    And I add notes "Brand: Žalia pieva"
    And I click "Save Changes"
    Then I should see "Item updated successfully"
    And I should see "Organic Milk 1L" in the list
    And quantity should show "2"
    And notes should show "Brand: Žalia pieva"

  Scenario: Check off completed items
    Given I have items in my list:
      | Description    | Quantity | Checked |
      | Milk 1L        | 1        | No      |
      | Bread          | 2        | No      |
      | Apples 1kg     | 1        | No      |
    When I check off "Milk 1L"
    Then "Milk 1L" should be marked as checked
    And the list should show "1 of 3 items completed"
    And checked items should appear with strikethrough

  Scenario: Uncheck completed items
    Given I have a checked item "Milk 1L" in my list
    When I uncheck "Milk 1L"
    Then "Milk 1L" should be marked as unchecked
    And the item should appear normally without strikethrough

  Scenario: Delete item from list
    Given I have an item "Old Item" in my list
    When I click "Delete" on "Old Item"
    And I confirm the deletion
    Then I should see "Item removed successfully"
    And "Old Item" should not appear in the list

  Scenario: Reorder items in list
    Given I have items in my list:
      | Description | Position |
      | Milk        | 1        |
      | Bread       | 2        |
      | Apples      | 3        |
    When I drag "Apples" to position 1
    Then the items should be reordered as:
      | Description | Position |
      | Apples      | 1        |
      | Milk        | 2        |
      | Bread       | 3        |

  Scenario: Bulk operations on items
    Given I have 5 items in my list, all unchecked
    When I select all items
    And I click "Mark All as Checked"
    Then all 5 items should be marked as checked
    And the list should show "5 of 5 items completed"

    When I click "Clear All Checked Items"
    And I confirm the action
    Then all checked items should be removed
    And the list should be empty

  Scenario: Item quantity validation
    When I try to add an item with quantity 0
    Then I should see "Quantity must be at least 1"
    And the item should not be added

    When I try to add an item with quantity over 999
    Then I should see "Quantity cannot exceed 999"
    And the item should not be added

  Scenario: Duplicate item handling
    Given I have an item "Milk 1L" with quantity 1 in my list
    When I try to add another item "Milk 1L"
    Then I should see "Item already exists. Would you like to increase quantity?"
    When I click "Yes, increase quantity"
    Then the existing item quantity should become 2
    And there should still be only one "Milk 1L" item

  Scenario: Product availability tracking
    Given I have an item "Seasonal Product" linked to a specific product
    And that product becomes unavailable
    When I view my shopping list
    Then "Seasonal Product" should be marked as "Currently unavailable"
    And I should see suggested alternatives if available

  Scenario: Price tracking for linked items
    Given I have an item "Bread" linked to a product with current price 1.50€
    And the product price changes to 1.80€
    When I view my shopping list
    Then I should see the updated price 1.80€
    And I should see a note "Price increased by 0.30€"

  Scenario: Smart item suggestions
    Given I type "mil" in the item description field
    Then I should see suggestions:
      | Suggestion           | Source        |
      | Milk 1L             | Previous items |
      | Milk 2% fat 1L      | Popular items  |
      | Pienas 1L - Rimi    | Current flyers |
    When I click on "Milk 1L" suggestion
    Then the description field should be filled with "Milk 1L"

  Scenario: Item notes and categories
    Given I am adding an item "Chicken breast 1kg"
    When I add notes "For dinner tomorrow"
    And I set category to "Meat & Fish"
    And I click "Add Item"
    Then the item should be added with the notes
    And it should be categorized under "Meat & Fish"
    And items in the list should be grouped by category