package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestAccessHandler_List_Success(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockAccessess{
		ListFunc: func(ctx context.Context) (map[int]*model.AccessTree, *errors.BusinessError) {
			accessesID := 4
			rolesID := 5
			return map[int]*model.AccessTree{
				0: &model.AccessTree{
					ID:    4,
					Name:  "accesses",
					Alias: "accesses",
					Childrens: []model.Access{
						{ID: 6, ParentID: &accessesID, Name: "GET /accesses", Alias: "accesses::list"},
					},
				},
				1: &model.AccessTree{
					ID:    5,
					Name:  "roles",
					Alias: "roles",
					Childrens: []model.Access{
						{ID: 11, ParentID: &rolesID, Name: "DELETE /roles/{id}", Alias: "roles::delete"},
						{ID: 13, ParentID: &rolesID, Name: "DELETE /roles/{id}/access/{access_id}", Alias: "roles::revoke"},
					},
				},
			}, nil
		},
	}

	handler := handler.NewAccessHandler(log, validate, mockService)

	req := httptest.NewRequest(http.MethodGet, "/accesses", nil)
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

	access1, ok := data[1].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "roles", access1["name"])
}

func TestAccessHandler_List_Error(t *testing.T) {
	log := mockpkg.NewMockLogger()
	validate := validator.New()
	mockService := &mocksvc.MockAccessess{
		ListFunc: func(ctx context.Context) (map[int]*model.AccessTree, *errors.BusinessError) {
			return nil, errors.NotFound()
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/accesses", nil)
	req = req.WithContext(context.Background())

	w := httptest.NewRecorder()

	handler := handler.NewAccessHandler(log, validate, mockService)
	handler.List(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "E002", resp.Status)
	assert.Equal(t, "Resource not found", resp.Message)
}
