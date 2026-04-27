# Mutation Testing Setup

## Overview

This project now includes **mutation testing** as part of its pre-commit quality gate. Mutation testing goes beyond code coverage by actually modifying the code to ensure tests can detect real faults.

## How It Works

### Mutation Score Calculation

```
Mutation Score = (Killed Mutations / Total Mutations) × 100%
```

- **Killed Mutation**: A mutation that causes tests to fail (good - tests caught the fault)
- **Survived Mutation**: A mutation that doesn't affect tests (bad - tests missed the fault)

### What Gets Mutated

The mutation script targets these patterns in `./services` and `./db` packages:

1. `==` → `!=` (equality to inequality)
2. `!=` → `==` (inequality to equality)
3. `&&` → `||` (AND to OR)
4. `||` → `&&` (OR to AND)
5. `return true` → `return false`
6. `return false` → `return true`
7. `> ` → `>= ` (greater than to greater-equal)
8. `< ` → `<= ` (less than to less-equal)
9. `== 0` → `== 1` (zero to one)
10. `len() > 0` → `len() > 1` (empty check to non-empty)

## Current Status

- **Mutation Score**: 36% (4 killed, 7 survived out of 11 mutations)
- **Minimum Threshold**: 30% (configurable via `MIN_MUTATION_SCORE`)
- **Packages Tested**: `./services`, `./db`
- **Max Mutations per Run**: 30 (for performance)

## Performance Optimizations

1. **Limited Scope**: Only tests critical packages (`services`, `db`)
2. **Mutation Cap**: Maximum 30 mutations per run
3. **Timeout**: 30 seconds per mutation test
4. **Skip Option**: Set `SKIP_MUTATION=1` to bypass for fast commits

## How to Use

### Run Mutation Testing Manually

```bash
# Full run
bash scripts/check_mutation.sh

# With custom threshold
MIN_MUTATION_SCORE=50 bash scripts/check_mutation.sh

# Skip mutation testing
SKIP_MUTATION=1 bash scripts/check_mutation.sh
```

### In Pre-Commit Hook

Mutation testing runs automatically after:
1. Unit tests
2. Coverage check
3. CCN check

If mutation score < threshold, the commit is blocked.

## Improving Mutation Score

### Current Survived Mutations

The following mutations are surviving (tests not catching them):

1. **OR to AND mutations** in `tracker.go`, `food_matcher.go`, `mock_db.go`
   - Tests don't verify both conditions in OR expressions
   - **Fix**: Add tests that require both conditions to be true

2. **GT to GTE mutations** in `food_matcher.go`, `mock_db.go`
   - Tests don't check boundary conditions
   - **Fix**: Add tests for exact boundary values (e.g., `x > 5` should fail at `x = 5`)

3. **LT to LTE mutations** in `food_matcher.go`
   - Similar to above, missing boundary tests
   - **Fix**: Test values at and below the boundary

### Recommended Test Improvements

#### 1. Add Boundary Condition Tests

```go
// Instead of just testing x > 5
func TestGreaterThan(t *testing.T) {
    result := check(6)
    assert.True(t, result)
}

// Add boundary test
func TestGreaterThanBoundary(t *testing.T) {
    result := check(5)  // Should be false
    assert.False(t, result)
}
```

#### 2. Test Both Sides of Logical Operators

```go
// For: if a || b
func TestOrOperator(t *testing.T) {
    // Test a=true, b=false
    // Test a=false, b=true
    // Test a=true, b=true
    // Test a=false, b=false (should fail)
}
```

#### 3. Add More Assertions

```go
// Weak test
func TestFunction(t *testing.T) {
    err := DoSomething()
    assert.NoError(t, err)  // Only checks no error
}

// Strong test
func TestFunction(t *testing.T) {
    result, err := DoSomething()
    assert.NoError(t, err)
    assert.Equal(t, expected, result)  // Also checks result
    assert.Greater(t, result.Value, 0) // Checks specific properties
}
```

#### 4. Test Error Paths

```go
func TestErrorHandling(t *testing.T) {
    // Test with invalid input
    _, err := DoSomethingInvalid()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected message")
}
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MIN_MUTATION_SCORE` | 60 | Minimum required mutation score (%) |
| `MAX_MUTATIONS` | 30 | Maximum mutations to test per run |
| `SKIP_MUTATION` | 0 | Set to 1 to skip mutation testing |

### Target Packages

Edit `CRITICAL_PACKAGES` in `scripts/check_mutation.sh`:

```bash
CRITICAL_PACKAGES="./services ./db"  # Currently tested packages
```

## Future Enhancements

1. **Increase Threshold**: Gradually increase from 30% → 50% → 70% as tests improve
2. **Add More Packages**: Include `./tui` and `./commands` once they have better tests
3. **Mutation Types**: Add more mutation operators (arithmetic, loop, etc.)
4. **Mutation Report**: Generate detailed HTML report showing which lines survived
5. **CI Integration**: Run full mutation suite in CI (not just pre-commit)

## Troubleshooting

### "No mutations were generated"

- Check that test files exist for the package
- Verify the package contains mutatable patterns (comparisons, logical operators)
- Add more test cases if code is too simple

### "Mutation score too low"

- Review survived mutations in the log
- Add tests that specifically target those code paths
- Ensure tests have proper assertions (not just error checks)
- Test edge cases and boundary conditions

### Performance Issues

- Reduce `MAX_MUTATIONS` (e.g., to 20)
- Increase timeout if mutations are taking too long
- Use `SKIP_MUTATION=1` for quick iterations

## Acceptance Criteria

✅ Mutation testing runs locally  
✅ Mutation score is calculated correctly  
✅ Pre-commit fails if score < 30%  
✅ Mutation runs only on selected packages  
✅ Output is clear and readable  
✅ Existing pipeline remains functional  
✅ Can skip mutation with environment variable  

## Next Steps

1. **Immediate**: Keep threshold at 30% while improving tests
2. **Short-term**: Add boundary condition tests to reach 50%
3. **Medium-term**: Improve logical operator tests to reach 60%
4. **Long-term**: Target 70%+ mutation score with comprehensive test coverage

---

**Remember**: High mutation score means your tests can detect real faults, not just cover code. Focus on quality over quantity!
