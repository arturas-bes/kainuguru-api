package wizard

import (
	"math"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// ScoreSuggestion calculates a deterministic score for a suggestion
// This is a pure function: same inputs always produce same output
// Scoring weights (per constitution): brand=3.0, store=2.0, size=1.0, price=1.0
func ScoreSuggestion(
	suggestion *models.Suggestion,
	originalProduct *models.Product,
	userStorePreferences map[int]bool,
	weights ScoringWeights,
) float64 {
	var score float64

	// 1. Brand match score (weight: 3.0)
	// Perfect brand match = 1.0, no match = 0.0
	if suggestion.Brand != nil && originalProduct != nil && originalProduct.Brand != nil {
		if *suggestion.Brand == *originalProduct.Brand {
			score += weights.Brand * 1.0
		}
	}

	// 2. Store preference score (weight: 2.0)
	// Preferred store = 1.0, not preferred = 0.5
	if userStorePreferences[suggestion.StoreID] {
		score += weights.Store * 1.0
	} else {
		score += weights.Store * 0.5
	}

	// 3. Size similarity score (weight: 1.0)
	// Calculate size match based on unit and size value
	sizeScore := calculateSizeScore(
		originalProduct,
		suggestion.SizeValue,
		suggestion.SizeUnit,
	)
	score += weights.Size * sizeScore

	// 4. Price score (weight: 1.0)
	// Lower price = higher score, normalized by original price
	priceScore := calculatePriceScore(
		originalProduct,
		suggestion.Price,
	)
	score += weights.Price * priceScore

	// Note: Suggestion.Score is already populated from FuzzySearchProducts
	// We're calculating additional scoring on top of search relevance

	return score
}

// calculateSizeScore computes size similarity between original and suggested product
// Returns 1.0 for exact match, 0.5 for same unit but different size, 0.0 for different unit
func calculateSizeScore(original *models.Product, suggestedSize *float64, suggestedUnit *string) float64 {
	if original == nil {
		return 0.0
	}

	// Extract original size info from UnitSize or Weight/Volume fields
	// For now, use simplified logic - TODO: improve parsing in Phase 3
	if suggestedUnit == nil || suggestedSize == nil {
		return 0.5 // Unknown sizes = neutral score
	}

	// Check if units match (simplified check)
	originalUnitType := original.UnitType
	if originalUnitType == nil {
		return 0.5 // Unknown original unit = neutral
	}

	if *originalUnitType != *suggestedUnit {
		return 0.0 // Different units (kg vs L) = no match
	}

	// Same unit - for now return high score
	// TODO: Add proper size value parsing and comparison in Phase 3
	return 0.8
}

// calculatePriceScore computes price attractiveness
// Lower price than original = higher score (up to 1.0)
// Same price = 0.5
// Higher price = lower score (down to 0.0)
func calculatePriceScore(original *models.Product, suggestedPrice float64) float64 {
	if original == nil {
		return 0.5 // Default neutral score
	}

	originalPrice := original.CurrentPrice

	if originalPrice == 0 || suggestedPrice == 0 {
		return 0.5 // Unknown prices = neutral
	}

	// Calculate price ratio
	if suggestedPrice <= originalPrice {
		// Suggested is cheaper or same - reward this
		// 10% cheaper = 0.7 score, 20% cheaper = 0.8, 50% cheaper = 1.0
		priceDiff := (originalPrice - suggestedPrice) / originalPrice
		return 0.5 + (priceDiff * 0.5) // Scale from 0.5 to 1.0
	} else {
		// Suggested is more expensive - penalize
		// 10% more = 0.4, 20% more = 0.3, 50% more = 0.0
		priceDiff := (suggestedPrice - originalPrice) / originalPrice
		penalty := math.Min(priceDiff*0.5, 0.5) // Cap penalty at 0.5
		return 0.5 - penalty                    // Scale from 0.5 to 0.0
	}
}

// RankSuggestions sorts suggestions by score in descending order
// Returns a new slice with ranked suggestions
// Sort order (per constitution): TotalScore DESC, PriceCompare ASC, ProductID ASC
func RankSuggestions(
	suggestions []*models.Suggestion,
	originalProduct *models.Product,
	userStorePreferences map[int]bool,
	weights ScoringWeights,
) []*models.Suggestion {
	// Calculate scores for all suggestions
	for _, suggestion := range suggestions {
		totalScore := ScoreSuggestion(suggestion, originalProduct, userStorePreferences, weights)

		// Populate ScoreBreakdown with component scores
		suggestion.ScoreBreakdown.TotalScore = totalScore
		suggestion.Score = totalScore

		// Calculate price difference for tie-breaking
		if originalProduct != nil && originalProduct.CurrentPrice > 0 {
			suggestion.PriceDifference = suggestion.Price - originalProduct.CurrentPrice
		}
	}

	// Sort by TotalScore DESC, PriceDifference ASC, FlyerProductID ASC
	// Using bubble sort for determinism (same scores maintain consistent ordering)
	for i := 0; i < len(suggestions); i++ {
		for j := i + 1; j < len(suggestions); j++ {
			shouldSwap := false

			// Primary sort: Score DESC (higher is better)
			if suggestions[j].ScoreBreakdown.TotalScore > suggestions[i].ScoreBreakdown.TotalScore {
				shouldSwap = true
			} else if suggestions[j].ScoreBreakdown.TotalScore == suggestions[i].ScoreBreakdown.TotalScore {
				// Tie-breaker 1: PriceDifference ASC (cheaper is better)
				if suggestions[j].PriceDifference < suggestions[i].PriceDifference {
					shouldSwap = true
				} else if suggestions[j].PriceDifference == suggestions[i].PriceDifference {
					// Tie-breaker 2: FlyerProductID ASC (determinism)
					if suggestions[j].FlyerProductID < suggestions[i].FlyerProductID {
						shouldSwap = true
					}
				}
			}

			if shouldSwap {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	return suggestions
}
