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

func TestUserHandler_List_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		ListFunc: func(ctx context.Context, search, order, sort string, limit, page int) ([]model.User, model.Pagination, *errors.BusinessError) {
			return []model.User{
					{ID: "1"},
				},
				model.Pagination{
					Page:  1,
					Limit: 10,
					Count: 5,
				},
				nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.([]any)
	assert.True(t, ok)
	assert.Equal(t, 1, len(data))

	user0, ok := data[0].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "1", user0["id"])

	meta, ok := resp.Meta.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "name", meta["order"])
	assert.Equal(t, "asc", meta["sort"])

	pagination, ok := meta["pagination"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, float64(5), pagination["total"])
}

func TestUserHandler_List_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		ListFunc: func(ctx context.Context, search, order, sort string, limit, page int) ([]model.User, model.Pagination, *errors.BusinessError) {
			return nil, model.Pagination{}, errors.Unauthorized()
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.List(w, req)

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

func TestUserHandler_Create_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		CreateFunc: func(ctx context.Context, user *model.User) *errors.BusinessError {
			user.ID = "1"
			return nil
		},
	}

	reqBody := dto.UserRequest{
		Name:     "admin",
		Username: "admin",
		Password: "secret1234",
		Email:    "admin@example.com",
		IsActive: true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Created", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "1", data["id"])
}

func TestUserHandler_Create_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		CreateFunc: func(ctx context.Context, user *model.User) *errors.BusinessError {
			return errors.Forbidden()
		},
	}

	reqBody := dto.UserRequest{
		Name:     "admin",
		Username: "admin",
		Password: "secret1234",
		Email:    "admin@example.com",
		IsActive: true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Create(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.ForbiddenCode, resp.Status)
	assert.Equal(t, errors.ForbiddenMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestUserHandler_Create_ValidateError(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{}

	reqBody := dto.UserRequest{
		Name: "admin",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, errors.InvalidInputMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, "Username is required", data["username"])
}

func TestUserHandler_Create_InvalidJSON(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{}

	jsonBody := []byte(`{"name": "admin")`)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Create(w, req)

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

func TestUserHandler_FindByID_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, *errors.BusinessError) {
			return &model.User{ID: "1"}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.FindByID(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "1", data["id"])
}

func TestUserHandler_FindByID_NotFound(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, *errors.BusinessError) {
			return nil, errors.NotFound()
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.FindByID(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.NotFoundCode, resp.Status)
	assert.Equal(t, errors.NotFoundMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestUserHandler_FindByID_MissingID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{}

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.FindByID(w, req)

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

func TestUserHandler_Update_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		UpdateFunc: func(ctx context.Context, user *model.User) *errors.BusinessError {
			return nil
		},
	}

	reqBody := dto.UserUpdateRequest{
		Name:     "admin",
		IsActive: true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "B1", resp.Status)
	assert.Equal(t, "Success", resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "1", data["id"])
}

func TestUserHandler_Update_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		UpdateFunc: func(ctx context.Context, user *model.User) *errors.BusinessError {
			return errors.NotFound()
		},
	}

	reqBody := dto.UserUpdateRequest{
		Name:     "admin",
		IsActive: true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.NotFoundCode, resp.Status)
	assert.Equal(t, errors.NotFoundMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestUserHandler_Update_ValidateError(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{}

	reqBody := dto.UserUpdateRequest{
		Name:     "",
		IsActive: true,
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.InvalidInputCode, resp.Status)
	assert.Equal(t, errors.InvalidInputMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, data["name"], "Name is required")
}

func TestUserHandler_Update_InvalidJSON(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{}

	jsonBody := []byte(`{"name": "admin")`)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(jsonBody))
	req.SetPathValue("id", "1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Update(w, req)

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

func TestUserHandler_Update_MissingID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{}

	req := httptest.NewRequest(http.MethodPut, "/users/1", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Update(w, req)

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

func TestUserHandler_Delete_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		DeletFunc: func(ctx context.Context, id string) *errors.BusinessError {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req.SetPathValue("id", "user-1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
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

func TestUserHandler_Delete_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{
		DeletFunc: func(ctx context.Context, id string) *errors.BusinessError {
			return errors.NotFound()
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req.SetPathValue("id", "user-1")
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
	handler.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.NotFoundCode, resp.Status)
	assert.Equal(t, errors.NotFoundMessage, resp.Message)

	data, ok := resp.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, map[string]any{}, data)
}

func TestUserHandler_Delete_MissingID(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	svc := &mocksvc.MockUsers{}

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewUserHandler(log, validate, svc)
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
