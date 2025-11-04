# Good Testing Patterns Reference

## Core Principles

1. **Test Behavior, Not Implementation**: Focus on what the code does, not how it does it
2. **Only Mock What You Don't Own**: Mock external dependencies, use real code for everything else
3. **Verify Outcomes, Not Logs**: Check actual side effects and state changes
4. **Design for Failure**: Every test must be able to fail when code is broken
5. **Keep Tests Independent**: Each test runs in isolation without order dependencies

## Universal Good Patterns

### 1. Behavioral Verification Pattern

**Description**: Verify actual outcomes and side effects rather than implementation details.

#### Go Example

```go
func TestPasswordResetSendsEmail(t *testing.T) {
    // Arrange
    mockEmailClient := mocks.NewEmailClient()
    service := auth.NewPasswordResetService(mockEmailClient)

    // Act
    err := service.ResetPassword("user@example.com")

    // Assert - Verify behavior
    require.NoError(t, err)

    calls := mockEmailClient.GetCalls()
    assert.Equal(t, 1, len(calls), "should send exactly one email")
    assert.Equal(t, "user@example.com", calls[0].To)
    assert.Contains(t, calls[0].Body, "reset-token")
    assert.Equal(t, "Password Reset", calls[0].Subject)
}
```

#### PHP Example

```php
public function testOrderProcessingSendsConfirmation() {
    // Arrange
    $mockMailer = $this->createMock(MailerInterface::class);
    $mockPayment = $this->createMock(PaymentGatewayInterface::class);

    $mockMailer->expects($this->once())
        ->method('send')
        ->with($this->callback(function($email) {
            return $email->getTo() === 'customer@example.com' &&
                   $email->getSubject() === 'Order Confirmation';
        }));

    $orderService = new OrderService($mockMailer, $mockPayment);

    // Act
    $order = $orderService->processOrder($orderData);

    // Assert - Verify side effects
    $this->assertDatabaseHas('orders', [
        'id' => $order->id,
        'status' => 'confirmed'
    ]);
}
```

### 2. Test Data Builder Pattern

**Description**: Use builders to create test data clearly and maintainably.

#### Go Example

```go
type UserBuilder struct {
    user User
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        user: User{
            ID:     uuid.New(),
            Status: "active",
        },
    }
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.user.Email = email
    return b
}

func (b *UserBuilder) WithStatus(status string) *UserBuilder {
    b.user.Status = status
    return b
}

func (b *UserBuilder) Build() User {
    return b.user
}

// Usage in tests
func TestUserDeactivation(t *testing.T) {
    activeUser := NewUserBuilder().
        WithEmail("test@example.com").
        WithStatus("active").
        Build()

    service := NewUserService()
    err := service.Deactivate(activeUser)

    assert.NoError(t, err)
}
```

### 3. Table-Driven Tests Pattern

**Description**: Use data tables to test multiple scenarios efficiently.

#### Go Example

```go
func TestCalculateDiscount(t *testing.T) {
    testCases := []struct {
        name     string
        amount   float64
        customer CustomerType
        want     float64
        wantErr  bool
    }{
        {
            name:     "regular customer small order",
            amount:   50.00,
            customer: CustomerRegular,
            want:     50.00,
            wantErr:  false,
        },
        {
            name:     "premium customer large order",
            amount:   500.00,
            customer: CustomerPremium,
            want:     450.00, // 10% discount
            wantErr:  false,
        },
        {
            name:     "negative amount",
            amount:   -10.00,
            customer: CustomerRegular,
            want:     0,
            wantErr:  true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            got, err := CalculateDiscount(tc.amount, tc.customer)

            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.want, got)
            }
        })
    }
}
```

#### PHP Example

```php
/**
 * @dataProvider discountCalculationProvider
 */
public function testCalculateDiscount($amount, $customerType, $expected) {
    $calculator = new DiscountCalculator();
    $result = $calculator->calculate($amount, $customerType);
    $this->assertEquals($expected, $result);
}

public function discountCalculationProvider() {
    return [
        'regular customer' => [50.00, 'regular', 50.00],
        'premium customer' => [500.00, 'premium', 450.00],
        'vip customer' => [100.00, 'vip', 80.00],
    ];
}
```

### 4. Real Infrastructure Testing Pattern

**Description**: Use real infrastructure or proper test doubles, not mocks.

#### Docker Compose Example

```yaml
version: "3.8"
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5432:5432"

  localstack:
    image: localstack/localstack
    environment:
      SERVICES: sqs,s3
    ports:
      - "4566:4566"

  redis:
    image: redis:7
    ports:
      - "6379:6379"
```

