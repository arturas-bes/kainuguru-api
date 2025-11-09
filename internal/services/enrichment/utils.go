package enrichment

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/kainuguru/kainuguru-api/internal/services/ai"
	"golang.org/x/text/unicode/norm"
)

// normalizeText normalizes Lithuanian text for search
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Remove extra whitespace
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	
	return text
}

// normalizeLithuanian normalizes Lithuanian text including diacritics
func normalizeLithuanian(text string) string {
	// First do basic normalization
	text = normalizeText(text)
	
	// Normalize Unicode (NFD = decomposed form)
	text = norm.NFD.String(text)
	
	// Remove combining marks but preserve Lithuanian letters
	var result strings.Builder
	for _, r := range text {
		if !unicode.Is(unicode.Mn, r) {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// parsePrice parses a price string to float64
func parsePrice(priceStr string) (float64, error) {
	if priceStr == "" {
		return 0, fmt.Errorf("empty price string")
	}
	
	// Remove currency symbols and whitespace
	priceStr = strings.TrimSpace(priceStr)
	priceStr = strings.ReplaceAll(priceStr, "€", "")
	priceStr = strings.ReplaceAll(priceStr, "EUR", "")
	priceStr = strings.TrimSpace(priceStr)
	
	// Replace comma with dot for decimal separator
	priceStr = strings.ReplaceAll(priceStr, ",", ".")
	
	// Parse as float
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid price format: %s", priceStr)
	}
	
	if price < 0 {
		return 0, fmt.Errorf("negative price: %f", price)
	}
	
	return price, nil
}

// calculateDiscount calculates discount percentage
func CalculateDiscount(original, current float64) float64 {
	if original <= 0 {
		return 0
	}
	return ((original - current) / original) * 100
}

// generateSearchVector generates a search vector for full-text search
func GenerateSearchVector(normalizedName string) string {
	// Simple implementation - can be enhanced with PostgreSQL tsvector
	words := strings.Fields(normalizedName)
	return strings.Join(words, " & ")
}

// validateProduct validates a product model
func ValidateProduct(name string, price float64) error {
	if name == "" {
		return fmt.Errorf("product name is required")
	}
	
	if len(name) < 3 {
		return fmt.Errorf("product name too short: %s", name)
	}
	
	if len(name) > 150 {
		return fmt.Errorf("product name too long: %s", name)
	}
	
	if price <= 0 {
		return fmt.Errorf("invalid price: %f", price)
	}
	
	if price > 9999.99 {
		return fmt.Errorf("price too high: %f", price)
	}
	
	return nil
}

// isValidUnitType checks if unit type is valid
func isValidUnitType(unit string) bool {
	validUnits := map[string]bool{
		"kg":   true,
		"g":    true,
		"l":    true,
		"ml":   true,
		"vnt":  true,
		"vnt.": true,
		"pak":  true,
		"pak.": true,
		"dėž":  true,
		"dėž.": true,
	}
	
	return validUnits[strings.ToLower(unit)]
}

// standardizeUnit standardizes unit types
func StandardizeUnit(unit string) string {
	unit = strings.ToLower(strings.TrimSpace(unit))
	
	// Mapping of various unit formats to standard ones
	unitMap := map[string]string{
		"kilogramas": "kg",
		"kg.":        "kg",
		"gramas":     "g",
		"gr":         "g",
		"g.":         "g",
		"litras":     "l",
		"ltr":        "l",
		"l.":         "l",
		"mililitras": "ml",
		"ml.":        "ml",
		"vienetų":    "vnt.",
		"vienetas":   "vnt.",
		"vnt":        "vnt.",
		"pakuotė":    "pak.",
		"pak":        "pak.",
		"dėžutė":     "dėž.",
		"dėž":        "dėž.",
	}
	
	if standard, ok := unitMap[unit]; ok {
		return standard
	}
	
	return unit
}

// extractBrand extracts brand from product name
func extractBrand(name string) *string {
	// Simple heuristic - uppercase words are often brands
	words := strings.Fields(name)
	for _, word := range words {
		if len(word) > 2 && strings.ToUpper(word) == word {
			return &word
		}
	}
	return nil
}

// categorizeProduct attempts to categorize a product based on its name
func categorizeProduct(name string) *string {
	name = strings.ToLower(name)
	
	categories := map[string][]string{
		"mėsa ir žuvis": {
			"kiauliena", "jautiena", "vištiena", "kalakutiena",
			"lašiša", "menkė", "silkė", "dešra", "kumpis",
		},
		"pieno produktai": {
			"pienas", "kefyras", "jogurtas", "sūris", "sviestas",
			"grietinė", "varškė",
		},
		"duona ir konditerija": {
			"duona", "batonas", "pyragas", "tortas", "sausainiai",
		},
		"vaisiai ir daržovės": {
			"obuoliai", "bananai", "pomidorai", "agurkai", "bulvės",
		},
		"gėrimai": {
			"vanduo", "sultys", "gėrimas", "kava", "arbata",
		},
	}
	
	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(name, keyword) {
				return &category
			}
		}
	}
	
	return nil
}

