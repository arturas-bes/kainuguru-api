package ai

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidatorConfig holds configuration for the extraction validator
type ValidatorConfig struct {
	MinConfidence            float64  `json:"min_confidence"`
	MaxPriceThreshold        float64  `json:"max_price_threshold"`
	RequiredFields           []string `json:"required_fields"`
	EnablePriceValidation    bool     `json:"enable_price_validation"`
	EnableCategoryValidation bool     `json:"enable_category_validation"`
	StrictMode               bool     `json:"strict_mode"`
}

// DefaultValidatorConfig returns sensible validation defaults
func DefaultValidatorConfig() ValidatorConfig {
	return ValidatorConfig{
		MinConfidence:            0.7,
		MaxPriceThreshold:        1000.0, // €1000 per product seems reasonable max
		RequiredFields:           []string{"name", "price"},
		EnablePriceValidation:    true,
		EnableCategoryValidation: true,
		StrictMode:               false,
	}
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid         bool                   `json:"is_valid"`
	Score           float64                `json:"score"`
	ValidProducts   []ExtractedProduct     `json:"valid_products"`
	InvalidProducts []InvalidProduct       `json:"invalid_products"`
	Issues          []ValidationIssue      `json:"issues"`
	Corrections     []ValidationCorrection `json:"corrections"`
	Statistics      ValidationStatistics   `json:"statistics"`
	ValidatedAt     time.Time              `json:"validated_at"`
}

// InvalidProduct represents a product that failed validation
type InvalidProduct struct {
	Product ExtractedProduct `json:"product"`
	Reasons []string         `json:"reasons"`
}