#### Go Integration Test

```go
func TestMessageProcessingIntegration(t *testing.T) {
    // Use real infrastructure
    db := setupTestDatabase(t)
    defer db.Close()

    sqsClient := setupLocalStackSQS(t)
    queueURL := createTestQueue(t, sqsClient)

    // Real components, not mocks
    repo := repository.NewUserRepository(db)
    queue := messaging.NewSQSQueue(sqsClient, queueURL)
    processor := NewMessageProcessor(repo, queue)

    // Test real flow
    message := createTestMessage()
    err := queue.Send(message)
    require.NoError(t, err)

    err = processor.ProcessNext()
    require.NoError(t, err)

    // Verify side effects
    user, err := repo.FindByID(message.UserID)
    assert.NoError(t, err)
    assert.Equal(t, "processed", user.Status)

    // Verify message consumed
    messages := queue.Receive(1)
    assert.Empty(t, messages, "message should be deleted after processing")
}
```

### 5. Error Testing Pattern

**Description**: Always test error scenarios and edge cases.

```go
func TestServiceErrorHandling(t *testing.T) {
    t.Run("handles API timeout", func(t *testing.T) {
        mockAPI := mocks.NewAPIClient()
        mockAPI.SimulateTimeout()

        service := NewService(mockAPI)
        err := service.Process()

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "timeout")

        // Verify graceful degradation
        assert.True(t, service.IsInDegradedMode())
    })

    t.Run("handles malformed response", func(t *testing.T) {
        mockAPI := mocks.NewAPIClient()
        mockAPI.ReturnMalformedJSON()

        service := NewService(mockAPI)
        err := service.Process()

        assert.Error(t, err)
        assert.IsType(t, &json.SyntaxError{}, errors.Unwrap(err))
    })

    t.Run("retries on transient errors", func(t *testing.T) {
        mockAPI := mocks.NewAPIClient()
        mockAPI.FailNTimes(2) // Fail twice, then succeed

        service := NewService(mockAPI)
        err := service.Process()

        assert.NoError(t, err)
        assert.Equal(t, 3, mockAPI.CallCount())
    })
}
```

## BDD/Gherkin Good Patterns

### Feature File Structure

```gherkin
Feature: Password Reset
  As a user
  I want to reset my password
  So that I can regain access to my account

  Background:
    Given the email service is available
    And the user "test@example.com" exists

  Scenario: Successful password reset
    When I request a password reset for "test@example.com"
    Then a reset email should be sent to "test@example.com"
    And the email should contain a valid reset token
    And the token should expire in 1 hour

  Scenario: Password reset for non-existent user
    When I request a password reset for "nonexistent@example.com"
    Then no email should be sent
    And I should receive a generic success message

  Scenario: Rate limiting password resets
    Given I have requested 3 password resets in the last hour
    When I request another password reset
    Then I should receive a rate limit error
    And no additional email should be sent
```

### Step Definition Pattern

```go
type passwordResetContext struct {
    mockEmailClient *mocks.EmailClient
    userRepo        *repository.UserRepository
    resetService    *auth.PasswordResetService
    lastError       error
}

func (ctx *passwordResetContext) iRequestPasswordResetFor(email string) error {
    ctx.lastError = ctx.resetService.RequestReset(email)
    return nil // Don't fail the step here, check in Then
}

func (ctx *passwordResetContext) emailShouldBeSentTo(email string) error {
    calls := ctx.mockEmailClient.GetEmailsSentTo(email)
    if len(calls) == 0 {
        return fmt.Errorf("no email sent to %s", email)
    }
    return nil
}

func (ctx *passwordResetContext) emailShouldContainValidResetToken() error {
    calls := ctx.mockEmailClient.GetCalls()
    if len(calls) == 0 {
        return errors.New("no emails sent")
    }

    token := extractToken(calls[0].Body)
    if !isValidJWT(token) {
        return errors.New("invalid reset token format")
    }

    claims, err := validateToken(token)
    if err != nil {
        return fmt.Errorf("invalid token: %w", err)
    }

    if claims.Type != "password_reset" {
        return fmt.Errorf("wrong token type: %s", claims.Type)
    }

    return nil
}
```

## Testing Pyramid

### Unit Tests (70%)

- Test individual functions/methods
- Mock external dependencies only
- Fast execution (<100ms per test)
- Focus on edge cases and error handling

### Integration Tests (20%)

- Test components working together
- Use real infrastructure (containerized)
- Verify data flow through system
- Test configuration and setup

### End-to-End Tests (10%)

- Test complete user journeys
- Run against deployed environment
- Verify business requirements
- Smoke test critical paths
