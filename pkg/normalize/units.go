package normalize

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// UnitType represents different types of units
type UnitType string

const (
	UnitTypeWeight  UnitType = "weight"
	UnitTypeVolume  UnitType = "volume"
	UnitTypeCount   UnitType = "count"
	UnitTypeLength  UnitType = "length"
	UnitTypeArea    UnitType = "area"
	UnitTypeUnknown UnitType = "unknown"
)

// Unit represents a measurement unit with value and type
type Unit struct {
	Value      float64  `json:"value"`
	Unit       string   `json:"unit"`
	Type       UnitType `json:"type"`
	Original   string   `json:"original"`
	Normalized string   `json:"normalized"`
	BaseValue  float64  `json:"base_value"` // Value in base unit (g, ml, etc.)
	BaseUnit   string   `json:"base_unit"`  // Base unit (g, ml, vnt, etc.)
}

// UnitExtractor extracts and normalizes units from Lithuanian text
type UnitExtractor struct {
	weightPatterns    []*regexp.Regexp
	volumePatterns    []*regexp.Regexp
	countPatterns     []*regexp.Regexp
	lengthPatterns    []*regexp.Regexp
	areaPatterns      []*regexp.Regexp
	weightConversions map[string]float64
	volumeConversions map[string]float64
	lengthConversions map[string]float64
}

// NewUnitExtractor creates a new unit extractor
func NewUnitExtractor() *UnitExtractor {
	return &UnitExtractor{
		weightPatterns:    createWeightPatterns(),
		volumePatterns:    createVolumePatterns(),
		countPatterns:     createCountPatterns(),
		lengthPatterns:    createLengthPatterns(),
		areaPatterns:      createAreaPatterns(),
		weightConversions: createWeightConversions(),
		volumeConversions: createVolumeConversions(),
		lengthConversions: createLengthConversions(),
	}
}

// createWeightPatterns creates regex patterns for weight units
func createWeightPatterns() []*regexp.Regexp {
	patterns := []string{
		// Standard formats
		`(\d+(?:[,.]?\d+)?)\s*(kg|kilogram[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(g|gram[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(t|ton[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(mg|miligram[ųuai]*)\b`,

		// Lithuanian forms
		`(\d+(?:[,.]?\d+)?)\s*(kilogram[ųuai]|kg)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(gram[ųuai]|g)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(ton[ųuai]|t)\b`,

		// Compound formats
		`(\d+)\s*kg\s*(\d+)\s*g`,
		`(\d+)\s*kilogram[ųuai]*\s*(\d+)\s*gram[ųuai]*`,

		// Range formats
		`(\d+(?:[,.]?\d+)?)\s*[-–]\s*(\d+(?:[,.]?\d+)?)\s*(kg|g|kilogram[ųuai]*|gram[ųuai]*)\b`,
	}

	var regexes []*regexp.Regexp
	for _, pattern := range patterns {
		regexes = append(regexes, regexp.MustCompile(`(?i)`+pattern))
	}
	return regexes
}

// createVolumePatterns creates regex patterns for volume units
func createVolumePatterns() []*regexp.Regexp {
	patterns := []string{
		// Standard formats
		`(\d+(?:[,.]?\d+)?)\s*(l|litr[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(ml|mililitr[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(cl|centilitr[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(dl|decilitr[ųuai]*)\b`,

		// Lithuanian forms
		`(\d+(?:[,.]?\d+)?)\s*(litr[ųuai]|l)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(mililitr[ųuai]|ml)\b`,

		// Compound formats
		`(\d+)\s*l\s*(\d+)\s*ml`,
		`(\d+)\s*litr[ųuai]*\s*(\d+)\s*mililitr[ųuai]*`,

		// Range formats
		`(\d+(?:[,.]?\d+)?)\s*[-–]\s*(\d+(?:[,.]?\d+)?)\s*(l|ml|litr[ųuai]*|mililitr[ųuai]*)\b`,
	}

	var regexes []*regexp.Regexp
	for _, pattern := range patterns {
		regexes = append(regexes, regexp.MustCompile(`(?i)`+pattern))
	}
	return regexes
}

