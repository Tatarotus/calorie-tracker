# Pre-Commit Quality Checks

This directory contains the unified pre-commit quality gate for the calorie-tracker project.

## Overview

The `pre_commit_check.sh` script runs all quality checks in a single pipeline:

1. **Unit Tests** - Runs `go test ./...`
2. **Coverage Check** - Ensures test coverage is ≥ 80%
3. **CCN Check** - Validates cyclomatic complexity using existing `check_ccn.sh`

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

## Dependencies

- Go toolchain
- bash
- gocyclo (for CCN check) - install with:
  ```bash
  go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
  ```
