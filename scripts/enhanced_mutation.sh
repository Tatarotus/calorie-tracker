#!/bin/bash

# Enhanced Mutation Testing Script for Go
# This script performs targeted mutation testing on critical packages
# and calculates the mutation score

# Configuration
MIN_MUTATION_SCORE=${MIN_MUTATION_SCORE:-30}
CRITICAL_PACKAGES="./services ./db"
TEMP_DIR="/tmp/go-mutation-$$"
LOG_FILE="$TEMP_DIR/mutation.log"
MAX_MUTATIONS=${MAX_MUTATIONS:-100}  # Increased for better statistical relevance

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if we should skip mutation testing
if [ "${SKIP_MUTATION:-0}" == "1" ]; then
    echo -e "${YELLOW}⊘ Skipping mutation testing (SKIP_MUTATION=1)${NC}"
    exit 0
fi

# Create temp directory
mkdir -p "$TEMP_DIR"
trap "rm -rf $TEMP_DIR" EXIT

echo "Running mutation testing on critical packages..."
echo "Packages: $CRITICAL_PACKAGES"
echo "Minimum mutation score: ${MIN_MUTATION_SCORE}%"
echo "Max mutations to test: $MAX_MUTATIONS"
echo ""

# Initialize counters
TOTAL_MUTATIONS=0
KILLED_MUTATIONS=0
SURVIVED_MUTATIONS=0

# Function to run tests and check if they pass
run_tests() {
    cd /home/sam/PARA/1.Projects/Code/calorie-tracker
    if go test $CRITICAL_PACKAGES -timeout 30s > /dev/null 2>&1; then
        return 0  # Tests pass
    else
        return 1  # Tests fail
    fi
}

# Function to create a mutation and test it
create_and_test_mutation() {
    local file=$1
    local line_num=$2
    local pattern=$3
    local replacement=$4
    local mutation_id=$5
    
    # Check if we've reached max mutations
    if [ $TOTAL_MUTATIONS -ge $MAX_MUTATIONS ]; then
        return 1  # Stop signal
    fi
    
    # Create backup
    cp "$file" "$TEMP_DIR/backup"
    
    # Apply mutation using sed
    sed -i "${line_num}s/${pattern}/${replacement}/" "$file"
    
    # Run tests
    if run_tests; then
        # Tests still pass - mutation survived
        echo "SURVIVED: $mutation_id in $file at line $line_num" >> "$LOG_FILE"
        SURVIVED_MUTATIONS=$((SURVIVED_MUTATIONS + 1))
    else
        # Tests failed - mutation killed
        echo "KILLED: $mutation_id in $file at line $line_num" >> "$LOG_FILE"
        KILLED_MUTATIONS=$((KILLED_MUTATIONS + 1))
    fi
    
    # Restore original
    cp "$TEMP_DIR/backup" "$file"
    TOTAL_MUTATIONS=$((TOTAL_MUTATIONS + 1))
    
    # Progress indicator
    if [ $((TOTAL_MUTATIONS % 5)) -eq 0 ]; then
        echo -n "."
    fi
    
    return 0
}

# Function to add mutations for a file
add_mutations_for_file() {
    local file=$1
    local filename=$(basename "$file")
    
    # Skip test files
    if [[ "$filename" == *_test.go ]]; then
        return
    fi
    
    # Skip files without test files
    local test_file="${file%.*}_test.go"
    if [ ! -f "$test_file" ]; then
        return
    fi
    
    # Read the entire file content
    local content=$(cat "$file")
    
    # Mutation 1: Change == to != (but not in imports)
    if echo "$content" | grep -qE '[^"]==' | grep -vE '^\s*import'; then
        create_and_test_mutation "$file" "==" "!=" "EQ_to_NEQ" || true
    fi
    
    # Mutation 2: Change != to ==
    if echo "$content" | grep -qE '!='; then
        create_and_test_mutation "$file" "!=" "==" "NEQ_to_EQ" || true
    fi
    
    # Mutation 3: Change && to ||
    if echo "$content" | grep -qE '&&'; then
        create_and_test_mutation "$file" "&&" "||" "AND_to_OR" || true
    fi
    
    # Mutation 4: Change || to &&
    if echo "$content" | grep -qE '||'; then
        create_and_test_mutation "$file" "||" "&&" "OR_to_AND" || true
    fi
    
    # Mutation 5: Change > to >=
    if echo "$content" | grep -qE '[^<>]>[^=]'; then
        create_and_test_mutation "$file" "> " ">= " "GT_to_GTE" || true
    fi
    
    # Mutation 6: Change < to <=
    if echo "$content" | grep -qE '[^<>]<[^=]'; then
        create_and_test_mutation "$file" "< " "<= " "LT_to_LTE" || true
    fi
    
    # Mutation 7: Change len() == 0 to len() != 0
    if echo "$content" | grep -qE 'len\([^)]*\) == 0'; then
        create_and_test_mutation "$file" "== 0" "!= 0" "LEN_EQ_0_to_NEQ_0" || true
    fi
    
    # Mutation 8: Change len() > 0 to len() > 1
    if echo "$content" | grep -qE 'len\([^)]*\) > 0'; then
        create_and_test_mutation "$file" "> 0" "> 1" "LEN_GT_0_to_GT_1" || true
    fi
}

# Process each critical package
echo "Generating mutations..."
for package in $CRITICAL_PACKAGES; do
    # Find all Go files in the package
    while IFS= read -r -d '' file; do
        add_mutations_for_file "$file"
        if [ $TOTAL_MUTATIONS -ge $MAX_MUTATIONS ]; then
            break
        fi
    done < <(find "$package" -name "*.go" -not -name "*_test.go" -print0 2>/dev/null)
    if [ $TOTAL_MUTATIONS -ge $MAX_MUTATIONS ]; then
        break
    fi
done

echo ""
echo ""

# Calculate mutation score
if [ $TOTAL_MUTATIONS -eq 0 ]; then
    echo -e "${YELLOW}⚠ No mutations were generated (no applicable code patterns found)${NC}"
    echo "This may indicate:"
    echo "  - No test files exist for the packages"
    echo "  - Code doesn't contain mutatable patterns"
    echo "  - Consider adding more test cases"
    exit 0
fi

MUTATION_SCORE=$((KILLED_MUTATIONS * 100 / TOTAL_MUTATIONS))

echo "=========================================="
echo "Mutation Testing Results"
echo "=========================================="
echo "Total mutations:  $TOTAL_MUTATIONS"
echo "Killed:           $KILLED_MUTATIONS"
echo "Survived:         $SURVIVED_MUTATIONS"
echo "Mutation score:   ${MUTATION_SCORE}%"
echo "=========================================="
echo ""

# Check against threshold
if [ $MUTATION_SCORE -ge $MIN_MUTATION_SCORE ]; then
    echo -e "${GREEN}✔ Mutation score OK (>= ${MIN_MUTATION_SCORE}%)${NC}"
    exit 0
else
    echo -e "${RED}✘ Mutation score too low: ${MUTATION_SCORE}% (required: ${MIN_MUTATION_SCORE}%)${NC}"
    echo ""
    echo "To improve mutation score:"
    echo "  1. Add more assertions in tests (not just error checks)"
    echo "  2. Test edge cases and boundary conditions"
    echo "  3. Test negative paths (error scenarios)"
    echo "  4. Verify test data covers all code paths"
    echo ""
    echo "Sample survived mutations (see $LOG_FILE for details):"
    head -10 "$LOG_FILE" 2>/dev/null || echo "  (none logged)"
    exit 1
fi