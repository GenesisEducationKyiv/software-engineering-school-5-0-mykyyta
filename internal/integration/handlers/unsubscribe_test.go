//go:build integration

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"weatherApi/internal/app"
	"weatherApi/internal/config"
	"weatherApi/internal/handlers"
	"weatherApi/internal/integration/testutils"
	"weatherApi/internal/subscription"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnsubscribeHandler_ValidToken_UnsubscribesUserSuccessfully(t *testing.T) {
	ctx := context.Background()

	pg, err := testutils.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	token := "mock-token-test@example.com"
	email := "test@example.com"

	emailProvider := &testutils.FakeEmailProvider{}
	tokenProvider := &testutils.FakeTokenProvider{}
	weatherProvider := &testutils.FakeWeatherProvider{Valid: true}

	providers := app.ProviderSet{
		EmailProvider:   emailProvider,
		TokenProvider:   tokenProvider,
		WeatherProvider: weatherProvider,
	}
	services := app.BuildServices(pg.DB, &config.Config{BaseURL: "http://localhost:8080"}, &providers)

	err = subscription.NewSubscriptionRepository(pg.DB.Gorm).Create(ctx, &subscription.Subscription{
		ID:             uuid.NewString(),
		Email:          email,
		City:           "Kyiv",
		Frequency:      "daily",
		Token:          token,
		IsConfirmed:    true,
		IsUnsubscribed: false,
	})
	require.NoError(t, err)

	handler := handlers.NewUnsubscribeHandler(services.SubService)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/unsubscribe/:token", handler.Handle)

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, "GET", "/api/unsubscribe/"+token, nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "Unsubscribed successfully")

	sub, err := subscription.NewSubscriptionRepository(pg.DB.Gorm).GetByEmail(ctx, email)
	require.NoError(t, err)
	require.True(t, sub.IsUnsubscribed)
}
