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

func TestRoleHandler_List_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		ListFunc: func(ctx context.Context) ([]model.Role, *errors.BusinessError) {
			return []model.Role{
				{ID: 1, Name: "kasir"},
				{ID: 2, Name: "finance"},
			}, nil
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.([]any)
	assert.True(t, ok)
	assert.Equal(t, 2, len(data))

	role0, ok := data[0].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "kasir", role0["name"])
}

func TestRoleHandler_List_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		ListFunc: func(ctx context.Context) ([]model.Role, *errors.BusinessError) {
			return nil, errors.NotFound()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.List(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E002", resp.Status)
	assert.Equal(t, "Resource not found", resp.Message)
}

func TestRoleHandler_Create_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		CreateFunc: func(ctx context.Context, role *model.Role) *errors.BusinessError {
			role.ID = 101
			return nil
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request body
	reqBody := dto.RoleRequest{
		Name: "Super Admin",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Created", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, float64(101), data["id"])
	assert.Equal(t, "Super Admin", data["name"])
}

func TestRoleHandler_Create_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		CreateFunc: func(ctx context.Context, role *model.Role) *errors.BusinessError {
			return errors.InternalServerError()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request body
	reqBody := dto.RoleRequest{
		Name: "Super Admin",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E000", resp.Status)
	assert.Equal(t, "Internal Server Error", resp.Message)
}

func TestRoleHandler_Create_ValidateError(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		CreateFunc: func(ctx context.Context, role *model.Role) *errors.BusinessError {
			return errors.InternalServerError()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	// Create request body
	reqBody := dto.RoleRequest{
		Name: "",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E001", resp.Status)
	assert.Equal(t, "Invalid input", resp.Message)
}

func TestRoleHandler_Create_InvalidJsonError(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		CreateFunc: func(ctx context.Context, role *model.Role) *errors.BusinessError {
			return errors.InternalServerError()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)
	invalidJSON := []byte(`{"name": "Admin"`) // invalid closing bracket

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBuffer(invalidJSON))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E001", resp.Status)
	assert.Equal(t, "Invalid input", resp.Message)
}

func TestRoleHandler_FindByID_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, *errors.BusinessError) {
			return &model.Role{ID: 1, Name: "admin"}, nil
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodGet, "/roles/1", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.FindByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), data["id"])
	assert.Equal(t, "admin", data["name"])
}

func TestRoleHandler_FindByID_NotFound(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, *errors.BusinessError) {
			return nil, errors.NotFound()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodGet, "/roles/1", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.FindByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.NotFoundCode, resp.Status)
	assert.Equal(t, errors.NotFoundMessage, resp.Message)
}

func TestRoleHandler_FindByID_InvalidID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodGet, "/roles/1", nil)
	req.SetPathValue("id", "satu")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.FindByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Invalid id", resp.Message)
}

func TestRoleHandler_FindByID_MissingID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodGet, "/roles/1", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.FindByID(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Missing id parameter", resp.Message)
}

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

func TestRoleHandler_Delete_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		DeleteFunc: func(ctx context.Context, id int) *errors.BusinessError {
			return nil
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Delete_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		DeleteFunc: func(ctx context.Context, id int) *errors.BusinessError {
			return errors.InternalServerError()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InternalServerErrorCode, resp.Status)
	assert.Equal(t, errors.InternalServerErrorMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Delete_InvalidID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1", nil)
	req.SetPathValue("id", "satu")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Invalid id", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Delete_MissingID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Delete(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Missing id parameter", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Grant_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		GrantFunc: func(ctx context.Context, roleID, accessID int) *errors.BusinessError {
			return nil
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodPost, "/roles/1/access/12", nil)
	req.SetPathValue("id", "1")
	req.SetPathValue("access_id", "12")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Grant(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Grant_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		GrantFunc: func(ctx context.Context, roleID, accessID int) *errors.BusinessError {
			return errors.InternalServerError()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodPost, "/roles/1/access/12", nil)
	req.SetPathValue("id", "1")
	req.SetPathValue("access_id", "12")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Grant(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InternalServerErrorCode, resp.Status)
	assert.Equal(t, errors.InternalServerErrorMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Grant_InvalidAccessID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodPost, "/roles/1/access/12", nil)
	req.SetPathValue("id", "1")
	req.SetPathValue("access_id", "dua-belas")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Grant(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Invalid access_id", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Grant_InvalidID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodPost, "/roles/1/access/12", nil)
	req.SetPathValue("id", "satu")
	req.SetPathValue("access_id", "dua-belas")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Grant(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Invalid id", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Grant_MissingAccessID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodPost, "/roles/1/access/12", nil)
	req.SetPathValue("id", "satu")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Grant(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Missing access_id parameter", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Grant_MissingID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodPost, "/roles/1/access/12", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Grant(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Missing id parameter", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Revoke_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		RevokeFunc: func(ctx context.Context, roleID, accessID int) *errors.BusinessError {
			return nil
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1/access/12", nil)
	req.SetPathValue("id", "1")
	req.SetPathValue("access_id", "12")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Revoke(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Revoke_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()

	mockService := &mocksvc.MockRoles{
		RevokeFunc: func(ctx context.Context, roleID, accessID int) *errors.BusinessError {
			return errors.InternalServerError()
		},
	}

	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1/access/12", nil)
	req.SetPathValue("id", "1")
	req.SetPathValue("access_id", "12")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Revoke(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InternalServerErrorCode, resp.Status)
	assert.Equal(t, errors.InternalServerErrorMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Revoke_InvalidAccessID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1/access/12", nil)
	req.SetPathValue("id", "1")
	req.SetPathValue("access_id", "dua-belas")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Revoke(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Invalid access_id", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Revoke_InvalidID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1/access/12", nil)
	req.SetPathValue("id", "satu")
	req.SetPathValue("access_id", "dua-belas")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Revoke(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Invalid id", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Revoke_MissingAccessID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1/access/12", nil)
	req.SetPathValue("id", "satu")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Revoke(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Missing access_id parameter", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestRoleHandler_Revoke_MissingID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockRoles{}
	handler := handler.NewRoleHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodDelete, "/roles/1/access/12", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler.Revoke(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, "Missing id parameter", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}
