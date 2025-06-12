package jwtutil

import (
	"errors"
	"os"
	"time"
	"weatherApi/config"

	"github.com/golang-jwt/jwt/v5"
)

func Generate(email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

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
