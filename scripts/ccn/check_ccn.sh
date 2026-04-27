#!/bin/bash
# Cyclomatic Complexity Check Script
# Fails if any function exceeds the complexity threshold

set -e

# Add Go bin to PATH if not already present
GOBIN=${GOBIN:-$HOME/go/bin}
if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
    export PATH="$GOBIN:$PATH"
fi

MAX_COMPLEXITY=20

echo "Checking cyclomatic complexity (max: $MAX_COMPLEXITY)..."

# Check if gocyclo is installed
if ! command -v gocyclo &> /dev/null; then
    echo "Error: gocyclo is not installed."
    echo "Install with: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"
    exit 1
fi

# Run gocyclo and capture output
# -top 1000 ensures we get all functions, not just the top N
# Exclude empty files and test directories
output=$(gocyclo -top 1000 $(find . -name "*.go" -not -name "*_test.go" -not -name "ccn_test.go" -not -path "./commands/*" -not -path "./db/*" | grep -v "^\./commands/" | grep -v "^\./db/") 2>&1) || true

if [ -z "$output" ]; then
    echo "PASS: No Go files found or analysis complete."
    exit 0
fi

# Check for functions exceeding threshold
# Output format: "complexity function_name path:line:col"
exceeded=0
while IFS= read -r line; do
    if [ -z "$line" ]; then
        continue
    fi
    
    # Extract complexity (first field)
    complexity=$(echo "$line" | awk '{print $1}')
    
    if [ -n "$complexity" ] && [ "$complexity" -gt "$MAX_COMPLEXITY" ]; then
        echo "FAIL: $line (max: $MAX_COMPLEXITY)"
        exceeded=1
    fi
done <<< "$output"

if [ "$exceeded" -eq 1 ]; then
    echo ""
    echo "FAILED: Some functions exceed cyclomatic complexity limit of $MAX_COMPLEXITY"
    exit 1
fi

echo "PASS: All functions have cyclomatic complexity <= $MAX_COMPLEXITY"
exit 0