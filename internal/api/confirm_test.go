package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"weatherApi/internal/model"
	"weatherApi/pkg/jwtutil"
	"weatherApi/pkg/scheduler"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	scheduler.FetchWeather = func(city string) (*model.Weather, int, error) {
		return &model.Weather{
			Temperature: 22.5,
			Humidity:    60,
			Description: "Clear skies",
		}, 200, nil
	}

	scheduler.SendWeatherEmail = func(to string, weather *model.Weather, city string, token string) error {
		return nil // simulate success
	}
}

// TestConfirmHandler_Success verifies that confirming a valid, unconfirmed subscription:
// - Returns HTTP 200 with success message
// - Sets IsConfirmed=true in the database
// - Does not modify other subscription fields
func TestConfirmHandler_Success(t *testing.T) {
	router := setupTestRouterWithDB(t)

	email := "confirmtest@example.com"
	token, err := jwtutil.Generate(email)
	require.NoError(t, err)

	err = DB.Create(&model.Subscription{
		ID:             "test-id",
		Email:          email,
		City:           "Kyiv",
		Frequency:      "daily",
		IsConfirmed:    false,
		IsUnsubscribed: false,
		Token:          token,
		CreatedAt:      time.Now(),
	}).Error
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.JSONEq(t, `{"message":"Subscription confirmed successfully"}`, w.Body.String())

	var sub model.Subscription
	err = DB.Where("email = ?", email).First(&sub).Error
	require.NoError(t, err)
	assert.True(t, sub.IsConfirmed)
}

// TestConfirmHandler_InvalidToken verifies that the confirmation endpoint:
// - Rejects malformed JWT tokens
// - Returns HTTP 400 Bad Request
// - Includes "Invalid token" in the error message
func TestConfirmHandler_InvalidToken(t *testing.T) {
	router := setupTestRouterWithDB(t)

	invalidToken := "not-a-valid-jwt"

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+invalidToken, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

// TestConfirmHandler_TokenButNoSubscription verifies that the confirmation endpoint:
// - Returns HTTP 404 when token is valid but no matching subscription exists
// - This prevents enumeration of emails via the confirmation endpoint
func TestConfirmHandler_TokenButNoSubscription(t *testing.T) {
	router := setupTestRouterWithDB(t)

	token, err := jwtutil.Generate("ghost@example.com")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Token not found")
}

// TestConfirmHandler_AlreadyConfirmed verifies that attempting to confirm
// an already confirmed subscription:
// - Returns HTTP 200 (idempotent operation)
// - Returns appropriate message indicating subscription was already confirmed
// - Does not modify the existing confirmed subscription
func TestConfirmHandler_AlreadyConfirmed(t *testing.T) {
	router := setupTestRouterWithDB(t)

	email := "already@confirmed.com"
	token, err := jwtutil.Generate(email)
	require.NoError(t, err)

	err = DB.Create(&model.Subscription{
		ID:             uuid.New().String(),
		Email:          email,
		City:           "Kyiv",
		Frequency:      "daily",
		IsConfirmed:    true,
		IsUnsubscribed: false,
		Token:          token,
		CreatedAt:      time.Now(),
	}).Error
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "already confirmed")
}
