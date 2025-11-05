package normalize

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// LithuanianNormalizer handles Lithuanian text normalization
type LithuanianNormalizer struct {
	diacriticMap map[rune]rune
	stopWords    map[string]bool
	brandMap     map[string]string
	patterns     map[string]*regexp.Regexp
}

// NewLithuanianNormalizer creates a new Lithuanian text normalizer
func NewLithuanianNormalizer() *LithuanianNormalizer {
	return &LithuanianNormalizer{
		diacriticMap: createDiacriticMap(),
		stopWords:    createStopWords(),
		brandMap:     createBrandMap(),
		patterns:     createPatterns(),
	}
}

// createDiacriticMap creates mapping for Lithuanian diacritics
func createDiacriticMap() map[rune]rune {
	return map[rune]rune{
		'ą': 'a', 'Ą': 'A',
		'č': 'c', 'Č': 'C',
		'ę': 'e', 'Ę': 'E',
		'ė': 'e', 'Ė': 'E',
		'į': 'i', 'Į': 'I',
		'š': 's', 'Š': 'S',
		'ų': 'u', 'Ų': 'U',
		'ū': 'u', 'Ū': 'U',
		'ž': 'z', 'Ž': 'Z',
	}
}

// createStopWords creates a set of Lithuanian stop words
func createStopWords() map[string]bool {
	stopWords := []string{
		"ir", "arba", "bet", "kad", "kaip", "su", "be", "po", "per", "nuo", "iki",
		"už", "į", "iš", "ant", "po", "prie", "tarp", "dėl", "pagal", "apie",
		"bei", "taip", "pat", "jau", "dar", "tik", "net", "vis", "kiek", "kur",
		"kurie", "kurios", "kuris", "kuri", "koks", "kokia", "kokie", "kokios",
		"šis", "ši", "šie", "šios", "tas", "ta", "tie", "tos",
		"mano", "tavo", "jo", "jos", "mūsų", "jūsų", "jų",
		"aš", "tu", "jis", "ji", "mes", "jūs", "jie", "jos",
		"yra", "buvo", "bus", "būti", "eiti", "ateiti", "daryti",
		"didelė", "didelis", "dideli", "didelės", "maža", "mažas", "maži", "mažos",
		"gera", "geras", "geri", "geros", "bloga", "blogas", "blogi", "blogos",
		"nauja", "naujas", "nauji", "naujos", "sena", "senas", "seni", "senos",
	}

	stopWordsMap := make(map[string]bool)
	for _, word := range stopWords {
		stopWordsMap[strings.ToLower(word)] = true
	}
	return stopWordsMap
}

// createBrandMap creates mapping for common brand variations
func createBrandMap() map[string]string {
	return map[string]string{
		"vilkyskis":    "Vilkyškis",
		"vilkyskiu":    "Vilkyškis",
		"zemaitijos":   "Žemaitijos",
		"zemaitiju":    "Žemaitijos",
		"kedainiu":     "Kėdainių",
		"kedainu":      "Kėdainių",
		"ukmerges":     "Ukmergės",
		"ukmerge":      "Ukmergės",
		"panevezio":    "Panevėžio",
		"panevezys":    "Panevėžys",
		"marijampoles": "Marijampolės",
		"marijampole":  "Marijampolė",
		"alytaus":      "Alytaus",
		"alytus":       "Alytus",
		"kauno":        "Kauno",
		"kaunas":       "Kaunas",
		"vilniaus":     "Vilniaus",
		"vilnius":      "Vilnius",
		"klaipedos":    "Klaipėdos",
		"klaipeda":     "Klaipėda",
		"siauliu":      "Šiaulių",
		"siauliai":     "Šiauliai",
	}
}

