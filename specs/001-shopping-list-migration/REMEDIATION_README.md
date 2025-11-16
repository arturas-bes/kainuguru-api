# Remediation Package for /speckit.analyze Findings

**Generated**: 2025-11-15
**Feature**: Shopping List Migration Wizard
**Analysis**: 18 findings (8 CRITICAL, 4 HIGH, 6 MEDIUM/LOW)

---

## 📦 What's Included

This remediation package contains everything needed to fix all critical and high-priority issues identified by `/speckit.analyze`:

1. **apply-fixes.sh** - Automated script for simple find/replace fixes
2. **REMEDIATION_MANUAL_EDITS.md** - Detailed manual editing instructions
3. **plan.md.fixed** - Complete updated plan with all fixes
4. **spec.md.fixed** - Complete updated spec with confidence range fix

---

## 🚀 Quick Start (Recommended Approach)

```bash
cd specs/001-shopping-list-migration

# Step 1: Run automated fixes (2 minutes)
./apply-fixes.sh

# Step 2: Apply manual edits (25-30 minutes)
# Open REMEDIATION_MANUAL_EDITS.md and follow instructions
# OR use the pre-fixed plan.md.fixed

# Step 3: Validate changes
grep -c "internal/repository/" tasks.md  # Should be 0
grep "Go 1.24" plan.md                    # Should match
grep "0.0-1.0" spec.md                    # Should match
grep "T011a\|T011b\|T008a\|T042a\|T042b\|T002a\|T002b" tasks.md  # Should match 7 times
```

---

## 📋 What Gets Fixed

### Automated (apply-fixes.sh)
- ✅ **C3** - Go version: 1.21+ → 1.24+ in plan.md
- ✅ **C4** - Repository paths: `internal/repository/` → `internal/repositories/` (18 occurrences)
- ✅ **C5** - Confidence scores: 0-100% → 0.0-1.0 scale in spec.md
- ✅ **C2** - Testing exception: Added constitution compliance section to plan.md

### Manual Edits Required
- ⚠️ **C1** - Error handling: Add 2 new tasks, update 15+ existing tasks
- ⚠️ **C6** - Service constructors: Update 3 service task descriptions
- ⚠️ **C7** - Repository patterns: Add 1 new task, update 3 repository tasks
- ⚠️ **C9** - List locking: Add 2 new tasks for FR-016 compliance
- ⚠️ **U2** - Dataset versioning: Add 2 new tasks, update 1 existing task

**Total**: 7 new tasks, ~20 task description updates

---

## 📖 Files Explained

### apply-fixes.sh
**What it does**: Automatically fixes simple issues
- Creates .original backups of plan.md, spec.md, tasks.md
- Updates Go version to 1.24+
- Fixes all repository path references
- Standardizes confidence score range
- Adds testing exception section
- Updates test policy reference

**Runtime**: ~5 seconds
**Manual effort**: 0 minutes (fully automated)

### REMEDIATION_MANUAL_EDITS.md
**What it does**: Provides copy-paste instructions for complex fixes
- Organized by finding ID (C1, C7, C9, U2, etc.)
- Exact line numbers and search patterns
- Complete task descriptions ready to paste
- Validation commands for each fix

**Manual effort**: 25-30 minutes of careful copy-paste

### plan.md.fixed
**What it does**: Complete replacement file with all fixes
- Go 1.24+ version
- Testing exception section
- Updated monitoring metrics
- Reference to spec.md for success criteria
- Constitution VI & VII compliance notes

**How to use**:
```bash
mv plan.md plan.md.old
mv plan.md.fixed plan.md
```

### spec.md.fixed
**What it does**: Spec with confidence range fix
- FR-006: 0-100% → 0.0-1.0 scale

**How to use**:
```bash
mv spec.md spec.md.old
mv spec.md.fixed spec.md
```

---

## 🎯 Recommended Workflow

### Option A: Hybrid (Fast + Complete)
1. Run `apply-fixes.sh` (automated simple fixes)
2. Apply manual edits from `REMEDIATION_MANUAL_EDITS.md`
3. Validate with provided commands

**Total time**: 30-35 minutes

