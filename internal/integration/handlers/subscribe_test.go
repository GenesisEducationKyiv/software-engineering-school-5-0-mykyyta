package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"weatherApi/internal/subscription"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"weatherApi/internal/app"
	"weatherApi/internal/handlers"
	"weatherApi/internal/integration/testutils"
)

func TestSubscribeHandler_Success(t *testing.T) {
	ctx := context.Background()

	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	emailProvider := &testutils.FakeEmailProvider{}
	tokenProvider := &testutils.FakeTokenProvider{}
	weatherProvider := &testutils.FakeWeatherProvider{Valid: true}

	builder := &app.ServiceBuilder{
		DB:              pg.DB,
		BaseURL:         "http://localhost:8080",
		EmailProvider:   emailProvider,
		TokenProvider:   tokenProvider,
		WeatherProvider: weatherProvider,
	}
	services, err := builder.BuildServices()
	require.NoError(t, err)

	handler := handlers.NewSubscribeHandler(services.SubService)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/subscribe", handler.Handle)

	reqBody := map[string]string{
		"email":     "test@example.com",
		"city":      "Kyiv",
		"frequency": "daily",
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", "/api/subscribe", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "Confirmation email sent.")

	repo := subscription.NewSubscriptionRepository(pg.DB.Gorm)
	sub, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, sub)
	require.Equal(t, "Kyiv", sub.City)
	require.Equal(t, "daily", sub.Frequency)
	require.False(t, sub.IsConfirmed)
	require.False(t, sub.IsUnsubscribed)
	require.NotEmpty(t, sub.Token)
	require.NotEmpty(t, sub.ID)
	require.WithinDuration(t, time.Now(), sub.CreatedAt, time.Minute)

	require.True(t, emailProvider.Sent)
	require.Equal(t, "test@example.com", emailProvider.To)
	require.Contains(t, emailProvider.Plain, sub.Token)
	require.Contains(t, emailProvider.Plain, "http://localhost:8080/api/confirm/")
	require.Contains(t, emailProvider.HTML, sub.Token)
	require.Contains(t, emailProvider.HTML, `<a href="http://localhost:8080/api/confirm/`)
}
