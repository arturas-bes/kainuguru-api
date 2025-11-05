package normalize

import (
	"regexp"
	"strings"
	"unicode"
)

// BrandInfo represents information about a brand
type BrandInfo struct {
	Name          string   `json:"name"`
	CanonicalName string   `json:"canonical_name"`
	Aliases       []string `json:"aliases"`
	Country       string   `json:"country"`
	Category      string   `json:"category"`
	Confidence    float64  `json:"confidence"`
	IsLithuanian  bool     `json:"is_lithuanian"`
}

// BrandMapper handles brand name normalization and mapping
type BrandMapper struct {
	brandMap            map[string]BrandInfo
	aliasMap            map[string]string // alias -> canonical name
	patterns            map[string]*regexp.Regexp
	lithuanianBrands    map[string]BrandInfo
	internationalBrands map[string]BrandInfo
}

// NewBrandMapper creates a new brand mapper
func NewBrandMapper() *BrandMapper {
	bm := &BrandMapper{
		brandMap:            make(map[string]BrandInfo),
		aliasMap:            make(map[string]string),
		patterns:            make(map[string]*regexp.Regexp),
		lithuanianBrands:    make(map[string]BrandInfo),
		internationalBrands: make(map[string]BrandInfo),
	}

	bm.initializeBrands()
	bm.initializePatterns()

	return bm
}

