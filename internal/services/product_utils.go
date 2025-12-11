package services

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/kainuguru/kainuguru-api/pkg/errors"
)

// NormalizeProductText normalizes Lithuanian text for search
func NormalizeProductText(text string) string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return text
}

// sanitizeForTsvector removes special characters that cause tsvector parsing errors
// These characters are special in PostgreSQL tsvector syntax: : & | ! ( ) * '
func sanitizeForTsvector(text string) string {
	// Remove tsvector special characters
	replacer := strings.NewReplacer(
		":", "", // word position indicator
		"&", "", // AND operator
		"|", "", // OR operator
		"!", "", // NOT operator (negation)
		"(", "", // grouping
		")", "", // grouping
		"*", "", // prefix search
		"'", "", // phrase quoting
	)
	return replacer.Replace(text)
}

// GenerateSearchVector generates a search vector for full-text search
func GenerateSearchVector(normalizedName string) string {
	// Sanitize to remove tsvector special characters
	sanitized := sanitizeForTsvector(normalizedName)
	words := strings.Fields(sanitized)
	// Filter out empty words after sanitization
	var validWords []string
	for _, w := range words {
		w = strings.TrimSpace(w)
		if w != "" {
			validWords = append(validWords, w)
		}
	}
	if len(validWords) == 0 {
		return ""
	}
	return strings.Join(validWords, " & ")
}

// ValidateProduct validates a product model
// Note: price can be 0 for category/brand_line promotions that only have discount_percent
func ValidateProduct(name string, price float64) error {
	if name == "" {
		return errors.Validation("product name is required")
	}
	if len(name) < 3 {
		return errors.ValidationF("product name too short: %s", name)
	}
	if len(name) > 150 {
		return errors.ValidationF("product name too long: %s", name)
	}
	// Allow price = 0 for percent-only promotions (category/brand_line)
	// Only reject negative prices
	if price < 0 {
		return errors.ValidationF("invalid price: %f", price)
	}
	if price > 9999.99 {
		return errors.ValidationF("price too high: %f", price)
	}
	return nil
}

// CalculateDiscount calculates discount percentage
func CalculateDiscount(original, current float64) float64 {
	if original <= 0 {
		return 0
	}
	return ((original - current) / original) * 100
}

// StandardizeUnit standardizes unit types
func StandardizeUnit(unit string) string {
	unit = strings.ToLower(strings.TrimSpace(unit))

	unitMap := map[string]string{
		"kilogramas": "kg", "kg.": "kg",
		"gramas": "g", "gr": "g", "g.": "g",
		"litras": "l", "ltr": "l", "l.": "l",
		"mililitras": "ml", "ml.": "ml",
		"vienetų": "vnt.", "vienetas": "vnt.", "vnt": "vnt.",
		"pakuotė": "pak.", "pak": "pak.",
		"dėžutė": "dėž.", "dėž": "dėž.",
	}

	if standard, ok := unitMap[unit]; ok {
		return standard
	}
	return unit
}

// ParsePrice parses a price string to float64
func ParsePrice(priceStr string) (float64, error) {
	if priceStr == "" {
		return 0, errors.Validation("empty price string")
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
		return 0, errors.ValidationF("invalid price format: %s", priceStr)
	}

	if price < 0 {
		return 0, errors.ValidationF("negative price: %f", price)
	}

	return price, nil
}
