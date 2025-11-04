# Severity Classification Guide

## Overview

This guide defines severity levels for test compliance violations. Use these classifications when generating validation reports to ensure consistent and actionable feedback.

## Severity Levels

### CRITICAL - Test Validity Compromised

**Definition**: Violations that make tests worthless or actively harmful by providing false confidence.

**Action Required**: Must be fixed immediately. Tests should not be merged with critical violations.

**Examples**:

1. **Mocked Application Code**

   - Mocking your own workflows, services, or business logic
   - Test passes even when actual code is completely broken
   - Impact: Zero confidence in implementation correctness

2. **Log-Based Assertions**

   - Primary verification through log messages
   - No verification of actual behavior or side effects
   - Impact: Implementation can be broken while logs remain

3. **Tests That Cannot Fail**

   - No assertions or only tautological assertions
   - Hardcoded success responses
   - Impact: Test provides no value, cannot catch regressions

4. **Fake Test Infrastructure**
   - Mocking infrastructure components (DB, queues, caches)
   - Not using real or proper test doubles (LocalStack, TestContainers)
   - Impact: Integration issues go undetected

### HIGH - Significant Quality Issues

**Definition**: Violations that significantly reduce test effectiveness or reliability.

**Action Required**: Should be fixed before merge. Exceptions require explicit justification.

**Examples**:

1. **Missing Error Scenarios**

   - Only testing happy path
   - No verification of error handling
   - Impact: Production failures from unhandled errors

2. **Test Dependencies**

   - Tests require specific execution order
   - Shared state between tests
   - Impact: Flaky tests, false positives/negatives

3. **Incomplete Coverage**

   - Missing critical edge cases
   - Not testing boundary conditions
   - Impact: Bugs in edge cases go undetected

4. **Implicit Contracts**
   - Magic values without explanation
   - Hidden dependencies on test data
   - Impact: Tests become unmaintainable, brittle

### MEDIUM - Code Quality Issues

**Definition**: Violations that affect maintainability, readability, or efficiency.

**Action Required**: Should be addressed but not blocking. Fix in follow-up if needed.

**Examples**:

1. **Dead Code**

   - Unused helper functions
   - Commented-out tests
   - Impact: Confusion, maintenance burden

2. **Poor Naming**

   - Test names don't describe what they verify
   - Generic names like "TestWorkflow1", "TestWorkflow2"
   - Impact: Hard to understand test purpose

3. **Redundant Comments**

   - Comments stating the obvious
   - Outdated comments
   - Impact: Noise, potential confusion

4. **Code Duplication**
   - Repeated test setup
   - Copy-pasted assertions
   - Impact: Maintenance burden, inconsistency risk

### LOW - Style and Convention Issues

**Definition**: Minor issues that don't affect test functionality but violate conventions.

**Action Required**: Optional fixes. Can be addressed during refactoring.

**Examples**:

1. **Formatting Issues**

   - Inconsistent indentation
   - Line length violations
   - Impact: Reduced readability

2. **Missing Documentation**

   - Complex tests without explanatory comments
   - No description of test purpose
   - Impact: Harder to understand intent

3. **Naming Conventions**
   - Not following language conventions
   - Inconsistent naming patterns
   - Impact: Reduced consistency

## Decision Matrix

Use this matrix to determine severity when multiple factors are present:

| Violation Type | Can Test Fail? | Verifies Behavior? | Uses Real Code? | Severity |
| -------------- | -------------- | ------------------ | --------------- | -------- |
| Any            | No             | -                  | -               | CRITICAL |
| Any            | Yes            | No                 | -               | CRITICAL |
| Any            | Yes            | Yes                | No              | CRITICAL |
| Any            | Yes            | Yes                | Yes             | Review\* |

\*Review for HIGH/MEDIUM/LOW based on other factors

## Severity Escalation

Certain combinations escalate severity:

1. **Multiple MEDIUM violations in same test → HIGH**

   - Cumulative effect degrades test quality significantly

2. **Pattern of violations across test suite → Escalate one level**

   - Systematic issues require systematic fixes

3. **Violations in critical business logic → Escalate one level**
   - Payment processing, authentication, data integrity

## Examples by Severity

### CRITICAL Example

```go
// Test that cannot fail
func TestUserCreation(t *testing.T) {
    user := User{Name: "test"}
    // No assertions at all!
}

// Mocking own business logic
func TestWorkflow(t *testing.T) {
    mockWorkflow := &MockWorkflow{} // Mocking own code
    mockWorkflow.Execute()
    assert.True(t, true) // Meaningless
}
```

### HIGH Example

```go
// Missing error cases
func TestDivide(t *testing.T) {
    result := Divide(10, 2)
    assert.Equal(t, 5, result)
    // What about divide by zero?
}

// Test dependency
var sharedCounter int

func TestOne(t *testing.T) {
    sharedCounter++
    assert.Equal(t, 1, sharedCounter)
}

func TestTwo(t *testing.T) {
    sharedCounter++
    assert.Equal(t, 2, sharedCounter) // Fails if TestOne didn't run first
}
```

### MEDIUM Example

```go
// Dead code
func unusedHelper() string {
    return "never called"
}

// Poor naming
func TestThing(t *testing.T) {
    DoStuff()
    CheckStuff()
}

// Redundant comment
func TestAddition(t *testing.T) {
    // This tests addition
    result := Add(1, 2)
    assert.Equal(t, 3, result)
}
```

### LOW Example

```go
// Formatting issue
func TestSomething(t *testing.T) {
result := Calculate()
    assert.Equal(t,expected,result) // Spacing issues
}

// Missing convention
func test_with_underscores(t *testing.T) { // Should be TestWithUnderscores
    // ...
}
```

## Reporting Guidelines

When reporting violations:

1. **Start with CRITICAL violations**

   - These block all progress
   - Provide clear fix examples

2. **Group by severity**

   - Makes prioritization clear
   - Helps focus effort

3. **Provide context**

   - Why this matters
   - Impact if not fixed
   - Suggested fix

4. **Be specific**
   - Include file and line numbers
   - Show bad and good examples
   - Explain the difference

## Quick Reference

| Severity | Fix Timeline | Blocks Merge? | Review Needed? |
| -------- | ------------ | ------------- | -------------- |
| CRITICAL | Immediately  | Yes           | Always         |
| HIGH     | Before merge | Usually       | Yes            |
| MEDIUM   | This sprint  | No            | If multiple    |
| LOW      | Eventually   | No            | No             |
