//go:build integration

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	di2 "monolith/internal/app/di"
	"monolith/internal/config"
	"monolith/internal/domain"
	"monolith/internal/subscription/repo"

	testutils2 "weatherApi/test/integration/testutils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestConfirmHandler_ValidToken_ConfirmsSubscriptionSuccessfully(t *testing.T) {
	ctx := context.Background()

	pg, err := testutils2.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	emailProvider := &testutils2.FakeEmailProvider{}
	tokenProvider := &testutils2.FakeTokenProvider{}
	weatherProvider := &testutils2.FakeWeatherProvider{Valid: true}

	providers := di2.Providers{
		EmailProvider:        emailProvider,
		TokenProvider:        tokenProvider,
		WeatherChainProvider: weatherProvider,
	}
	services := di2.BuildServices(pg.DB, &config.Config{BaseURL: "http://localhost:8080"}, providers)

	err = repo.NewRepo(pg.DB.Gorm).Create(ctx, &domain.Subscription{
		ID:             uuid.NewString(),
		Email:          "test@example.com",
		City:           "Kyiv",
		Frequency:      "daily",
		Token:          "mock-token-test@example.com",
		IsConfirmed:    false,
		IsUnsubscribed: false,
	})
	require.NoError(t, err)

	handler := subscription.NewConfirm(services.SubService)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/confirm/:token", handler.Handle)

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, "GET", "/api/confirm/mock-token-test@example.com", nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "Subscription confirmed successfully")

	sub, err := repo.NewRepo(pg.DB.Gorm).GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	require.True(t, sub.IsConfirmed)
}