// ValidationIssue represents a validation issue found
type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"` // critical, warning, info
	Description string `json:"description"`
	Field       string `json:"field,omitempty"`
	Value       string `json:"value,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ValidationCorrection represents an automatic correction made
type ValidationCorrection struct {
	Field          string `json:"field"`
	OriginalValue  string `json:"original_value"`
	CorrectedValue string `json:"corrected_value"`
	Reason         string `json:"reason"`
}

// ValidationStatistics represents validation statistics
type ValidationStatistics struct {
	TotalProducts     int     `json:"total_products"`
	ValidProducts     int     `json:"valid_products"`
	InvalidProducts   int     `json:"invalid_products"`
	CorrectedProducts int     `json:"corrected_products"`
	ValidRate         float64 `json:"valid_rate"`
	AverageConfidence float64 `json:"average_confidence"`
}

// ExtractionValidator validates and corrects extraction results
type ExtractionValidator struct {
	config          ValidatorConfig
	priceRegex      *regexp.Regexp
	unitRegex       *regexp.Regexp
	categoryMap     map[string]string
	lithuanianRegex *regexp.Regexp
}

// NewExtractionValidator creates a new extraction validator
func NewExtractionValidator(config ValidatorConfig) *ExtractionValidator {
	return &ExtractionValidator{
		config:          config,
		priceRegex:      regexp.MustCompile(`^(\d+[,.]\d{2}|\d+)\s*€?$`),
		unitRegex:       regexp.MustCompile(`^(kg|g|l|ml|vnt\.|pak\.|dėž\.|m|cm|mm)\s*\.?$`),
		lithuanianRegex: regexp.MustCompile(`[ąčęėįšųūž]`),
		categoryMap:     createCategoryMap(),
	}
}

// createCategoryMap creates a mapping for category normalization
func createCategoryMap() map[string]string {
	return map[string]string{
		"mesa":       "mėsa ir žuvis",
		"žuvis":      "mėsa ir žuvis",
		"pienas":     "pieno produktai",
		"sūris":      "pieno produktai",
		"jogurtas":   "pieno produktai",
		"duona":      "duona ir konditerija",
		"pyragai":    "duona ir konditerija",
		"vaisiai":    "vaisiai ir daržovės",
		"daržovės":   "vaisiai ir daržovės",
		"gėrimai":    "gėrimai",
		"vanduo":     "gėrimai",
		"sultys":     "gėrimai",
		"šaldyti":    "šaldyti produktai",
		"konservai":  "konservai",
		"kruopos":    "kruopos ir makaronai",
		"makaronai":  "kruopos ir makaronai",
		"saldumynai": "saldumynai",
		"šokoladas":  "saldumynai",
		"higiena":    "higienos prekės",
		"namų":       "namų ūkio prekės",
		"alkoholis":  "alkoholiniai gėrimai",
		"alus":       "alkoholiniai gėrimai",
		"vynas":      "alkoholiniai gėrimai",
	}
}

// ValidateExtraction validates an extraction result
func (v *ExtractionValidator) ValidateExtraction(result *ExtractionResult) *ValidationResult {
	validation := &ValidationResult{
		ValidatedAt: time.Now(),
		Issues:      []ValidationIssue{},
		Corrections: []ValidationCorrection{},
	}

	var validProducts []ExtractedProduct
	var invalidProducts []InvalidProduct

	for _, product := range result.Products {
		validated, issues, corrections := v.validateProduct(product)

		if len(issues) == 0 || (!v.config.StrictMode && v.hasOnlyCriticalIssues(issues)) {
			validProducts = append(validProducts, validated)
		} else {
			invalidProducts = append(invalidProducts, InvalidProduct{
				Product: product,
				Reasons: v.extractReasons(issues),
			})
		}

		// Collect all issues and corrections
		validation.Issues = append(validation.Issues, issues...)
		validation.Corrections = append(validation.Corrections, corrections...)
	}

	validation.ValidProducts = validProducts
	validation.InvalidProducts = invalidProducts
	validation.Statistics = v.calculateStatistics(result.Products, validProducts, invalidProducts)
	validation.Score = v.calculateOverallScore(validation.Statistics, validation.Issues)
	validation.IsValid = validation.Score >= 0.7 // 70% threshold

	return validation
}

// validateProduct validates a single product
func (v *ExtractionValidator) validateProduct(product ExtractedProduct) (ExtractedProduct, []ValidationIssue, []ValidationCorrection) {
	var issues []ValidationIssue
	var corrections []ValidationCorrection
	validated := product

	// Validate required fields
	for _, field := range v.config.RequiredFields {
		if err := v.validateRequiredField(product, field); err != nil {
			issues = append(issues, ValidationIssue{
				Type:        "missing_required_field",
				Severity:    "critical",
				Description: fmt.Sprintf("Required field '%s' is missing or empty", field),
				Field:       field,
			})
		}
	}

	// Validate and correct name
	if nameIssues, nameCorrections := v.validateName(validated.Name); len(nameIssues) > 0 || len(nameCorrections) > 0 {
		issues = append(issues, nameIssues...)
		corrections = append(corrections, nameCorrections...)
		if len(nameCorrections) > 0 {
			validated.Name = nameCorrections[len(nameCorrections)-1].CorrectedValue
		}
	}

	// Validate and correct price
	if v.config.EnablePriceValidation {
		if priceIssues, priceCorrections := v.validatePrice(validated.Price); len(priceIssues) > 0 || len(priceCorrections) > 0 {
			issues = append(issues, priceIssues...)
			corrections = append(corrections, priceCorrections...)
			if len(priceCorrections) > 0 {
				validated.Price = priceCorrections[len(priceCorrections)-1].CorrectedValue
			}
		}
	}

	// Validate and correct unit
	if unitIssues, unitCorrections := v.validateUnit(validated.Unit); len(unitIssues) > 0 || len(unitCorrections) > 0 {
		issues = append(issues, unitIssues...)
		corrections = append(corrections, unitCorrections...)
		if len(unitCorrections) > 0 {
			validated.Unit = unitCorrections[len(unitCorrections)-1].CorrectedValue
		}
	}

	// Validate and correct category
	if v.config.EnableCategoryValidation {
		if categoryIssues, categoryCorrections := v.validateCategory(validated.Category, validated.Name); len(categoryIssues) > 0 || len(categoryCorrections) > 0 {
			issues = append(issues, categoryIssues...)
			corrections = append(corrections, categoryCorrections...)
			if len(categoryCorrections) > 0 {
				validated.Category = categoryCorrections[len(categoryCorrections)-1].CorrectedValue
			}
		}
	}

	// Validate confidence
	if validated.Confidence < v.config.MinConfidence {
		issues = append(issues, ValidationIssue{
			Type:        "low_confidence",
			Severity:    "warning",
			Description: fmt.Sprintf("Product confidence %.2f is below threshold %.2f", validated.Confidence, v.config.MinConfidence),
			Field:       "confidence",
			Value:       fmt.Sprintf("%.2f", validated.Confidence),
		})
	}

	return validated, issues, corrections
}

// validateRequiredField checks if a required field is present
func (v *ExtractionValidator) validateRequiredField(product ExtractedProduct, field string) error {
	switch field {
	case "name":
		if strings.TrimSpace(product.Name) == "" {
			return fmt.Errorf("name is required")
		}
	case "price":
		if strings.TrimSpace(product.Price) == "" {
			return fmt.Errorf("price is required")
		}
	case "unit":
		if strings.TrimSpace(product.Unit) == "" {
			return fmt.Errorf("unit is required")
		}
	case "category":
		if strings.TrimSpace(product.Category) == "" {
			return fmt.Errorf("category is required")
		}
	}
	return nil
}

// validateName validates and corrects product name
func (v *ExtractionValidator) validateName(name string) ([]ValidationIssue, []ValidationCorrection) {
	var issues []ValidationIssue
	var corrections []ValidationCorrection

	if len(name) < 2 {
		issues = append(issues, ValidationIssue{
			Type:        "name_too_short",
			Severity:    "critical",
			Description: "Product name is too short",
			Field:       "name",
			Value:       name,
		})
	}

	if len(name) > 200 {
		issues = append(issues, ValidationIssue{
			Type:        "name_too_long",
			Severity:    "warning",
			Description: "Product name is unusually long",
			Field:       "name",
			Value:       name,
		})
	}

	// Check for Lithuanian characters (good sign for Lithuanian products)
	if !v.lithuanianRegex.MatchString(name) {
		issues = append(issues, ValidationIssue{
			Type:        "no_lithuanian_chars",
			Severity:    "info",
			Description: "Product name does not contain Lithuanian characters",
			Field:       "name",
			Value:       name,
		})
	}

	// Auto-correct common issues
	corrected := name
	if corrected != strings.TrimSpace(corrected) {
		corrected = strings.TrimSpace(corrected)
		corrections = append(corrections, ValidationCorrection{
			Field:          "name",
			OriginalValue:  name,
			CorrectedValue: corrected,
			Reason:         "removed leading/trailing whitespace",
		})
	}

	// Normalize multiple spaces
	normalized := regexp.MustCompile(`\s+`).ReplaceAllString(corrected, " ")
	if normalized != corrected {
		corrections = append(corrections, ValidationCorrection{
			Field:          "name",
			OriginalValue:  corrected,
			CorrectedValue: normalized,
			Reason:         "normalized multiple spaces",
		})
	}

	return issues, corrections
}

// validatePrice validates and corrects product price
func (v *ExtractionValidator) validatePrice(price string) ([]ValidationIssue, []ValidationCorrection) {
	var issues []ValidationIssue
	var corrections []ValidationCorrection

	if price == "" {
		issues = append(issues, ValidationIssue{
			Type:        "empty_price",
			Severity:    "critical",
			Description: "Price is empty",
			Field:       "price",
		})
		return issues, corrections
	}

	// Extract numeric value for validation
	numericPrice := v.extractNumericPrice(price)
	if numericPrice <= 0 {
		issues = append(issues, ValidationIssue{
			Type:        "invalid_price_format",
			Severity:    "critical",
			Description: "Price format is invalid or price is not positive",
			Field:       "price",
			Value:       price,
		})
	} else if numericPrice > v.config.MaxPriceThreshold {
		issues = append(issues, ValidationIssue{
			Type:        "price_too_high",
			Severity:    "warning",
			Description: fmt.Sprintf("Price %.2f exceeds reasonable threshold %.2f", numericPrice, v.config.MaxPriceThreshold),
			Field:       "price",
			Value:       price,
		})
	}

	// Auto-correct price format
	corrected := v.normalizePriceFormat(price)
	if corrected != price {
		corrections = append(corrections, ValidationCorrection{
			Field:          "price",
			OriginalValue:  price,
			CorrectedValue: corrected,
			Reason:         "normalized price format",
		})
	}

	return issues, corrections
}

// validateUnit validates and corrects product unit
func (v *ExtractionValidator) validateUnit(unit string) ([]ValidationIssue, []ValidationCorrection) {
	var issues []ValidationIssue
	var corrections []ValidationCorrection

	if unit == "" {
		issues = append(issues, ValidationIssue{
			Type:        "empty_unit",
			Severity:    "warning",
			Description: "Unit is empty",
			Field:       "unit",
		})
		return issues, corrections
	}

	// Check if unit matches expected format
	if !v.unitRegex.MatchString(strings.ToLower(unit)) {
		issues = append(issues, ValidationIssue{
			Type:        "invalid_unit_format",
			Severity:    "warning",
			Description: "Unit format doesn't match expected pattern",
			Field:       "unit",
			Value:       unit,
			Suggestion:  "Expected formats: kg, g, l, ml, vnt., pak., dėž.",
		})
	}

	// Auto-correct unit format
	corrected := v.normalizeUnit(unit)
	if corrected != unit {
		corrections = append(corrections, ValidationCorrection{
			Field:          "unit",
			OriginalValue:  unit,
			CorrectedValue: corrected,
			Reason:         "normalized unit format",
		})
	}

	return issues, corrections
}

// validateCategory validates and corrects product category
func (v *ExtractionValidator) validateCategory(category, productName string) ([]ValidationIssue, []ValidationCorrection) {
	var issues []ValidationIssue
	var corrections []ValidationCorrection

	if category == "" {
		// Try to infer category from product name
		inferredCategory := v.inferCategoryFromName(productName)
		if inferredCategory != "" {
			corrections = append(corrections, ValidationCorrection{
				Field:          "category",
				OriginalValue:  category,
				CorrectedValue: inferredCategory,
				Reason:         "inferred from product name",
			})
		} else {
			issues = append(issues, ValidationIssue{
				Type:        "empty_category",
				Severity:    "warning",
				Description: "Category is empty and could not be inferred",
				Field:       "category",
			})
		}
		return issues, corrections
	}

	// Check if category is in our known categories
	normalizedCategory := v.normalizeCategoryName(category)
	if normalizedCategory != category {
		corrections = append(corrections, ValidationCorrection{
			Field:          "category",
			OriginalValue:  category,
			CorrectedValue: normalizedCategory,
			Reason:         "normalized to known category",
		})
	}

	return issues, corrections
}

// Helper methods

func (v *ExtractionValidator) extractNumericPrice(price string) float64 {
	// Remove currency symbols and spaces
	cleaned := regexp.MustCompile(`[€\s]`).ReplaceAllString(price, "")

	// Replace comma with dot for parsing
	cleaned = strings.ReplaceAll(cleaned, ",", ".")

	// Parse as float
	if value, err := strconv.ParseFloat(cleaned, 64); err == nil {
		return value
	}

	return 0
}

func (v *ExtractionValidator) normalizePriceFormat(price string) string {
	// Extract numeric part and normalize format
	numericPrice := v.extractNumericPrice(price)
	if numericPrice > 0 {
		return fmt.Sprintf("%.2f €", numericPrice)
	}
	return price
}

func (v *ExtractionValidator) normalizeUnit(unit string) string {
	unit = strings.TrimSpace(strings.ToLower(unit))

	// Common unit normalizations
	unitMap := map[string]string{
		"kilogramas": "kg",
		"kg.":        "kg",
		"gramas":     "g",
		"g.":         "g",
		"litras":     "l",
		"l.":         "l",
		"mililitras": "ml",
		"ml.":        "ml",
		"vienetų":    "vnt.",
		"vienetai":   "vnt.",
		"vienetas":   "vnt.",
		"vnt":        "vnt.",
		"pakuotė":    "pak.",
		"pak":        "pak.",
		"dėžė":       "dėž.",
		"dez":        "dėž.",
	}

	if normalized, exists := unitMap[unit]; exists {
		return normalized
	}

	return unit
}

func (v *ExtractionValidator) normalizeCategoryName(category string) string {
	category = strings.TrimSpace(strings.ToLower(category))

	// Check direct mapping
	if mapped, exists := v.categoryMap[category]; exists {
		return mapped
	}

	// Check partial matches
	for key, value := range v.categoryMap {
		if strings.Contains(category, key) || strings.Contains(key, category) {
			return value
		}
	}

	return category
}

func (v *ExtractionValidator) inferCategoryFromName(productName string) string {
	productName = strings.ToLower(productName)

	for keyword, category := range v.categoryMap {
		if strings.Contains(productName, keyword) {
			return category
		}
	}

	return ""
}

func (v *ExtractionValidator) hasOnlyCriticalIssues(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "critical" {
			return false
		}
	}
	return true
}

func (v *ExtractionValidator) extractReasons(issues []ValidationIssue) []string {
	var reasons []string
	for _, issue := range issues {
		reasons = append(reasons, issue.Description)
	}
	return reasons
}

func (v *ExtractionValidator) calculateStatistics(original, valid []ExtractedProduct, invalid []InvalidProduct) ValidationStatistics {
	stats := ValidationStatistics{
		TotalProducts:   len(original),
		ValidProducts:   len(valid),
		InvalidProducts: len(invalid),
	}

	if stats.TotalProducts > 0 {
		stats.ValidRate = float64(stats.ValidProducts) / float64(stats.TotalProducts)
	}

	// Calculate average confidence for valid products
	var totalConfidence float64
	for _, product := range valid {
		totalConfidence += product.Confidence
	}

	if len(valid) > 0 {
		stats.AverageConfidence = totalConfidence / float64(len(valid))
	}

	return stats
}

func (v *ExtractionValidator) calculateOverallScore(stats ValidationStatistics, issues []ValidationIssue) float64 {
	baseScore := stats.ValidRate

	// Penalize based on issues
	var penalty float64
	for _, issue := range issues {
		switch issue.Severity {
		case "critical":
			penalty += 0.1
		case "warning":
			penalty += 0.05
		case "info":
			penalty += 0.01
		}
	}

	score := baseScore - penalty
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}
