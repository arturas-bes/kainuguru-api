Feature: Password Reset
  As a user who forgot their password
  I want to reset my password securely
  So that I can regain access to my account

  Background:
    Given the system is running
    And the database is clean
    And a user exists with email "forgetful@example.com" and password "OldPassword123!"

  Scenario: Request password reset
    Given I am not logged in
    When I request a password reset for "forgetful@example.com"
    Then the request should be successful
    And a password reset token should be generated
    And the token should have an expiration time
    And I should receive a confirmation message
    # Note: In real implementation, an email would be sent

  Scenario: Request password reset for non-existent user
    Given I am not logged in
    When I request a password reset for "nonexistent@example.com"
    Then the request should appear successful for security reasons
    But no password reset token should be generated
    And I should receive the same confirmation message

  Scenario: Request password reset for inactive user
    Given a user exists with email "inactive@example.com" and password "OldPassword123!"
    And the user is marked as inactive
    When I request a password reset for "inactive@example.com"
    Then the request should appear successful for security reasons
    But no password reset token should be generated

  Scenario: Complete password reset with valid token
    Given I have requested a password reset for "forgetful@example.com"
    And I have a valid reset token
    When I reset my password to "NewPassword456!" using the token
    Then the password reset should be successful
    And my password should be updated
    And the reset token should be invalidated
    And all existing sessions should be terminated
    And I should be able to login with the new password

  Scenario: Password reset with expired token
    Given I have requested a password reset for "forgetful@example.com"
    And I have a reset token that has expired
    When I try to reset my password to "NewPassword456!" using the expired token
    Then the password reset should fail
    And I should receive an error about the token being expired
    And my password should remain unchanged

  Scenario: Password reset with invalid token
    Given I am not logged in
    When I try to reset my password to "NewPassword456!" using an invalid token
    Then the password reset should fail
    And I should receive an error about the token being invalid

  Scenario: Password reset with used token
    Given I have requested a password reset for "forgetful@example.com"
    And I have successfully used a reset token
    When I try to use the same token again
    Then the password reset should fail
    And I should receive an error about the token being invalid

  Scenario: Password reset with weak new password
    Given I have requested a password reset for "forgetful@example.com"
    And I have a valid reset token
    When I reset my password to "123" using the token
    Then the password reset should fail
    And I should receive an error about password requirements
    And my password should remain unchanged

  Scenario: Multiple password reset requests
    Given I have requested a password reset for "forgetful@example.com"
    When I request another password reset for the same email
    Then the new request should be successful
    And the previous reset token should be invalidated
    And only the latest token should be valid

  Scenario: Password reset token expiration time
    Given I have requested a password reset for "forgetful@example.com"
    Then the reset token should expire in 1 hour
    And the token should include creation timestamp
    And the token should be securely hashed in the database

  Scenario: Password reset security considerations
    Given I have requested a password reset for "forgetful@example.com"
    And I have a valid reset token
    When I complete the password reset
    Then all active sessions should be terminated
    And all refresh tokens should be invalidated
    And the user should be forced to login again
    And the password change should be logged for security audit

  Scenario: Rate limiting password reset requests
    Given I have made multiple password reset requests
    When I try to make another request immediately
    Then the request should be rate limited
    And I should receive an error about too many requests
    And I should be told to wait before trying again

  Scenario: Password reset confirmation flow
    Given I have requested a password reset for "forgetful@example.com"
    When I provide the reset token and new password
    And I confirm the new password
    Then both passwords should match
    And the reset should be successful
    And I should receive a confirmation of the password change

  Scenario: Verify old password cannot be used after reset
    Given I have successfully reset my password from "OldPassword123!" to "NewPassword456!"
    When I try to login with the old password "OldPassword123!"
    Then the login should fail
    And I should receive an error about invalid credentials

  Scenario: Login after successful password reset
    Given I have successfully reset my password to "NewPassword456!"
    When I login with email "forgetful@example.com" and password "NewPassword456!"
    Then the login should be successful
    And I should receive a JWT token
    And a new session should be created