// initializeBrands initializes the brand database
func (bm *BrandMapper) initializeBrands() {
	// Lithuanian brands
	lithuanianBrands := []BrandInfo{
		{
			Name:          "Žemaitijos",
			CanonicalName: "Žemaitijos",
			Aliases:       []string{"zemaitijos", "zemaitiju", "žemaitiju", "zemaitijų"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Vilkyškis",
			CanonicalName: "Vilkyškis",
			Aliases:       []string{"vilkyskis", "vilkyskio", "vilkyškio"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Kėdainių",
			CanonicalName: "Kėdainių",
			Aliases:       []string{"kedainiu", "kedainių", "kedainu"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Marijampolės",
			CanonicalName: "Marijampolės",
			Aliases:       []string{"marijampoles", "marijampolės", "marijampole"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Ukmergės",
			CanonicalName: "Ukmergės",
			Aliases:       []string{"ukmerges", "ukmergės", "ukmerge"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Panevėžio",
			CanonicalName: "Panevėžio",
			Aliases:       []string{"panevezio", "panevėžio", "panevezys", "panevėžys"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Kauno",
			CanonicalName: "Kauno",
			Aliases:       []string{"kauno", "kaunas"},
			Country:       "Lithuania",
			Category:      "meat",
			IsLithuanian:  true,
		},
		{
			Name:          "Alytaus",
			CanonicalName: "Alytaus",
			Aliases:       []string{"alytaus", "alytus"},
			Country:       "Lithuania",
			Category:      "meat",
			IsLithuanian:  true,
		},
		{
			Name:          "Šiaulių",
			CanonicalName: "Šiaulių",
			Aliases:       []string{"siauliu", "šiaulių", "siauliai", "šiauliai"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Klaipėdos",
			CanonicalName: "Klaipėdos",
			Aliases:       []string{"klaipedos", "klaipėdos", "klaipeda", "klaipėda"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Vilniaus",
			CanonicalName: "Vilniaus",
			Aliases:       []string{"vilniaus", "vilnius"},
			Country:       "Lithuania",
			Category:      "bakery",
			IsLithuanian:  true,
		},
		{
			Name:          "Mantinga",
			CanonicalName: "Mantinga",
			Aliases:       []string{"mantinga", "mantingos"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Pieno žvaigždės",
			CanonicalName: "Pieno žvaigždės",
			Aliases:       []string{"pieno žvaigždės", "pieno zvaigzdes", "pieno žvaigzdes"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Dvaro",
			CanonicalName: "Dvaro",
			Aliases:       []string{"dvaro", "dvaras"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
		{
			Name:          "Rokiškio",
			CanonicalName: "Rokiškio",
			Aliases:       []string{"rokiskio", "rokiškio", "rokiskis", "rokiškis"},
			Country:       "Lithuania",
			Category:      "dairy",
			IsLithuanian:  true,
		},
	}

	// International brands
	internationalBrands := []BrandInfo{
		{
			Name:          "Coca-Cola",
			CanonicalName: "Coca-Cola",
			Aliases:       []string{"coca cola", "coke", "coca-cola"},
			Country:       "USA",
			Category:      "beverages",
			IsLithuanian:  false,
		},
		{
			Name:          "Pepsi",
			CanonicalName: "Pepsi",
			Aliases:       []string{"pepsi", "pepsi-cola"},
			Country:       "USA",
			Category:      "beverages",
			IsLithuanian:  false,
		},
		{
			Name:          "Nestlé",
			CanonicalName: "Nestlé",
			Aliases:       []string{"nestle", "nestlé"},
			Country:       "Switzerland",
			Category:      "food",
			IsLithuanian:  false,
		},
		{
			Name:          "Danone",
			CanonicalName: "Danone",
			Aliases:       []string{"danone", "dannon"},
			Country:       "France",
			Category:      "dairy",
			IsLithuanian:  false,
		},
		{
			Name:          "Barilla",
			CanonicalName: "Barilla",
			Aliases:       []string{"barilla"},
			Country:       "Italy",
			Category:      "pasta",
			IsLithuanian:  false,
		},
		{
			Name:          "Ferrero",
			CanonicalName: "Ferrero",
			Aliases:       []string{"ferrero"},
			Country:       "Italy",
			Category:      "confectionery",
			IsLithuanian:  false,
		},
		{
			Name:          "Unilever",
			CanonicalName: "Unilever",
			Aliases:       []string{"unilever"},
			Country:       "Netherlands",
			Category:      "household",
			IsLithuanian:  false,
		},
		{
			Name:          "Procter & Gamble",
			CanonicalName: "Procter & Gamble",
			Aliases:       []string{"p&g", "procter & gamble", "procter and gamble"},
			Country:       "USA",
			Category:      "household",
			IsLithuanian:  false,
		},
		{
			Name:          "Heinz",
			CanonicalName: "Heinz",
			Aliases:       []string{"heinz"},
			Country:       "USA",
			Category:      "condiments",
			IsLithuanian:  false,
		},
		{
			Name:          "Kraft",
			CanonicalName: "Kraft",
			Aliases:       []string{"kraft"},
			Country:       "USA",
			Category:      "food",
			IsLithuanian:  false,
		},
		{
			Name:          "Kellogg's",
			CanonicalName: "Kellogg's",
			Aliases:       []string{"kelloggs", "kellogg's", "kellogg"},
			Country:       "USA",
			Category:      "cereal",
			IsLithuanian:  false,
		},
		{
			Name:          "General Mills",
			CanonicalName: "General Mills",
			Aliases:       []string{"general mills"},
			Country:       "USA",
			Category:      "cereal",
			IsLithuanian:  false,
		},
	}

	// Register Lithuanian brands
	for _, brand := range lithuanianBrands {
		bm.registerBrand(brand)
		bm.lithuanianBrands[strings.ToLower(brand.CanonicalName)] = brand
	}

	// Register international brands
	for _, brand := range internationalBrands {
		bm.registerBrand(brand)
		bm.internationalBrands[strings.ToLower(brand.CanonicalName)] = brand
	}
}

// registerBrand registers a brand and its aliases
func (bm *BrandMapper) registerBrand(brand BrandInfo) {
	canonical := strings.ToLower(brand.CanonicalName)
	bm.brandMap[canonical] = brand

	// Register aliases
	for _, alias := range brand.Aliases {
		aliasKey := strings.ToLower(alias)
		bm.aliasMap[aliasKey] = brand.CanonicalName
	}

	// Register the main name as an alias too
	mainKey := strings.ToLower(brand.Name)
	if mainKey != canonical {
		bm.aliasMap[mainKey] = brand.CanonicalName
	}
}

// initializePatterns initializes regex patterns for brand detection
func (bm *BrandMapper) initializePatterns() {
	bm.patterns["brand_marker"] = regexp.MustCompile(`(?i)\b(tm|®|©|\(r\)|\(tm\)|\(c\))\b`)
	bm.patterns["lithuanian_ending"] = regexp.MustCompile(`(?i)(ų|ui|ės|ėjų|ose|ais|ams)$`)
	bm.patterns["whitespace"] = regexp.MustCompile(`\s+`)
	bm.patterns["punctuation"] = regexp.MustCompile(`[^\w\s\-ąčęėįšųūžĄČĘĖĮŠŲŪŽ]`)
}

// ExtractBrands extracts brand names from text
func (bm *BrandMapper) ExtractBrands(text string) []BrandInfo {
	if text == "" {
		return []BrandInfo{}
	}

	var brands []BrandInfo
	seen := make(map[string]bool)

	// Clean text
	cleaned := bm.cleanText(text)
	words := strings.Fields(cleaned)

	// Check individual words and phrases
	for i := 0; i < len(words); i++ {
		// Check single word
		if brand := bm.findBrandByWord(words[i]); brand != nil {
			key := strings.ToLower(brand.CanonicalName)
			if !seen[key] {
				brands = append(brands, *brand)
				seen[key] = true
			}
		}

		// Check two-word combinations
		if i < len(words)-1 {
			phrase := words[i] + " " + words[i+1]
			if brand := bm.findBrandByPhrase(phrase); brand != nil {
				key := strings.ToLower(brand.CanonicalName)
				if !seen[key] {
					brands = append(brands, *brand)
					seen[key] = true
				}
			}
		}

		// Check three-word combinations
		if i < len(words)-2 {
			phrase := words[i] + " " + words[i+1] + " " + words[i+2]
			if brand := bm.findBrandByPhrase(phrase); brand != nil {
				key := strings.ToLower(brand.CanonicalName)
				if !seen[key] {
					brands = append(brands, *brand)
					seen[key] = true
				}
			}
		}
	}

	return brands
}

// findBrandByWord finds a brand by a single word
func (bm *BrandMapper) findBrandByWord(word string) *BrandInfo {
	word = strings.ToLower(strings.TrimSpace(word))

	// Direct match
	if canonical, exists := bm.aliasMap[word]; exists {
		if brand, found := bm.brandMap[strings.ToLower(canonical)]; found {
			return &brand
		}
	}

	// Try with Lithuanian endings removed
	cleanWord := bm.removeLithuanianEndings(word)
	if cleanWord != word {
		if canonical, exists := bm.aliasMap[cleanWord]; exists {
			if brand, found := bm.brandMap[strings.ToLower(canonical)]; found {
				return &brand
			}
		}
	}

	return nil
}

// findBrandByPhrase finds a brand by a phrase
func (bm *BrandMapper) findBrandByPhrase(phrase string) *BrandInfo {
	phrase = strings.ToLower(strings.TrimSpace(phrase))

	// Direct match
	if canonical, exists := bm.aliasMap[phrase]; exists {
		if brand, found := bm.brandMap[strings.ToLower(canonical)]; found {
			return &brand
		}
	}

	return nil
}

// NormalizeBrandName normalizes a brand name
func (bm *BrandMapper) NormalizeBrandName(name string) string {
	if name == "" {
		return name
	}

	cleaned := bm.cleanText(name)

	// Try to find exact match
	if canonical, exists := bm.aliasMap[strings.ToLower(cleaned)]; exists {
		return canonical
	}

	// Try with Lithuanian endings removed
	words := strings.Fields(cleaned)
	var normalizedWords []string

	for _, word := range words {
		cleanWord := bm.removeLithuanianEndings(strings.ToLower(word))
		if canonical, exists := bm.aliasMap[cleanWord]; exists {
			normalizedWords = append(normalizedWords, canonical)
		} else {
			// Capitalize first letter for unknown brands
			if len(word) > 0 {
				normalizedWords = append(normalizedWords, strings.Title(word))
			}
		}
	}

	if len(normalizedWords) > 0 {
		return strings.Join(normalizedWords, " ")
	}

	return name
}

// GetBrandInfo returns detailed information about a brand
func (bm *BrandMapper) GetBrandInfo(name string) *BrandInfo {
	name = strings.ToLower(strings.TrimSpace(name))

	// Direct lookup
	if brand, exists := bm.brandMap[name]; exists {
		return &brand
	}

	// Try through alias
	if canonical, exists := bm.aliasMap[name]; exists {
		if brand, found := bm.brandMap[strings.ToLower(canonical)]; found {
			return &brand
		}
	}

	return nil
}

// IsLithuanianBrand checks if a brand is Lithuanian
func (bm *BrandMapper) IsLithuanianBrand(name string) bool {
	brand := bm.GetBrandInfo(name)
	return brand != nil && brand.IsLithuanian
}

// GetBrandsByCategory returns brands in a specific category
func (bm *BrandMapper) GetBrandsByCategory(category string) []BrandInfo {
	var brands []BrandInfo

	for _, brand := range bm.brandMap {
		if strings.EqualFold(brand.Category, category) {
			brands = append(brands, brand)
		}
	}

	return brands
}

// GetLithuanianBrands returns all Lithuanian brands
func (bm *BrandMapper) GetLithuanianBrands() []BrandInfo {
	var brands []BrandInfo

	for _, brand := range bm.lithuanianBrands {
		brands = append(brands, brand)
	}

	return brands
}

// cleanText cleans text for brand detection
func (bm *BrandMapper) cleanText(text string) string {
	// Remove brand markers
	cleaned := bm.patterns["brand_marker"].ReplaceAllString(text, "")

	// Remove excessive punctuation
	cleaned = bm.patterns["punctuation"].ReplaceAllString(cleaned, " ")

	// Normalize whitespace
	cleaned = bm.patterns["whitespace"].ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}

// removeLithuanianEndings removes Lithuanian grammatical endings
func (bm *BrandMapper) removeLithuanianEndings(word string) string {
	endings := []string{"ų", "ui", "ės", "ėjų", "ose", "ais", "ams", "uose", "iems", "ose"}

	for _, ending := range endings {
		if strings.HasSuffix(word, ending) && len(word) > len(ending)+2 {
			return word[:len(word)-len(ending)]
		}
	}

	return word
}

// CalculateBrandConfidence calculates confidence score for brand detection
func (bm *BrandMapper) CalculateBrandConfidence(originalText, detectedBrand string) float64 {
	if originalText == "" || detectedBrand == "" {
		return 0.0
	}

	confidence := 0.5 // Base confidence

	// Exact match
	if strings.EqualFold(originalText, detectedBrand) {
		confidence += 0.4
	}

	// Brand is in known database
	if bm.GetBrandInfo(detectedBrand) != nil {
		confidence += 0.3
	}

	// Brand has proper capitalization
	if bm.hasProperCapitalization(detectedBrand) {
		confidence += 0.1
	}

	// Brand contains Lithuanian characters (for Lithuanian brands)
	if bm.containsLithuanianChars(detectedBrand) {
		confidence += 0.1
	}

	// Ensure confidence is between 0 and 1
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// hasProperCapitalization checks if text has proper brand capitalization
func (bm *BrandMapper) hasProperCapitalization(text string) bool {
	if text == "" {
		return false
	}

	// Should start with capital letter
	firstRune := rune(text[0])
	return unicode.IsUpper(firstRune)
}

// containsLithuanianChars checks if text contains Lithuanian characters
func (bm *BrandMapper) containsLithuanianChars(text string) bool {
	lithuanianChars := "ąčęėįšųūžĄČĘĖĮŠŲŪŽ"

	for _, char := range text {
		if strings.ContainsRune(lithuanianChars, char) {
			return true
		}
	}

	return false
}

// AddCustomBrand adds a custom brand to the mapper
func (bm *BrandMapper) AddCustomBrand(brand BrandInfo) {
	bm.registerBrand(brand)

	if brand.IsLithuanian {
		bm.lithuanianBrands[strings.ToLower(brand.CanonicalName)] = brand
	} else {
		bm.internationalBrands[strings.ToLower(brand.CanonicalName)] = brand
	}
}

// GetSimilarBrands finds brands similar to the given name
func (bm *BrandMapper) GetSimilarBrands(name string, threshold float64) []BrandInfo {
	var similar []BrandInfo
	name = strings.ToLower(name)

	for _, brand := range bm.brandMap {
		similarity := bm.calculateStringSimilarity(name, strings.ToLower(brand.Name))
		if similarity >= threshold {
			brandCopy := brand
			brandCopy.Confidence = similarity
			similar = append(similar, brandCopy)
		}

		// Check aliases too
		for _, alias := range brand.Aliases {
			similarity := bm.calculateStringSimilarity(name, strings.ToLower(alias))
			if similarity >= threshold {
				brandCopy := brand
				brandCopy.Confidence = similarity
				similar = append(similar, brandCopy)
				break // Don't add the same brand multiple times
			}
		}
	}

	return similar
}

// calculateStringSimilarity calculates similarity between two strings
func (bm *BrandMapper) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// Simple Jaccard similarity based on character n-grams
	ngrams1 := bm.getNgrams(s1, 2)
	ngrams2 := bm.getNgrams(s2, 2)

	intersection := 0
	for ngram := range ngrams1 {
		if ngrams2[ngram] {
			intersection++
		}
	}

	union := len(ngrams1) + len(ngrams2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// getNgrams generates character n-grams from a string
func (bm *BrandMapper) getNgrams(s string, n int) map[string]bool {
	ngrams := make(map[string]bool)

	if len(s) < n {
		ngrams[s] = true
		return ngrams
	}

	for i := 0; i <= len(s)-n; i++ {
		ngrams[s[i:i+n]] = true
	}

	return ngrams
}