// createCountPatterns creates regex patterns for count units
func createCountPatterns() []*regexp.Regexp {
	patterns := []string{
		// Standard formats
		`(\d+)\s*(vnt\.?|vienet[ųuai]*)\b`,
		`(\d+)\s*(pak\.?|pakuot[ėęjų]*)\b`,
		`(\d+)\s*(dėž\.?|dėž[ėęjų]*)\b`,
		`(\d+)\s*(bot\.?|butelis|buteliai|butelių)\b`,
		`(\d+)\s*(skard\.?|skardin[ėęjų]*)\b`,

		// Lithuanian forms
		`(\d+)\s*(vienet[ųuai]*|vnt\.?)\b`,
		`(\d+)\s*(pakuot[ėęjų]*|pak\.?)\b`,
		`(\d+)\s*(dėž[ėęjų]*|dėž\.?)\b`,

		// Special cases
		`(\d+)\s*(gabals?|gabalai|gabalų)\b`,
		`(\d+)\s*(port[ųuai]*|porcij[ųuai]*)\b`,
		`(\d+)\s*(dalys|dalių|dalis)\b`,

		// Range formats
		`(\d+)\s*[-–]\s*(\d+)\s*(vnt\.?|vienet[ųuai]*|pak\.?|pakuot[ėęjų]*)\b`,
	}

	var regexes []*regexp.Regexp
	for _, pattern := range patterns {
		regexes = append(regexes, regexp.MustCompile(`(?i)`+pattern))
	}
	return regexes
}

// createLengthPatterns creates regex patterns for length units
func createLengthPatterns() []*regexp.Regexp {
	patterns := []string{
		`(\d+(?:[,.]?\d+)?)\s*(m|metr[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(cm|centimetr[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(mm|milimetr[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(km|kilometr[ųuai]*)\b`,
	}

	var regexes []*regexp.Regexp
	for _, pattern := range patterns {
		regexes = append(regexes, regexp.MustCompile(`(?i)`+pattern))
	}
	return regexes
}

// createAreaPatterns creates regex patterns for area units
func createAreaPatterns() []*regexp.Regexp {
	patterns := []string{
		`(\d+(?:[,.]?\d+)?)\s*(m²|m2|kvadratini[ųuai]*\s*metr[ųuai]*)\b`,
		`(\d+(?:[,.]?\d+)?)\s*(cm²|cm2|kvadratini[ųuai]*\s*centimetr[ųuai]*)\b`,
	}

	var regexes []*regexp.Regexp
	for _, pattern := range patterns {
		regexes = append(regexes, regexp.MustCompile(`(?i)`+pattern))
	}
	return regexes
}

// createWeightConversions creates weight conversion factors to grams
func createWeightConversions() map[string]float64 {
	return map[string]float64{
		"g":          1.0,
		"gram":       1.0,
		"gramai":     1.0,
		"gramų":      1.0,
		"kg":         1000.0,
		"kilogram":   1000.0,
		"kilogramai": 1000.0,
		"kilogramų":  1000.0,
		"t":          1000000.0,
		"tona":       1000000.0,
		"tonos":      1000000.0,
		"tonų":       1000000.0,
		"mg":         0.001,
		"miligram":   0.001,
		"miligramai": 0.001,
		"miligramų":  0.001,
	}
}

// createVolumeConversions creates volume conversion factors to milliliters
func createVolumeConversions() map[string]float64 {
	return map[string]float64{
		"ml":          1.0,
		"mililitras":  1.0,
		"mililitrai":  1.0,
		"mililitrų":   1.0,
		"l":           1000.0,
		"litras":      1000.0,
		"litrai":      1000.0,
		"litrų":       1000.0,
		"cl":          10.0,
		"centilitras": 10.0,
		"centilitrai": 10.0,
		"centilitrų":  10.0,
		"dl":          100.0,
		"decilitras":  100.0,
		"decilitrai":  100.0,
		"decilitrų":   100.0,
	}
}

// createLengthConversions creates length conversion factors to millimeters
func createLengthConversions() map[string]float64 {
	return map[string]float64{
		"mm":          1.0,
		"milimetras":  1.0,
		"milimetrai":  1.0,
		"milimetrų":   1.0,
		"cm":          10.0,
		"centimetras": 10.0,
		"centimetrai": 10.0,
		"centimetrų":  10.0,
		"m":           1000.0,
		"metras":      1000.0,
		"metrai":      1000.0,
		"metrų":       1000.0,
		"km":          1000000.0,
		"kilometras":  1000000.0,
		"kilometrai":  1000000.0,
		"kilometrų":   1000000.0,
	}
}

