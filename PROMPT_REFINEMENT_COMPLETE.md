# AI Prompt Refinement - Complete ✓

## Date: 2025-11-09

## Problem Statement
The AI extraction system was missing many valid promotion modules because it required all promotions to have a price. This broke extraction for:
- Category-level discounts (e.g., "Makaronams -25%")
- Brand-line promotions (e.g., "VIGO šiukšlių maišams -50%")
- Percent-only modules with no specific product price

## Changes Made

### 1. Updated AI Prompts (`internal/services/ai/prompt_builder.go`)

#### ProductExtractionPrompt:
- Added explicit "PRICE OR PERCENT — NOT BOTH REQUIRED" rule
- Clarified promotion types (single_product, category, brand_line, equipment, bundle, loyalty)
- Added "PERCENT-ONLY MODULES" guidance
- Added "MULTIPLE BADGES" handling (keep stronger discount)
- Removed implicit price requirement

#### DetectionPrompt (Pass 1):
- Added "PERCENT-ONLY MODULES ARE VALID" emphasis
- Added "CATEGORY PROMOTIONS" examples
- Added "MULTIPLE BADGES" guidance
- Made clear that modules with only percent badges should be extracted

#### FillDetailsPrompt (Pass 2):
- Added "RESPECT PASS-1 FINDINGS" rule (don't invent prices)
- Added "PERCENT-ONLY IS VALID" emphasis
- Added "MULTIPLE BADGES" handling
- Reinforced "NO HALLUCINATION" for prices

### 2. Updated Extraction Logic (`internal/services/ai/extractor.go`)

#### isValidPromotion:
- Changed validation to accept promotions with EITHER price OR percent
- Added comments clarifying that percent-only modules are valid
- Kept bundle and loyalty markers as valid triggers

#### calculatePromotionConfidence:
- Updated to give confidence score to percent-only promotions (0.15)
- Made clear that either price or percent contributes to confidence

### 3. Updated Service Layer (`internal/services/enrichment/service.go`)

#### convertToProducts:
- Changed to use `result.Promotions` instead of `result.Products`
- Added conversion for ALL promotions (including percent-only)
- Handles discount_pct field directly for percent-only promotions
- Falls back to legacy Products for compatibility

#### assessQuality:
- Updated to count Promotions instead of Products
- Changed error messages to reference "promotions" not "products"
- Uses Promotions for confidence calculation

#### Added extractPromotionTags (in `utils.go`):
- New function to extract tags from Promotion objects
- Mirrors functionality of extractProductTags

### 4. Updated Validation (`internal/services/product_utils.go`)

#### ValidateProduct:
- Changed to allow price = 0 (for percent-only promotions)
- Only rejects negative prices
- Added comment explaining why 0 is valid

## Test Results

### Before Changes:
- **4 products** extracted from 3 pages
- Only products with explicit prices were captured

### After Changes:
- **13 products** extracted from 3 pages
- All expected modules captured:
  1. Pampers sauskelnėms (30%, no price) ✓
  2. VIGO šiukšlių maišams (50%, no price) ✓
  3. BILLA BIO vaikų tyrelėms (25%, no price) ✓
  4. SUDOCREM kremas (4.99 €, 33%) ✓
  5. Bananai (0.99 €) ✓
  6. Vytintiems mėsos gaminiams (30%, no price) ✓
  7. Jogurtams PIENO ŽVAIGŽDĖS (20%, no price) ✓
  8. Makaronams (25%, no price) ✓
  9. Padažams, kečupams (25%, no price) ✓
  10. Guminukams (25%, no price) ✓
  11. Vaisvandeniams (25%, no price) ✓
  12. Moterų higienos priemonėms (25%, no price) ✓
  13. Kapsulinis kavos aparatas LAVAZZA (39.99 €, 67%) ✓

## Key Improvements

1. **Percent-only promotions now work**: Category and brand-line promotions without specific prices are properly extracted
2. **Exact text extraction**: Product names use exact Lithuanian text from flyers
3. **Brand extraction improved**: Brands are properly identified even in category promotions
4. **Multiple badge handling**: When multiple discount badges appear, the stronger one is kept
5. **Special tags captured**: IKI EXPRESS, TIK, SUPER KAINA tags are properly extracted

## Database Schema
No changes to models were required - the existing schema already supported:
- `current_price` as numeric (nullable, can be 0)
- `discount_percent` as numeric (nullable)
- Products can exist with either price or discount_percent or both

## Testing
Run tests with:
```bash
./bin/reset-products
./bin/enrich-flyers --store=iki --max-pages=3 --debug
```

Verify results:
```sql
SELECT id, name, current_price, discount_percent, special_discount, brand, category 
FROM products ORDER BY id;
```

## Conclusion
✓ All 13 test cases pass
✓ The old "must have price" rule has been removed
✓ System now properly extracts category/brand_line promotions with percent-only discounts
✓ No model changes were required
✓ Backward compatible with existing price-based products
