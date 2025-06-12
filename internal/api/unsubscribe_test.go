package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"weatherApi/internal/subscription"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mock service ---.
type mockUnsubscribeService struct {
	unsubscribeFunc func(ctx context.Context, token string) error
}

func (m *mockUnsubscribeService) Unsubscribe(ctx context.Context, token string) error {
	return m.unsubscribeFunc(ctx, token)
}

// --- Setup router ---.
func setupUnsubscribeRouter(s unsubscribeService) *gin.Engine {
	handler := NewUnsubscribeHandler(s)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/unsubscribe/:token", handler.Handle)
	return r
}

// --- Tests ---

func TestUnsubscribeHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service := &mockUnsubscribeService{
			unsubscribeFunc: func(ctx context.Context, token string) error {
				return nil
			},
		}
		router := setupUnsubscribeRouter(service)

		req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/some-valid-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Unsubscribed successfully"}`, w.Body.String())
	})

	t.Run("InvalidToken", func(t *testing.T) {
		service := &mockUnsubscribeService{
			unsubscribeFunc: func(ctx context.Context, token string) error {
				return subscription.ErrInvalidToken
			},
		}
		router := setupUnsubscribeRouter(service)

		req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/not-a-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"Invalid token"}`, w.Body.String())
	})

	t.Run("NotFound", func(t *testing.T) {
		service := &mockUnsubscribeService{
			unsubscribeFunc: func(ctx context.Context, token string) error {
				return subscription.ErrSubscriptionNotFound
			},
		}
		router := setupUnsubscribeRouter(service)

		req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/ghost-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"Subscription not found"}`, w.Body.String())
	})

	t.Run("InternalError", func(t *testing.T) {
		service := &mockUnsubscribeService{
			unsubscribeFunc: func(ctx context.Context, token string) error {
				return errors.New("unexpected DB error")
			},
		}
		router := setupUnsubscribeRouter(service)

		req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/token123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Something went wrong")
	})
}
