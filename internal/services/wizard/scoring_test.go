package wizard

import (
	"testing"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// TestScoreSuggestion_Determinism verifies that ScoreSuggestion returns identical scores for identical inputs
func TestScoreSuggestion_Determinism(t *testing.T) {
	tests := []struct {
		name                 string
		suggestion           *models.Suggestion
		originalProduct      *models.Product
		userStorePreferences map[int]bool
		weights              ScoringWeights
		expectedScore        float64
		description          string
	}{
		{
			name: "perfect_brand_match_preferred_store",
			suggestion: &models.Suggestion{
				Brand:     strPtr("Coca-Cola"),
				StoreID:   1,
				Price:     2.50,
				SizeValue: floatPtr(1.5),
				SizeUnit:  strPtr("L"),
			},
			originalProduct: &models.Product{
				Brand:        strPtr("Coca-Cola"),
				CurrentPrice: 3.00,
				UnitType:     strPtr("L"),
			},
			userStorePreferences: map[int]bool{1: true},
			weights:              DefaultScoringWeights(),
			expectedScore:        6.4, // brand(3.0) + store(2.0) + size(0.8) + price(~0.6) = ~6.4
			description:          "Perfect brand match in preferred store with cheaper price",
		},
		{
			name: "brand_match_non_preferred_store",
			suggestion: &models.Suggestion{
				Brand:     strPtr("Coca-Cola"),
				StoreID:   2,
				Price:     2.50,
				SizeValue: floatPtr(1.5),
				SizeUnit:  strPtr("L"),
			},
			originalProduct: &models.Product{
				Brand:        strPtr("Coca-Cola"),
				CurrentPrice: 3.00,
				UnitType:     strPtr("L"),
			},
			userStorePreferences: map[int]bool{1: true},
			weights:              DefaultScoringWeights(),
			expectedScore:        5.4, // brand(3.0) + store(1.0) + size(0.8) + price(~0.6) = ~5.4
			description:          "Brand match but not in preferred store",
		},
		{
			name: "no_brand_match_preferred_store",
			suggestion: &models.Suggestion{
				Brand:     strPtr("Pepsi"),
				StoreID:   1,
				Price:     2.50,
				SizeValue: floatPtr(1.5),
				SizeUnit:  strPtr("L"),
			},
			originalProduct: &models.Product{
				Brand:        strPtr("Coca-Cola"),
				CurrentPrice: 3.00,
				UnitType:     strPtr("L"),
			},
			userStorePreferences: map[int]bool{1: true},
			weights:              DefaultScoringWeights(),
			expectedScore:        3.4, // brand(0.0) + store(2.0) + size(0.8) + price(~0.6) = ~3.4
			description:          "Different brand in preferred store",
		},
		{
			name: "more_expensive_product",
			suggestion: &models.Suggestion{
				Brand:     strPtr("Coca-Cola"),
				StoreID:   1,
				Price:     3.50,
				SizeValue: floatPtr(1.5),
				SizeUnit:  strPtr("L"),
			},
			originalProduct: &models.Product{
				Brand:        strPtr("Coca-Cola"),
				CurrentPrice: 3.00,
				UnitType:     strPtr("L"),
			},
			userStorePreferences: map[int]bool{1: true},
			weights:              DefaultScoringWeights(),
			expectedScore:        5.8,
			description:          "Same brand but more expensive",
		},
		{
			name: "nil_brand_handling",
			suggestion: &models.Suggestion{
				Brand:     nil,
				StoreID:   1,
				Price:     2.50,
				SizeValue: floatPtr(1.5),
				SizeUnit:  strPtr("L"),
			},
			originalProduct: &models.Product{
				Brand:        strPtr("Coca-Cola"),
				CurrentPrice: 3.00,
				UnitType:     strPtr("L"),
			},
			userStorePreferences: map[int]bool{1: true},
			weights:              DefaultScoringWeights(),
			expectedScore:        3.4, // brand(0.0) + store(2.0) + size(0.8) + price(~0.6) = ~3.4
			description:          "Handles nil brand gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score1 := ScoreSuggestion(tt.suggestion, tt.originalProduct, tt.userStorePreferences, tt.weights)
			score2 := ScoreSuggestion(tt.suggestion, tt.originalProduct, tt.userStorePreferences, tt.weights)
			score3 := ScoreSuggestion(tt.suggestion, tt.originalProduct, tt.userStorePreferences, tt.weights)

			if score1 != score2 || score2 != score3 {
				t.Errorf("ScoreSuggestion is not deterministic: got scores %v, %v, %v", score1, score2, score3)
			}

			maxPossibleScore := tt.weights.Brand + tt.weights.Store + tt.weights.Size + tt.weights.Price
			if score1 < 0.0 || score1 > maxPossibleScore {
				t.Errorf("Score out of bounds: got %v, want 0.0 to %v", score1, maxPossibleScore)
			}

			scoreDiff := abs(score1 - tt.expectedScore)
			if scoreDiff > 0.5 {
				t.Errorf("%s: expected score ~%v, got %v (diff: %v)", tt.description, tt.expectedScore, score1, scoreDiff)
			}

			t.Logf("✓ %s: score=%v (expected ~%v, diff=%v)", tt.description, score1, tt.expectedScore, scoreDiff)
		})
	}
}

