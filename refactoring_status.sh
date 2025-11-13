#!/bin/bash

# Refactoring Status Check Script
# Run this to get current codebase metrics and identify refactoring priorities

echo "========================================="
echo "   KAINUGURU API REFACTORING STATUS"
echo "========================================="
echo ""
echo "Date: $(date)"
echo "Branch: $(git branch --show-current)"
echo ""

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

echo "📊 CODEBASE METRICS"
echo "-------------------"

# Count Go files and lines
go_files=$(find . -name "*.go" -not -path "./vendor/*" | wc -l | tr -d ' ')
go_lines=$(find . -name "*.go" -not -path "./vendor/*" -exec cat {} \; | wc -l | tr -d ' ')
test_files=$(find . -name "*_test.go" -not -path "./vendor/*" | wc -l | tr -d ' ')

echo "• Go files: $go_files"
echo "• Total LOC: $go_lines"
echo "• Test files: $test_files"

# Test coverage
echo ""
echo "📈 TEST COVERAGE"
echo "----------------"
if coverage=$(go test -cover ./... 2>&1 | grep -E "coverage:.*%" | awk '{sum += $2; count++} END {if (count > 0) print sum/count; else print 0}'); then
    echo "• Average coverage: ${coverage}%"
else
    echo "• Coverage check failed"
fi

# Count TODOs
echo ""
echo "📝 TECHNICAL DEBT"
echo "-----------------"
todos=$(grep -r "TODO" --include="*.go" | wc -l | tr -d ' ')
fixmes=$(grep -r "FIXME" --include="*.go" | wc -l | tr -d ' ')
echo "• TODO comments: $todos"
echo "• FIXME comments: $fixmes"

# Large files
echo ""
echo "📁 LARGE FILES (>500 LOC)"
echo "-------------------------"
find . -name "*.go" -not -path "./vendor/*" -exec wc -l {} \; | awk '$1 > 500 {print $2 ": " $1 " lines"}' | sort -t: -k2 -rn | head -5

# Complex functions (if gocyclo is installed)
if command_exists gocyclo; then
    echo ""
    echo "🔧 COMPLEX FUNCTIONS (Cyclomatic > 10)"
    echo "--------------------------------------"
    gocyclo -over 10 . 2>/dev/null | head -5
fi

# Duplication check (if dupl is installed)
if command_exists dupl; then
    echo ""
    echo "📋 CODE DUPLICATION"
    echo "-------------------"
    dupl_count=$(dupl -t 50 . 2>/dev/null | grep -c "found")
    echo "• Duplicate blocks found: $dupl_count"
fi

# Critical issues
echo ""
echo "🚨 CRITICAL ISSUES"
echo "------------------"

# Check for context.Background() in handlers
context_abuse=$(grep -r "context.Background()" --include="*.go" | grep -v "_test.go" | wc -l | tr -d ' ')
echo "• context.Background() usage: $context_abuse instances"

# Check for unhandled errors
unhandled=$(grep -r "_ =" --include="*.go" | grep -E "err|error" | wc -l | tr -d ' ')
echo "• Ignored errors: $unhandled instances"

# Check repository structure confusion
if [ -d "repositories" ] && [ -d "repository" ]; then
    echo "• ⚠️ Both 'repositories/' and 'repository/' directories exist!"
fi

# Check for placeholder service
if [ -f "services/placeholder_service.go" ]; then
    placeholder_lines=$(wc -l services/placeholder_service.go | awk '{print $1}')
    echo "• ⚠️ Placeholder service exists: $placeholder_lines LOC of dead code"
fi

echo ""
echo "🎯 TOP REFACTORING PRIORITIES"
echo "-----------------------------"

priority_count=0

# Priority 1: Context abuse
if [ "$context_abuse" -gt 0 ]; then
    ((priority_count++))
    echo "$priority_count. Fix context.Background() usage ($context_abuse instances)"
fi

# Priority 2: Test coverage
if (( $(echo "$coverage < 30" | bc -l) )); then
    ((priority_count++))
    echo "$priority_count. Increase test coverage (currently ${coverage}%)"
fi

# Priority 3: Large files
large_files=$(find . -name "*.go" -not -path "./vendor/*" -exec wc -l {} \; | awk '$1 > 500' | wc -l | tr -d ' ')
if [ "$large_files" -gt 0 ]; then
    ((priority_count++))
    echo "$priority_count. Split large files ($large_files files > 500 LOC)"
fi

# Priority 4: Duplicate directories
if [ -d "repositories" ] && [ -d "repository" ]; then
    ((priority_count++))
    echo "$priority_count. Consolidate repository directories"
fi

# Priority 5: Dead code
if [ -f "services/placeholder_service.go" ]; then
    ((priority_count++))
    echo "$priority_count. Remove placeholder service"
fi

echo ""
echo "📚 REFACTORING GUIDES"
echo "--------------------"
echo "• General Guidelines: REFACTORING_GUIDELINES.md"
echo "• Agent Instructions: AGENT_REFACTORING_INSTRUCTIONS.md"
echo "• Go Patterns: GO_REFACTORING_PATTERNS.md"
echo "• Analysis Docs: CODEBASE_ANALYSIS.md"
echo ""

echo "💡 QUICK ACTIONS"
echo "----------------"
echo "1. Run tests: go test ./..."
echo "2. Check coverage: go test -cover ./..."
echo "3. Find TODOs: grep -r 'TODO' --include='*.go'"
echo "4. Find large files: find . -name '*.go' -exec wc -l {} + | sort -rn | head"
echo ""

echo "========================================="
echo "Run specific refactoring with: go run refactor.go <pattern>"
echo "========================================="