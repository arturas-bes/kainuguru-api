# Test Anti-Patterns Reference

## Critical Anti-Patterns

### 1. Mocking Application Code

**Description**: Mocking your own business logic, workflows, or handlers defeats the purpose of integration testing.

**Examples Across Languages:**

#### Go

```go
// ❌ WRONG - Mocking own workflow
type mockWorkflow struct{}
func (m *mockWorkflow) Execute(client APIClient) error {
    return nil // Always succeeds
}

// ✅ CORRECT - Use real workflow, mock external API
workflow := NewWorkflow(payload)
mockAPIClient := mocks.NewExternalAPIClient()
err := workflow.Execute(mockAPIClient)
```

#### PHP

```php
// ❌ WRONG - Mocking own service
$mockService = $this->createMock(UserService::class);
$mockService->method('createUser')->willReturn(true);

// ✅ CORRECT - Use real service, mock external dependency
$mockHttpClient = $this->createMock(HttpClientInterface::class);
$userService = new UserService($mockHttpClient);
$result = $userService->createUser($data);
```

#### JavaScript/TypeScript

```javascript
// ❌ WRONG - Mocking own repository
const mockRepo = jest.fn().mockResolvedValue({ id: 1 });

// ✅ CORRECT - Use real repository, mock database driver
const mockDbConnection = jest.mock("pg");
const repo = new UserRepository(mockDbConnection);
```

### 2. Log-Based Assertions

**Description**: Verifying log messages instead of actual behavior.

**Examples:**

```go
// ❌ WRONG
if !strings.Contains(logs, "workflow completed") {
    t.Fatal("workflow did not complete")
}

// ✅ CORRECT
// Verify actual side effects
calls := mockClient.GetCalls()
assert.Equal(t, 1, len(calls))
assert.Equal(t, expectedPayload, calls[0].Body)
```

```php
// ❌ WRONG
$this->assertStringContainsString('User created', $this->getLogOutput());

// ✅ CORRECT
$this->assertDatabaseHas('users', ['email' => 'test@example.com']);
$this->assertEquals(1, $mockMailer->getSentCount());
```

### 3. Tests That Cannot Fail

**Description**: Tests that always pass regardless of implementation.

**Examples:**

```go
// ❌ WRONG - No assertions
func TestProcess(t *testing.T) {
    processor := NewProcessor()
    processor.Process(data)
    // Test always passes!
}

// ✅ CORRECT - Meaningful assertions
func TestProcess(t *testing.T) {
    processor := NewProcessor()
    result, err := processor.Process(data)
    require.NoError(t, err)
    assert.Equal(t, expected, result)
    assert.True(t, processor.IsComplete())
}
```

### 4. Testing Test Infrastructure

**Description**: Writing tests that verify test helpers or test utilities work.

**Examples:**

```gherkin
# ❌ WRONG - Testing the test helper
Scenario: Queue helper creates queue
  Given I create a test queue named "test-queue"
  Then the queue should exist

# ✅ CORRECT - Testing actual functionality
Scenario: Message processor handles invalid messages
  Given a message with invalid JSON in the queue
  When the processor consumes the message
  Then the message should be moved to dead letter queue
```

### 5. Magic Value Contracts

**Description**: Implicit data contracts where test behavior depends on magic values.

**Examples:**

```go
// ❌ WRONG - Magic value determines behavior
// In test data: {"phone": "12345678900"} // Why this number?
// In mock:
if phone[len(phone)-1] == '0' {
    return "subscribed"
}

// ✅ CORRECT - Explicit configuration
const SubscribedPhone = "12345678900"
mock.ConfigurePhone(SubscribedPhone, SubscriptionStatus.Active)
```

### 6. Duplicate Test Cases

**Description**: Multiple tests verifying the exact same behavior.

**Examples:**

```go
// ❌ WRONG - Three tests doing the same thing
func TestUserCreation(t *testing.T) { ... }
func TestCreateUser(t *testing.T) { ... }     // Same as above
func TestNewUser(t *testing.T) { ... }        // Same as above

// ✅ CORRECT - One comprehensive test or table-driven test
func TestUserCreation(t *testing.T) {
    testCases := []struct{
        name string
        input User
        want Result
    }{
        {"valid user", validUser, success},
        {"invalid email", invalidEmail, errorResult},
        {"duplicate user", existingUser, duplicateError},
    }
    // ...
}
```

### 7. Over-Mocking

**Description**: Mocking components that can run in-process without issues.

**Examples:**

```go
// ❌ WRONG - Mocking simple in-memory components
mockCache := &MockInMemoryCache{}
mockRouter := &MockGinRouter{}

// ✅ CORRECT - Use real in-process components
cache := cache.NewInMemoryCache()
router := gin.New()
handler := NewHandler(cache, mockExternalAPI) // Only mock external
```

### 8. Hardcoded Fake Responses

**Description**: Test responses that don't reflect real behavior.

**Examples:**

```php
// ❌ WRONG - Always returning success
class FakePaymentGateway {
    public function charge($amount) {
        return ['status' => 'success', 'id' => '123'];
    }
}

// ✅ CORRECT - Configurable test double
class TestPaymentGateway {
    private $responses = [];

    public function willReturn($response) {
        $this->responses[] = $response;
    }

    public function charge($amount) {
        return array_shift($this->responses);
    }
}
```

## Common Patterns by Test Type

### BDD/Gherkin Tests

- ❌ Step definitions with no assertions
- ❌ Scenarios that test happy path only
- ❌ Given/When/Then that don't map to real actions
- ✅ Clear business scenarios with full coverage
- ✅ Reusable step definitions with proper assertions

### Unit Tests

- ❌ Testing private methods directly
- ❌ Tests coupled to implementation details
- ❌ No edge case coverage
- ✅ Testing public interface only
- ✅ Black-box testing approach
- ✅ Comprehensive edge case coverage

### Integration Tests

- ❌ Mocking infrastructure (DB, queues)
- ❌ Not testing error paths
- ❌ Tests requiring specific execution order
- ✅ Using test containers or local infrastructure
- ✅ Testing both success and failure scenarios
- ✅ Independent, isolated tests

## Language-Specific Anti-Patterns

### Go

- ❌ Not using `t.Parallel()` for independent tests
- ❌ Ignoring errors in tests
- ❌ Not using subtests for related scenarios
- ❌ Global variables in tests

### PHP

- ❌ Not using data providers for parametrized tests
- ❌ Database tests without transactions
- ❌ Not resetting global state between tests
- ❌ Testing framework internals

### JavaScript/TypeScript

- ❌ Not handling async properly in tests
- ❌ Modifying global objects (Date, Math)
- ❌ Tests with setTimeout/setInterval
- ❌ Not cleaning up after tests (listeners, timers)
