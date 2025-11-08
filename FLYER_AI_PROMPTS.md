# Production-Ready AI Prompts for Lithuanian Flyer Enrichment System

## Table of Contents
1. [Main Product Extraction Prompt](#1-main-product-extraction-prompt)
2. [Validation Prompt](#2-validation-prompt)
3. [Category Classification Prompt](#3-category-classification-prompt)
4. [Quality Check Prompt](#4-quality-check-prompt)
5. [Fallback OCR Prompt](#5-fallback-ocr-prompt)
6. [Price Analysis Prompt](#6-price-analysis-prompt)
7. [Testing Strategies](#testing-strategies)
8. [Optimization Recommendations](#optimization-recommendations)
9. [Token Usage Estimates](#token-usage-estimates)
10. [Success Metrics](#success-metrics)

---

## 1. Main Product Extraction Prompt

### The Prompt
```
You are an expert AI system specialized in extracting product information from Lithuanian retail flyer images. You have extensive knowledge of Lithuanian grocery products, pricing formats, and retail terminology.

TASK: Extract ALL visible products with prices from this flyer page image.

CONTEXT:
- Store: {store_name} ({store_code})
- Page: {page_number}
- Expected language: Lithuanian
- Currency: EUR (€)

EXTRACTION REQUIREMENTS:

For each product, extract:
1. name: Product name EXACTLY as written (preserve Lithuanian diacritics: ą, č, ę, ė, į, š, ų, ū, ž)
2. price: Current selling price (format: "X,XX €" or "X.XX €")
3. original_price: Original price before discount (if visible)
4. unit: Quantity/weight (kg, g, l, ml, vnt., pak., dėž.)
5. brand: Brand name if visible (e.g., "Dvaro", "Rokiškio", "Svalya")
6. category: One of: ["mėsa ir žuvis", "pieno produktai", "duona ir konditerija", "vaisiai ir daržovės", "gėrimai", "šaldyti produktai", "konservai", "kruopos ir makaronai", "saldumynai", "higienos prekės", "namų ūkio prekės", "alkoholiniai gėrimai", "kita"]
7. discount_percentage: If shown (e.g., "-25%")
8. discount_type: "percentage" | "absolute" | "bundle" | "loyalty"
9. confidence: Your confidence level (0.0-1.0)
10. bounding_box: {x: 0-1, y: 0-1, width: 0-1, height: 0-1}
11. position: {row: 1-N, column: 1-N, zone: "header"|"main"|"footer"|"sidebar"}

SPECIAL INSTRUCTIONS:

Price Handling:
- Look for both crossed-out and current prices
- Identify loyalty card prices (e.g., "su EŽYS kortele")
- Handle multi-buy offers (e.g., "3 vnt. už 5 €")
- Detect per-kg prices (e.g., "12,99 €/kg")

Lithuanian Text:
- Preserve ALL diacritical marks
- Common units: "vnt." (unit), "pak." (package), "dėž." (box)
- Watch for: "AKCIJA", "SUPER KAINA", "NAUJIENA", "TIKTAI"

Store-Specific Patterns:
- IKI: Yellow price tags, "IKI KAINOS" labels
- MAXIMA: Red accents, "X" card prices
- RIMI: Orange highlights, member prices

Quality Checks:
- Minimum 5 products expected per page
- Skip decorative elements and headers
- Ignore products without clear prices
- Flag bundles as single items with bundle notation

OUTPUT FORMAT:
{
  "extraction_metadata": {
    "total_products_found": number,
    "extraction_confidence": 0.0-1.0,
    "page_quality": "high"|"medium"|"low",
    "warnings": []
  },
  "products": [
    {
      "name": "Kiaulienos sprandinė KREKENAVOS, 1 kg",
      "price": "4,59 €",
      "original_price": "5,99 €",
      "unit": "1 kg",
      "brand": "KREKENAVOS",
      "category": "mėsa ir žuvis",
      "discount_percentage": "-23%",
      "discount_type": "percentage",
      "confidence": 0.95,
      "bounding_box": {
        "x": 0.125,
        "y": 0.234,
        "width": 0.185,
        "height": 0.142
      },
      "position": {
        "row": 2,
        "column": 1,
        "zone": "main"
      }
    }
  ]
}

IMPORTANT:
- Return ONLY valid JSON
- Include ALL products with visible prices
- Use null for missing optional fields
- Coordinates are normalized (0=top/left, 1=bottom/right)
```

### Implementation Notes
- Uses role-based expertise establishment
- Includes store-specific visual cue recognition
- Handles Lithuanian diacritics preservation
- Provides structured confidence scoring
- Includes spatial information for UI rendering

---

## 2. Validation Prompt

### The Prompt
```
You are a quality assurance specialist for Lithuanian retail data. Validate and correct the extracted product information.

EXTRACTED DATA TO VALIDATE:
{extracted_json}

VALIDATION RULES:

1. Price Validation:
   - Format must be "X,XX €" or "X.XX €"
   - Price must be positive and realistic (0.01-9999.99)
   - Original price must be higher than current price if both exist
   - Calculate discount percentage: ((original - current) / original) * 100

2. Lithuanian Text:
   - Verify diacritical marks are present where expected
   - Common words requiring diacritics: "mėsa", "pienas", "duona", "daržovės"
   - Brand names: Check against known Lithuanian brands

3. Units Standardization:
   INPUT -> OUTPUT
   - "kilogramas", "kg." -> "kg"
   - "gramas", "gr" -> "g"
   - "litras", "ltr" -> "l"
   - "mililitras" -> "ml"
   - "vienetų", "vienetas" -> "vnt."
   - "pakuotė" -> "pak."
   - "dėžutė" -> "dėž."

4. Category Validation:
   Product keywords -> Category mapping:
   - ["kiauliena", "jautiena", "vištiena", "žuvis", "lašiša"] -> "mėsa ir žuvis"
   - ["pienas", "kefyras", "jogurtas", "sūris", "sviestas"] -> "pieno produktai"
   - ["duona", "batonas", "pyragas", "tortas"] -> "duona ir konditerija"
   - ["obuoliai", "bananai", "pomidorai", "agurkai"] -> "vaisiai ir daržovės"

5. Data Completeness:
   Required fields: name, price
   Flag for review if:
   - Name < 3 characters or > 150 characters
   - Price format invalid
   - Confidence < 0.6
   - Suspicious patterns: "test", "error", "unknown"

6. Logical Consistency:
   - Discount percentage should match price difference
   - Bundle prices should indicate quantity
   - Category should match product type

OUTPUT FORMAT:
{
  "validation_summary": {
    "total_validated": number,
    "corrections_made": number,
    "products_removed": number,
    "requires_manual_review": boolean
  },
  "validated_products": [
    {
      ...all original fields...,
      "validation_status": "valid"|"corrected"|"flagged",
      "corrections": ["price_format", "unit_normalized", "category_fixed"],
      "validation_score": 0.0-1.0
    }
  ],
  "removed_products": [
    {
      "original_name": "string",
      "removal_reason": "no_price"|"invalid_data"|"duplicate"|"low_confidence"
    }
  ],
  "manual_review_needed": [
    {
      "product_name": "string",
      "review_reason": "string",
      "suggested_action": "string"
    }
  ]
}

Apply all validation rules and return corrected data.
```

### Implementation Notes
- Multi-stage validation process
- Lithuanian-specific text validation
- Automatic unit standardization
- Confidence-based flagging system
- Detailed correction tracking

---

## 3. Category Classification Prompt

### The Prompt
```
You are a Lithuanian retail categorization expert. Classify this product into the most appropriate category.

PRODUCT INFORMATION:
Name: {product_name}
Brand: {brand}
Unit: {unit}
Price: {price}

AVAILABLE CATEGORIES WITH KEYWORDS:

1. "mėsa ir žuvis"
   Keywords: kiauliena, jautiena, vištiena, kalakutiena, antiena, aviena, triušiena, lašiša, menkė, silkė, karpis, upėtakis, krevečių, krabų, dešra, kumpis, šoninė, faršas, kotletai, kepsnys

2. "pieno produktai"
   Keywords: pienas, kefyras, jogurtas, varškė, sūris, sviestas, grietinė, grietinėlė, glaistytas sūrelis, ryžių pienas, sojų pienas, mocarela, fermentinis, pasukos, rūgpienis

3. "duona ir konditerija"
   Keywords: duona, batonas, kepalėlis, lavašas, duonelė, bandelė, pyragas, tortas, napoleonas, eklerai, spurga, riestainiai, sausainiai, vafliai, krekeriai, grissini

4. "vaisiai ir daržovės"
   Keywords: obuoliai, bananai, apelsinai, mandarinai, citrina, ananasas, vynuogės, braškės, mėlynės, bulvės, morkos, kopūstai, pomidorai, agurkai, svogūnai, česnakai, salotos, brokoliai

5. "gėrimai"
   Keywords: vanduo, sultys, gėrimas, limonadas, energetinis, kava, arbata, kakava, mineralinis, gazuotas, natūralus, nektaras, kompota, morsas

6. "šaldyti produktai"
   Keywords: ledai, šaldytos daržovės, šaldytos uogos, šaldyta pica, koldūnai, virtiniai, šaldytas, užšaldytas, greitai užšaldyta

7. "konservai"
   Keywords: konservuoti, pomidorų pasta, padažas, marinuoti agurkai, šprotai, aliejuje, savo sultyse, stiklainis, skardinė

8. "kruopos ir makaronai"
   Keywords: ryžiai, grikiai, makaronai, spagečiai, avižinės, manai, perlinės, kruopos, kuskusas, vermišeliai

9. "saldumynai"
   Keywords: šokoladas, saldainiai, karamelė, šokoladukai, batonėlis, marmeladas, zefyras, chalva, medus, uogienė, džemas

10. "higienos prekės"
    Keywords: šampūnas, muilas, dantų pasta, tualetinis popierius, servetėlės, šluostės, dezodorantas, gelis, kremas

11. "namų ūkio prekės"
    Keywords: ploviklis, valiklis, skalbiklis, indų, grindų, langų, kempinėlė, šepetys, šiukšlių maišai

12. "alkoholiniai gėrimai"
    Keywords: alus, vynas, degtinė, brendis, viskis, šampanas, sidras, likeris, spiritinis

13. "kita"
    Default category for unmatched products

CLASSIFICATION LOGIC:
1. Exact keyword match (highest priority)
2. Partial keyword match in product name
3. Brand association (e.g., "Švyturys" -> "alkoholiniai gėrimai")
4. Unit hints (e.g., "0.5 l" often beverages, "kg" often meat/produce)
5. Price range patterns

OUTPUT:
{
  "product_name": "{original_name}",
  "selected_category": "category_name",
  "confidence": 0.95,
  "matching_keywords": ["keyword1", "keyword2"],
  "reasoning": "Produktas '{name}' atitinka kategoriją '{category}' nes turi raktažodžius: {keywords}",
  "alternative_category": "second_best_match",
  "alternative_confidence": 0.65
}

Classify the product and explain your reasoning in Lithuanian.
```

### Implementation Notes
- Comprehensive keyword mapping
- Multi-tier matching logic
- Confidence scoring with alternatives
- Lithuanian reasoning explanation
- Brand-aware classification

---

## 4. Quality Check Prompt

### The Prompt
```
You are a quality control specialist for flyer page analysis. Assess the overall quality and completeness of this extracted data.

PAGE DATA:
{extracted_page_data}

METADATA:
Store: {store_code}
Page Number: {page_number}
Extraction Timestamp: {timestamp}

QUALITY ASSESSMENT CRITERIA:

1. Quantity Check:
   - Minimum expected: 5 products per page
   - Typical range: 8-25 products
   - Maximum reasonable: 40 products

2. Data Completeness (per product):
   - Critical fields: name, price (100% required)
   - Important fields: unit, category (80% expected)
   - Optional fields: brand, discount (varies)

3. Price Distribution Analysis:
   - Check for reasonable price range (0.19 € to 299.99 €)
   - Verify discount percentages (typically 5-50%)
   - Flag suspicious patterns (all same price, all round numbers)

4. Text Quality:
   - Lithuanian text presence
   - Reasonable product name lengths (5-100 chars)
   - No placeholder text ("Lorem ipsum", "test", etc.)

5. Spatial Coverage:
   - Products distributed across page (not all in one corner)
   - Bounding boxes don't overlap excessively
   - Coverage area: expect 40-80% of page area

6. Category Distribution:
   - Unusually homogeneous (all same category) may indicate error
   - Empty flyer pages often have <3 products

7. Common Issues to Detect:
   - Duplicate products (same name, same price)
   - Header/footer text extracted as products
   - Page numbers or dates extracted as prices
   - Advertisement slogans as product names

DECISION TREE:
```
Products < 3 -> "empty_page"
Products 3-4 + low confidence -> "needs_review"
Products >= 5 + high confidence -> "good_quality"
Extraction errors > 30% -> "failed_extraction"
Suspicious patterns -> "manual_review"
```

OUTPUT FORMAT:
{
  "quality_assessment": {
    "overall_quality": "high"|"medium"|"low"|"failed",
    "quality_score": 0.0-1.0,
    "product_count": number,
    "page_coverage": 0.0-1.0,
    "data_completeness": 0.0-1.0,
    "confidence_average": 0.0-1.0
  },
  "quality_flags": {
    "has_minimum_products": boolean,
    "has_valid_prices": boolean,
    "has_lithuanian_text": boolean,
    "has_reasonable_distribution": boolean,
    "has_diverse_categories": boolean
  },
  "issues_detected": [
    {
      "issue_type": "low_product_count"|"duplicate_products"|"price_errors"|"text_quality"|"spatial_issues",
      "severity": "critical"|"high"|"medium"|"low",
      "description": "Detailed description in Lithuanian",
      "affected_products": [indices],
      "recommendation": "Suggested action"
    }
  ],
  "extraction_quality_metrics": {
    "price_format_accuracy": 0.0-1.0,
    "category_assignment_rate": 0.0-1.0,
    "unit_extraction_rate": 0.0-1.0,
    "discount_detection_accuracy": 0.0-1.0
  },
  "recommendation": "approve"|"review"|"reprocess"|"reject",
  "reprocessing_hints": {
    "suggested_approach": "enhanced_ocr"|"manual_review"|"different_model",
    "problem_areas": ["specific issues to address"]
  }
}

Provide comprehensive quality assessment with actionable recommendations.
```

### Implementation Notes
- Multi-dimensional quality scoring
- Automatic issue detection
- Actionable recommendations
- Reprocessing guidance
- Statistical validation

---

## 5. Fallback OCR Prompt

### The Prompt
```
You are a text extraction specialist. Extract ALL readable text from this Lithuanian retail flyer image when structured extraction fails.

EXTRACTION MODE: Simplified OCR Fallback

TASK: Extract all visible text, focusing on recovering basic product and price information.

TEXT EXTRACTION PRIORITIES:
1. Product names (any text that appears to be a product)
2. Prices (any number with €, Eur, or decimal points)
3. Percentages (discounts like -20%, -30%)
4. Units (kg, g, l, ml, vnt.)
5. Brand names (typically in capitals)
6. Promotional text (AKCIJA, SUPER KAINA, etc.)

SCANNING PATTERN:
- Top to bottom, left to right
- Group related text by proximity
- Maintain relative positioning

PRICE PATTERNS TO DETECT:
- "X,XX €" or "X.XX €"
- "X,XX" near € symbol
- "€ X,XX"
- Crossed out prices (original prices)
- "nuo X,XX €" (from X.XX €)

LITHUANIAN TEXT MARKERS:
Look for common Lithuanian product words:
- Pienas, mėsa, duona, sūris, sviestas
- Kiaulienos, jautienos, vištienos
- Švieži, šaldyti, kepti, rūkyti
- Common endings: -as, -is, -us, -ė, -iai

OUTPUT FORMAT:
{
  "extraction_mode": "fallback_ocr",
  "extraction_confidence": 0.0-1.0,
  "text_blocks": [
    {
      "block_id": 1,
      "text_content": "Extracted text here",
      "block_type": "product"|"price"|"header"|"promo"|"unknown",
      "position_hint": "top-left"|"top-center"|"top-right"|"middle-left"|"center"|"middle-right"|"bottom-left"|"bottom-center"|"bottom-right",
      "associated_elements": {
        "probable_price": "X,XX €",
        "probable_unit": "kg",
        "probable_discount": "-25%"
      }
    }
  ],
  "extracted_prices": [
    "4,99 €",
    "2,49 €",
    "12,90 €"
  ],
  "extracted_products": [
    "Pienas ŽEMAITIJOS 2.5%",
    "Sviestas ROKIŠKIO 82%"
  ],
  "raw_text_dump": "All text concatenated for backup processing",
  "recovery_statistics": {
    "total_text_blocks": number,
    "prices_found": number,
    "products_identified": number,
    "lithuanian_text_detected": boolean
  }
}

IMPORTANT:
- This is a fallback method - prioritize text recovery over structure
- Group likely related text elements
- Don't attempt complex parsing - just extract text
- Flag this data as requiring manual review
```

### Implementation Notes
- Simplified extraction for robustness
- Pattern-based text grouping
- Position hints for manual review
- Raw text preservation
- Recovery statistics

---

## 6. Price Analysis Prompt

### The Prompt
```
You are a pricing analyst specializing in Lithuanian retail. Analyze and validate pricing information from extracted flyer data.

PRODUCTS TO ANALYZE:
{products_json}

STORE CONTEXT:
Store: {store_name}
Date: {flyer_date}
Region: Lithuania

PRICE ANALYSIS TASKS:

1. Format Validation:
   - Standard format: "X,XX €" (comma as decimal separator)
   - Alternative accepted: "X.XX €"
   - Range validation: 0.01 € to 9999.99 €
   - Check for missing currency symbols

2. Discount Calculation:
   Original: 5,99 € → Current: 3,99 €
   Discount = ((5.99 - 3.99) / 5.99) × 100 = 33.39%
   Verify all discount percentages match calculations

3. Price Reasonableness by Category:
   Category → Expected range:
   - "pieno produktai": 0.39 € - 15.99 €
   - "mėsa ir žuvis": 2.99 € - 49.99 €
   - "vaisiai ir daržovės": 0.29 € - 9.99 €
   - "alkoholiniai gėrimai": 0.99 € - 199.99 €
   - "higienos prekės": 0.49 € - 29.99 €

4. Bundle Analysis:
   Detect patterns:
   - "2 vnt. už X €" → calculate unit price
   - "3+1 nemokamai" → calculate effective discount
   - "Antras -50%" → calculate average unit price

5. Price-Per-Unit Calculation:
   Product: "Obuoliai, 2 kg" Price: "3,98 €"
   → Price per kg: 1,99 €/kg
   → Compare to market average

6. Loyalty Price Detection:
   - "Su kortele" prices
   - Member-only discounts
   - App-exclusive prices

7. Temporal Price Validation:
   - Compare to historical prices if available
   - Flag unusual price changes (>50% change)

8. Competition Benchmarking:
   Common products reference prices:
   - Pienas 2.5% 1L: ~0.89-1.29 €
   - Duona (batonas): ~0.59-0.99 €
   - Kiaušiniai 10 vnt.: ~1.49-2.49 €

ANALYSIS OUTPUT:
{
  "price_analysis_summary": {
    "total_products_analyzed": number,
    "valid_prices": number,
    "corrected_prices": number,
    "suspicious_prices": number,
    "average_discount_percentage": number
  },
  "price_statistics": {
    "min_price": "0,29 €",
    "max_price": "45,99 €",
    "median_price": "2,99 €",
    "average_price": "4,57 €",
    "price_distribution": {
      "under_1_eur": number,
      "1_to_5_eur": number,
      "5_to_10_eur": number,
      "10_to_20_eur": number,
      "over_20_eur": number
    }
  },
  "discount_analysis": {
    "products_on_discount": number,
    "average_discount": percentage,
    "max_discount": percentage,
    "discount_distribution": {
      "5_to_10_percent": number,
      "10_to_25_percent": number,
      "25_to_50_percent": number,
      "over_50_percent": number
    }
  },
  "price_corrections": [
    {
      "product_name": "string",
      "original_price": "string",
      "corrected_price": "string",
      "correction_reason": "format"|"calculation"|"outlier",
      "confidence": 0.0-1.0
    }
  ],
  "bundle_deals": [
    {
      "description": "2 vnt. Jogurtas ACTIVIA",
      "bundle_price": "3,98 €",
      "unit_price": "1,99 €",
      "savings": "0,60 €",
      "discount_percentage": 13.1
    }
  ],
  "price_per_unit_calculations": [
    {
      "product": "string",
      "total_price": "string",
      "quantity": "string",
      "price_per_unit": "string",
      "unit_type": "kg"|"l"|"vnt.",
      "market_comparison": "below_average"|"average"|"above_average"
    }
  ],
  "suspicious_prices": [
    {
      "product": "string",
      "price": "string",
      "issue": "too_high"|"too_low"|"format_error"|"calculation_mismatch",
      "expected_range": "X,XX € - Y,YY €",
      "recommendation": "manual_review"|"auto_correct"
    }
  ],
  "pricing_insights": {
    "most_discounted_category": "string",
    "best_deals": ["product1", "product2"],
    "price_competitiveness": "very_competitive"|"competitive"|"average"|"expensive",
    "promotional_intensity": "high"|"medium"|"low"
  }
}

Provide comprehensive pricing analysis with corrections and insights.
```

### Implementation Notes
- Mathematical discount validation
- Category-based price validation
- Bundle deal calculation
- Market price benchmarking
- Competitive insights generation

---

## Testing Strategies

### 1. Unit Testing Each Prompt
```python
def test_product_extraction_prompt():
    """Test cases for main extraction prompt"""
    test_cases = [
        {
            "name": "Standard product page",
            "image": "standard_flyer_page.jpg",
            "expected_min_products": 8,
            "expected_categories": ["mėsa ir žuvis", "pieno produktai"]
        },
        {
            "name": "Sale-heavy page",
            "image": "discount_page.jpg",
            "expected_discount_products": 5,
            "verify_discount_calculations": True
        },
        {
            "name": "Bundle offers page",
            "image": "bundle_page.jpg",
            "expected_bundle_deals": 3
        }
    ]

    for case in test_cases:
        result = extract_products(case["image"])
        assert len(result["products"]) >= case.get("expected_min_products", 5)
        # Additional assertions...
```

### 2. Lithuanian Language Validation
```python
def test_lithuanian_text_preservation():
    """Ensure diacritics are preserved"""
    test_products = [
        "Šviežia kiaulienos šoninė",
        "Varškės sūreliai ŽEMAITIJOS",
        "Česnakų duonelė"
    ]
    # Test extraction maintains diacritics
```

### 3. Price Format Testing
```python
def test_price_formats():
    """Test various price format handling"""
    price_formats = [
        ("2,99 €", True),
        ("2.99 €", True),
        ("€ 2,99", True),
        ("2,99", False),  # Missing currency
        ("2,999 €", False)  # Wrong format
    ]
```

### 4. Integration Testing
- Test full pipeline: Image → Extraction → Validation → Storage
- Test with real flyer images from each store
- Verify database schema compatibility

### 5. Performance Testing
```python
def test_extraction_performance():
    """Monitor extraction speed and token usage"""
    metrics = {
        "avg_time_per_page": 0,
        "avg_tokens_per_page": 0,
        "success_rate": 0
    }
```

---

## Optimization Recommendations

### 1. Prompt Optimization

**Token Reduction Strategies:**
- Use prompt compression (remove redundant instructions)
- Implement prompt caching for repeated elements
- Use shorter JSON keys in production ("n" instead of "name")

**Improved Example:**
```python
# Compressed production prompt (saves ~30% tokens)
COMPRESSED_PROMPT = """
Extract products from Lithuanian flyer. Return JSON:
{"p":[{"n":"name","pr":"price","u":"unit","c":"category"}]}
Categories: mėsa,pienas,duona,vaisiai,gėrimai,šaldyti,konservai,kruopos,saldumynai,higiena,namų,alkoholis,kita
"""
```

### 2. Batching Strategy

```python
def optimize_batch_processing():
    """Process multiple pages efficiently"""
    config = {
        "batch_size": 5,  # Process 5 pages at once
        "parallel_requests": 3,  # Max concurrent API calls
        "retry_strategy": "exponential_backoff",
        "cache_duration": 3600  # 1 hour cache
    }
```

### 3. Cost Optimization

**Model Selection:**
```python
MODEL_SELECTION = {
    "high_quality_extraction": "gpt-4-vision-preview",  # Best accuracy
    "standard_extraction": "gpt-4-vision",  # Good balance
    "quick_validation": "gpt-3.5-turbo",  # Text-only validation
    "fallback_ocr": "gpt-4-vision"  # Robust fallback
}
```

**Caching Strategy:**
```python
def implement_smart_caching():
    """Cache extraction results intelligently"""
    cache_rules = {
        "successful_extraction": 24 * 3600,  # 24 hours
        "failed_extraction": 3600,  # 1 hour retry
        "validation_results": 7 * 24 * 3600  # 1 week
    }
```

### 4. Error Recovery

```python
def implement_graceful_degradation():
    """Fallback strategy for failures"""
    pipeline = [
        ("main_extraction", 0.95),  # Try main prompt
        ("simplified_extraction", 0.80),  # Simpler prompt
        ("ocr_fallback", 0.60),  # Basic OCR
        ("manual_queue", 0.0)  # Queue for manual review
    ]
```

---

## Token Usage Estimates

### Per-Prompt Token Estimates

| Prompt Type | Input Tokens | Output Tokens | Total | Cost (GPT-4V) |
|------------|--------------|---------------|--------|---------------|
| Main Extraction | ~1,200 | ~800 | ~2,000 | $0.06 |
| Validation | ~600 | ~400 | ~1,000 | $0.03 |
| Category Classification | ~300 | ~100 | ~400 | $0.012 |
| Quality Check | ~500 | ~300 | ~800 | $0.024 |
| OCR Fallback | ~400 | ~600 | ~1,000 | $0.03 |
| Price Analysis | ~700 | ~500 | ~1,200 | $0.036 |

### Full Pipeline Estimates

**Per Flyer Page:**
- Main extraction + Validation: ~3,000 tokens ($0.09)
- With quality checks: ~3,800 tokens ($0.114)
- With all validations: ~5,000 tokens ($0.15)

**Per Complete Flyer (20 pages):**
- Standard processing: ~60,000 tokens ($1.80)
- With full validation: ~100,000 tokens ($3.00)

### Monthly Estimates

**Assuming 100 flyers/month, 20 pages each:**
- Token usage: ~6-10 million tokens
- Estimated cost: $180-$300/month
- With caching (30% reduction): $126-$210/month

---

## Success Metrics

### 1. Extraction Accuracy Metrics

```python
ACCURACY_TARGETS = {
    "product_name_accuracy": 0.95,  # 95% correct product names
    "price_accuracy": 0.98,  # 98% correct prices
    "category_accuracy": 0.90,  # 90% correct categories
    "discount_accuracy": 0.92,  # 92% correct discount calculations
    "overall_accuracy": 0.93  # 93% overall accuracy
}
```

### 2. Performance Metrics

```python
PERFORMANCE_TARGETS = {
    "avg_extraction_time": 3.0,  # seconds per page
    "success_rate": 0.95,  # 95% successful extractions
    "retry_rate": 0.05,  # <5% require retry
    "manual_review_rate": 0.02,  # <2% need manual review
    "api_error_rate": 0.01  # <1% API errors
}
```

### 3. Business Metrics

```python
BUSINESS_METRICS = {
    "products_extracted_per_day": 10000,
    "flyers_processed_per_hour": 5,
    "cost_per_product": 0.003,  # $0.003 per product
    "time_saved_vs_manual": 0.95,  # 95% time reduction
    "data_completeness": 0.92  # 92% of fields populated
}
```

### 4. Quality Metrics

```python
QUALITY_METRICS = {
    "lithuanian_text_preservation": 0.98,
    "bounding_box_accuracy": 0.85,
    "duplicate_detection_rate": 0.99,
    "false_positive_rate": 0.02,
    "price_validation_accuracy": 0.97
}
```

### 5. Monitoring Dashboard

```python
def create_monitoring_dashboard():
    """Real-time metrics monitoring"""
    return {
        "daily_metrics": {
            "total_pages_processed": 0,
            "successful_extractions": 0,
            "failed_extractions": 0,
            "average_products_per_page": 0,
            "total_tokens_used": 0,
            "total_cost": 0
        },
        "alerts": {
            "low_extraction_rate": "< 5 products/page",
            "high_failure_rate": "> 10% failures",
            "cost_overrun": "> $10/day",
            "api_errors": "> 5% error rate"
        },
        "trending_metrics": {
            "extraction_accuracy_7d": [],
            "cost_per_product_7d": [],
            "processing_speed_7d": []
        }
    }
```

---

## Implementation Priorities

### Phase 1: Core Implementation (Week 1)
1. Implement main extraction prompt
2. Basic validation prompt
3. Database integration
4. Error handling

### Phase 2: Optimization (Week 2)
1. Category classification
2. Price analysis
3. Caching layer
4. Batch processing

### Phase 3: Quality & Scaling (Week 3)
1. Quality check system
2. OCR fallback
3. Monitoring dashboard
4. Performance optimization

### Phase 4: Production Readiness (Week 4)
1. A/B testing different prompts
2. Cost optimization
3. Manual review interface
4. Documentation & training

---

## Conclusion

These production-ready prompts provide:
- **Comprehensive extraction** with Lithuanian language support
- **Multi-stage validation** ensuring data quality
- **Cost-effective processing** with optimization strategies
- **Robust error handling** with fallback mechanisms
- **Clear success metrics** for monitoring and improvement

The system is designed to handle the specific challenges of Lithuanian retail flyers while maintaining high accuracy and cost efficiency. The modular approach allows for incremental implementation and continuous optimization based on real-world performance data.