// createPatterns creates regex patterns for text processing
func createPatterns() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"whitespace":     regexp.MustCompile(`\s+`),
		"punctuation":    regexp.MustCompile(`[^\w\s\-ąčęėįšųūž]`),
		"numbers":        regexp.MustCompile(`\d+`),
		"price":          regexp.MustCompile(`(\d+)[,.](\d{2})\s*€?`),
		"percentage":     regexp.MustCompile(`(\d+)[,.]?(\d*)\s*%`),
		"weight":         regexp.MustCompile(`(\d+)[,.]?(\d*)\s*(kg|g|gram[ui]?|kilogram[ui]?)`),
		"volume":         regexp.MustCompile(`(\d+)[,.]?(\d*)\s*(l|ml|litr[ui]?|mililitr[ui]?)`),
		"packaging":      regexp.MustCompile(`(\d+)\s*(vnt|vienet[ui]?|pak|pakuot[ėę]s?|dėž[ėę]s?)`),
		"brand_marker":   regexp.MustCompile(`(?i)(^|\s)(tm|®|©|\(r\))\s*`),
		"size_marker":    regexp.MustCompile(`(?i)(dydis|izmeros|matmenys|ilgis|plotis|aukstis)`),
		"quality_marker": regexp.MustCompile(`(?i)(kokybe|kokybes|premium|auksciausia|geriausia)`),
	}
}

// NormalizeText performs comprehensive text normalization
func (ln *LithuanianNormalizer) NormalizeText(text string) string {
	if text == "" {
		return text
	}

	// Step 1: Clean basic formatting
	normalized := strings.TrimSpace(text)
	normalized = ln.patterns["whitespace"].ReplaceAllString(normalized, " ")

	// Step 2: Normalize Lithuanian characters for search
	normalized = ln.normalizeCase(normalized)

	// Step 3: Remove unnecessary punctuation (but keep hyphens and Lithuanian chars)
	normalized = ln.cleanPunctuation(normalized)

	// Step 4: Normalize common brand variations
	normalized = ln.normalizeBrands(normalized)

	// Step 5: Final cleanup
	normalized = strings.TrimSpace(normalized)

	return normalized
}

// NormalizeForSearch creates a search-friendly version of text
func (ln *LithuanianNormalizer) NormalizeForSearch(text string) string {
	if text == "" {
		return text
	}

	// Start with basic normalization
	normalized := ln.NormalizeText(text)

	// Convert to lowercase for search
	normalized = strings.ToLower(normalized)

	// Remove diacritics for broader matching
	normalized = ln.removeDiacritics(normalized)

	// Remove stop words
	normalized = ln.removeStopWords(normalized)

	// Normalize units and measurements
	normalized = ln.normalizeUnits(normalized)

	return strings.TrimSpace(normalized)
}

// NormalizeProductName specifically normalizes product names
func (ln *LithuanianNormalizer) NormalizeProductName(name string) string {
	if name == "" {
		return name
	}

	// Basic normalization
	normalized := ln.NormalizeText(name)

	// Remove brand markers
	normalized = ln.patterns["brand_marker"].ReplaceAllString(normalized, " ")

	// Remove size/quality markers that don't add value
	normalized = ln.patterns["size_marker"].ReplaceAllString(normalized, " ")

	// Normalize units consistently
	normalized = ln.normalizeUnits(normalized)

	// Final cleanup
	normalized = ln.patterns["whitespace"].ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)

	return normalized
}

// ExtractKeywords extracts meaningful keywords from text
func (ln *LithuanianNormalizer) ExtractKeywords(text string) []string {
	if text == "" {
		return []string{}
	}

	// Normalize for processing
	normalized := ln.NormalizeForSearch(text)

	// Split into words
	words := strings.Fields(normalized)

	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		// Skip very short words
		if len(word) < 3 {
			continue
		}

		// Skip stop words
		if ln.stopWords[word] {
			continue
		}

		// Skip if already seen
		if seen[word] {
			continue
		}

		// Skip pure numbers
		if ln.patterns["numbers"].MatchString(word) && len(word) < 4 {
			continue
		}

		keywords = append(keywords, word)
		seen[word] = true
	}

	return keywords
}

