package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"workshop/internal/dto"
	"workshop/internal/handler"
	"workshop/internal/model"
	"workshop/mock/mockpkg"
	"workshop/mock/mocksvc"
	"workshop/pkg/errors"
	"workshop/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Login_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockAuths{
		LoginFunc: func(ctx context.Context, email, password string) (string, *model.User, []string, *errors.BusinessError) {
			return "my-token", &model.User{ID: "user-id"}, []string{"root"}, nil
		},
	}

	// Create request body
	reqBody := dto.LoginRequest{
		Username: "admin@example.com",
		Password: "secret",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewAuthHandler(log, validate, mockService)
	handler.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "my-token", data["token"])

	user, ok := data["user"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "user-id", user["id"])

	permissions, ok := data["permissions"].([]any)
	assert.True(t, ok)
	assert.Equal(t, 1, len(permissions))
	assert.Equal(t, "root", permissions[0])
}

func TestAuthHandler_Login_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockAuths{
		LoginFunc: func(ctx context.Context, email, password string) (string, *model.User, []string, *errors.BusinessError) {
			return "", nil, []string{}, errors.Unauthorized()
		},
	}

	reqBody := dto.LoginRequest{
		Username: "admin@example.com",
		Password: "secret",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewAuthHandler(log, validate, mockService)
	handler.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.UnauthorizedCode, resp.Status)
	assert.Equal(t, errors.UnauthorizedMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockAuths{}

	reqBody := dto.LoginRequest{
		Username: "admin",
		Password: "secret",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewAuthHandler(log, validate, mockService)
	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, errors.InvalidInputMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, "Username must be a valid email address", data["username"])
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockAuths{}

	invalidJSON := []byte(`{"name": "Admin"`)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(invalidJSON))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewAuthHandler(log, validate, mockService)
	handler.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, errors.InvalidInputMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}
