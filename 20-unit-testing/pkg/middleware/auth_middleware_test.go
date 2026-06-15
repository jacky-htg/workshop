package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"workshop/mock/mockpkg"
	"workshop/mock/mockrepo"
	"workshop/pkg/app"
	"workshop/pkg/errors"
	"workshop/pkg/middleware"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jacky-htg/go-libs/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_MissingAuthHeader(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{}

	handler := middleware.Auth(db, log, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Check response contains unauthorized error
	assert.Contains(t, w.Body.String(), errors.UnauthorizedCode)
}

func TestAuth_InvalidAuthHeaderFormat(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{}
	handler := middleware.Auth(db, log, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{}
	handler := middleware.Auth(db, log, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_NoPermission(t *testing.T) {
	myToken, err := token.ClaimToken(map[string]any{
		"email": "admin@example.com",
		"id":    "user-123",
	}, 5)
	require.NoError(t, err)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		HasPermissionFunc: func(ctx context.Context, email, routePath, routeGroup string) bool {
			return false
		},
	}
	handler := middleware.Auth(db, log, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+myToken)
	ctx := req.Context()
	ctx = context.WithValue(ctx, app.MyCtx("route-path"), "/test")
	ctx = context.WithValue(ctx, app.MyCtx("route-group"), "test")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuth_HasPermission(t *testing.T) {
	myToken, err := token.ClaimToken(map[string]any{
		"email": "admin@example.com",
		"id":    "user-123",
	}, 5)
	require.NoError(t, err)

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		HasPermissionFunc: func(ctx context.Context, email, routePath, routeGroup string) bool {
			return true
		},
	}
	handler := middleware.Auth(db, log, repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+myToken)
	ctx := req.Context()
	ctx = context.WithValue(ctx, app.MyCtx("route-path"), "/test")
	ctx = context.WithValue(ctx, app.MyCtx("route-group"), "test")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
