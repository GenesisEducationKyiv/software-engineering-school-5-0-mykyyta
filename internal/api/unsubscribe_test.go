package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"weatherApi/internal/subscription"

	"weatherApi/internal/jwtutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnsubscribeHandler_Success verifies that a valid unsubscribe request
// properly updates the subscription status and returns success.
func TestUnsubscribeHandler_Success(t *testing.T) {
	router := setupTestRouterWithDB(t)

	email := "user@unsubscribe.com"
	token, err := jwtutil.Generate(email)
	require.NoError(t, err)

	// Create an active subscription for testing
	err = DB.Create(&subscription.Subscription{
		ID:             uuid.NewString(),
		Email:          email,
		City:           "Kyiv",
		Frequency:      "daily",
		IsConfirmed:    true,
		IsUnsubscribed: false,
		Token:          token,
		CreatedAt:      time.Now(),
	}).Error
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"Unsubscribed successfully"}`, w.Body.String())
}

// TestUnsubscribeHandler_InvalidToken tests the handler's response
// when provided with a malformed or invalid JWT token.
func TestUnsubscribeHandler_InvalidToken(t *testing.T) {
	router := setupTestRouterWithDB(t)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/not-a-token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid token"}`, w.Body.String())
}

// TestUnsubscribeHandler_NotFound verifies that the handler returns
// appropriate error when attempting to unsubscribe with a valid token
// but no matching subscription in the database.
func TestUnsubscribeHandler_NotFound(t *testing.T) {
	router := setupTestRouterWithDB(t)

	token, err := jwtutil.Generate("ghost@nowhere.com")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"Token not found"}`, w.Body.String())
}

// TestUnsubscribeHandler_AlreadyUnsubscribed ensures that attempting to
// unsubscribe an already unsubscribed subscription returns a helpful message
// rather than an error.
func TestUnsubscribeHandler_AlreadyUnsubscribed(t *testing.T) {
	router := setupTestRouterWithDB(t)

	email := "already@unsubscribed.com"
	token, err := jwtutil.Generate(email)
	require.NoError(t, err)

	// Create a subscription that's already unsubscribed
	err = DB.Create(&subscription.Subscription{
		ID:             uuid.NewString(),
		Email:          email,
		City:           "Kyiv",
		Frequency:      "daily",
		IsConfirmed:    true,
		IsUnsubscribed: true,
		Token:          token,
		CreatedAt:      time.Now(),
	}).Error
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"You are already unsubscribed"}`, w.Body.String())
}
