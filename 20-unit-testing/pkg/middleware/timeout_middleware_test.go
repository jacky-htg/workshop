package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"workshop/mock/mockpkg"
	"workshop/pkg/middleware"

	"github.com/stretchr/testify/assert"
)

func TestTimeout_NormalRequest(t *testing.T) {
	log := mockpkg.NewMockLogger()

	handler := middleware.Timeout(log, 1*time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTimeout_RequestExceedsTimeout(t *testing.T) {
	log := mockpkg.NewMockLogger()

	handler := middleware.Timeout(log, 50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond) // Simulate slow processing
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should timeout and return Gateway Timeout
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
}

func TestTimeout_RequestFinishesBeforeTimeout(t *testing.T) {
	log := mockpkg.NewMockLogger()

	handler := middleware.Timeout(log, 100*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond) // Fast enough
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	handler.ServeHTTP(w, req)
	elapsed := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, elapsed, 100*time.Millisecond)
}
