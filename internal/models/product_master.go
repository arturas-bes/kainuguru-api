package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type ProductMaster struct {
	bun.BaseModel `bun:"table:product_masters,alias:pm"`

	ID                  int    `bun:"id,pk,autoincrement" json:"id"`
	CanonicalName       string `bun:"canonical_name,unique,notnull" json:"canonical_name"`
	NormalizedName      string `bun:"normalized_name,notnull" json:"normalized_name"`
	Brand               string `bun:"brand,notnull" json:"brand"`
	Category            string `bun:"category,notnull" json:"category"`
	Subcategory         *string `bun:"subcategory" json:"subcategory,omitempty"`

	// Product specifications
	StandardUnitSize    *string `bun:"standard_unit_size" json:"standard_unit_size,omitempty"`
	StandardUnitType    *string `bun:"standard_unit_type" json:"standard_unit_type,omitempty"`
	StandardPackageSize *string `bun:"standard_package_size" json:"standard_package_size,omitempty"`
	StandardWeight      *string `bun:"standard_weight" json:"standard_weight,omitempty"`
	StandardVolume      *string `bun:"standard_volume" json:"standard_volume,omitempty"`

	// Matching and validation
	MatchingKeywords    json.RawMessage `bun:"matching_keywords,type:jsonb" json:"matching_keywords"`
	AlternativeNames    json.RawMessage `bun:"alternative_names,type:jsonb" json:"alternative_names"`
	ExclusionKeywords   json.RawMessage `bun:"exclusion_keywords,type:jsonb" json:"exclusion_keywords"`

	// Quality and confidence metrics
	ConfidenceScore     float64 `bun:"confidence_score,default:0.0" json:"confidence_score"`
	MatchedProducts     int     `bun:"matched_products,default:0" json:"matched_products"`
	SuccessfulMatches   int     `bun:"successful_matches,default:0" json:"successful_matches"`
	FailedMatches       int     `bun:"failed_matches,default:0" json:"failed_matches"`

	// Status and lifecycle
	Status              string     `bun:"status,default:'active'" json:"status"`
	IsVerified          bool       `bun:"is_verified,default:false" json:"is_verified"`
	LastMatchedAt       *time.Time `bun:"last_matched_at" json:"last_matched_at,omitempty"`
	VerifiedAt          *time.Time `bun:"verified_at" json:"verified_at,omitempty"`
	VerifiedBy          *string    `bun:"verified_by" json:"verified_by,omitempty"`

	// Search optimization
	SearchVector        string `bun:"search_vector" json:"-"`

	// Timestamps
	CreatedAt           time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt           time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Relations
	Products            []*Product `bun:"rel:has-many,join:id=product_master_id" json:"products,omitempty"`
}

// ProductMasterStatus represents possible product master statuses
type ProductMasterStatus string

const (
	ProductMasterStatusActive     ProductMasterStatus = "active"
	ProductMasterStatusInactive   ProductMasterStatus = "inactive"
	ProductMasterStatusPending    ProductMasterStatus = "pending"
	ProductMasterStatusDuplicate  ProductMasterStatus = "duplicate"
	ProductMasterStatusDeprecated ProductMasterStatus = "deprecated"
)

// MatchingKeywords represents keywords used for product matching
type MatchingKeywords struct {
	Primary   []string `json:"primary"`
	Secondary []string `json:"secondary"`
	Brand     []string `json:"brand"`
	Category  []string `json:"category"`
}

// AlternativeNames represents alternative names for the product
type AlternativeNames struct {
	Names     []string `json:"names"`
	Aliases   []string `json:"aliases"`
	Abbreviations []string `json:"abbreviations"`
}

// ExclusionKeywords represents keywords that exclude this product from matching
type ExclusionKeywords struct {
	Exclusions []string `json:"exclusions"`
	Conflicts  []string `json:"conflicts"`
}

// GetMatchingKeywords parses the matching_keywords JSON field
func (pm *ProductMaster) GetMatchingKeywords() (MatchingKeywords, error) {
	var keywords MatchingKeywords
	if pm.MatchingKeywords != nil {
		err := json.Unmarshal(pm.MatchingKeywords, &keywords)
		return keywords, err
	}
	return keywords, nil
}

