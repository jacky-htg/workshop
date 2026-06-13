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

func TestRoleHandler_Update_Success(t *testing.T) {
	// Setup
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	// Mock service
	mockService := &mocksvc.MockRoles{
		UpdateFunc: func(ctx context.Context, role *model.Role) *errors.BusinessError {
			// Verify the role data
			assert.Equal(t, 1, role.ID)
			assert.Equal(t, "Super Admin", role.Name)
			return nil
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request body
	reqBody := dto.RoleRequest{
		Name: "Super Admin",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create request
	req := httptest.NewRequest(http.MethodPut, "/roles/1", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute
	handler.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	// Verify response data
	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "Super Admin", data["name"])
}

func TestRoleHandler_Update_InvalidID(t *testing.T) {
	// Setup
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request with invalid ID
	reqBody := dto.RoleRequest{Name: "Admin"}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/roles/invalid", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "invalid") // Invalid ID (not a number)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	// Execute
	handler.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E001", resp.Status)
	assert.Contains(t, resp.Message, "Invalid id")
}

func TestRoleHandler_Update_MissingID(t *testing.T) {
	// Setup
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request without ID
	reqBody := dto.RoleRequest{Name: "Admin"}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/roles/", bytes.NewBuffer(jsonBody))
	// No path value set for "id"
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	// Execute
	handler.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E001", resp.Status)
	assert.Contains(t, resp.Message, "Missing id parameter")
}

func TestRoleHandler_Update_InvalidJSON(t *testing.T) {
	// Setup
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request with invalid JSON
	invalidJSON := []byte(`{"name": "Admin"`) // Missing closing brace

	req := httptest.NewRequest(http.MethodPut, "/roles/1", bytes.NewBuffer(invalidJSON))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	// Execute
	handler.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E001", resp.Status)
}

func TestRoleHandler_Update_ValidationError(t *testing.T) {
	// Setup
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request with empty name (assuming validation requires name)
	reqBody := dto.RoleRequest{
		Name: "", // Empty name
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/roles/1", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	// Execute
	handler.Update(w, req)

	// Assert - should return validation error if Name field has validation tag
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E001", resp.Status)
}

func TestRoleHandler_Update_ServiceNotFoundError(t *testing.T) {
	// Setup
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	// Mock service that returns not found error
	mockService := &mocksvc.MockRoles{
		UpdateFunc: func(ctx context.Context, role *model.Role) *errors.BusinessError {
			return errors.NotFound("role not found")
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request
	reqBody := dto.RoleRequest{Name: "Updated Name"}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/roles/999", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "999")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	// Execute
	handler.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E002", resp.Status)
	assert.Contains(t, resp.Message, "role not found")
}

func TestRoleHandler_Update_ServiceInternalServerError(t *testing.T) {
	// Setup
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	// Mock service that returns internal server error
	mockService := &mocksvc.MockRoles{
		UpdateFunc: func(ctx context.Context, role *model.Role) *errors.BusinessError {
			return errors.InternalServerError("database connection failed")
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request
	reqBody := dto.RoleRequest{Name: "Updated Name"}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/roles/1", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	// Execute
	handler.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E000", resp.Status)
	assert.Contains(t, resp.Message, "database connection failed")
}
