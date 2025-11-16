#!/bin/bash
# Validation script for /speckit.analyze remediation
# Checks that all critical fixes have been applied

set -e

SPEC_DIR="/Users/arturas/Dev/kainuguru_all/kainuguru-api/specs/001-shopping-list-migration"
cd "$SPEC_DIR"

echo "=== Validating /speckit.analyze Remediation ==="
echo

PASS=0
FAIL=0

# Helper function
check() {
    local name="$1"
    local cmd="$2"
    local expected="$3"

    echo -n "[$name] "
    if eval "$cmd"; then
        echo "✅ PASS"
        ((PASS++))
    else
        echo "❌ FAIL (expected: $expected)"
        ((FAIL++))
    fi
}

# C4: Repository paths fixed
check "Repository paths" \
    "[ \$(grep -c 'internal/repository/' tasks.md 2>/dev/null || echo 0) -eq 0 ]" \
    "0 occurrences of 'internal/repository/'"

# C3: Go version updated
check "Go version" \
    "grep -q 'Go 1.24' plan.md" \
    "Go 1.24+ in plan.md"

# C5: Confidence range standardized
check "Confidence range" \
    "grep -q '0.0-1.0 scale' spec.md" \
    "0.0-1.0 scale in spec.md FR-006"

# C2: Testing exception documented
check "Testing exception" \
    "grep -q 'Testing Exception' plan.md" \
    "Testing Exception section in plan.md"

# C1: Error handling tasks added
ERR_COUNT=$(grep -c "T011a\|T011b" tasks.md 2>/dev/null || echo 0)
check "Error handling tasks" \
    "[ $ERR_COUNT -eq 2 ]" \
    "2 error tasks (T011a, T011b)"

# C7: Base repository task added
check "Base repository task" \
    "grep -q 'T008a' tasks.md" \
    "T008a task exists"

# C9: Locking tasks added
LOCK_COUNT=$(grep -c "T042a\|T042b" tasks.md 2>/dev/null || echo 0)
check "Locking tasks" \
    "[ $LOCK_COUNT -eq 2 ]" \
    "2 locking tasks (T042a, T042b)"

# U2: Dataset version tasks added
DS_COUNT=$(grep -c "T002a\|T002b" tasks.md 2>/dev/null || echo 0)
check "Dataset version tasks" \
    "[ $DS_COUNT -eq 2 ]" \
    "2 dataset tasks (T002a, T002b)"

# Task count verification
TASK_COUNT=$(grep -c "^- \[ \] T[0-9]" tasks.md 2>/dev/null || echo 0)
check "Total task count" \
    "[ $TASK_COUNT -eq 107 ]" \
    "107 tasks (found: $TASK_COUNT)"

# Constitution repository directory exists
check "Repository directory" \
    "[ -d '../../internal/repositories' ]" \
    "internal/repositories/ directory exists"

echo
echo "=== Summary ==="
echo "✅ Passed: $PASS/10"
echo "❌ Failed: $FAIL/10"
echo

if [ $FAIL -eq 0 ]; then
    echo "🎉 All validations passed! Remediation complete."
    echo
    echo "Next steps:"
    echo "  1. Review changes: git diff"
    echo "  2. Commit fixes: git add . && git commit -m 'fix: apply speckit.analyze remediation'"
    echo "  3. Proceed with implementation: /speckit.implement"
    exit 0
else
    echo "⚠️  Some validations failed. Review REMEDIATION_MANUAL_EDITS.md"
    echo
    echo "Common issues:"
    if [ $ERR_COUNT -lt 2 ]; then
        echo "  - Missing error handling tasks (C1): Add T011a, T011b"
    fi
    if ! grep -q "T008a" tasks.md 2>/dev/null; then
        echo "  - Missing base repository task (C7): Add T008a"
    fi
    if [ $LOCK_COUNT -lt 2 ]; then
        echo "  - Missing locking tasks (C9): Add T042a, T042b"
    fi
    if [ $DS_COUNT -lt 2 ]; then
        echo "  - Missing dataset version tasks (U2): Add T002a, T002b"
    fi
    if [ $TASK_COUNT -ne 107 ]; then
        echo "  - Task count mismatch: Expected 107, got $TASK_COUNT"
    fi
    exit 1
fi
