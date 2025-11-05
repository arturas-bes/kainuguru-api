Feature: User Authentication
  As a registered user
  I want to login and logout
  So that I can access my account securely

  Background:
    Given the system is running
    And the database is clean
    And a user exists with email "test@example.com" and password "SecurePass123!"

  Scenario: Successful login
    Given I am not logged in
    When I login with email "test@example.com" and password "SecurePass123!"
    Then the login should be successful
    And I should receive a JWT token
    And I should receive user information
    And a session should be created
    And the session should be active
    And the user last login should be updated

  Scenario: Login with invalid email
    Given I am not logged in
    When I login with email "nonexistent@example.com" and password "SecurePass123!"
    Then the login should fail
    And I should receive an error about invalid credentials

  Scenario: Login with incorrect password
    Given I am not logged in
    When I login with email "test@example.com" and password "WrongPassword"
    Then the login should fail
    And I should receive an error about invalid credentials
    And no session should be created

  Scenario: Login with inactive user
    Given a user exists with email "inactive@example.com" and password "SecurePass123!"
    And the user is marked as inactive
    When I login with email "inactive@example.com" and password "SecurePass123!"
    Then the login should fail
    And I should receive an error about account being inactive

  Scenario: Login creates session with metadata
    Given I am not logged in
    When I login with email "test@example.com" and password "SecurePass123!"
    And I provide user agent "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
    And I provide IP address "192.168.1.100"
    And I provide device type "web"
    Then the login should be successful
    And a session should be created with the provided metadata
    And the session should have the correct IP address
    And the session should have the correct user agent
    And the session should have device type "web"

  Scenario: Login with browser information
    Given I am not logged in
    When I login with email "test@example.com" and password "SecurePass123!"
    And I provide browser information:
      | name     | Chrome    |
      | version  | 91.0.4472 |
      | platform | Windows   |
      | isMobile | false     |
    Then the login should be successful
    And the session should have the browser information stored

  Scenario: Successful logout
    Given I am logged in as "test@example.com"
    When I logout
    Then the logout should be successful
    And my session should be marked as inactive
    And my JWT token should be invalidated

  Scenario: Logout from specific session
    Given I am logged in as "test@example.com" on multiple devices
    When I logout from the current session
    Then the logout should be successful
    And only the current session should be marked as inactive
    And other sessions should remain active

  Scenario: Logout from all sessions
    Given I am logged in as "test@example.com" on multiple devices
    When I logout from all sessions
    Then the logout should be successful
    And all my sessions should be marked as inactive

  Scenario: Access with expired token
    Given I am logged in as "test@example.com"
    And my JWT token has expired
    When I try to access a protected resource
    Then the access should be denied
    And I should receive an error about expired token

  Scenario: Refresh token flow
    Given I am logged in as "test@example.com"
    And I have a refresh token
    When my access token expires
    And I use the refresh token to get a new access token
    Then I should receive a new valid access token
    And my session should remain active
    And the session last used time should be updated

  Scenario: Session timeout
    Given I am logged in as "test@example.com"
    And my session has been inactive for more than the timeout period
    When I try to access a protected resource
    Then the access should be denied
    And my session should be marked as expired

  Scenario: Multiple concurrent sessions
    Given I am logged in as "test@example.com" on a web browser
    When I login with the same credentials on a mobile device
    Then both logins should be successful
    And I should have multiple active sessions
    And each session should have different device information

  Scenario: Session management
    Given I am logged in as "test@example.com"
    When I view my active sessions
    Then I should see a list of my sessions
    And each session should show device information
    And each session should show login time
    And each session should show last activity time

  Scenario: Security alert on new device login
    Given I am logged in as "test@example.com" on my usual device
    When I login from a new device with different IP and browser
    Then the login should be successful
    And the session should be flagged for security review
    And the session should include location information if available