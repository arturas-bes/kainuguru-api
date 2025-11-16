package services

import (
	"strings"
	"testing"
)

func TestNormalizeProductText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase conversion",
			input:    "PIENAS",
			expected: "pienas",
		},
		{
			name:     "trim spaces",
			input:    "  pienas  ",
			expected: "pienas",
		},
		{
			name:     "collapse multiple spaces",
			input:    "pienas   2.5%",
			expected: "pienas 2.5%",
		},
		{
			name:     "combined normalization",
			input:    "  PIENAS   2.5%  RIEB.  ",
			expected: "pienas 2.5% rieb.",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeProductText(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeProductText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateSearchVector(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single word",
			input:    "pienas",
			expected: "pienas",
		},
		{
			name:     "multiple words",
			input:    "pienas 2.5% rieb.",
			expected: "pienas & 2.5% & rieb.",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSearchVector(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateSearchVector(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateProduct(t *testing.T) {
	tests := []struct {
		name      string
		prodName  string
		price     float64
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid product",
			prodName:  "Pienas 2.5%",
			price:     1.99,
			expectErr: false,
		},
		{
			name:      "empty name",
			prodName:  "",
			price:     1.99,
			expectErr: true,
			errMsg:    "product name is required",
		},
		{
			name:      "name too short",
			prodName:  "ab",
			price:     1.99,
			expectErr: true,
			errMsg:    "product name too short",
		},
		{
			name:      "name too long",
			prodName:  "a" + string(make([]byte, 150)),
			price:     1.99,
			expectErr: true,
			errMsg:    "product name too long",
		},
		{
			name:      "zero price allowed",
			prodName:  "Akcija -20%",
			price:     0,
			expectErr: false,
		},
		{
			name:      "negative price",
			prodName:  "Pienas",
			price:     -1.99,
			expectErr: true,
			errMsg:    "invalid price",
		},
		{
			name:      "price too high",
			prodName:  "Pienas",
			price:     10000.00,
			expectErr: true,
			errMsg:    "price too high",
		},
		{
			name:      "valid high price",
			prodName:  "Pienas",
			price:     9999.99,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProduct(tt.prodName, tt.price)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ValidateProduct(%q, %f) expected error containing %q, got nil", tt.prodName, tt.price, tt.errMsg)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateProduct(%q, %f) error = %q, want error containing %q", tt.prodName, tt.price, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateProduct(%q, %f) unexpected error: %v", tt.prodName, tt.price, err)
				}
			}
		})
	}
}

func TestCalculateDiscount(t *testing.T) {
	tests := []struct {
		name     string
		original float64
		current  float64
		expected float64
	}{
		{
			name:     "50% discount",
			original: 2.00,
			current:  1.00,
			expected: 50.0,
		},
		{
			name:     "25% discount",
			original: 4.00,
			current:  3.00,
			expected: 25.0,
		},
		{
			name:     "no discount",
			original: 1.99,
			current:  1.99,
			expected: 0.0,
		},
		{
			name:     "zero original price",
			original: 0,
			current:  1.00,
			expected: 0.0,
		},
		{
			name:     "negative original price",
			original: -1.00,
			current:  1.00,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDiscount(tt.original, tt.current)
			if result != tt.expected {
				t.Errorf("CalculateDiscount(%f, %f) = %f, want %f", tt.original, tt.current, result, tt.expected)
			}
		})
	}
}

func TestStandardizeUnit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "kilogramas to kg",
			input:    "kilogramas",
			expected: "kg",
		},
		{
			name:     "kg. to kg",
			input:    "kg.",
			expected: "kg",
		},
		{
			name:     "gramas to g",
			input:    "gramas",
			expected: "g",
		},
		{
			name:     "litras to l",
			input:    "litras",
			expected: "l",
		},
		{
			name:     "mililitras to ml",
			input:    "mililitras",
			expected: "ml",
		},
		{
			name:     "vienetų to vnt.",
			input:    "vienetų",
			expected: "vnt.",
		},
		{
			name:     "uppercase conversion",
			input:    "KILOGRAMAS",
			expected: "kg",
		},
		{
			name:     "trim spaces",
			input:    "  litras  ",
			expected: "l",
		},
		{
			name:     "unknown unit unchanged",
			input:    "custom",
			expected: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StandardizeUnit(tt.input)
			if result != tt.expected {
				t.Errorf("StandardizeUnit(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParsePrice(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		expectErr bool
		errMsg    string
	}{
		{
			name:      "simple price",
			input:     "1.99",
			expected:  1.99,
			expectErr: false,
		},
		{
			name:      "price with euro symbol",
			input:     "€1.99",
			expected:  1.99,
			expectErr: false,
		},
		{
			name:      "price with EUR",
			input:     "1.99 EUR",
			expected:  1.99,
			expectErr: false,
		},
		{
			name:      "price with comma separator",
			input:     "1,99",
			expected:  1.99,
			expectErr: false,
		},
		{
			name:      "price with spaces",
			input:     "  1.99  ",
			expected:  1.99,
			expectErr: false,
		},
		{
			name:      "zero price",
			input:     "0",
			expected:  0,
			expectErr: false,
		},
		{
			name:      "empty string",
			input:     "",
			expected:  0,
			expectErr: true,
			errMsg:    "empty price string",
		},
		{
			name:      "invalid format",
			input:     "abc",
			expected:  0,
			expectErr: true,
			errMsg:    "invalid price format",
		},
		{
			name:      "negative price",
			input:     "-1.99",
			expected:  0,
			expectErr: true,
			errMsg:    "negative price",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePrice(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("ParsePrice(%q) expected error containing %q, got nil", tt.input, tt.errMsg)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParsePrice(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ParsePrice(%q) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ParsePrice(%q) = %f, want %f", tt.input, result, tt.expected)
				}
			}
		})
	}
}

