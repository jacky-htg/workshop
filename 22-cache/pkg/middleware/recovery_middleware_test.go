package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"workshop/mock/mockpkg"
	"workshop/pkg/middleware"

	"github.com/stretchr/testify/assert"
)

func TestRecovery_NormalRequest(t *testing.T) {
	log := mockpkg.NewMockLogger()

	// Create handler that doesn't panic
	handler := middleware.Recovery(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestRecovery_PanicRecovery(t *testing.T) {
	log := mockpkg.NewMockLogger()

	// Create handler that panics
	handler := middleware.Recovery(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected error")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// This should not panic because recovery middleware catches it
	handler.ServeHTTP(w, req)

	// Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Response body should contain error response
	assert.Contains(t, w.Body.String(), "Internal Server Error")
}

func TestRecovery_ChainedMiddleware(t *testing.T) {
	log := mockpkg.NewMockLogger()

	callOrder := []string{}

	// Create chain: Recovery -> Middleware1 -> Handler
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "middleware1")
			next.ServeHTTP(w, r)
		})
	}

	handler := middleware.Recovery(log)(middleware1(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callOrder = append(callOrder, "handler")
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"middleware1", "handler"}, callOrder)
}
