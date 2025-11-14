package matching

import (
	"context"
	"math"
	"strings"

	"github.com/kainuguru/kainuguru-api/internal/models"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/kainuguru/kainuguru-api/pkg/normalize"
	"github.com/uptrace/bun"
)

type MatchResult struct {
	Master     *models.ProductMaster
	Score      float64
	Method     string
	Confidence float64
}

type MatchStrategy interface {
	Score(product *models.Product, master *models.ProductMaster) float64
	Weight() float64
	Name() string
}

type CompositeMatcher struct {
	db         *bun.DB
	normalizer *normalize.LithuanianNormalizer
	strategies []MatchStrategy
}

func NewCompositeMatcher(db *bun.DB) *CompositeMatcher {
	matcher := &CompositeMatcher{
		db:         db,
		normalizer: normalize.NewLithuanianNormalizer(),
	}

	matcher.strategies = []MatchStrategy{
		&ExactNameMatcher{weight: 1.0},
		&FuzzyNameMatcher{weight: 0.8},
		&BrandCategoryMatcher{weight: 0.6},
		&BarcodeMatcher{weight: 1.0},
	}

	return matcher
}

func (m *CompositeMatcher) FindBestMatches(ctx context.Context, product *models.Product, limit int) ([]*MatchResult, error) {
	var candidates []*models.ProductMaster

	query := m.db.NewSelect().
		Model(&candidates).
		Where("pm.status = ?", models.ProductMasterStatusActive).
		Where("pm.confidence_score >= ?", 0.3)

	if product.Brand != nil && *product.Brand != "" {
		query = query.Where("(pm.brand = ? OR pm.brand IS NULL)", *product.Brand)
	}

	if product.Category != nil && *product.Category != "" {
		query = query.Where("(pm.category = ? OR pm.category IS NULL)", *product.Category)
	}

	query = query.Limit(100)

	err := query.Scan(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to fetch master candidates")
	}

	var results []*MatchResult
	for _, master := range candidates {
		score := m.calculateCompositeScore(product, master)
		if score >= 0.5 {
			results = append(results, &MatchResult{
				Master:     master,
				Score:      score,
				Method:     "composite",
				Confidence: score,
			})
		}
	}

	results = sortByScore(results)

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func (m *CompositeMatcher) calculateCompositeScore(product *models.Product, master *models.ProductMaster) float64 {
	var totalScore float64
	var totalWeight float64

	for _, strategy := range m.strategies {
		score := strategy.Score(product, master)
		weight := strategy.Weight()
		totalScore += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0
	}

	return totalScore / totalWeight
}

type ExactNameMatcher struct {
	weight float64
}

func (m *ExactNameMatcher) Score(product *models.Product, master *models.ProductMaster) float64 {
	if product.NormalizedName == master.NormalizedName {
		return 1.0
	}

	for _, altName := range master.AlternativeNames {
		if strings.EqualFold(product.NormalizedName, altName) {
			return 0.95
		}
	}

	return 0.0
}

func (m *ExactNameMatcher) Weight() float64 {
	return m.weight
}

func (m *ExactNameMatcher) Name() string {
	return "exact_name"
}

type FuzzyNameMatcher struct {
	weight float64
}

func (m *FuzzyNameMatcher) Score(product *models.Product, master *models.ProductMaster) float64 {
	similarity := calculateTrigramSimilarity(product.NormalizedName, master.NormalizedName)

	for _, altName := range master.AlternativeNames {
		altSim := calculateTrigramSimilarity(product.NormalizedName, altName)
		if altSim > similarity {
			similarity = altSim
		}
	}

	if similarity < 0.7 {
		return 0.0
	}

	return similarity
}

func (m *FuzzyNameMatcher) Weight() float64 {
	return m.weight
}

func (m *FuzzyNameMatcher) Name() string {
	return "fuzzy_name"
}

type BrandCategoryMatcher struct {
	weight float64
}

func (m *BrandCategoryMatcher) Score(product *models.Product, master *models.ProductMaster) float64 {
	score := 0.0

	if product.Brand != nil && master.Brand != nil {
		if strings.EqualFold(*product.Brand, *master.Brand) {
			score += 0.5
		}
	}

	if product.Category != nil && master.Category != nil {
		if strings.EqualFold(*product.Category, *master.Category) {
			score += 0.3
		}
	}

	productWords := strings.Fields(product.NormalizedName)
	masterWords := strings.Fields(master.NormalizedName)
	commonWords := countCommonWords(productWords, masterWords)
	if len(productWords) > 0 {
		wordScore := float64(commonWords) / float64(max(len(productWords), len(masterWords)))
		score += wordScore * 0.2
	}

	return score
}

func (m *BrandCategoryMatcher) Weight() float64 {
	return m.weight
}

func (m *BrandCategoryMatcher) Name() string {
	return "brand_category"
}

type BarcodeMatcher struct {
	weight float64
}

func (m *BarcodeMatcher) Score(product *models.Product, master *models.ProductMaster) float64 {
	return 0.0
}

func (m *BarcodeMatcher) Weight() float64 {
	return m.weight
}

func (m *BarcodeMatcher) Name() string {
	return "barcode"
}

func calculateTrigramSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	trigrams1 := getTrigrams(s1)
	trigrams2 := getTrigrams(s2)

	if len(trigrams1) == 0 || len(trigrams2) == 0 {
		return 0.0
	}

	intersection := 0
	for trigram := range trigrams1 {
		if trigrams2[trigram] {
			intersection++
		}
	}

	union := len(trigrams1) + len(trigrams2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

func getTrigrams(s string) map[string]bool {
	trigrams := make(map[string]bool)
	s = "  " + strings.ToLower(s) + "  "

	for i := 0; i < len(s)-2; i++ {
		trigram := s[i : i+3]
		trigrams[trigram] = true
	}

	return trigrams
}

func countCommonWords(words1, words2 []string) int {
	wordMap := make(map[string]bool)
	for _, word := range words1 {
		wordMap[strings.ToLower(word)] = true
	}

	count := 0
	for _, word := range words2 {
		if wordMap[strings.ToLower(word)] {
			count++
		}
	}

	return count
}

func sortByScore(results []*MatchResult) []*MatchResult {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	return results
}

func max(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}