// normalizeCase handles case normalization while preserving Lithuanian specifics
func (ln *LithuanianNormalizer) normalizeCase(text string) string {
	// Convert to title case for proper nouns, but preserve Lithuanian formatting
	words := strings.Fields(text)
	var normalized []string

	for _, word := range words {
		// Check if it's likely a brand or proper noun (starts with capital)
		if len(word) > 0 && unicode.IsUpper(rune(word[0])) {
			// Keep first letter uppercase, rest lowercase
			if len(word) > 1 {
				word = string(unicode.ToUpper(rune(word[0]))) + strings.ToLower(word[1:])
			}
		} else {
			word = strings.ToLower(word)
		}
		normalized = append(normalized, word)
	}

	return strings.Join(normalized, " ")
}

// cleanPunctuation removes unnecessary punctuation while preserving important chars
func (ln *LithuanianNormalizer) cleanPunctuation(text string) string {
	// Remove most punctuation but keep hyphens, apostrophes, and Lithuanian chars
	cleaned := regexp.MustCompile(`[^\w\s\-'ąčęėįšųūžĄČĘĖĮŠŲŪŽ]`).ReplaceAllString(text, " ")
	return ln.patterns["whitespace"].ReplaceAllString(cleaned, " ")
}

// normalizeBrands normalizes common brand name variations
func (ln *LithuanianNormalizer) normalizeBrands(text string) string {
	words := strings.Fields(text)
	var normalized []string

	for _, word := range words {
		lowerWord := strings.ToLower(word)
		if brandName, exists := ln.brandMap[lowerWord]; exists {
			normalized = append(normalized, brandName)
		} else {
			normalized = append(normalized, word)
		}
	}

	return strings.Join(normalized, " ")
}

