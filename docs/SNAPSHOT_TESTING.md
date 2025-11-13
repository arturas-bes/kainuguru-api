# GraphQL Snapshot Testing

## Overview

Snapshot tests capture the exact JSON structure of GraphQL connection payloads to detect unintended schema or resolver changes. This prevents accidental breaking changes to the API contract.

## What Are Snapshot Tests?

Snapshot tests compare the current output of GraphQL resolvers against "golden" reference files stored in version control. If the output changes, the test fails—forcing explicit review of the change.

### Covered Connections

The following GraphQL pagination connections have snapshot tests:

1. **Product Connection** (`testdata/product_connection.json`)
2. **Flyer Connection** (`testdata/flyer_connection.json`)
3. **Flyer Page Connection** (`testdata/flyer_page_connection.json`)
4. **Shopping List Connection** (`testdata/shopping_list_connection.json`)
5. **Shopping List Item Connection** (`testdata/shopping_list_item_connection.json`)
6. **Product Master Connection** (`testdata/product_master_connection.json`)
7. **Price History Connection** (`testdata/price_history_connection.json`)

## Running Snapshot Tests

### Quick Check

```bash
# Run all snapshot tests
make test-snapshots
```

This runs only the GraphQL snapshot tests, verifying that connection structures haven't changed.

### Full Test Suite

```bash
# Run all tests (includes snapshots)
make test
```

### CI Integration

When setting up CI/CD, include:

```yaml
# Example GitHub Actions
- name: Run snapshot tests
  run: make test-snapshots
```

```yaml
# Example GitLab CI
test:snapshots:
  script:
    - make test-snapshots
```

## Updating Snapshots

### When to Update

Update snapshots ONLY when you've intentionally changed:
- GraphQL schema fields
- Resolver return structures
- Pagination logic
- Connection edge formats

### How to Update

```bash
# Update all snapshot files
make update-snapshots

# Review changes
git diff internal/graphql/resolvers/testdata/

# If changes are correct, commit them
git add internal/graphql/resolvers/testdata/
git commit -m "refactor(graphql): update connection snapshots for X change"
```

### Update Workflow

1. **Make Your Change**: Modify resolver or schema
2. **Run Tests**: `make test-snapshots` will fail
3. **Review Failure**: Examine what changed in the diff
4. **Update Snapshots**: `make update-snapshots`
5. **Review Golden Files**: `git diff` to see exact changes
6. **Verify Intentional**: Ensure changes match your intent
7. **Commit Together**: Commit code changes + snapshot updates in same PR

## Snapshot Test Architecture

### Test File Structure

```
internal/graphql/resolvers/
├── pagination_snapshot_test.go    # Test definitions
└── testdata/                      # Golden reference files
    ├── product_connection.json
    ├── flyer_connection.json
    ├── flyer_page_connection.json
    ├── shopping_list_connection.json
    ├── shopping_list_item_connection.json
    ├── product_master_connection.json
    └── price_history_connection.json
```

### How It Works

1. **Test Setup**: Each test creates sample data
2. **Build Connection**: Calls connection builder (e.g., `buildProductConnection`)
3. **Marshal JSON**: Converts connection to formatted JSON
4. **Compare**: Byte-for-byte comparison with golden file
5. **Report Diff**: If mismatch, shows expected vs actual

### Example Test

```go
func TestProductConnectionSnapshot(t *testing.T) {
    t.Parallel()
    products := []*models.Product{
        {ID: 1, Name: "Organic Milk", ValidFrom: testTime(2024, 1, 1)},
        {ID: 2, Name: "Barista Oat Milk", ValidFrom: testTime(2024, 1, 8)},
    }
    conn := buildProductConnection(products, 1, 5, 42)
    assertSnapshot(t, "product_connection", conn)
}
```

## Troubleshooting

### Test Fails After Intentional Change

**Problem**: You changed resolver logic, snapshot test fails
**Solution**:
```bash
make update-snapshots
git diff internal/graphql/resolvers/testdata/
# Review, then commit if correct
```

### Test Fails Unexpectedly

**Problem**: You didn't change anything, but test fails
**Investigation**:
1. Check if someone else updated code without updating snapshots
2. Check for timestamp fields (should use fixed test times)
3. Check for randomness in data generation
4. Run `git diff testdata/` to see what changed

**Fix**: Either revert unintended change or update snapshot if change is valid

### Snapshot Out of Sync

**Problem**: Golden file doesn't match current schema
**Solution**:
```bash
# Regenerate all snapshots from current code
make update-snapshots

# Review ALL changes carefully
git diff internal/graphql/resolvers/testdata/

# Commit if intentional, otherwise fix code
```

