package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// jwtService implements JWTService
type jwtService struct {
	config *AuthConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(config *AuthConfig) JWTService {
	return &jwtService{
		config: config,
	}
}

// GenerateTokenPair generates access and refresh token pair
func (j *jwtService) GenerateTokenPair(userID uuid.UUID, sessionID uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(j.config.AccessTokenExpiry)
	refreshExpiry := now.Add(j.config.RefreshTokenExpiry)

	// Generate access token
	accessToken, err := j.generateToken(userID, sessionID, "", "access", now, accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(userID, sessionID, "", "refresh", now, refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken validates an access token and returns claims
func (j *jwtService) ValidateAccessToken(token string) (*TokenClaims, error) {
	return j.validateToken(token, "access")
}

// ValidateRefreshToken validates a refresh token and returns claims
func (j *jwtService) ValidateRefreshToken(token string) (*TokenClaims, error) {
	return j.validateToken(token, "refresh")
}

// GetTokenHash returns SHA256 hash of the token for database storage
func (j *jwtService) GetTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// ExtractClaims extracts claims from token without validation (for debugging)
func (j *jwtService) ExtractClaims(token string) (*TokenClaims, error) {
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(token, &jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return j.mapClaimsToTokenClaims(*claims)
}

// generateToken generates a JWT token with specified claims
func (j *jwtService) generateToken(userID, sessionID uuid.UUID, email, tokenType string, issuedAt, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"sid":   sessionID.String(),
		"email": email,
		"type":  tokenType,
		"iat":   issuedAt.Unix(),
		"exp":   expiresAt.Unix(),
		"aud":   j.config.TokenAudience,
		"iss":   j.config.TokenIssuer,
		"jti":   uuid.New().String(), // JWT ID for uniqueness
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.config.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// validateToken validates a JWT token and returns claims
func (j *jwtService) validateToken(tokenString, expectedType string) (*TokenClaims, error) {
	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	tokenClaims, err := j.mapClaimsToTokenClaims(claims)
	if err != nil {
		return nil, err
	}

	// Validate token type
	if tokenClaims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, tokenClaims.TokenType)
	}

	// Validate expiry
	if time.Now().After(tokenClaims.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// Validate audience and issuer
	if tokenClaims.Audience != j.config.TokenAudience {
		return nil, fmt.Errorf("invalid token audience")
	}

	if tokenClaims.Issuer != j.config.TokenIssuer {
		return nil, fmt.Errorf("invalid token issuer")
	}

	return tokenClaims, nil
}

// mapClaimsToTokenClaims converts JWT MapClaims to TokenClaims
func (j *jwtService) mapClaimsToTokenClaims(claims jwt.MapClaims) (*TokenClaims, error) {
	// Extract and validate required claims
	subStr, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing subject claim")
	}

	userID, err := uuid.Parse(subStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in subject claim: %w", err)
	}

	sidStr, ok := claims["sid"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid or missing session ID claim")
	}

	sessionID, err := uuid.Parse(sidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID claim: %w", err)
	}

	email, _ := claims["email"].(string) // Optional
	tokenType, _ := claims["type"].(string)
	audience, _ := claims["aud"].(string)
	issuer, _ := claims["iss"].(string)

	// Parse timestamps
	var issuedAt, expiresAt time.Time

	if iat, ok := claims["iat"].(float64); ok {
		issuedAt = time.Unix(int64(iat), 0)
	}

	if exp, ok := claims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(exp), 0)
	} else {
		return nil, fmt.Errorf("missing or invalid expiry claim")
	}

	return &TokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		Email:     email,
		TokenType: tokenType,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		Subject:   subStr,
		Audience:  audience,
		Issuer:    issuer,
	}, nil
}

// ValidateTokenStructure performs basic structural validation without signature verification
func (j *jwtService) ValidateTokenStructure(tokenString string) error {
	// Basic JWT structure validation (header.payload.signature)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid token structure")
	}

	// Try to decode without verification
	_, _, err := new(jwt.Parser).ParseUnverified(tokenString, &jwt.MapClaims{})
	if err != nil {
		return fmt.Errorf("invalid token format: %w", err)
	}

	return nil
}

// GetTokenExpiry extracts expiry time from token without full validation
func (j *jwtService) GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := j.ExtractClaims(tokenString)
	if err != nil {
		return time.Time{}, err
	}
	return claims.ExpiresAt, nil
}

// IsTokenExpired checks if token is expired without full validation
func (j *jwtService) IsTokenExpired(tokenString string) (bool, error) {
	expiry, err := j.GetTokenExpiry(tokenString)
	if err != nil {
		return true, err
	}
	return time.Now().After(expiry), nil
}
