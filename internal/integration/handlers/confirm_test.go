//go:build integration

package handlers_test

import (
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
	"golang.org/x/net/context"
)

func TestConfirmHandler_ValidToken_ConfirmsSubscriptionSuccessfully(t *testing.T) {
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

	providers := app.ProviderSet{
		EmailProvider:        emailProvider,
		TokenProvider:        tokenProvider,
		WeatherChainProvider: weatherProvider,
	}
	services := app.BuildServices(pg.DB, &config.Config{BaseURL: "http://localhost:8080"}, &providers)

	err = subscription.NewRepo(pg.DB.Gorm).Create(ctx, &subscription.Subscription{
		ID:             uuid.NewString(),
		Email:          "test@example.com",
		City:           "Kyiv",
		Frequency:      "daily",
		Token:          "mock-token-test@example.com",
		IsConfirmed:    false,
		IsUnsubscribed: false,
	})
	require.NoError(t, err)

	handler := handlers.NewConfirm(services.SubService)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/confirm/:token", handler.Handle)

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, "GET", "/api/confirm/mock-token-test@example.com", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "Subscription confirmed successfully")

	sub, err := subscription.NewRepo(pg.DB.Gorm).GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	require.True(t, sub.IsConfirmed)
}
