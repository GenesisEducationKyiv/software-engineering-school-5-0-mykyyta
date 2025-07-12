//go:build integration

package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	di2 "weatherApi/monolith/internal/app/di"
	"weatherApi/monolith/internal/config"
	"weatherApi/monolith/internal/domain"
	"weatherApi/monolith/internal/subscription/repo"

	testutils2 "weatherApi/test/integration/testutils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnsubscribeHandler_ValidToken_UnsubscribesUserSuccessfully(t *testing.T) {
	ctx := context.Background()

	pg, err := testutils2.StartPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	}()

	token := "mock-token-test@example.com"
	email := "test@example.com"

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
		Email:          email,
		City:           "Kyiv",
		Frequency:      "daily",
		Token:          token,
		IsConfirmed:    true,
		IsUnsubscribed: false,
	})
	require.NoError(t, err)

	handler := subscription.NewUnsubscribe(services.SubService)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/unsubscribe/:token", handler.Handle)

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, "GET", "/api/unsubscribe/"+token, nil)
	require.NoError(t, err)
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	require.Contains(t, w.Body.String(), "Unsubscribed successfully")

	sub, err := repo.NewRepo(pg.DB.Gorm).GetByEmail(ctx, email)
	require.NoError(t, err)
	require.True(t, sub.IsUnsubscribed)
}
