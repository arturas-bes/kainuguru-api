package enrichment

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

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
