Feature: Email Notifications System
  As a user
  I want to receive email notifications for important events
  So that I stay informed about my account and shopping activities

  Background:
    Given the email service is configured with SMTP
    And email templates are loaded

  Scenario: Welcome email on registration
    When a new user registers with email "jonas@example.com"
    Then a welcome email should be sent to "jonas@example.com"
    And the email should contain:
      | field        | value                          |
      | subject      | Welcome to Kainuguru!          |
      | greeting     | Hello, Jonas!                  |
      | content      | Thank you for joining          |
      | call_to_action | Start browsing flyers        |
    And the email should be in Lithuanian language
    And the email template should be "welcome"

  Scenario: Email verification after registration
    When a new user registers with email "petras@example.com"
    Then a verification email should be sent
    And the email should contain a verification link
    And the link should be valid for 24 hours
    And clicking the link should verify the account

  Scenario: Password reset email
    Given a user exists with email "marija@example.com"
    When the user requests a password reset
    Then a reset email should be sent to "marija@example.com"
    And the email should contain:
      | field          | value                             |
      | subject        | Reset Your Password               |
      | reset_link     | https://app.com/reset?token=...   |
      | expiry_info    | Valid for 1 hour                  |
      | security_note  | Ignore if you didn't request this |
    And the reset token should expire in 1 hour

  Scenario: Price alert notification
    Given user "jonas@example.com" has a price alert for "Pienas 2.5% 1L" at "1.85"
    And the current price drops to "1.79" at "Maxima"
    When the price alert checker runs
    Then an alert email should be sent
    And the email should contain:
      | field         | value                        |
      | subject       | Price Alert: Pienas 2.5% 1L |
      | product_name  | Pienas 2.5% 1L              |
      | target_price  | €1.85                        |
      | current_price | €1.79                        |
      | store         | Maxima                       |
      | savings       | Save €0.06                   |

  Scenario: Shopping list migration notification
    Given user "jonas@example.com" has 5 expired items in their shopping list
    When the migration worker completes
    Then a migration summary email should be sent
    And the email should contain:
      | field              | value                          |
      | subject            | Your Shopping List Updated     |
      | items_updated      | 5 items                        |
      | price_changes      | 2 price increases, 1 decrease  |
      | action_required    | None                           |

  Scenario: Weekly deals digest
    Given user "jonas@example.com" has opted into weekly digests
    And new flyers were published this week from 3 stores
    When the weekly digest job runs on Monday
    Then a digest email should be sent
    And the email should contain:
      | field              | value                      |
      | subject            | This Week's Best Deals     |
      | new_flyers         | 3 stores                   |
      | top_deals          | List of 10 best discounts  |
      | personalized       | Based on your interests    |

  Scenario: Email delivery retry on failure
    Given an email sending fails with "Connection timeout"
    When the failure is detected
    Then the email should be queued for retry
    And retry should occur with exponential backoff
    And maximum 3 retry attempts should be made
    And if all retries fail, admin should be notified

  Scenario: Unsubscribe from notifications
    Given user "jonas@example.com" receives marketing emails
    When the user clicks "Unsubscribe" link
    Then the user preferences should be updated
    And no more marketing emails should be sent
    And a confirmation email should be sent
    And transactional emails should still be sent

  Scenario: Batch email sending for multiple users
    Given 100 users have price alerts triggered
    When the notification job runs
    Then emails should be sent in batches of 10
    And sending should respect rate limits
    And the process should complete within 5 minutes
    And all delivery statuses should be tracked

  Scenario: Email template customization
    Given the welcome email template exists
    When an admin updates the template with new content
    Then new registrations should use the updated template
    And the template should support variables:
      | variable      | example_value   |
      | {{user_name}} | Jonas           |
      | {{app_url}}   | https://app.com |
      | {{logo_url}}  | https://cdn...  |

  Scenario: Multi-language email support
    Given user "john@example.com" has language preference "English"
    And user "jonas@example.com" has language preference "Lithuanian"
    When both users register
    Then "john@example.com" should receive email in English
    And "jonas@example.com" should receive email in Lithuanian
    And subjects should be in respective languages

  Scenario: Email open and click tracking
    Given user "jonas@example.com" receives a price alert email
    When the user opens the email
    Then an open event should be tracked
    And when the user clicks "View Product"
    Then a click event should be tracked
    And analytics should record engagement

  Scenario: Failed email notification alerting
    Given 10 consecutive emails fail to send
    When the failure threshold is reached
    Then an alert should be sent to system admins
    And the email service should be marked as degraded
    And health check endpoint should reflect the issue

  Scenario: Email delivery status verification
    Given an email was sent to "jonas@example.com"
    When I check the email status
    Then I should see:
      | field          | value               |
      | status         | delivered           |
      | sent_at        | 2025-11-08 10:30    |
      | delivered_at   | 2025-11-08 10:31    |
      | opens          | 1                   |
      | clicks         | 0                   |

  Scenario: Email sending during system maintenance
    Given the system is in maintenance mode
    When critical emails need to be sent
    Then transactional emails should still be delivered
    And marketing emails should be queued
    And queued emails should send after maintenance

  Scenario: Personalized recommendations email
    Given user "jonas@example.com" frequently buys dairy products
    And new dairy deals are available
    When the personalization engine runs
    Then a personalized deals email should be sent
    And the email should prioritize dairy products
    And include "Recommended for you" section