### Option B: File Replacement (Fastest for plan/spec)
1. Replace plan.md with plan.md.fixed
2. Replace spec.md with spec.md.fixed
3. Apply manual edits to tasks.md from `REMEDIATION_MANUAL_EDITS.md`
4. Validate

**Total time**: 25-30 minutes

### Option C: Regenerate Tasks (Cleanest)
1. Run `apply-fixes.sh` to fix plan.md and spec.md
2. Delete tasks.md
3. Run `/speckit.tasks` again to regenerate with updated plan/spec
4. Manually add the 7 new constitution compliance tasks

**Total time**: 10 minutes + `/speckit.tasks` runtime

---

## ✅ Validation

After applying all fixes, run these commands:

```bash
# Check for remaining issues
./validate-fixes.sh

# Or manually:
echo "Checking repository paths..."
test $(grep -c "internal/repository/" tasks.md) -eq 0 && echo "✓ PASS" || echo "✗ FAIL"

echo "Checking Go version..."
grep -q "Go 1.24" plan.md && echo "✓ PASS" || echo "✗ FAIL"

echo "Checking confidence range..."
grep -q "0.0-1.0 scale" spec.md && echo "✓ PASS" || echo "✗ FAIL"

echo "Checking error handling tasks..."
test $(grep -c "T011a\|T011b" tasks.md) -eq 2 && echo "✓ PASS" || echo "✗ FAIL"

echo "Checking base repository task..."
grep -q "T008a" tasks.md && echo "✓ PASS" || echo "✗ FAIL"

echo "Checking locking tasks..."
test $(grep -c "T042a\|T042b" tasks.md) -eq 2 && echo "✓ PASS" || echo "✗ FAIL"

echo "Checking dataset version tasks..."
test $(grep -c "T002a\|T002b" tasks.md) -eq 2 && echo "✓ PASS" || echo "✗ FAIL"

echo "Checking task count..."
TASK_COUNT=$(grep -c "^- \[ \] T[0-9]" tasks.md)
test $TASK_COUNT -eq 107 && echo "✓ PASS ($TASK_COUNT tasks)" || echo "⚠ CHECK ($TASK_COUNT tasks, expected 107)"
```

---

## 🔄 Rollback

If you need to revert changes:

```bash
# If you used apply-fixes.sh:
mv plan.md.original plan.md
mv spec.md.original spec.md
mv tasks.md.original tasks.md

# If you used file replacement:
mv plan.md.old plan.md
mv spec.md.old spec.md
```

---

## 📊 Before/After Summary

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Go Version | 1.21+ | 1.24+ | Constitution compliance |
| Confidence Range | 0-100% | 0.0-1.0 | Consistency with code |
| Repository Paths | Mixed | internal/repositories/ | Matches codebase |
| Error Handling | Implicit | Explicit pkg/errors | Constitution VI |
| Testing Policy | Excluded | Exception documented | Constitution VII |
| Total Tasks | 100 | 107 | +7 for compliance |
| Critical Issues | 8 | 0 | All resolved |
| High Issues | 4 | 0 | All resolved |
| Constitution Violations | 5 | 0 | All resolved |

---

## ❓ FAQ

**Q: Can I skip the manual edits?**
A: You can skip medium/low priority items (U2, U3, A2), but CRITICAL and HIGH items (C1, C7, C9) should be addressed for constitution compliance.

**Q: What if I just want to fix critical issues?**
A: Run `apply-fixes.sh` + manually add just C1 (error handling), C7 (repository pattern), and C9 (locking). See REMEDIATION_MANUAL_EDITS.md sections for those IDs.

**Q: Should I regenerate tasks.md instead?**
A: Yes, if you prefer a clean slate. Fix plan.md and spec.md first, then run `/speckit.tasks` again.

**Q: How do I know which fixes are mandatory?**
A: Anything marked CRITICAL is mandatory for constitution compliance. HIGH is strongly recommended. MEDIUM can be deferred.

---

## 📞 Support

If you encounter issues:
1. Check the validation commands in REMEDIATION_MANUAL_EDITS.md
2. Review the full analysis report from `/speckit.analyze`
3. Use diff to compare changes: `diff plan.md.original plan.md`

---

**Ready to proceed?** Start with `./apply-fixes.sh` 🚀
