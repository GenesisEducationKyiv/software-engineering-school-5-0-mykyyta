package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestJWTService_GenerateAndParse_Valid(t *testing.T) {
	svc := NewJWT("secret123")

	token, err := svc.Generate("user@example.com")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	email, err := svc.Parse(token)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", email)
}

func TestJWTService_Generate_MissingSecret(t *testing.T) {
	svc := NewJWT("")

	token, err := svc.Generate("user@example.com")
	require.Error(t, err)
	assert.Empty(t, token)
}

func TestJWTService_Parse_InvalidToken(t *testing.T) {
	svc := NewJWT("secret123")

	_, err := svc.Parse("not-a-real-token")
	assert.Error(t, err)
}

func TestJWTService_Parse_TamperedToken(t *testing.T) {
	svc := NewJWT("secret123")
	tamperedSvc := NewJWT("othersecret")

	token, err := tamperedSvc.Generate("user@example.com")
	require.NoError(t, err)

	_, err = svc.Parse(token)
	assert.Error(t, err)
}

func TestJWTService_Parse_ExpiredToken(t *testing.T) {
	svc := NewJWT("secret123")

	claims := jwt.MapClaims{
		"email": "user@example.com",
		"exp":   time.Now().Add(-1 * time.Hour).Unix(), // expired
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(svc.secret))
	require.NoError(t, err)

	_, err = svc.Parse(tokenStr)
	assert.Error(t, err)
}

func TestJWTService_Parse_MissingEmailClaim(t *testing.T) {
	svc := NewJWT("secret123")

	claims := jwt.MapClaims{
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(svc.secret))
	require.NoError(t, err)

	email, err := svc.Parse(tokenStr)
	assert.Error(t, err)
	assert.Empty(t, email)
}

func TestJWTService_Parse_EmailIsNotString(t *testing.T) {
	svc := NewJWT("secret123")

	claims := jwt.MapClaims{
		"email": 12345,
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(svc.secret))
	require.NoError(t, err)

	email, err := svc.Parse(tokenStr)
	assert.Error(t, err)
	assert.Empty(t, email)
}
