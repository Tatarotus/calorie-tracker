#!/bin/bash
# Unified Pre-Commit Quality Gate
# Runs all quality checks in a single pipeline
# Fails commit if ANY check fails

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
MIN_COVERAGE=4  # Initial target - incrementally increase as tests are added
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Track overall status
FAILED=0

echo "=========================================="
echo "  Running Pre-Commit Quality Checks"
echo "=========================================="
echo ""

# -----------------------------------------------------------------------------
# 1. Run Unit Tests
# -----------------------------------------------------------------------------
echo "🧪 Running unit tests..."
if ! go test ./... -v; then
    echo -e "${RED}✗ FAILED: Unit tests failed${NC}"
    FAILED=1
else
    echo -e "${GREEN}✔ Tests passed${NC}"
fi
echo ""

# Exit immediately if tests failed
if [ "$FAILED" -eq 1 ]; then
    echo -e "${RED}Commit blocked: Unit tests failed${NC}"
    exit 1
fi

# -----------------------------------------------------------------------------
# 2. Coverage Check
# -----------------------------------------------------------------------------
echo "📊 Checking test coverage (minimum: ${MIN_COVERAGE}%)..."

# Generate coverage profile
if ! go test ./... -coverprofile=coverage.out > /dev/null 2>&1; then
    echo -e "${RED}✗ FAILED: Could not generate coverage profile${NC}"
    exit 1
fi

# Extract total coverage percentage
COVERAGE=$(go tool cover -func=coverage.out 2>/dev/null | grep "total:" | awk '{print $3}' | tr -d '%')

if [ -z "$COVERAGE" ]; then
    echo -e "${RED}✗ FAILED: Could not extract coverage percentage${NC}"
    exit 1
fi

# Compare coverage (using bc for floating point comparison)
if command -v bc &> /dev/null; then
    IS_SUFFICIENT=$(echo "$COVERAGE >= $MIN_COVERAGE" | bc -l)
else
    # Fallback: use awk for comparison
    IS_SUFFICIENT=$(echo "$COVERAGE $MIN_COVERAGE" | awk '{if ($1 >= $2) print 1; else print 0}')
fi

if [ "$IS_SUFFICIENT" -eq 0 ]; then
    echo -e "${RED}✗ FAILED: Coverage ${COVERAGE}% is below minimum ${MIN_COVERAGE}%${NC}"
    FAILED=1
else
    echo -e "${GREEN}✔ Coverage OK (${COVERAGE}% >= ${MIN_COVERAGE}%)${NC}"
fi
echo ""

# Exit immediately if coverage check failed
if [ "$FAILED" -eq 1 ]; then
    echo -e "${RED}Commit blocked: Coverage requirement not met${NC}"
    exit 1
fi

# -----------------------------------------------------------------------------
# 3. CCN (Cyclomatic Complexity) Check
# -----------------------------------------------------------------------------
echo "📈 Checking cyclomatic complexity..."

# Run the existing CCN check script
if ! bash "$PROJECT_ROOT/check_ccn.sh"; then
    echo -e "${RED}✗ FAILED: CCN check failed${NC}"
    FAILED=1
else
    echo -e "${GREEN}✔ CCN OK${NC}"
fi
echo ""

# Exit immediately if CCN check failed
if [ "$FAILED" -eq 1 ]; then
    echo -e "${RED}Commit blocked: Cyclomatic complexity threshold exceeded${NC}"
    exit 1
fi

# -----------------------------------------------------------------------------
# All Checks Passed
# -----------------------------------------------------------------------------
echo "=========================================="
echo -e "${GREEN}  All Quality Checks Passed!${NC}"
echo "=========================================="
echo "✔ Tests passed"
echo "✔ Coverage OK (>= ${MIN_COVERAGE}%)"
echo "✔ CCN OK"
echo "✔ Commit allowed"
echo ""

exit 0