// TestRankSuggestions_Determinism verifies that RankSuggestions produces consistent ordering
func TestRankSuggestions_Determinism(t *testing.T) {
	originalProduct := &models.Product{
		Brand:        strPtr("Coca-Cola"),
		CurrentPrice: 3.00,
		UnitType:     strPtr("L"),
	}

	suggestions := []*models.Suggestion{
		{
			FlyerProductID: 1,
			Brand:          strPtr("Coca-Cola"),
			StoreID:        1,
			Price:          2.50,
		},
		{
			FlyerProductID: 2,
			Brand:          strPtr("Pepsi"),
			StoreID:        1,
			Price:          2.30,
		},
		{
			FlyerProductID: 3,
			Brand:          strPtr("Coca-Cola"),
			StoreID:        2,
			Price:          2.60,
		},
	}

	userStorePreferences := map[int]bool{1: true}
	weights := DefaultScoringWeights()

	result1 := RankSuggestions(cloneSuggestions(suggestions), originalProduct, userStorePreferences, weights)
	result2 := RankSuggestions(cloneSuggestions(suggestions), originalProduct, userStorePreferences, weights)
	result3 := RankSuggestions(cloneSuggestions(suggestions), originalProduct, userStorePreferences, weights)

	if !sameSuggestionOrder(result1, result2) || !sameSuggestionOrder(result2, result3) {
		t.Errorf("RankSuggestions is not deterministic")
		t.Logf("Run 1: %v", suggestionIDs(result1))
		t.Logf("Run 2: %v", suggestionIDs(result2))
		t.Logf("Run 3: %v", suggestionIDs(result3))
	}

	for i := 0; i < len(result1)-1; i++ {
		current := result1[i]
		next := result1[i+1]

		if current.ScoreBreakdown.TotalScore < next.ScoreBreakdown.TotalScore {
			t.Errorf("Ordering violation: suggestion %d (score=%v) should come after %d (score=%v)",
				current.FlyerProductID, current.ScoreBreakdown.TotalScore,
				next.FlyerProductID, next.ScoreBreakdown.TotalScore)
		}

		if current.ScoreBreakdown.TotalScore == next.ScoreBreakdown.TotalScore {
			if current.PriceDifference > next.PriceDifference {
				t.Errorf("Tie-breaker violation: suggestion %d (price diff=%v) should come after %d (price diff=%v)",
					current.FlyerProductID, current.PriceDifference,
					next.FlyerProductID, next.PriceDifference)
			}

			if current.PriceDifference == next.PriceDifference {
				if current.FlyerProductID > next.FlyerProductID {
					t.Errorf("ID tie-breaker violation: suggestion %d should come after %d",
						current.FlyerProductID, next.FlyerProductID)
				}
			}
		}
	}

	t.Logf("✓ RankSuggestions is deterministic, order: %v", suggestionIDs(result1))
}

// TestScoreSuggestion_WeightContribution verifies individual weight contributions
func TestScoreSuggestion_WeightContribution(t *testing.T) {
	tests := []struct {
		name            string
		modifyWeights   func(*ScoringWeights)
		expectedChange  string
		verifyCondition func(score float64) bool
	}{
		{
			name: "zero_brand_weight",
			modifyWeights: func(w *ScoringWeights) {
				w.Brand = 0.0
			},
			expectedChange:  "Brand match should not contribute to score",
			verifyCondition: func(score float64) bool { return score < 5.0 },
		},
		{
			name: "double_brand_weight",
			modifyWeights: func(w *ScoringWeights) {
				w.Brand = 6.0
			},
			expectedChange:  "Brand weight doubled should increase brand match score",
			verifyCondition: func(score float64) bool { return score > 8.0 },
		},
		{
			name: "zero_store_weight",
			modifyWeights: func(w *ScoringWeights) {
				w.Store = 0.0
			},
			expectedChange:  "Store preference should not contribute",
			verifyCondition: func(score float64) bool { return score < 6.0 },
		},
	}

	baseSuggestion := &models.Suggestion{
		Brand:   strPtr("Coca-Cola"),
		StoreID: 1,
		Price:   2.50,
	}
	baseProduct := &models.Product{
		Brand:        strPtr("Coca-Cola"),
		CurrentPrice: 3.00,
	}
	basePreferences := map[int]bool{1: true}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights := DefaultScoringWeights()
			tt.modifyWeights(&weights)

			score := ScoreSuggestion(baseSuggestion, baseProduct, basePreferences, weights)

			if !tt.verifyCondition(score) {
				t.Errorf("%s: score=%v did not meet expected condition", tt.expectedChange, score)
			} else {
				t.Logf("✓ %s: score=%v", tt.expectedChange, score)
			}
		})
	}
}

// Helper functions

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func cloneSuggestions(suggestions []*models.Suggestion) []*models.Suggestion {
	cloned := make([]*models.Suggestion, len(suggestions))
	for i, s := range suggestions {
		clone := *s
		cloned[i] = &clone
	}
	return cloned
}

func sameSuggestionOrder(a, b []*models.Suggestion) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].FlyerProductID != b[i].FlyerProductID {
			return false
		}
	}
	return true
}

func suggestionIDs(suggestions []*models.Suggestion) []int64 {
	ids := make([]int64, len(suggestions))
	for i, s := range suggestions {
		ids[i] = s.FlyerProductID
	}
	return ids
}