// ExtractUnits extracts all units from text
func (ue *UnitExtractor) ExtractUnits(text string) []Unit {
	var units []Unit

	// Extract weight units
	units = append(units, ue.extractByType(text, UnitTypeWeight, ue.weightPatterns, ue.weightConversions, "g")...)

	// Extract volume units
	units = append(units, ue.extractByType(text, UnitTypeVolume, ue.volumePatterns, ue.volumeConversions, "ml")...)

	// Extract count units
	units = append(units, ue.extractByType(text, UnitTypeCount, ue.countPatterns, nil, "vnt.")...)

	// Extract length units
	units = append(units, ue.extractByType(text, UnitTypeLength, ue.lengthPatterns, ue.lengthConversions, "mm")...)

	// Extract area units
	units = append(units, ue.extractByType(text, UnitTypeArea, ue.areaPatterns, nil, "m²")...)

	return units
}

// extractByType extracts units of a specific type
func (ue *UnitExtractor) extractByType(text string, unitType UnitType, patterns []*regexp.Regexp, conversions map[string]float64, baseUnit string) []Unit {
	var units []Unit

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				unit := ue.parseUnit(match, unitType, conversions, baseUnit)
				if unit.Value > 0 {
					units = append(units, unit)
				}
			}
		}
	}

	return units
}

// parseUnit parses a unit from regex match
func (ue *UnitExtractor) parseUnit(match []string, unitType UnitType, conversions map[string]float64, baseUnit string) Unit {
	unit := Unit{
		Type:     unitType,
		Original: match[0],
	}

	// Parse value
	valueStr := strings.ReplaceAll(match[1], ",", ".")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		unit.Value = value
	}

	// Parse unit string
	if len(match) >= 3 {
		unit.Unit = strings.ToLower(strings.TrimSpace(match[2]))
	}

	// Normalize unit
	unit.Normalized = ue.normalizeUnit(unit.Unit, unitType)

	// Calculate base value
	if conversions != nil {
		if factor, exists := conversions[unit.Unit]; exists {
			unit.BaseValue = unit.Value * factor
			unit.BaseUnit = baseUnit
		} else {
			// Try without Lithuanian endings
			cleanUnit := ue.cleanLithuanianEndings(unit.Unit)
			if factor, exists := conversions[cleanUnit]; exists {
				unit.BaseValue = unit.Value * factor
				unit.BaseUnit = baseUnit
			}
		}
	} else {
		unit.BaseValue = unit.Value
		unit.BaseUnit = unit.Normalized
	}

	return unit
}

// normalizeUnit normalizes unit string
func (ue *UnitExtractor) normalizeUnit(unit string, unitType UnitType) string {
	unit = strings.ToLower(strings.TrimSpace(unit))

	switch unitType {
	case UnitTypeWeight:
		return ue.normalizeWeightUnit(unit)
	case UnitTypeVolume:
		return ue.normalizeVolumeUnit(unit)
	case UnitTypeCount:
		return ue.normalizeCountUnit(unit)
	case UnitTypeLength:
		return ue.normalizeLengthUnit(unit)
	case UnitTypeArea:
		return ue.normalizeAreaUnit(unit)
	}

	return unit
}

// normalizeWeightUnit normalizes weight unit
func (ue *UnitExtractor) normalizeWeightUnit(unit string) string {
	unit = ue.cleanLithuanianEndings(unit)

	switch {
	case strings.Contains(unit, "kg") || strings.Contains(unit, "kilogram"):
		return "kg"
	case strings.Contains(unit, "g") && !strings.Contains(unit, "kg"):
		return "g"
	case strings.Contains(unit, "t") || strings.Contains(unit, "ton"):
		return "t"
	case strings.Contains(unit, "mg") || strings.Contains(unit, "miligram"):
		return "mg"
	}

	return unit
}

// normalizeVolumeUnit normalizes volume unit
func (ue *UnitExtractor) normalizeVolumeUnit(unit string) string {
	unit = ue.cleanLithuanianEndings(unit)

	switch {
	case unit == "l" || strings.Contains(unit, "litr"):
		return "l"
	case unit == "ml" || strings.Contains(unit, "mililitr"):
		return "ml"
	case unit == "cl" || strings.Contains(unit, "centilitr"):
		return "cl"
	case unit == "dl" || strings.Contains(unit, "decilitr"):
		return "dl"
	}

	return unit
}

// normalizeCountUnit normalizes count unit
func (ue *UnitExtractor) normalizeCountUnit(unit string) string {
	unit = ue.cleanLithuanianEndings(unit)

	switch {
	case strings.Contains(unit, "vnt") || strings.Contains(unit, "vienet"):
		return "vnt."
	case strings.Contains(unit, "pak") || strings.Contains(unit, "pakuot"):
		return "pak."
	case strings.Contains(unit, "dėž"):
		return "dėž."
	case strings.Contains(unit, "bot") || strings.Contains(unit, "butel"):
		return "bot."
	case strings.Contains(unit, "skard"):
		return "skard."
	case strings.Contains(unit, "gabal"):
		return "gab."
	case strings.Contains(unit, "port") || strings.Contains(unit, "porcij"):
		return "port."
	}

	return unit
}

