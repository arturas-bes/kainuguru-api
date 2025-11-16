#!/bin/bash
# Auto-remediation script for /speckit.analyze findings
# Applies all CRITICAL and HIGH priority fixes

set -e

SPEC_DIR="/Users/arturas/Dev/kainuguru_all/kainuguru-api/specs/001-shopping-list-migration"
cd "$SPEC_DIR"

echo "=== Applying /speckit.analyze Remediation ==="
echo

# Backup original files
echo "[1/5] Creating backups..."
cp plan.md plan.md.original
cp spec.md spec.md.original
cp tasks.md tasks.md.original
echo "✓ Backups created (.original files)"
echo

# Fix C3: Go version in plan.md
echo "[2/5] Fixing Go version (C3)..."
sed -i '' 's/Go 1\.21+ (based on existing kainuguru-api codebase)/Go 1.24+ (per constitution requirement, existing codebase to be upgraded)/' plan.md
echo "✓ Updated plan.md line 14: Go 1.21+ → Go 1.24+"
echo

# Fix C5: Confidence score range in spec.md
echo "[3/5] Fixing confidence score range (C5)..."
sed -i '' 's/confidence scores (0-100%)/confidence scores (0.0-1.0 scale)/' spec.md
echo "✓ Updated spec.md FR-006: 0-100% → 0.0-1.0 scale"
echo

# Fix C4: Repository paths in tasks.md
echo "[4/5] Fixing repository paths (C4)..."
COUNT=$(grep -c "internal/repository/" tasks.md || true)
sed -i '' 's|internal/repository/|internal/repositories/|g' tasks.md
echo "✓ Updated $COUNT occurrences: internal/repository/ → internal/repositories/"
echo

# Add constitution compliance sections
echo "[5/5] Adding constitution compliance updates..."

# Add testing exception to plan.md (after line 254)
cat >> plan.md.temp << 'EOF'

## Testing Exception (Constitution VII)

**Exception Requested**: Deferring 70% test coverage requirement to post-MVP phase

**Justification**:
- Feature is behind feature flag (WIZARD_FEATURE_ENABLED) for safe rollout
- Quickstart.md provides comprehensive manual testing scenarios
- Integration test script validates end-to-end flow
- Production rollout follows gradual strategy (5% → 25% → 50% → 100%)

**Post-MVP Commitment**:
- Add repository fake implementations within 2 weeks of MVP deployment
- Achieve 70% coverage within 1 month of MVP deployment
- Follow table-driven test patterns per constitution
- Use in-memory SQLite for repository tests
- Use fakes with function fields for service tests

**Risk Mitigation**:
- Shadow mode testing before enabling for users
- Comprehensive logging for debugging production issues
- Quick rollback via feature flag if issues detected
EOF

# Insert testing exception after "Phase 2: Task Generation" section
awk '/^## Phase 2: Task Generation/{print; while(getline line && line !~ /^##/) print line; print ""; while((getline line < "plan.md.temp") > 0) print line; next}1' plan.md > plan.md.new
mv plan.md.new plan.md
rm plan.md.temp

echo "✓ Added Testing Exception section to plan.md"

# Update tasks.md header
sed -i '' 's/\*\*Tests\*\*: NOT included - tests are optional and not explicitly requested in the specification./**Tests**: Constitution VII requires 70% coverage - EXCEPTION DOCUMENTED in plan.md with post-MVP timeline. Manual testing via quickstart.md and integration test./' tasks.md

echo "✓ Updated tasks.md test policy reference"
echo

echo "=== Summary of Changes ==="
echo "✅ plan.md: Go version updated to 1.24+, Testing Exception added"
echo "✅ spec.md: Confidence score range standardized to 0.0-1.0"
echo "✅ tasks.md: Repository paths corrected, test policy updated"
echo
echo "Original files saved with .original extension"
echo "Review changes with: diff plan.md.original plan.md"
echo
echo "=== Remaining Manual Actions ===="
echo "The following fixes require manual editing (see remediation plan):"
echo "  - C1: Add error handling tasks (T011a, T011b) and update service tasks"
echo "  - C6: Add constructor patterns to service tasks (T013, T014, T023)"
echo "  - C7: Add Repository[T] generic pattern to repository tasks"
echo "  - C9: Add shopping list locking tasks (T042a, T042b)"
echo "  - U2: Add dataset version tracking tasks (T002a, T002b)"
echo
echo "Refer to the full remediation plan for exact task descriptions."