// SetMatchingKeywords sets the matching_keywords JSON field
func (pm *ProductMaster) SetMatchingKeywords(keywords MatchingKeywords) error {
	data, err := json.Marshal(keywords)
	if err != nil {
		return err
	}
	pm.MatchingKeywords = data
	return nil
}

// GetAlternativeNames parses the alternative_names JSON field
func (pm *ProductMaster) GetAlternativeNames() (AlternativeNames, error) {
	var names AlternativeNames
	if pm.AlternativeNames != nil {
		err := json.Unmarshal(pm.AlternativeNames, &names)
		return names, err
	}
	return names, nil
}

// SetAlternativeNames sets the alternative_names JSON field
func (pm *ProductMaster) SetAlternativeNames(names AlternativeNames) error {
	data, err := json.Marshal(names)
	if err != nil {
		return err
	}
	pm.AlternativeNames = data
	return nil
}

// GetExclusionKeywords parses the exclusion_keywords JSON field
func (pm *ProductMaster) GetExclusionKeywords() (ExclusionKeywords, error) {
	var exclusions ExclusionKeywords
	if pm.ExclusionKeywords != nil {
		err := json.Unmarshal(pm.ExclusionKeywords, &exclusions)
		return exclusions, err
	}
	return exclusions, nil
}

// SetExclusionKeywords sets the exclusion_keywords JSON field
func (pm *ProductMaster) SetExclusionKeywords(exclusions ExclusionKeywords) error {
	data, err := json.Marshal(exclusions)
	if err != nil {
		return err
	}
	pm.ExclusionKeywords = data
	return nil
}

// NormalizeName creates a normalized version of the canonical name
func (pm *ProductMaster) NormalizeName() {
	normalized := strings.ToLower(pm.CanonicalName)

	// Lithuanian character normalization
	normalized = strings.ReplaceAll(normalized, "ą", "a")
	normalized = strings.ReplaceAll(normalized, "č", "c")
	normalized = strings.ReplaceAll(normalized, "ę", "e")
	normalized = strings.ReplaceAll(normalized, "ė", "e")
	normalized = strings.ReplaceAll(normalized, "į", "i")
	normalized = strings.ReplaceAll(normalized, "š", "s")
	normalized = strings.ReplaceAll(normalized, "ų", "u")
	normalized = strings.ReplaceAll(normalized, "ū", "u")
	normalized = strings.ReplaceAll(normalized, "ž", "z")

	// Remove punctuation and normalize spaces
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = strings.ReplaceAll(normalized, ".", " ")
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "/", " ")
	normalized = strings.ReplaceAll(normalized, "(", " ")
	normalized = strings.ReplaceAll(normalized, ")", " ")

	// Normalize multiple spaces
	for strings.Contains(normalized, "  ") {
		normalized = strings.ReplaceAll(normalized, "  ", " ")
	}

	pm.NormalizedName = strings.TrimSpace(normalized)
}

// IsActive checks if the product master is active
func (pm *ProductMaster) IsActive() bool {
	return pm.Status == string(ProductMasterStatusActive)
}

// IsManuallyVerified checks if the product master has been manually verified
func (pm *ProductMaster) IsManuallyVerified() bool {
	return pm.IsVerified && pm.VerifiedAt != nil
}

// GetMatchSuccessRate returns the success rate of product matching
func (pm *ProductMaster) GetMatchSuccessRate() float64 {
	if pm.MatchedProducts == 0 {
		return 0.0
	}
	return float64(pm.SuccessfulMatches) / float64(pm.MatchedProducts) * 100.0
}

// IncrementSuccessfulMatch increments successful match counter
func (pm *ProductMaster) IncrementSuccessfulMatch() {
	pm.SuccessfulMatches++
	pm.MatchedProducts++
	pm.LastMatchedAt = &time.Time{}
	*pm.LastMatchedAt = time.Now()
	pm.UpdateMatchingConfidence()
	pm.UpdatedAt = time.Now()
}

// IncrementFailedMatch increments failed match counter
func (pm *ProductMaster) IncrementFailedMatch() {
	pm.FailedMatches++
	pm.MatchedProducts++
	pm.UpdateMatchingConfidence()
	pm.UpdatedAt = time.Now()
}

