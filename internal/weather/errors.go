package weather

import "errors"

var ErrCityNotFound = errors.New("city not found")

var (
	ErrAPIUnavailable = errors.New("weather API unavailable")
	ErrRateLimited    = errors.New("weather API rate limited")
	ErrInvalidKey     = errors.New("invalid or missing API key")
	ErrMalformedData  = errors.New("unexpected API response")
)
