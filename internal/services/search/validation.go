package search

import (
	"strings"
	"unicode/utf8"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

// ValidateSearchRequest validates a search request
func ValidateSearchRequest(req *SearchRequest) error {
	if req == nil {
		return apperrors.Validation("search request cannot be nil")
	}

	// Validate query
	if err := validateQuery(req.Query); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid query")
	}

	// Validate price range
	if err := validatePriceRange(req.MinPrice, req.MaxPrice); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid price range")
	}

	// Validate pagination
	if err := validatePagination(req.Limit, req.Offset); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid pagination")
	}

	// Validate store IDs
	if err := validateStoreIDs(req.StoreIDs); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid store IDs")
	}

	// Validate category
	if err := validateCategory(req.Category); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid category")
	}

	return nil
}

// ValidateSuggestionRequest validates a suggestion request
func ValidateSuggestionRequest(req *SuggestionRequest) error {
	if req == nil {
		return apperrors.Validation("suggestion request cannot be nil")
	}

	if err := validatePartialQuery(req.PartialQuery); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid partial query")
	}

	if req.Limit < 1 || req.Limit > 20 {
		return apperrors.Validation("limit must be between 1 and 20")
	}

	return nil
}

// ValidateSimilarProductsRequest validates a similar products request
func ValidateSimilarProductsRequest(req *SimilarProductsRequest) error {
	if req == nil {
		return apperrors.Validation("similar products request cannot be nil")
	}

	if req.ProductID <= 0 {
		return apperrors.Validation("product ID must be greater than 0")
	}

	if req.Limit < 1 || req.Limit > 50 {
		return apperrors.Validation("limit must be between 1 and 50")
	}

	return nil
}

// ValidateCorrectionRequest validates a correction request
func ValidateCorrectionRequest(req *CorrectionRequest) error {
	if req == nil {
		return apperrors.Validation("correction request cannot be nil")
	}

	if err := validateQuery(req.Query); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeValidation, "invalid query")
	}

	if req.Limit < 1 || req.Limit > 10 {
		return apperrors.Validation("limit must be between 1 and 10")
	}

	return nil
}

func validateQuery(query string) error {
	query = strings.TrimSpace(query)

	if query == "" {
		return apperrors.Validation("query cannot be empty")
	}

	if !utf8.ValidString(query) {
		return apperrors.Validation("query must be valid UTF-8")
	}

	if utf8.RuneCountInString(query) > 255 {
		return apperrors.Validation("query cannot exceed 255 characters")
	}

	if utf8.RuneCountInString(query) < 1 {
		return apperrors.Validation("query must be at least 1 character")
	}

	// Check for potentially malicious patterns
	maliciousPatterns := []string{
		"<script", "javascript:", "vbscript:", "onload=", "onerror=",
		"--", "/*", "*/", "xp_", "sp_", "DROP ", "DELETE ", "INSERT ",
		"UPDATE ", "CREATE ", "ALTER ", "TRUNCATE ", "EXEC", "EXECUTE",
	}

	queryLower := strings.ToLower(query)
	for _, pattern := range maliciousPatterns {
		if strings.Contains(queryLower, strings.ToLower(pattern)) {
			return apperrors.Validation("query contains potentially malicious content")
		}
	}

	return nil
}

func validatePartialQuery(query string) error {
	query = strings.TrimSpace(query)

	if query == "" {
		return apperrors.Validation("partial query cannot be empty")
	}

	if !utf8.ValidString(query) {
		return apperrors.Validation("partial query must be valid UTF-8")
	}

	if utf8.RuneCountInString(query) > 100 {
		return apperrors.Validation("partial query cannot exceed 100 characters")
	}

	return nil
}

func validatePriceRange(minPrice, maxPrice *float64) error {
	if minPrice != nil && *minPrice < 0 {
		return apperrors.Validation("minimum price cannot be negative")
	}

	if maxPrice != nil && *maxPrice < 0 {
		return apperrors.Validation("maximum price cannot be negative")
	}

	if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {
		return apperrors.Validation("minimum price cannot be greater than maximum price")
	}

	if minPrice != nil && *minPrice > 10000 {
		return apperrors.Validation("minimum price cannot exceed 10000")
	}

	if maxPrice != nil && *maxPrice > 10000 {
		return apperrors.Validation("maximum price cannot exceed 10000")
	}

	return nil
}

func validatePagination(limit, offset int) error {
	if limit < 1 {
		return apperrors.Validation("limit must be at least 1")
	}

	if limit > 100 {
		return apperrors.Validation("limit cannot exceed 100")
	}

	if offset < 0 {
		return apperrors.Validation("offset cannot be negative")
	}

	if offset > 10000 {
		return apperrors.Validation("offset cannot exceed 10000")
	}

	return nil
}

func validateStoreIDs(storeIDs []int) error {
	if len(storeIDs) > 50 {
		return apperrors.Validation("cannot specify more than 50 store IDs")
	}

	for _, id := range storeIDs {
		if id <= 0 {
			return apperrors.Validation("store ID must be greater than 0")
		}
	}

	return nil
}

func validateCategory(category string) error {
	if category == "" {
		return nil // Category is optional
	}

	if !utf8.ValidString(category) {
		return apperrors.Validation("category must be valid UTF-8")
	}

	if utf8.RuneCountInString(category) > 100 {
		return apperrors.Validation("category cannot exceed 100 characters")
	}

	return nil
}

// SanitizeQuery removes or replaces potentially problematic characters
func SanitizeQuery(query string) string {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Remove or replace control characters
	query = strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1 // Remove control characters
		}
		return r
	}, query)

	// Normalize multiple whitespace to single space
	query = strings.Join(strings.Fields(query), " ")

	return query
}

// NormalizeSearchQuery normalizes a search query for better matching
func NormalizeSearchQuery(query string) string {
	query = SanitizeQuery(query)
	query = strings.ToLower(query)

	// Replace Lithuanian diacritics for broader matching
	replacements := map[string]string{
		"ą": "a", "č": "c", "ę": "e", "ė": "e", "į": "i",
		"š": "s", "ų": "u", "ū": "u", "ž": "z",
	}

	for old, new := range replacements {
		query = strings.ReplaceAll(query, old, new)
	}

	return query
}
