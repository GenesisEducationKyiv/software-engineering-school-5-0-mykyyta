package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"weatherApi/internal/subscription"

	"weatherApi/internal/jwtutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Setup: TestMain для JWT_SECRET ---.
func TestMain(m *testing.M) {
	_ = os.Setenv("JWT_SECRET", "test-secret")
	os.Exit(m.Run())
}

// mockConfirmService реалізує confirmService (використовується в ConfirmHandler).
type mockConfirmService struct {
	ConfirmFunc func(ctx context.Context, token string) error
}

func (m *mockConfirmService) Confirm(ctx context.Context, token string) error {
	return m.ConfirmFunc(ctx, token)
}

func setupRouterWithMockService(t *testing.T, mock *mockConfirmService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewConfirmHandler(mock)

	r := gin.Default()
	r.GET("/api/confirm/:token", handler.Handle)
	return r
}

func TestConfirmHandler_Success(t *testing.T) {
	token, err := jwtutil.Generate("confirmtest@example.com")
	require.NoError(t, err)

	mock := &mockConfirmService{
		ConfirmFunc: func(ctx context.Context, tkn string) error {
			assert.Equal(t, token, tkn)
			return nil
		},
	}

	router := setupRouterWithMockService(t, mock)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"Subscription confirmed successfully"}`, w.Body.String())
}

func TestConfirmHandler_InvalidToken(t *testing.T) {
	mock := &mockConfirmService{
		ConfirmFunc: func(ctx context.Context, token string) error {
			return subscription.ErrInvalidToken
		},
	}

	router := setupRouterWithMockService(t, mock)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/invalid-token", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"Invalid token"}`, w.Body.String())
}

func TestConfirmHandler_TokenButNoSubscription(t *testing.T) {
	token, err := jwtutil.Generate("ghost@example.com")
	require.NoError(t, err)

	mock := &mockConfirmService{
		ConfirmFunc: func(ctx context.Context, token string) error {
			return subscription.ErrSubscriptionNotFound
		},
	}

	router := setupRouterWithMockService(t, mock)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"Subscription not found"}`, w.Body.String())
}

func TestConfirmHandler_AlreadyConfirmed(t *testing.T) {
	token, err := jwtutil.Generate("already@confirmed.com")
	require.NoError(t, err)

	mock := &mockConfirmService{
		ConfirmFunc: func(ctx context.Context, token string) error {
			// Already confirmed — no error
			return nil
		},
	}

	router := setupRouterWithMockService(t, mock)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"Subscription confirmed successfully"}`, w.Body.String())
}
