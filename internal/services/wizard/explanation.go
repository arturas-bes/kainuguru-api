package wizard

import (
	"fmt"
	"math"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// GenerateExplanation creates a human-readable explanation for a suggestion
// Examples:
//   - "Same brand, similar size, €0.50 cheaper"
//   - "Similar product, different brand, 10% more expensive"
//   - "Same brand, exact size match, best price in your preferred store"
func GenerateExplanation(
	suggestion *models.Suggestion,
	originalProduct *models.Product,
	isPreferredStore bool,
) string {
	parts := []string{}

	// 1. Brand comparison
	brandMatch := false
	if suggestion.Brand != nil && originalProduct.Brand != nil {
		if *suggestion.Brand == *originalProduct.Brand {
			parts = append(parts, "Same brand")
			brandMatch = true
		} else {
			parts = append(parts, "Different brand")
		}
	} else {
		parts = append(parts, "Similar product")
	}

	// 2. Size comparison
	if suggestion.SizeValue != nil && suggestion.SizeUnit != nil {
		if originalProduct.UnitType != nil && *originalProduct.UnitType == *suggestion.SizeUnit {
			parts = append(parts, "similar size")
		} else {
			parts = append(parts, "different size")
		}
	}

	// 3. Price comparison
	priceExplanation := generatePriceExplanation(originalProduct.CurrentPrice, suggestion.Price)
	if priceExplanation != "" {
		parts = append(parts, priceExplanation)
	}

	// 4. Store preference note (optional)
	if isPreferredStore && brandMatch {
		parts = append(parts, "at your preferred store")
	}

	// Join parts with commas
	explanation := ""
	for i, part := range parts {
		if i == 0 {
			explanation = part
		} else if i == len(parts)-1 {
			explanation += ", " + part
		} else {
			explanation += ", " + part
		}
	}

	return explanation
}

// generatePriceExplanation creates price comparison text
// Examples: "€0.50 cheaper", "10% more expensive", "same price"
func generatePriceExplanation(originalPrice, suggestedPrice float64) string {
	if originalPrice == 0 || suggestedPrice == 0 {
		return "" // Can't compare unknown prices
	}

	priceDiff := suggestedPrice - originalPrice

	// Same price (within 0.01 tolerance)
	if math.Abs(priceDiff) < 0.01 {
		return "same price"
	}

	// Cheaper
	if priceDiff < 0 {
		absDiff := math.Abs(priceDiff)
		if absDiff < 1.0 {
			// Small difference: show exact cents
			return fmt.Sprintf("€%.2f cheaper", absDiff)
		} else {
			// Larger difference: show percentage
			percentDiff := (absDiff / originalPrice) * 100
			return fmt.Sprintf("%.0f%% cheaper (€%.2f)", percentDiff, absDiff)
		}
	}

	// More expensive
	if priceDiff > 1.0 {
		percentDiff := (priceDiff / originalPrice) * 100
		return fmt.Sprintf("%.0f%% more expensive (€%.2f)", percentDiff, priceDiff)
	} else {
		return fmt.Sprintf("€%.2f more expensive", priceDiff)
	}
}

// GenerateBulkExplanation creates explanation for bulk accept scenario
func GenerateBulkExplanation(totalItems, replacedItems, keptItems, removedItems int) string {
	if replacedItems == totalItems {
		return fmt.Sprintf("All %d items replaced with suggested alternatives", totalItems)
	}

	parts := []string{}
	if replacedItems > 0 {
		parts = append(parts, fmt.Sprintf("%d replaced", replacedItems))
	}
	if keptItems > 0 {
		parts = append(parts, fmt.Sprintf("%d kept for manual search", keptItems))
	}
	if removedItems > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", removedItems))
	}

	explanation := ""
	for i, part := range parts {
		if i == 0 {
			explanation = part
		} else if i == len(parts)-1 {
			explanation += " and " + part
		} else {
			explanation += ", " + part
		}
	}

	return explanation
}
