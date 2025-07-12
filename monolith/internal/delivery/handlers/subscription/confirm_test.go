package subscription

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"weatherApi/monolith/internal/subscription"
	"weatherApi/monolith/internal/token/jwt"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("JWT_SECRET", "test-secret")
	os.Exit(m.Run())
}

// --- mockConfirmService реалізує domain-level ConfirmService ---.
type mockConfirmService struct {
	ConfirmFunc func(ctx context.Context, token string) error
}

func (m *mockConfirmService) Confirm(ctx context.Context, token string) error {
	return m.ConfirmFunc(ctx, token)
}

func setupRouterWithMock(t *testing.T, mock *mockConfirmService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	handler := NewConfirm(mock)
	r := gin.Default()
	r.GET("/api/confirm/:token", handler.Handle)
	return r
}

func TestConfirmHandler(t *testing.T) {
	jwt := jwt.NewJWT("test-secret")

	t.Run("Success", func(t *testing.T) {
		token, err := jwt.Generate("confirmtest@example.com")
		require.NoError(t, err)

		mock := &mockConfirmService{
			ConfirmFunc: func(ctx context.Context, inputToken string) error {
				assert.Equal(t, token, inputToken)
				return nil
			},
		}

		router := setupRouterWithMock(t, mock)

		req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Subscription confirmed successfully"}`, w.Body.String())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		mock := &mockConfirmService{
			ConfirmFunc: func(ctx context.Context, token string) error {
				return subscription.ErrInvalidToken
			},
		}

		router := setupRouterWithMock(t, mock)

		req := httptest.NewRequest(http.MethodGet, "/api/confirm/invalid-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"Invalid token"}`, w.Body.String())
	})

	t.Run("TokenButNoSubscription", func(t *testing.T) {
		token, err := jwt.Generate("ghost@example.com")
		require.NoError(t, err)

		mock := &mockConfirmService{
			ConfirmFunc: func(ctx context.Context, token string) error {
				return subscription.ErrSubscriptionNotFound
			},
		}

		router := setupRouterWithMock(t, mock)

		req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"Subscription not found"}`, w.Body.String())
	})

	t.Run("AlreadyConfirmed", func(t *testing.T) {
		token, err := jwt.Generate("already@confirmed.com")
		require.NoError(t, err)

		mock := &mockConfirmService{
			ConfirmFunc: func(ctx context.Context, token string) error {
				return nil // Already confirmed
			},
		}

		router := setupRouterWithMock(t, mock)

		req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Subscription confirmed successfully"}`, w.Body.String())
	})

	t.Run("InternalError", func(t *testing.T) {
		token, err := jwt.Generate("broken@token.com")
		require.NoError(t, err)

		mock := &mockConfirmService{
			ConfirmFunc: func(ctx context.Context, token string) error {
				return assert.AnError // generic error
			},
		}

		router := setupRouterWithMock(t, mock)

		req := httptest.NewRequest(http.MethodGet, "/api/confirm/"+token, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error":"Something went wrong"}`, w.Body.String())
	})
}
