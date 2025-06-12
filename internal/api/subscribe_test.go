package api

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"weatherApi/internal/subscription"

	"weatherApi/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// mustSetEnv sets an environment variable and logs fatal error if it fails.
func mustSetEnv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		log.Fatalf("failed to set env %s: %v", key, err)
	}
}

// TestMain is the entry point for tests in this package.
func TestMain(m *testing.M) {
	mustSetEnv("SENDGRID_API_KEY", "dummy-key")
	mustSetEnv("EMAIL_FROM", "test@example.com")
	mustSetEnv("DB_URL", "dummy-db-url")
	mustSetEnv("JWT_SECRET", "dummy-jwt-secret")
	mustSetEnv("WEATHER_API_KEY", "dummy-weather-key")

	config.Reload()

	os.Exit(m.Run())
}

// setupTestRouterWithDB creates an in-memory SQLite database and initializes
// a test router with all API routes registered. It also sets up a mock city
// validator that accepts all cities.
func setupTestRouterWithDB(t *testing.T) *gin.Engine {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}

	err = db.AutoMigrate(&subscription.Subscription{})
	if err != nil {
		t.Fatalf("failed to migrate test DB: %v", err)
	}

	SetDB(db)

	cityValidator = func(ctx context.Context, city string) (bool, error) {
		return true, nil // Accept all cities in tests
	}

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	RegisterRoutes(r)
	return r
}

// - Creates subscription record in database.
func TestSubscribe_Success(t *testing.T) {
	router := setupTestRouterWithDB(t)

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("city", "Kyiv")
	form.Add("frequency", "daily")

	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expected := `{"message":"Subscription successful. Confirmation email sent."}`
	assert.JSONEq(t, expected, w.Body.String())
}

// - Contains "Invalid input" in error message.
func TestSubscribe_MissingEmail(t *testing.T) {
	router := setupTestRouterWithDB(t)

	form := url.Values{}
	form.Add("city", "Kyiv")
	form.Add("frequency", "daily")

	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid input")
}

// - Contains "Invalid input" in error message.
func TestSubscribe_InvalidFrequency(t *testing.T) {
	router := setupTestRouterWithDB(t)

	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("city", "Kyiv")
	form.Add("frequency", "weekly")

	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid input")
}

// - Does not create duplicate subscription in database.
func TestSubscribe_DuplicateEmail(t *testing.T) {
	router := setupTestRouterWithDB(t)

	err := DB.Create(&subscription.Subscription{
		ID:             uuid.New().String(),
		Email:          "duplicate@example.com",
		City:           "Kyiv",
		Frequency:      "daily",
		IsConfirmed:    true,
		IsUnsubscribed: false,
		Token:          "some-token",
		CreatedAt:      time.Now(),
	}).Error
	require.NoError(t, err)

	form := url.Values{}
	form.Add("email", "duplicate@example.com")
	form.Add("city", "Kyiv")
	form.Add("frequency", "daily")

	req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Email already subscribed")
}
