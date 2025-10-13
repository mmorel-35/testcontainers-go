# Cleanup Analysis: Mockery vs Testcontainers

## Executive Summary

After analyzing both mockery's cleanup approach and testcontainers' cleanup approach, this document provides recommendations on whether testcontainers should adopt mockery's cleanup pattern.

## Mockery's Cleanup Approach

### Key Characteristics

1. **Automatic Cleanup Registration**: Mockery generates a constructor function `newMockStrategyTarget` that automatically registers a cleanup function using `t.Cleanup()`:
   ```go
   func newMockStrategyTarget(t interface {
       mock.TestingT
       Cleanup(func())
   }) *mockStrategyTarget {
       mock := &mockStrategyTarget{}
       mock.Mock.Test(t)
       
       t.Cleanup(func() { mock.AssertExpectations(t) })
       
       return mock
   }
   ```

2. **Purpose**: The cleanup ensures that all mock expectations are verified at the end of the test, preventing false positives where a test passes but the mock was never actually called as expected.

3. **Timing**: The cleanup runs automatically after the test completes (success or failure), via Go's `t.Cleanup()` mechanism.

4. **Error Handling**: If expectations are not met, the test fails with clear assertion errors.

### Benefits of Mockery's Approach

- **Impossible to forget**: The cleanup is embedded in the constructor
- **Clear intent**: Mock expectations are always verified
- **Consistent pattern**: All generated mocks follow the same pattern
- **Test safety**: Prevents false positives from unmet expectations

## Testcontainers' Cleanup Approach

### Current Implementation

Testcontainers provides TWO cleanup patterns:

#### Pattern 1: `CleanupContainer` Helper (Recommended)
```go
func CleanupContainer(tb testing.TB, ctr Container, options ...TerminateOption) {
    tb.Helper()
    
    tb.Cleanup(func() {
        noErrorOrIgnored(tb, TerminateContainer(ctr, options...))
    })
}
```

Usage example:
```go
func TestIntegrationNginxLatestReturn(t *testing.T) {
    ctx := context.Background()
    
    nginxC, err := startContainer(ctx)
    testcontainers.CleanupContainer(t, nginxC)  // Called immediately after creation
    require.NoError(t, err)
    
    // ... test code ...
}
```

#### Pattern 2: Manual defer (Used in examples for documentation)
```go
func ExampleRun() {
    ctx := context.Background()
    
    mockserverContainer, err := mockserver.Run(ctx, "mockserver/mockserver:5.15.0")
    defer func() {
        if err := testcontainers.TerminateContainer(mockserverContainer); err != nil {
            log.Printf("failed to terminate container: %s", err)
        }
    }()
    if err != nil {
        log.Printf("failed to start container: %s", err)
        return
    }
    // ...
}
```

### Key Differences from Mockery

1. **Manual invocation**: Developer must explicitly call `CleanupContainer()`
2. **Flexible timing**: Can be called immediately after creation or later
3. **Error tolerance**: Uses `isCleanupSafe()` to ignore expected cleanup errors:
   - Container not found (already removed)
   - Container already being terminated
4. **Resource cleanup**: Terminates actual Docker containers (external resources) vs verifying expectations (internal state)

## Analysis: Should Testcontainers Adopt Mockery's Pattern?

### Key Differences to Consider

| Aspect | Mockery Mocks | Testcontainers Containers |
|--------|---------------|---------------------------|
| Resource Type | In-memory test doubles | External Docker containers |
| Cleanup Purpose | Verify expectations | Free external resources |
| Failure Impact | Test logic errors | Resource leaks |
| Error Expectations | Always expect clean assertions | Must handle "already gone" cases |
| Creation Pattern | Constructor function | Multiple creation methods |

### Recommendation: **NO** - Keep Current Pattern

Testcontainers should **NOT** adopt mockery's automatic cleanup pattern embedded in constructors. Here's why:

#### Reasons Against Adoption

1. **Multiple Creation Patterns**: 
   - Testcontainers has multiple ways to create containers: `Run()`, `GenericContainer()`, module-specific functions
   - Mockery has a single constructor pattern
   - Embedding cleanup in all creation methods would be complex and inconsistent

2. **Error Handling Requirements Differ**:
   - Containers may legitimately fail to terminate (not found, already terminating)
   - Mockery expectations should always pass cleanly
   - Current `isCleanupSafe()` logic is sophisticated and necessary

3. **Explicit is Better**:
   - Container lifecycle is a critical concern that should be visible
   - `CleanupContainer(t, ctr)` makes it clear that cleanup will happen
   - Developers should be aware they're managing external resources

4. **Flexibility Needed**:
   - Some tests need custom cleanup logic or timing
   - Tests may want to inspect containers after test body completes
   - Deferred cleanup provides more control

5. **Current Pattern Works Well**:
   - `CleanupContainer()` is already using `t.Cleanup()` (same mechanism as mockery)
   - Pattern is well-documented and consistently used across codebase
   - Follows Go testing best practices

#### What Could Be Improved

Instead of adopting mockery's pattern, consider these improvements:

1. **Better Documentation**: Emphasize that `CleanupContainer()` should be called immediately after container creation
   
2. **Linting Support**: Add a linter rule to detect missing cleanup calls

3. **Consistent Pattern**: Ensure all examples and tests use `CleanupContainer()` consistently

## Mockery's Cleanup in Context

The key insight is that **mockery's cleanup serves a different purpose**:
- It's about **assertion verification**, not resource cleanup
- It prevents **false positives** in tests
- It's verifying **test logic**, not cleaning up external state

For testcontainers:
- Cleanup is about **resource management**
- It prevents **resource leaks**
- It's managing **external infrastructure**, not test state

## Current State of Testcontainers Cleanup

### Good Patterns Found
```go
// Test files - using CleanupContainer helper
func testContainerStart(t *testing.T) {
    ctx := context.Background()
    ctr, err := Run(ctx, nginxAlpineImage, WithExposedPorts(nginxDefaultPort))
    CleanupContainer(t, ctr)  // ✅ Good: Called immediately
    require.NoError(t, err)
}
```

### Pattern for Examples (documentation)
```go
// Example files - using defer for clarity in documentation
func ExampleRun() {
    mockserverContainer, err := mockserver.Run(ctx, "...")
    defer func() {
        if err := testcontainers.TerminateContainer(mockserverContainer); err != nil {
            log.Printf("failed to terminate container: %s", err)
        }
    }()
    // Example code...
}
```

## Conclusion

**Recommendation**: Keep testcontainers' current cleanup approach with `CleanupContainer()` helper function.

**Rationale**:
1. Mockery's pattern is optimized for mock expectation verification, not resource cleanup
2. Testcontainers' pattern appropriately handles external resource management
3. Current pattern provides necessary flexibility and error handling
4. Both patterns use `t.Cleanup()`, but for fundamentally different purposes
5. Explicit cleanup calls make resource management visible and intentional

**Action Items**:
1. ✅ Document best practices for using `CleanupContainer()` 
2. ✅ Ensure consistent usage across all test files
3. ✅ Consider adding linter rules to catch missing cleanup
4. ✅ Maintain the dual pattern: `CleanupContainer()` for tests, defer for examples

The current approach is appropriate and should be maintained.
