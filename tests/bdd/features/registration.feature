Feature: User Registration
  As a new user
  I want to register an account
  So that I can access personalized features

  Background:
    Given the system is running
    And the database is clean

  Scenario: Successful user registration
    Given I am not logged in
    When I register with email "test@example.com" and password "SecurePass123!"
    And I provide full name "Jonas Lietuvonis"
    And I set preferred language to "lt"
    Then the registration should be successful
    And I should receive a user ID
    And the user should not be verified initially
    And the user should have default preferences
    And the user should be active

  Scenario: Registration with duplicate email
    Given a user exists with email "existing@example.com"
    When I register with email "existing@example.com" and password "SecurePass123!"
    Then the registration should fail
    And I should receive an error about duplicate email

  Scenario: Registration with invalid email
    Given I am not logged in
    When I register with email "invalid-email" and password "SecurePass123!"
    Then the registration should fail
    And I should receive an error about invalid email format

  Scenario: Registration with weak password
    Given I am not logged in
    When I register with email "test@example.com" and password "123"
    Then the registration should fail
    And I should receive an error about password requirements

  Scenario: Registration with Lithuanian preferences
    Given I am not logged in
    When I register with email "vilnius@example.lt" and password "SecurePass123!"
    And I set preferred language to "lt"
    And I set full name "Antanas Vilniaus"
    Then the registration should be successful
    And the user preferred language should be "lt"
    And the user full name should be "Antanas Vilniaus"

  Scenario: Registration with custom metadata
    Given I am not logged in
    When I register with email "custom@example.com" and password "SecurePass123!"
    And I set custom preferences:
      | theme              | dark  |
      | currency           | EUR   |
      | emailNotifications | false |
    Then the registration should be successful
    And the user should have the custom preferences

  Scenario: Registration creates default metadata
    Given I am not logged in
    When I register with email "default@example.com" and password "SecurePass123!"
    Then the registration should be successful
    And the user should have default preferences:
      | theme              | light |
      | currency           | EUR   |
      | emailNotifications | true  |
      | profilePublic      | false |

  Scenario: Registration without optional fields
    Given I am not logged in
    When I register with email "minimal@example.com" and password "SecurePass123!"
    Then the registration should be successful
    And the user full name should be null
    And the user preferred language should be "lt"

  Scenario: Registration with OAuth readiness
    Given I am not logged in
    When I register with email "oauth-ready@example.com" and password "SecurePass123!"
    Then the registration should be successful
    And the user OAuth provider should be null
    And the user OAuth ID should be null
    And the user should be ready for future OAuth linking

  Scenario: Password is properly hashed
    Given I am not logged in
    When I register with email "secure@example.com" and password "SecurePass123!"
    Then the registration should be successful
    And the password should be hashed using bcrypt
    And the plain password should not be stored
    And the password hash should be different from the plain password