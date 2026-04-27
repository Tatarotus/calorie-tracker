#!/bin/bash

# Mutation Testing Script for Go (Line-by-Line)
MIN_MUTATION_SCORE=${MIN_MUTATION_SCORE:-30}
CRITICAL_FILES="services/tracker.go services/food_matcher.go db/sqlite.go"
TEMP_DIR="${TEMP_DIR:-/tmp/go-mutation-$$}"
LOG_FILE="$TEMP_DIR/mutation.log"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

mkdir -p "$TEMP_DIR"
trap "rm -rf $TEMP_DIR" EXIT

echo "Running line-by-line mutation testing..."
echo "Files: $CRITICAL_FILES"
echo ""

TOTAL_MUTATIONS=0
KILLED_MUTATIONS=0
SURVIVED_MUTATIONS=0

run_tests() {
    if go test ./services ./db ./tests -timeout 20s > /dev/null 2>&1; then
        return 0 # survived
    else
        return 1 # killed
    fi
}

mutate_line() {
    local file=$1
    local line=$2
    local old=$3
    local new=$4
    local desc=$5

    cp "$file" "$TEMP_DIR/backup.go"
    # Use @ as delimiter to avoid issues with / in patterns
    sed -i "${line}s@${old}@${new}@" "$file"
    
    if run_tests; then
        echo "❌ SURVIVED: $desc in $file at line $line" >> "$LOG_FILE"
        SURVIVED_MUTATIONS=$((SURVIVED_MUTATIONS + 1))
    else
        echo "✅ KILLED: $desc in $file at line $line" >> "$LOG_FILE"
        KILLED_MUTATIONS=$((KILLED_MUTATIONS + 1))
    fi
    
    cp "$TEMP_DIR/backup.go" "$file"
    TOTAL_MUTATIONS=$((TOTAL_MUTATIONS + 1))
    echo -n "."
}

for file in $CRITICAL_FILES; do
    if [ ! -f "$file" ]; then continue; fi
    
    # Find all lines with operators and use a file/redirection to avoid subshell
    while IFS=: read -r line_num content; do
        # Basic mutations
        if [[ "$content" == *"=="* ]]; then mutate_line "$file" "$line_num" "==" "!=" "EQ_to_NEQ"; fi
        if [[ "$content" == *"!="* ]]; then mutate_line "$file" "$line_num" "!=" "==" "NEQ_to_EQ"; fi
        if [[ "$content" == *"&&"* ]]; then mutate_line "$file" "$line_num" "&&" "||" "AND_to_OR"; fi
        if [[ "$content" == *"||"* ]]; then mutate_line "$file" "$line_num" "||" "&&" "OR_to_AND"; fi
        if [[ "$content" == *"<"* && "$content" != *"<="* && "$content" != *"*="* ]]; then mutate_line "$file" "$line_num" "<" ">=" "LT_to_GTE"; fi
        if [[ "$content" == *">"* && "$content" != *">="* && "$content" != *"->"* ]]; then mutate_line "$file" "$line_num" ">" "<=" "GT_to_LTE"; fi
        if [[ "$content" == *"LIMIT 1"* ]]; then mutate_line "$file" "$line_num" "LIMIT 1" "LIMIT 2" "LIMIT_1_to_2"; fi
    done < <(grep -nE "&&|\|\||!=|==|<|>|LIMIT" "$file")
done

# Special case for tracker.go loop
if [ -f "services/tracker.go" ]; then
    line_num=$(grep -n "i >= 0" services/tracker.go | cut -d: -f1)
    if [ ! -z "$line_num" ]; then
        mutate_line "services/tracker.go" "$line_num" ">= 0" "> 0" "GTE0_to_GT0"
    fi
fi

echo ""
echo ""

if [ $TOTAL_MUTATIONS -eq 0 ]; then
    echo "No mutations generated."
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

if [ $MUTATION_SCORE -ge $MIN_MUTATION_SCORE ]; then
    echo -e "${GREEN}✔ Mutation score OK (>= ${MIN_MUTATION_SCORE}%)${NC}"
    exit 0
else
    echo -e "${RED}✘ Mutation score too low: ${MUTATION_SCORE}% (required: ${MIN_MUTATION_SCORE}%)${NC}"
    echo "Top survivors:"
    grep "SURVIVED" "$LOG_FILE" | head -10
    exit 1
fi
