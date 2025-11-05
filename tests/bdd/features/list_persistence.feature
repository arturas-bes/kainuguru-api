Feature: Shopping List Data Persistence
  As a registered user
  I want my shopping lists and items to be reliably saved and synchronized
  So that I don't lose my data when switching devices or after network interruptions

  Background:
    Given I am a registered user with email "user@example.com"
    And I am logged in with valid credentials
    And there are stores with current flyers available

  Scenario: Shopping list auto-save functionality
    Given I am creating a new shopping list
    When I enter list name "Auto-save Test"
    And I wait for 2 seconds without clicking save
    Then the list should be automatically saved as draft
    And I should see "Draft saved" indicator
    When I add description "Test description"
    And I wait for 2 seconds
    Then the description should be auto-saved
    And the last saved timestamp should be updated

  Scenario: List item auto-save during editing
    Given I have a shopping list "Weekly Groceries"
    And I am adding an item "Milk 1L"
    When I set quantity to 2
    And I start typing notes "Low fat"
    And I wait for 2 seconds without finishing
    Then the partial item should be saved as draft
    When I complete the notes "Low fat preferred"
    And I click away from the item
    Then the complete item should be saved
    And there should be no draft data remaining

  Scenario: Offline data persistence
    Given I have a shopping list "Offline Test" with 3 items
    When my internet connection is lost
    And I add a new item "Offline Item"
    And I check off an existing item "Milk"
    And I edit another item to "Updated Bread"
    Then all changes should be stored locally
    And I should see "Changes saved locally" indicator
    When my internet connection is restored
    Then all local changes should sync to the server
    And I should see "Synchronized" confirmation
    And the server should have all my changes

  Scenario: Cross-device synchronization
    Given I have a shopping list "Sync Test" on device A
    And the list contains items:
      | Description | Quantity | Checked |
      | Milk        | 1        | No      |
      | Bread       | 2        | Yes     |
    When I open the same list on device B
    Then I should see the same items with correct states
    When I add "Apples" on device B
    And device A refreshes after 5 seconds
    Then device A should show the new "Apples" item
    And both devices should be in sync

  Scenario: Concurrent editing conflict resolution
    Given I have a shopping list "Conflict Test" open on two devices
    When I edit item "Milk" to "Organic Milk" on device A
    And simultaneously edit the same item to "Almond Milk" on device B
    And both devices save at the same time
    Then I should see a conflict resolution dialog
    And I should be able to choose which version to keep
    When I select "Almond Milk" as the final version
    Then both devices should show "Almond Milk"
    And the server should store "Almond Milk"

  Scenario: Data recovery after app crash
    Given I am editing a shopping list "Recovery Test"
    And I have unsaved changes:
      | Action | Item      | Detail        |
      | Add    | New Item  | Quantity: 3   |
      | Edit   | Old Item  | Changed name  |
      | Check  | Done Item | Marked done   |
    When the application crashes unexpectedly
    And I restart the application
    Then I should see a "Recover unsaved changes?" prompt
    When I click "Recover"
    Then all my unsaved changes should be restored
    And I should be able to continue editing

  Scenario: Large list performance and persistence
    Given I have a shopping list with 500 items
    When I add another item "Item 501"
    Then the item should be saved within 2 seconds
    And the UI should remain responsive
    When I scroll through the entire list
    Then all items should load smoothly
    And the application should not crash or slow down

  Scenario: Data integrity validation
    Given I have shopping list data that becomes corrupted
    When the system detects data inconsistency
    Then I should see "Data corruption detected" warning
    And the system should automatically backup current data
    And attempt to restore from last known good state
    When restoration is successful
    Then I should see "Data restored successfully"
    And my lists should be accessible again

  Scenario: Backup and restore functionality
    Given I have 5 shopping lists with a total of 50 items
    When I trigger a manual backup
    Then all my lists and items should be backed up to the cloud
    And I should see "Backup completed successfully"
    When I delete all my local data
    And I trigger a restore from backup
    Then all 5 shopping lists should be restored
    And all 50 items should be in their correct lists
    And all item states (checked/unchecked) should be preserved

  Scenario: Storage quota management
    Given my account is approaching storage limits
    When I try to add a new shopping list
    Then I should see "Storage almost full" warning
    And I should be offered options to:
      | Option                    | Description                    |
      | Delete old lists         | Remove lists older than 6 months |
      | Archive completed lists  | Move checked lists to archive  |
      | Upgrade storage         | Purchase additional storage    |
    When I select "Archive completed lists"
    Then completed lists should be moved to archive
    And I should be able to create new lists

  Scenario: List sharing persistence
    Given I share my list "Shared List" with "friend@example.com"
    When my friend adds an item "Friend's Item"
    And I go offline for 1 hour
    And come back online
    Then I should see "Friend's Item" in the shared list
    And I should see "Added by friend@example.com" attribution
    When I edit "Friend's Item" to "Updated Item"
    Then my friend should see the update on their device
    And the change history should show both our edits

  Scenario: Database transaction integrity
    Given I am performing multiple operations simultaneously:
      | Operation | Target           | Detail           |
      | Create    | New list        | "Transaction Test" |
      | Add       | Item to new list| "Test Item"      |
      | Update    | Existing list   | Change name      |
      | Delete    | Old list        | Remove "Old"     |
    When one operation fails due to network error
    Then all related operations should be rolled back
    And the database should remain in consistent state
    And I should see appropriate error messages
    When I retry the operations
    Then they should complete successfully

  Scenario: Migration and versioning
    Given my data was created with app version 1.0
    When I upgrade to app version 2.0 with new data structure
    Then my existing data should be automatically migrated
    And all my lists and items should remain accessible
    And no data should be lost during migration
    When migration completes
    Then I should see "Data updated successfully" confirmation
    And the app should function normally with migrated data

  Scenario: Real-time updates and notifications
    Given I have a shared list "Family Groceries"
    When a family member adds "Emergency Item" to the list
    Then I should receive a real-time notification
    And the item should appear immediately without refresh
    When multiple family members edit different items simultaneously
    Then all changes should be synchronized in real-time
    And I should see live indicators of who is editing what

  Scenario: Partial sync recovery
    Given I have local changes that failed to sync:
      | List Name      | Change Type | Status  |
      | Weekly List    | Item added  | Failed  |
      | Monthly List   | Item edited | Failed  |
      | Shopping List  | List renamed| Success |
    When I manually trigger sync
    Then only failed changes should be retried
    And successful changes should not be duplicated
    When all changes sync successfully
    Then sync status should show "All changes synchronized"
    And there should be no pending operations