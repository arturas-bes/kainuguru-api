# GraphQL Code Generation - Known Issues

## Current Status

⚠️ **DO NOT run `gqlgen generate` unless absolutely necessary**

The project has a working GraphQL implementation, but regenerating the code with `gqlgen generate` introduces breaking changes.

## Problem Description

### DateTime Scalar Issue

The main issue is with the `DateTime` custom scalar configuration:

1. **Scalar Functions Exist**: Custom marshal/unmarshal functions are properly implemented in `internal/graphql/scalars/datetime.go`:
   - `MarshalDateTime(t time.Time) graphql.Marshaler`
   - `UnmarshalDateTime(v interface{}) (time.Time, error)`

2. **Configuration Issue**: gqlgen is not properly binding these functions when generating code, resulting in:
   ```
   ec.unmarshalInputDateTime undefined (type *executionContext has no field or method unmarshalInputDateTime)
   ec._DateTime undefined (type *executionContext has no field or method _DateTime)
   ```

3. **Generated Code**: When gqlgen regenerates, it creates methods that call non-existent functions on the execution context.

### Enum Mapping Issues

There are also pre-existing enum mapping issues:
```
cannot use data (variable of type []models.FlyerStatus) as []model.FlyerStatus value in assignment
```

## Current Configuration

The `gqlgen.yml` has been improved with:

```yaml
# Autobind to find custom scalar implementations
autobind:
  - github.com/kainuguru/kainuguru-api/internal/graphql/scalars

models:
  DateTime:
    model:
      - time.Time
```

However, this configuration alone doesn't fully resolve the issue.

## Working State

- **Last Working Binary**: `bin/api` from November 8, 2024
- **Last Working Generated File**: Commit `e5090bd` (before Phase 2 refactoring)
- **Current State**: Generated code has issues but was restored to a working version

## Recommendations

### For Development

1. **Don't regenerate** unless you:
   - Add new GraphQL types/mutations/queries
   - Need to update the schema in a breaking way

2. **Manual Resolvers**: All wizard resolvers are manually implemented in `internal/graphql/resolvers/wizard.go`, which means they don't require regeneration.

3. **Schema Changes**: If you must add to the wizard schema:
   - Edit `internal/graphql/schema/wizard.graphql`
   - Manually update `internal/graphql/resolvers/wizard.go`
   - Update `internal/graphql/model/models_gen.go` if needed
   - **Skip gqlgen generate**

### For Future Fixes

To properly fix this issue, you would need to:

1. **Update gqlgen version**: Check if newer versions handle custom scalars better
2. **Review Configuration**: Study gqlgen documentation for proper scalar configuration
3. **Add Scalar Bindings**: Ensure gqlgen knows about the scalar package functions
4. **Test Generation**: Test on a separate branch before applying to main code
5. **Fix Enum Mappings**: Resolve the enum type mismatch issues

## Alternative Approach

Consider using gqlgen's newer configuration options:

```yaml
# In gqlgen.yml - experimental approach
scalars:
  DateTime:
    input:
      type: time.Time
      marshaler: github.com/kainuguru/kainuguru-api/internal/graphql/scalars.UnmarshalDateTime
    output:
      type: time.Time
      marshaler: github.com/kainuguru/kainuguru-api/internal/graphql/scalars.MarshalDateTime
```

But this would require:
1. Testing with the current gqlgen version
2. Potentially updating dependencies
3. Comprehensive testing of all GraphQL operations

## Current Workaround

The project uses **manual resolver implementation** (via `filename_template: "-"` in gqlgen.yml), which means:

- Resolvers are hand-written in `internal/graphql/resolvers/`
- Changes to resolvers don't require regeneration
- Only schema additions would need attention

This is actually a **good practice** and provides more control over the implementation.

## Testing GraphQL Changes

To test if GraphQL changes work without regenerating:

```bash
# 1. Check if code compiles
go build ./internal/graphql/resolvers/...

# 2. Run resolver tests
go test ./internal/graphql/resolvers/...

# 3. Test with API server
go run ./cmd/api
# Then test queries/mutations with GraphQL client
```

## Related Files

- Configuration: `gqlgen.yml`
- Scalar Implementations: `internal/graphql/scalars/`
- Generated Code: `internal/graphql/generated/generated.go`
- Manual Resolvers: `internal/graphql/resolvers/`
- Schema: `internal/graphql/schema/*.graphql`

## Conclusion

The wizard implementation is **fully functional** without regenerating GraphQL code. The manual resolver approach provides better control and avoids generation issues.

If you need to regenerate in the future, be prepared to:
1. Debug scalar binding issues
2. Fix enum type mismatches
3. Test all GraphQL operations thoroughly
4. Potentially revert and use manual implementation

**Recommendation**: Stick with manual resolvers for critical features like the wizard.
