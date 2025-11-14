package matching

import (
	"testing"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

func TestExactNameMatcher(t *testing.T) {
	matcher := &ExactNameMatcher{weight: 1.0}

	tests := []struct {
		name          string
		product       *models.Product
		master        *models.ProductMaster
		expectedScore float64
		description   string
	}{
		{
			name: "exact_match",
			product: &models.Product{
				NormalizedName: "pienas 1l",
			},
			master: &models.ProductMaster{
				NormalizedName: "pienas 1l",
			},
			expectedScore: 1.0,
			description:   "Exact normalized name match should score 1.0",
		},
		{
			name: "no_match",
			product: &models.Product{
				NormalizedName: "pienas 1l",
			},
			master: &models.ProductMaster{
				NormalizedName: "duona 500g",
			},
			expectedScore: 0.0,
			description:   "Different products should score 0.0",
		},
		{
			name: "alternative_name_match",
			product: &models.Product{
				NormalizedName: "pienas fermentuotas",
			},
			master: &models.ProductMaster{
				NormalizedName:   "fermentuotas pienas",
				AlternativeNames: []string{"pienas fermentuotas", "pienas ruges"},
			},
			expectedScore: 0.95,
			description:   "Alternative name match should score 0.95",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := matcher.Score(tt.product, tt.master)
			if score != tt.expectedScore {
				t.Errorf("%s: expected score %.2f, got %.2f",
					tt.description, tt.expectedScore, score)
			}
		})
	}
}

func TestFuzzyNameMatcher(t *testing.T) {
	matcher := &FuzzyNameMatcher{weight: 0.8}

	tests := []struct {
		name        string
		product     *models.Product
		master      *models.ProductMaster
		minScore    float64
		maxScore    float64
		description string
	}{
		{
			name: "exact_match",
			product: &models.Product{
				NormalizedName: "pienas 1l",
			},
			master: &models.ProductMaster{
				NormalizedName: "pienas 1l",
			},
			minScore:    0.9,
			maxScore:    1.0,
			description: "Exact match should have very high score",
		},
		{
			name: "low_similarity_below_threshold",
			product: &models.Product{
				NormalizedName: "pienas fermentuotas 1l",
			},
			master: &models.ProductMaster{
				NormalizedName: "fermentuotas pienas 1 litras",
			},
			minScore:    0.0,
			maxScore:    0.7,
			description: "Products below 0.7 similarity threshold should return 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := matcher.Score(tt.product, tt.master)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("%s: expected score between %.2f and %.2f, got %.2f",
					tt.description, tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestTrigramSimilarity(t *testing.T) {
	tests := []struct {
		s1          string
		s2          string
		minScore    float64
		maxScore    float64
		description string
	}{
		{
			s1:          "pienas",
			s2:          "pienas",
			minScore:    1.0,
			maxScore:    1.0,
			description: "Identical strings should have similarity 1.0",
		},
		{
			s1:          "pienas",
			s2:          "duona",
			minScore:    0.0,
			maxScore:    0.1,
			description: "Completely different strings should have low similarity",
		},
		{
			s1:          "pienas 1l",
			s2:          "pienas 1 litras",
			minScore:    0.3,
			maxScore:    0.5,
			description: "Similar strings should have medium similarity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.s1+"_vs_"+tt.s2, func(t *testing.T) {
			score := calculateTrigramSimilarity(tt.s1, tt.s2)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("%s: expected similarity between %.2f and %.2f, got %.2f",
					tt.description, tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestBrandCategoryMatcher(t *testing.T) {
	matcher := &BrandCategoryMatcher{weight: 0.6}

	brand1 := "Rokiskio"
	brand2 := "Pieno Zvaigzdes"
	category := "Pieno produktai"

	tests := []struct {
		name        string
		product     *models.Product
		master      *models.ProductMaster
		minScore    float64
		description string
	}{
		{
			name: "same_brand_and_category",
			product: &models.Product{
				NormalizedName: "pienas 1l",
				Brand:          &brand1,
				Category:       &category,
			},
			master: &models.ProductMaster{
				NormalizedName: "pienas 1 litras",
				Brand:          &brand1,
				Category:       &category,
			},
			minScore:    0.8,
			description: "Same brand and category should score high",
		},
		{
			name: "different_brands",
			product: &models.Product{
				NormalizedName: "pienas 1l",
				Brand:          &brand1,
			},
			master: &models.ProductMaster{
				NormalizedName: "pienas 1l",
				Brand:          &brand2,
			},
			minScore:    0.0,
			description: "Different brands should not match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := matcher.Score(tt.product, tt.master)
			if score < tt.minScore {
				t.Errorf("%s: expected score >= %.2f, got %.2f",
					tt.description, tt.minScore, score)
			}
		})
	}
}
