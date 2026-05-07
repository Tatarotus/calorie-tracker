# Pre-Commit Quality Checks

This directory contains the unified pre-commit quality gate for the calorie-tracker project.

## Overview

The `pre_commit_check.sh` script runs all quality checks in a single pipeline:

1. **Unit Tests** - Runs `go test ./...`
2. **Coverage Check** - Ensures test coverage is ≥ 80%
3. **CCN Check** - Validates cyclomatic complexity using existing `check_ccn.sh`
4. **Module Size Check** - Ensures no Go file exceeds 500 lines
5. **Linting** - Runs `golangci-lint` to detect unused code, bad imports, potential bugs

## Installation

The Git pre-commit hook is already installed at `.git/hooks/pre-commit`. It automatically calls the unified script.

To manually install or reinstall the hook:

```bash
chmod +x .git/hooks/pre-commit
```

## Usage

### Automatic (on git commit)

Simply run:
```bash
git commit -m "your message"
```

The checks will run automatically before the commit is created.

### Manual Execution

To run all checks manually:
```bash
bash scripts/pre_commit_check.sh
```

## Output

### Success
```
==========================================
  All Quality Checks Passed!
==========================================
✔ Tests passed
✔ Coverage OK (>= 80%)
✔ CCN OK
✔ Mutation skipped (CI only)
✔ Module size OK
✔ Architecture OK
✔ Commit allowed
```

### Failure

If any check fails, the script will:
- Print which check failed
- Exit with status 1
- Block the commit

## Extending with New Checks

To add a new quality check (e.g., linting, mutation testing):

1. **Add the check logic** to `pre_commit_check.sh` after the existing checks
2. **Follow the pattern**:
   ```bash
   echo "🔍 Running new check..."
   if ! your_new_check_command; then
       echo -e "${RED}✗ FAILED: New check failed${NC}"
       FAILED=1
   else
       echo -e "${GREEN}✔ New check passed${NC}"
   fi
   echo ""
   ```
3. **Add early exit** if the check should block commits:
   ```bash
   if [ "$FAILED" -eq 1 ]; then
       echo -e "${RED}Commit blocked: New check failed${NC}"
       exit 1
   fi
   ```

## Configuration

- **Minimum Coverage**: 80% (defined in `pre_commit_check.sh`)
- **Max CCN**: 20 (defined in `check_ccn.sh` and `ccn_check.go`)
- **Max File Size**: 500 lines (defined in `pre_commit_check.sh`)
- **Mutation Testing**: Run in CI only (not pre-commit)

## Dependencies

- Go toolchain
- bash
- gocyclo (for CCN check) - install with:
  ```bash
  go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
  ```
- golangci-lint (for linting) - install with:
  ```bash
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

## Quality Dimensions

### 1. Test Coverage (≥ 80%)

Runs `go test ./... -coverprofile=coverage.out` and checks the total coverage percentage.

### 2. Cyclomatic Complexity (≤ 20)

Uses `gocyclo` to check that no function exceeds a complexity of 20.

### 3. Mutation Testing (≥ 60%, CI only)

Runs a custom mutation testing script that modifies operators and checks if tests catch the changes. This is intentionally skipped in pre-commit due to time constraints.

### 4. Module/File Size (≤ 500 lines)

Checks that no Go source file exceeds 500 lines of code.

### 5. Linting (Zero issues)

Runs `golangci-lint` with the following enabled linters:
- `govet` - Go vet analysis
- `staticcheck` - Static analysis
- `unused` - Unused code detection
- `ineffassign` - Ineffective assignments
- `gocyclo` - Cyclomatic complexity
- `gosimple` - Simplification suggestions
- `errcheck` - Unchecked errors
- `deadcode` - Dead code detection
- `structcheck` - Unused struct fields
- `varcheck` - Unused variables
- `typecheck` - Type checking
- `goimports` - Import formatting
