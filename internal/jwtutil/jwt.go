package jwtutil

import (
	"weatherApi/config"

	"github.com/golang-jwt/jwt/v5"
)

// Generate creates a JWT token with an email claim.
// The token has no expiration and must be verified manually if needed.
func Generate(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		// "exp" intentionally removed — token has no expiration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(config.C.JWTSecret))
}

// Parse validates the JWT token signature and extracts the email claim.
// It returns an error if the token is invalid, malformed, or missing the email.
func Parse(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.C.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", jwt.ErrTokenMalformed
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", jwt.ErrTokenMalformed
	}

	return email, nil
}
