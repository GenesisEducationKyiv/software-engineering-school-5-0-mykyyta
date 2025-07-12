package subscription

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"subscription/internal/domain"
	"subscription/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mock Service Implementation ---

type mockSubscribeService struct {
	subscribeFunc func(ctx context.Context, email, city string, frequency domain.Frequency) error
}

func (m *mockSubscribeService) Subscribe(ctx context.Context, email, city string, frequency domain.Frequency) error {
	return m.subscribeFunc(ctx, email, city, frequency)
}

// --- Test Setup ---

func setupTestRouter(handler Subscribe) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/subscribe", handler.Handle)
	return r
}

// --- Test Cases ---

func TestSubscribeHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service := &mockSubscribeService{
			subscribeFunc: func(ctx context.Context, email, city string, frequency domain.Frequency) error {
				assert.Equal(t, domain.FreqDaily, frequency)
				return nil
			},
		}
		handler := NewSubscribe(service)
		router := setupTestRouter(handler)

		form := url.Values{}
		form.Add("email", "test@example.com")
		form.Add("city", "Kyiv")
		form.Add("frequency", "daily")

		req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"message":"Subscription successful. Confirmation email sent."}`, w.Body.String())
	})

	t.Run("MissingEmail", func(t *testing.T) {
		service := &mockSubscribeService{}
		handler := NewSubscribe(service)
		router := setupTestRouter(handler)

		form := url.Values{}
		form.Add("city", "Kyiv")
		form.Add("frequency", "daily")

		req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
	})

	t.Run("InvalidFrequency", func(t *testing.T) {
		service := &mockSubscribeService{}
		handler := NewSubscribe(service)
		router := setupTestRouter(handler)

		form := url.Values{}
		form.Add("email", "test@example.com")
		form.Add("city", "Kyiv")
		form.Add("frequency", "weekly") // invalid

		req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
	})

	t.Run("DuplicateEmail", func(t *testing.T) {
		service := &mockSubscribeService{
			subscribeFunc: func(ctx context.Context, email, city string, frequency domain.Frequency) error {
				return service.ErrEmailAlreadyExists
			},
		}
		handler := NewSubscribe(service)
		router := setupTestRouter(handler)

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
	})

	t.Run("CityNotFound", func(t *testing.T) {
		service := &mockSubscribeService{
			subscribeFunc: func(ctx context.Context, email, city string, frequency domain.Frequency) error {
				return service.ErrCityNotFound
			},
		}
		handler := NewSubscribe(service)
		router := setupTestRouter(handler)

		form := url.Values{}
		form.Add("email", "test@example.com")
		form.Add("city", "Atlantis")
		form.Add("frequency", "daily")

		req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "City not found")
	})

	t.Run("InternalError", func(t *testing.T) {
		service := &mockSubscribeService{
			subscribeFunc: func(ctx context.Context, email, city string, frequency domain.Frequency) error {
				return errors.New("unexpected error")
			},
		}
		handler := NewSubscribe(service)
		router := setupTestRouter(handler)

		form := url.Values{}
		form.Add("email", "test@example.com")
		form.Add("city", "Kyiv")
		form.Add("frequency", "daily")

		req := httptest.NewRequest(http.MethodPost, "/api/subscribe", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Something went wrong")
	})
}