// normalizeLengthUnit normalizes length unit
func (ue *UnitExtractor) normalizeLengthUnit(unit string) string {
	unit = ue.cleanLithuanianEndings(unit)

	switch {
	case unit == "m" || strings.Contains(unit, "metr"):
		return "m"
	case unit == "cm" || strings.Contains(unit, "centimetr"):
		return "cm"
	case unit == "mm" || strings.Contains(unit, "milimetr"):
		return "mm"
	case unit == "km" || strings.Contains(unit, "kilometr"):
		return "km"
	}

	return unit
}

// normalizeAreaUnit normalizes area unit
func (ue *UnitExtractor) normalizeAreaUnit(unit string) string {
	unit = ue.cleanLithuanianEndings(unit)

	switch {
	case strings.Contains(unit, "m²") || strings.Contains(unit, "m2") || strings.Contains(unit, "kvadratini"):
		return "m²"
	case strings.Contains(unit, "cm²") || strings.Contains(unit, "cm2"):
		return "cm²"
	}

	return unit
}

// cleanLithuanianEndings removes Lithuanian grammatical endings
func (ue *UnitExtractor) cleanLithuanianEndings(unit string) string {
	// Remove common Lithuanian endings
	endings := []string{"ų", "ai", "ams", "uose", "ais"}
	for _, ending := range endings {
		if strings.HasSuffix(unit, ending) {
			return unit[:len(unit)-len(ending)]
		}
	}
	return unit
}

// GetPrimaryUnit returns the most significant unit from a list
func (ue *UnitExtractor) GetPrimaryUnit(units []Unit) *Unit {
	if len(units) == 0 {
		return nil
	}

	// Priority: weight > volume > count > length > area
	typePriority := map[UnitType]int{
		UnitTypeWeight: 1,
		UnitTypeVolume: 2,
		UnitTypeCount:  3,
		UnitTypeLength: 4,
		UnitTypeArea:   5,
	}

	var bestUnit *Unit
	bestPriority := 999

	for i := range units {
		if priority, exists := typePriority[units[i].Type]; exists && priority < bestPriority {
			bestPriority = priority
			bestUnit = &units[i]
		}
	}

	if bestUnit == nil {
		return &units[0]
	}

	return bestUnit
}

// ConvertUnit converts a unit to a different unit of the same type
func (ue *UnitExtractor) ConvertUnit(unit Unit, targetUnit string) (*Unit, error) {
	if unit.BaseValue == 0 || unit.BaseUnit == "" {
		return nil, fmt.Errorf("cannot convert unit without base value")
	}

	var conversions map[string]float64
	switch unit.Type {
	case UnitTypeWeight:
		conversions = ue.weightConversions
	case UnitTypeVolume:
		conversions = ue.volumeConversions
	case UnitTypeLength:
		conversions = ue.lengthConversions
	default:
		return nil, fmt.Errorf("conversion not supported for unit type %s", unit.Type)
	}

	targetFactor, exists := conversions[targetUnit]
	if !exists {
		return nil, fmt.Errorf("unknown target unit: %s", targetUnit)
	}

	convertedValue := unit.BaseValue / targetFactor

	converted := Unit{
		Value:      convertedValue,
		Unit:       targetUnit,
		Type:       unit.Type,
		Original:   unit.Original,
		Normalized: ue.normalizeUnit(targetUnit, unit.Type),
		BaseValue:  unit.BaseValue,
		BaseUnit:   unit.BaseUnit,
	}

	return &converted, nil
}

// CompareUnits compares two units of the same type
func (ue *UnitExtractor) CompareUnits(unit1, unit2 Unit) int {
	if unit1.Type != unit2.Type {
		return 0 // Cannot compare different types
	}

	if unit1.BaseValue < unit2.BaseValue {
		return -1
	} else if unit1.BaseValue > unit2.BaseValue {
		return 1
	}
	return 0
}

// FormatUnit formats a unit for display
func (ue *UnitExtractor) FormatUnit(unit Unit) string {
	if unit.Value == float64(int(unit.Value)) {
		return fmt.Sprintf("%d %s", int(unit.Value), unit.Normalized)
	}
	return fmt.Sprintf("%.1f %s", unit.Value, unit.Normalized)
}