// extractProductTags extracts relevant tags from product information
func extractProductTags(extracted ai.ExtractedProduct) []string {
	tags := []string{}
	
	// Add category as tag
	if extracted.Category != "" {
		tags = append(tags, strings.ToLower(extracted.Category))
	}
	
	// Add brand as tag if present
	if extracted.Brand != "" {
		tags = append(tags, strings.ToLower(extracted.Brand))
	}
	
	// Add discount-related tags
	if extracted.Discount != "" {
		tags = append(tags, "nuolaida")
		if strings.Contains(strings.ToLower(extracted.Discount), "akcija") {
			tags = append(tags, "akcija")
		}
	}
	
	// Add unit type tags
	if extracted.Unit != "" {
		unit := strings.ToLower(extracted.Unit)
		if strings.Contains(unit, "kg") || strings.Contains(unit, "g") {
			tags = append(tags, "svoris")
		} else if strings.Contains(unit, "l") || strings.Contains(unit, "ml") {
			tags = append(tags, "tūris")
		}
	}
	
	// Add tags based on product name keywords
	name := strings.ToLower(extracted.Name)
	
	// Organic/Bio products
	if strings.Contains(name, "ekologišk") || strings.Contains(name, "bio") {
		tags = append(tags, "ekologiškas")
	}
	
	// Fresh products
	if strings.Contains(name, "šviež") {
		tags = append(tags, "šviežias")
	}
	
	// Frozen products
	if strings.Contains(name, "šaldyt") {
		tags = append(tags, "šaldytas")
	}
	
	// Light/Diet products
	if strings.Contains(name, "lengv") || strings.Contains(name, "diet") {
		tags = append(tags, "lengvas")
	}
	
	// New products
	if strings.Contains(name, "nauj") {
		tags = append(tags, "naujiena")
	}
	
	// Remove duplicates
	tagMap := make(map[string]bool)
	uniqueTags := []string{}
	for _, tag := range tags {
		if !tagMap[tag] {
			tagMap[tag] = true
			uniqueTags = append(uniqueTags, tag)
		}
	}
	
	return uniqueTags
}

// normalizeProductNameForMaster removes brand and extracts generic product name
func normalizeProductNameForMaster(name string, brand *string) string {
	normalized := name
	
	// Remove brand from name if present
	if brand != nil && *brand != "" {
		brandUpper := strings.ToUpper(*brand)
		normalized = strings.ReplaceAll(normalized, brandUpper, "")
		normalized = strings.ReplaceAll(normalized, *brand, "")
	}
	
	// Remove common brand indicators
	normalized = regexp.MustCompile(`\b[A-ZĄČĘĖĮŠŲŪŽ]{2,}\b`).ReplaceAllString(normalized, "")
	
	// Clean up extra spaces
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)
	
	// Capitalize first letter
	if len(normalized) > 0 {
		runes := []rune(normalized)
		runes[0] = unicode.ToUpper(runes[0])
		normalized = string(runes)
	}
	
	return normalized
}
