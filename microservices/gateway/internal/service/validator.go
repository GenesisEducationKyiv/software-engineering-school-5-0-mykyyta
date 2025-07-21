package service

import (
	"fmt"
	"strings"
)

type securityValidator struct{}

func NewSecurityValidator() SecurityValidator {
	return &securityValidator{}
}

func (v *securityValidator) ValidateToken(token string) error {
	if len(token) < 8 || len(token) > 128 {
		return fmt.Errorf("invalid token length")
	}

	if strings.ContainsAny(token, "<>\"'&;") {
		return fmt.Errorf("invalid token format")
	}

	return nil
}

func (v *securityValidator) ValidateCity(city string) error {
	if len(city) == 0 || len(city) > 50 {
		return fmt.Errorf("invalid city length")
	}

	if strings.ContainsAny(city, "<>\"';") {
		return fmt.Errorf("invalid city format")
	}

	return nil
}

func (v *securityValidator) SanitizeInput(input string) string {
	input = strings.ReplaceAll(input, "<", "")
	input = strings.ReplaceAll(input, ">", "")
	input = strings.ReplaceAll(input, "\"", "")
	input = strings.ReplaceAll(input, "'", "")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "&", "")

	return strings.TrimSpace(input)
}
