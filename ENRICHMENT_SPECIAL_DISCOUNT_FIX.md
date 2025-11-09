# Special Discount Field - Fixed

## Date: 2025-11-09

## Issue
The `special_discount` field was not being populated during enrichment, even though:
- The database column existed
- The model had the field
- The service was converting it

## Root Cause
The AI prompt in `prompt_builder.go` did not include `special_discount` in:
1. The extraction requirements list
2. The output format example
3. The special instructions

This meant the AI was never instructed to extract special discount information.

## Fix Applied

### File: `internal/services/ai/prompt_builder.go`

**1. Added to Extraction Requirements:**
```go
9. special_discount: Special offers like "1+1", "2+1", "3 už 2 €", "SUPER KAINA" (extract exactly as shown)
```

**2. Updated OUTPUT FORMAT Example:**
```json
{
  "name": "Kiaulienos sprandinė KREKENAVOS, 1 kg",
  "price": "4,59 €",
  "original_price": "5,99 €",
  "unit": "1 kg",
  "brand": "KREKENAVOS",
  "category": "mėsa ir žuvis",
  "discount": "-23%",
  "discount_type": "percentage",
  "special_discount": "1+1",  // <- ADDED
  "confidence": 0.95,
  ...
}
```

**3. Enhanced Price Handling Instructions:**
```
Price Handling:
- Look for both crossed-out and current prices
- Identify loyalty card prices (e.g., "su EŽYS kortele")
- Handle multi-buy offers (e.g., "3 vnt. už 5 €") - extract to special_discount field
- Detect per-kg prices (e.g., "12,99 €/kg")
- Look for special promotions: "1+1", "2+1 GRATIS", "3 už 2 €", "SUPER KAINA", etc.
```

## Verification

### Test Run Results
```bash
./bin/enrich-flyers --store=iki --max-pages=3
```

**Products Extracted with Special Discounts:**
```sql
id  | name                     | special_discount
----|--------------------------|------------------
146 | Džiovinti mangai         | SUPER KAINA
147 | Šilauogės                | TIK
148 | Slyvos ANGELENO          | TIK
149 | Feichojos                | TIK
150 | CLEVER svogūnai          | SUPER KAINA
153 | Žaliosios cukinijos      | 1+1
```

## Types of Special Discounts Detected

1. **"SUPER KAINA"** - Super price promotion
2. **"TIK"** - Only/exclusive (often part of loyalty card promotions)
3. **"1+1"** - Buy one get one free
4. **"2+1"** - Buy two get one free
5. **"3 už 2 €"** - Three for 2 euros
6. **Other bundle offers** - Various multi-buy promotions

## GraphQL Exposure

The field is already exposed in the GraphQL schema:

```graphql
type ProductPrice {
  current: Float!
  original: Float
  discountPercent: Float
  specialDiscount: String  # Available for queries
}
```

## Status
✅ **FIXED** - Special discounts are now being extracted and stored correctly.

## Related Files
- `internal/services/ai/prompt_builder.go` - Prompt generation
- `internal/services/ai/extractor.go` - ExtractedProduct struct
- `internal/services/enrichment/service.go` - Product conversion
- `internal/models/product.go` - Product model
- `migrations/032_add_special_discount_to_products.sql` - Database schema
- `internal/graphql/schema/schema.graphql` - GraphQL schema