// UpdateMatchingConfidence recalculates confidence score based on match history
func (pm *ProductMaster) UpdateMatchingConfidence() {
	if pm.MatchedProducts == 0 {
		pm.ConfidenceScore = 0.0
		return
	}

	successRate := pm.GetMatchSuccessRate()

	// Base confidence on success rate
	baseConfidence := successRate / 100.0

	// Adjust for verification status
	if pm.IsVerified {
		baseConfidence = (baseConfidence + 1.0) / 2.0 // Boost verified products
	}

	// Adjust for match volume (more matches = higher confidence)
	volumeBoost := float64(pm.MatchedProducts) / (float64(pm.MatchedProducts) + 10.0)
	pm.ConfidenceScore = (baseConfidence + volumeBoost) / 2.0

	// Cap at 1.0
	if pm.ConfidenceScore > 1.0 {
		pm.ConfidenceScore = 1.0
	}
}

// VerifyProduct marks the product master as manually verified
func (pm *ProductMaster) VerifyProduct(verifierID string) {
	now := time.Now()
	pm.IsVerified = true
	pm.VerifiedAt = &now
	pm.VerifiedBy = &verifierID
	pm.Status = string(ProductMasterStatusActive)
	pm.UpdateMatchingConfidence()
	pm.UpdatedAt = now
}

// Deactivate marks the product master as inactive
func (pm *ProductMaster) Deactivate() {
	pm.Status = string(ProductMasterStatusInactive)
	pm.UpdatedAt = time.Now()
}

// MarkAsDuplicate marks the product master as a duplicate
func (pm *ProductMaster) MarkAsDuplicate() {
	pm.Status = string(ProductMasterStatusDuplicate)
	pm.UpdatedAt = time.Now()
}

// CanBeMatched checks if the product master can be used for matching
func (pm *ProductMaster) CanBeMatched() bool {
	return pm.IsActive() && pm.ConfidenceScore >= 0.3
}

// ShouldBeReviewed checks if the product master needs manual review
func (pm *ProductMaster) ShouldBeReviewed() bool {
	return !pm.IsVerified &&
		   (pm.ConfidenceScore < 0.7 ||
		    pm.GetMatchSuccessRate() < 70.0 ||
		    pm.MatchedProducts > 50)
}

// AddMatchingKeyword adds a new keyword to the matching keywords
func (pm *ProductMaster) AddMatchingKeyword(keyword string, keywordType string) error {
	keywords, err := pm.GetMatchingKeywords()
	if err != nil {
		return err
	}

	switch keywordType {
	case "primary":
		keywords.Primary = append(keywords.Primary, strings.ToLower(keyword))
	case "secondary":
		keywords.Secondary = append(keywords.Secondary, strings.ToLower(keyword))
	case "brand":
		keywords.Brand = append(keywords.Brand, strings.ToLower(keyword))
	case "category":
		keywords.Category = append(keywords.Category, strings.ToLower(keyword))
	}

	return pm.SetMatchingKeywords(keywords)
}

// AddAlternativeName adds a new alternative name
func (pm *ProductMaster) AddAlternativeName(name string, nameType string) error {
	names, err := pm.GetAlternativeNames()
	if err != nil {
		return err
	}

	switch nameType {
	case "name":
		names.Names = append(names.Names, name)
	case "alias":
		names.Aliases = append(names.Aliases, name)
	case "abbreviation":
		names.Abbreviations = append(names.Abbreviations, name)
	}

	return pm.SetAlternativeNames(names)
}

// GetAllMatchableTerms returns all terms that can be used for matching
func (pm *ProductMaster) GetAllMatchableTerms() []string {
	var terms []string

	terms = append(terms, pm.CanonicalName)
	terms = append(terms, pm.NormalizedName)
	terms = append(terms, pm.Brand)

	// Add keywords
	if keywords, err := pm.GetMatchingKeywords(); err == nil {
		terms = append(terms, keywords.Primary...)
		terms = append(terms, keywords.Secondary...)
		terms = append(terms, keywords.Brand...)
		terms = append(terms, keywords.Category...)
	}

	// Add alternative names
	if names, err := pm.GetAlternativeNames(); err == nil {
		terms = append(terms, names.Names...)
		terms = append(terms, names.Aliases...)
		terms = append(terms, names.Abbreviations...)
	}

	return terms
}

// TableName returns the table name for Bun
func (pm *ProductMaster) TableName() string {
	return "product_masters"
}