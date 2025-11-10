package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NormalizeProductText normalizes Lithuanian text for search
func NormalizeProductText(text string) string {
	text = strings.ToLower(text)
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return text
}

// GenerateSearchVector generates a search vector for full-text search
func GenerateSearchVector(normalizedName string) string {
	words := strings.Fields(normalizedName)
	return strings.Join(words, " & ")
}

// ValidateProduct validates a product model
// Note: price can be 0 for category/brand_line promotions that only have discount_percent
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
	// Allow price = 0 for percent-only promotions (category/brand_line)
	// Only reject negative prices
	if price < 0 {
		return fmt.Errorf("invalid price: %f", price)
	}
	if price > 9999.99 {
		return fmt.Errorf("price too high: %f", price)
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
