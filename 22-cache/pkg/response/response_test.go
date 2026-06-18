package response_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"workshop/mock/mockpkg"
	"workshop/pkg/errors"
	"workshop/pkg/response"

	"github.com/stretchr/testify/assert"
)

func TestSetOk_Simple(t *testing.T) {
	log := mockpkg.NewMockLogger()
	w := httptest.NewRecorder()
	ctx := context.Background()

	response.SetOk(ctx, log, w, map[string]string{"id": "1"})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)
}

func TestSetCreated_Simple(t *testing.T) {
	log := mockpkg.NewMockLogger()
	w := httptest.NewRecorder()
	ctx := context.Background()

	response.SetCreated(ctx, log, w, map[string]string{"id": "1"})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Created", resp.Message)
}

func TestSetError_NotFound(t *testing.T) {
	log := mockpkg.NewMockLogger()
	w := httptest.NewRecorder()
	ctx := context.Background()

	bizErr := errors.NotFound("User not found")
	response.SetError(ctx, log, w, bizErr, nil)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, errors.NotFoundCode, resp.Status)
	assert.Equal(t, "User not found", resp.Message)
}

func TestSetError_InvalidInput(t *testing.T) {
	log := mockpkg.NewMockLogger()
	w := httptest.NewRecorder()
	ctx := context.Background()

	bizErr := errors.InvalidInput("Email is required")
	response.SetError(ctx, log, w, bizErr, nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Email is required", resp.Message)
}

func TestSetError_Unauthorized(t *testing.T) {
	log := mockpkg.NewMockLogger()
	w := httptest.NewRecorder()
	ctx := context.Background()

	bizErr := errors.Unauthorized("Invalid token")
	response.SetError(ctx, log, w, bizErr, nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, errors.UnauthorizedCode, resp.Status)
}

func TestSetError_Forbidden(t *testing.T) {
	log := mockpkg.NewMockLogger()
	w := httptest.NewRecorder()
	ctx := context.Background()

	bizErr := errors.Forbidden("Access denied")
	response.SetError(ctx, log, w, bizErr, nil, "custom message")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp response.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, errors.ForbiddenCode, resp.Status)
}

func TestSetError_InternalServerError(t *testing.T) {
	log := mockpkg.NewMockLogger()
	w := httptest.NewRecorder()
	ctx := context.Background()

	bizErr := errors.InternalServerError("Database error")
	response.SetError(ctx, log, w, bizErr, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.StandardResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, errors.InternalServerErrorCode, resp.Status)
}