## Best Practices

### DO

✅ Update snapshots immediately after intentional schema changes
✅ Review snapshot diffs in code review
✅ Keep test data stable (no random values, fixed timestamps)
✅ Commit snapshots with the code that changed them
✅ Run snapshot tests before pushing
✅ Include snapshot checks in CI/CD

### DON'T

❌ Update snapshots without reviewing diffs
❌ Commit snapshot changes without explaining them
❌ Use random data in snapshot tests
❌ Ignore snapshot test failures
❌ Update snapshots to "make tests pass" without understanding why
❌ Skip snapshot tests in CI

## Integration with Development Workflow

### Pre-Commit

```bash
# Before committing resolver changes
make test-snapshots
```

### During Code Review

Reviewers should:
1. Check if `testdata/` files changed
2. Review JSON diffs for correctness
3. Verify changes match PR description
4. Ensure no unintended field removals

### CI/CD Pipeline

```yaml
# Recommended CI steps
steps:
  - checkout
  - install-dependencies
  - run: make test-snapshots  # Fail build if snapshots don't match
  - run: make test            # Run full test suite
```

### Pull Request Checklist

When modifying GraphQL resolvers:
- [ ] Snapshot tests pass OR
- [ ] Snapshot tests updated with `make update-snapshots`
- [ ] Snapshot diffs reviewed and match intentions
- [ ] Breaking changes documented in PR description
- [ ] API consumers notified if connection structure changed

## Schema Changes Impact

### Breaking Changes (Require Snapshot Update)

- Removing fields from connection edges
- Changing field types
- Renaming fields
- Changing pagination structure

### Non-Breaking Changes (May Not Require Update)

- Adding optional fields
- Adding new connection types (new test needed)
- Internal refactoring without output changes

## Advanced Usage

### Running Specific Snapshot Tests

```bash
# Run only product snapshot test
go test ./internal/graphql/resolvers -run TestProductConnectionSnapshot

# Run all connection snapshots
go test ./internal/graphql/resolvers -run Snapshot

# Update only flyer snapshots
go test ./internal/graphql/resolvers -run TestFlyerConnectionSnapshot -update_graphql_snapshots
```

### Adding New Snapshot Tests

1. Create test data
2. Build connection
3. Call `assertSnapshot(t, "name", connection)`
4. Run with `-update_graphql_snapshots` to generate golden file
5. Verify generated JSON is correct
6. Commit test + golden file together

Example:

```go
func TestNewConnectionSnapshot(t *testing.T) {
    t.Parallel()
    items := []*models.NewType{
        {ID: 1, Name: "Test"},
    }
    conn := buildNewConnection(items, 1, 0, 10)
    assertSnapshot(t, "new_connection", conn)
}
```

### Snapshot Metadata

Each golden file contains:
- Connection structure (`edges`, `pageInfo`, `totalCount`)
- Edge cursors (base64 encoded offsets)
- Node data (matching GraphQL schema)

## Maintenance

### Regular Tasks

- **Weekly**: Run `make test-snapshots` on main branch
- **Per Release**: Verify all snapshots match current schema
- **When Migrating**: Update snapshots for GraphQL pagination changes

### Snapshot File Size

Current snapshot files are small (500-1300 bytes each). Keep test data minimal:
- 2-3 items per connection test
- Fixed timestamps
- Simple, representative data

## FAQ

**Q: Why did my test fail when I didn't change anything?**
A: Check if pagination helper was modified. Run `git log internal/graphql/resolvers/helpers.go` to see recent changes.

**Q: Can I run snapshots without Docker?**
A: Yes, snapshot tests don't require database. They test connection builders directly.

**Q: What if I need to update only one snapshot?**
A: Run specific test with update flag: `go test ./internal/graphql/resolvers -run TestProductConnectionSnapshot -update_graphql_snapshots`

**Q: Should snapshots be in .gitignore?**
A: No! Golden files MUST be committed. They're the test expectations.

**Q: How do I know if a change is breaking?**
A: If snapshot diff shows field removals or type changes, it's likely breaking. Check GraphQL schema compatibility.

## Related Documentation

- [GraphQL Schema](../internal/graphql/schema.graphqls)
- [Pagination Helpers](../internal/graphql/resolvers/helpers.go)
- [Connection Builders](../internal/graphql/resolvers/connection_builders.go)
- [REFACTORING_ROADMAP.md](../REFACTORING_ROADMAP.md) - Phase 2.2

## References

- Snapshot tests follow the "Golden File" testing pattern
- Inspired by Jest snapshot testing for JavaScript
- Ensures GraphQL schema backwards compatibility
