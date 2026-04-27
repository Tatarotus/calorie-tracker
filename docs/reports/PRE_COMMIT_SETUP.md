# Pre-Commit Quality Gate Setup

## Overview

A unified pre-commit quality gate has been implemented to enforce code quality before any Git commit.

## Files Created

- `scripts/pre_commit_check.sh` - Main quality gate script
- `.git/hooks/pre-commit` - Git hook that calls the quality gate
- `scripts/README.md` - Documentation for the quality checks

## How It Works

When you run `git commit`, the following checks run automatically in order:

1. **Unit Tests** - `go test ./...`
2. **Coverage Check** - Must be ≥ 80%
3. **CCN Check** - Cyclomatic complexity must be ≤ 20

If ANY check fails, the commit is blocked immediately.

## Usage

### Automatic (on commit)
```bash
git commit -m "your message"
```

### Manual execution
```bash
bash scripts/pre_commit_check.sh
```

## Current Status

⚠️ **Note**: The current test coverage is 0% because there's only one test file (`ccn_test.go`) that doesn't cover the main application code. To pass the pre-commit checks, you need to:

1. Add unit tests for your code
2. Ensure coverage reaches ≥ 80%

## Extending with New Checks

To add a new check (e.g., linting):

1. Open `scripts/pre_commit_check.sh`
2. Add a new section following the existing pattern:
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
3. Add early exit if needed:
   ```bash
   if [ "$FAILED" -eq 1 ]; then
       echo -e "${RED}Commit blocked: New check failed${NC}"
       exit 1
   fi
   ```

## Dependencies

- Go toolchain
- bash
- gocyclo (for CCN check)

Install gocyclo:
```bash
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```
