package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func main() {
	// Use the test user ID from the database
	userID := uuid.MustParse("e97ccfaa-2dfa-4cae-b92a-ccdf3519d556") // test@example.com

	var sessionID uuid.UUID
	if len(os.Args) > 1 {
		sessionID = uuid.MustParse(os.Args[1])
	} else {
		sessionID = uuid.New()
	}

	// Get JWT secret from environment or use default
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secret-jwt-key-change-in-production"
	}

	// Create claims
	claims := jwt.MapClaims{
		"sub":  userID.String(),
		"sid":  sessionID.String(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
		"type": "access",
		"iss":  "kainuguru-auth",
		"aud":  "kainuguru-api",
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	fmt.Fprintf(os.Stderr, "Using secret starting with: %s\n", secret[:3])
	fmt.Fprintf(os.Stderr, "Claims: %+v\n", claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(tokenString)
}
