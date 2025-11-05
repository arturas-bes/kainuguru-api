package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// passwordService implements PasswordService
type passwordService struct {
	config *AuthConfig
}

// NewPasswordService creates a new password service
func NewPasswordService(config *AuthConfig) PasswordService {
	return &passwordService{
		config: config,
	}
}

// HashPassword hashes a password using bcrypt
func (p *passwordService) HashPassword(password string) (string, error) {
	if err := p.ValidatePasswordStrength(password); err != nil {
		return "", fmt.Errorf("password validation failed: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.config.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func (p *passwordService) VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return fmt.Errorf("invalid password")
		}
		return fmt.Errorf("password verification failed: %w", err)
	}
	return nil
}

// GenerateRandomPassword generates a cryptographically secure random password
func (p *passwordService) GenerateRandomPassword(length int) string {
	if length < 8 {
		length = 12 // Default to 12 characters minimum
	}

	// Define character sets
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	symbols := "!@#$%^&*()-_=+[]{}|;:,.<>?"

	// Ensure we have at least one character from each required set
	var password strings.Builder
	var allChars string

	// Add required characters based on config
	if p.config.PasswordRequireLower {
		password.WriteByte(lowercase[p.secureRandInt(len(lowercase))])
		allChars += lowercase
	}

	if p.config.PasswordRequireUpper {
		password.WriteByte(uppercase[p.secureRandInt(len(uppercase))])
		allChars += uppercase
	}

	if p.config.PasswordRequireNumber {
		password.WriteByte(numbers[p.secureRandInt(len(numbers))])
		allChars += numbers
	}

	if p.config.PasswordRequireSymbol {
		password.WriteByte(symbols[p.secureRandInt(len(symbols))])
		allChars += symbols
	}

	// If no requirements set, use all character types
	if allChars == "" {
		allChars = lowercase + uppercase + numbers + symbols
	}

	// Fill remaining length with random characters
	for password.Len() < length {
		password.WriteByte(allChars[p.secureRandInt(len(allChars))])
	}

	// Shuffle the password to avoid predictable patterns
	passwordBytes := []byte(password.String())
	for i := len(passwordBytes) - 1; i > 0; i-- {
		j := p.secureRandInt(i + 1)
		passwordBytes[i], passwordBytes[j] = passwordBytes[j], passwordBytes[i]
	}

	return string(passwordBytes)
}

// ValidatePasswordStrength validates password against configured requirements
func (p *passwordService) ValidatePasswordStrength(password string) error {
	if len(password) < p.config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters long", p.config.PasswordMinLength)
	}

	// Check for maximum reasonable length (prevent DoS)
	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	var hasLower, hasUpper, hasNumber, hasSymbol bool

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	// Check requirements
	if p.config.PasswordRequireLower && !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if p.config.PasswordRequireUpper && !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if p.config.PasswordRequireNumber && !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}

	if p.config.PasswordRequireSymbol && !hasSymbol {
		return fmt.Errorf("password must contain at least one symbol")
	}

	// Check for common weak patterns
	if err := p.checkWeakPatterns(password); err != nil {
		return err
	}

	return nil
}

// checkWeakPatterns checks for common weak password patterns
func (p *passwordService) checkWeakPatterns(password string) error {
	lower := strings.ToLower(password)

	// Common weak patterns
	weakPatterns := []string{
		"password", "123456", "qwerty", "admin", "root", "user",
		"test", "guest", "demo", "sample", "default", "login",
		"passw0rd", "p@ssw0rd", "123456789", "1234567890",
	}

	for _, weak := range weakPatterns {
		if strings.Contains(lower, weak) {
			return fmt.Errorf("password contains common weak pattern")
		}
	}

	// Check for repeated characters (more than 3 in a row)
	repeatedChar := regexp.MustCompile(`(.)\1{3,}`)
	if repeatedChar.MatchString(password) {
		return fmt.Errorf("password contains too many repeated characters")
	}

	// Check for sequential characters
	if p.hasSequentialChars(password, 4) {
		return fmt.Errorf("password contains sequential characters")
	}

	// Check for keyboard patterns
	keyboardPatterns := []string{
		"qwerty", "asdf", "zxcv", "1234", "abcd",
		"qwertyuiop", "asdfghjkl", "zxcvbnm",
	}

	for _, pattern := range keyboardPatterns {
		if strings.Contains(lower, pattern) {
			return fmt.Errorf("password contains keyboard pattern")
		}
	}

	return nil
}

// hasSequentialChars checks for sequential characters
func (p *passwordService) hasSequentialChars(password string, maxLength int) bool {
	if len(password) < maxLength {
		return false
	}

	for i := 0; i <= len(password)-maxLength; i++ {
		isSequential := true
		for j := 1; j < maxLength; j++ {
			if password[i+j] != password[i]+byte(j) && password[i+j] != password[i]-byte(j) {
				isSequential = false
				break
			}
		}
		if isSequential {
			return true
		}
	}

	return false
}

// secureRandInt generates a cryptographically secure random integer
func (p *passwordService) secureRandInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Fallback to a simpler method if crypto/rand fails
		// This should never happen in practice
		return 0
	}
	return int(n.Int64())
}

// EstimatePasswordStrength estimates password strength (0-100)
func (p *passwordService) EstimatePasswordStrength(password string) int {
	if password == "" {
		return 0
	}

	score := 0
	length := len(password)

	// Length scoring
	if length >= 8 {
		score += 25
	}
	if length >= 12 {
		score += 25
	}
	if length >= 16 {
		score += 10
	}

	// Character variety scoring
	var hasLower, hasUpper, hasNumber, hasSymbol bool
	uniqueChars := make(map[rune]bool)

	for _, char := range password {
		uniqueChars[char] = true
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	// Bonus for character types
	if hasLower {
		score += 5
	}
	if hasUpper {
		score += 5
	}
	if hasNumber {
		score += 5
	}
	if hasSymbol {
		score += 10
	}

	// Bonus for character diversity
	diversity := float64(len(uniqueChars)) / float64(length)
	score += int(diversity * 15)

	// Penalty for common patterns
	if p.checkWeakPatterns(password) != nil {
		score -= 20
	}

	// Ensure score is within bounds
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// IsPasswordCompromised checks if password appears in common breach lists
// Note: This is a simplified implementation. In production, you might want to
// check against services like HaveIBeenPwned API or maintain a local database
func (p *passwordService) IsPasswordCompromised(password string) bool {
	// List of most commonly compromised passwords
	commonPasswords := []string{
		"123456", "password", "123456789", "12345678", "12345",
		"111111", "1234567", "sunshine", "qwerty", "iloveyou",
		"princess", "admin", "welcome", "666666", "abc123",
		"football", "123123", "monkey", "654321", "!@#$%^&*",
		"charlie", "aa123456", "donald", "password1", "qwerty123",
	}

	lower := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lower == common {
			return true
		}
	}

	return false
}