// removeDiacritics removes Lithuanian diacritics for broader search matching
func (ln *LithuanianNormalizer) removeDiacritics(text string) string {
	var result strings.Builder
	for _, r := range text {
		if replacement, exists := ln.diacriticMap[r]; exists {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// removeStopWords removes common Lithuanian stop words
func (ln *LithuanianNormalizer) removeStopWords(text string) string {
	words := strings.Fields(text)
	var filtered []string

	for _, word := range words {
		if !ln.stopWords[strings.ToLower(word)] {
			filtered = append(filtered, word)
		}
	}

	return strings.Join(filtered, " ")
}

// normalizeUnits normalizes unit representations
func (ln *LithuanianNormalizer) normalizeUnits(text string) string {
	// Normalize weight units
	text = ln.patterns["weight"].ReplaceAllStringFunc(text, func(match string) string {
		return ln.normalizeWeightUnit(match)
	})

	// Normalize volume units
	text = ln.patterns["volume"].ReplaceAllStringFunc(text, func(match string) string {
		return ln.normalizeVolumeUnit(match)
	})

	// Normalize packaging units
	text = ln.patterns["packaging"].ReplaceAllStringFunc(text, func(match string) string {
		return ln.normalizePackagingUnit(match)
	})

	return text
}

// normalizeWeightUnit normalizes weight unit expressions
func (ln *LithuanianNormalizer) normalizeWeightUnit(match string) string {
	// Extract number and unit
	matches := ln.patterns["weight"].FindStringSubmatch(match)
	if len(matches) < 4 {
		return match
	}

	number := matches[1]
	decimal := matches[2]
	unit := strings.ToLower(matches[3])

	// Convert number
	var value float64
	if decimal != "" {
		if val, err := strconv.ParseFloat(number+"."+decimal, 64); err == nil {
			value = val
		}
	} else {
		if val, err := strconv.ParseFloat(number, 64); err == nil {
			value = val
		}
	}

	// Normalize unit
	switch {
	case strings.Contains(unit, "kg") || strings.Contains(unit, "kilogram"):
		if value == float64(int(value)) {
			return strconv.Itoa(int(value)) + " kg"
		}
		return strconv.FormatFloat(value, 'f', -1, 64) + " kg"
	case strings.Contains(unit, "g") || strings.Contains(unit, "gram"):
		if value == float64(int(value)) {
			return strconv.Itoa(int(value)) + " g"
		}
		return strconv.FormatFloat(value, 'f', -1, 64) + " g"
	}

	return match
}

// normalizeVolumeUnit normalizes volume unit expressions
func (ln *LithuanianNormalizer) normalizeVolumeUnit(match string) string {
	matches := ln.patterns["volume"].FindStringSubmatch(match)
	if len(matches) < 4 {
		return match
	}

	number := matches[1]
	decimal := matches[2]
	unit := strings.ToLower(matches[3])

	var value float64
	if decimal != "" {
		if val, err := strconv.ParseFloat(number+"."+decimal, 64); err == nil {
			value = val
		}
	} else {
		if val, err := strconv.ParseFloat(number, 64); err == nil {
			value = val
		}
	}

	switch {
	case strings.Contains(unit, "l") && !strings.Contains(unit, "ml"):
		if value == float64(int(value)) {
			return strconv.Itoa(int(value)) + " l"
		}
		return strconv.FormatFloat(value, 'f', -1, 64) + " l"
	case strings.Contains(unit, "ml") || strings.Contains(unit, "mililitr"):
		if value == float64(int(value)) {
			return strconv.Itoa(int(value)) + " ml"
		}
		return strconv.FormatFloat(value, 'f', -1, 64) + " ml"
	}

	return match
}

// normalizePackagingUnit normalizes packaging unit expressions
func (ln *LithuanianNormalizer) normalizePackagingUnit(match string) string {
	matches := ln.patterns["packaging"].FindStringSubmatch(match)
	if len(matches) < 3 {
		return match
	}

	number := matches[1]
	unit := strings.ToLower(matches[2])

	switch {
	case strings.Contains(unit, "vnt") || strings.Contains(unit, "vienet"):
		return number + " vnt."
	case strings.Contains(unit, "pak") || strings.Contains(unit, "pakuot"):
		return number + " pak."
	case strings.Contains(unit, "dėž"):
		return number + " dėž."
	}

	return match
}

// GetSimilarityScore calculates similarity between two Lithuanian texts
func (ln *LithuanianNormalizer) GetSimilarityScore(text1, text2 string) float64 {
	if text1 == "" || text2 == "" {
		return 0.0
	}

	// Normalize both texts for comparison
	norm1 := ln.NormalizeForSearch(text1)
	norm2 := ln.NormalizeForSearch(text2)

	// Extract keywords
	keywords1 := ln.ExtractKeywords(norm1)
	keywords2 := ln.ExtractKeywords(norm2)

	if len(keywords1) == 0 || len(keywords2) == 0 {
		return 0.0
	}

	// Calculate Jaccard similarity
	set1 := make(map[string]bool)
	for _, keyword := range keywords1 {
		set1[keyword] = true
	}

	intersection := 0
	for _, keyword := range keywords2 {
		if set1[keyword] {
			intersection++
		}
	}

	union := len(keywords1) + len(keywords2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// IsLithuanianText checks if text contains Lithuanian characteristics
func (ln *LithuanianNormalizer) IsLithuanianText(text string) bool {
	if text == "" {
		return false
	}

	// Check for Lithuanian diacritics
	lithuanianChars := 0
	totalChars := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			totalChars++
			if _, isLithuanian := ln.diacriticMap[r]; isLithuanian {
				lithuanianChars++
			}
		}
	}

	if totalChars == 0 {
		return false
	}

	// If more than 5% of letters are Lithuanian diacritics, likely Lithuanian
	lithuanianRatio := float64(lithuanianChars) / float64(totalChars)
	return lithuanianRatio > 0.05
}

// SplitIntoSentences splits Lithuanian text into sentences
func (ln *LithuanianNormalizer) SplitIntoSentences(text string) []string {
	// Lithuanian sentence endings
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(text, -1)

	var result []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			result = append(result, sentence)
		}
	}

	return result
}
