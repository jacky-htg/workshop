package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"workshop/internal/dto"

	"github.com/stretchr/testify/require"
)

func CreateRole(t *testing.T, token, name string) *dto.RoleResponse {
	return DoRequest(t, RequestConfig[dto.RoleResponse]{
		Method:         "POST",
		Path:           "/roles",
		Token:          token,
		Body:           dto.RoleRequest{Name: name},
		ExpectedStatus: http.StatusCreated,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Created",
		Validate: func(t *testing.T, data json.RawMessage) *dto.RoleResponse {
			var resp dto.RoleResponse
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			require.NotZero(t, resp.ID)
			require.Equal(t, name, resp.Name)
			return &resp
		},
	})
}

func GetRole(t *testing.T, token string, roleID int) *dto.RoleResponse {
	return DoRequest(t, RequestConfig[dto.RoleResponse]{
		Method:         "GET",
		Path:           fmt.Sprintf("/roles/%d", roleID),
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate: func(t *testing.T, data json.RawMessage) *dto.RoleResponse {
			var resp dto.RoleResponse
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			require.Equal(t, roleID, resp.ID)
			return &resp
		},
	})
}

func ListRoles(t *testing.T, token string) []dto.RoleResponse {
	var roles []dto.RoleResponse
	DoRequest(t, RequestConfig[any]{
		Method:         "GET",
		Path:           "/roles",
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate: func(t *testing.T, data json.RawMessage) *any {
			err := json.Unmarshal(data, &roles)
			require.NoError(t, err)
			return nil
		},
	})
	return roles
}

func UpdateRole(t *testing.T, token string, roleID int, newName string) *dto.RoleResponse {
	return DoRequest(t, RequestConfig[dto.RoleResponse]{
		Method:         "PUT",
		Path:           fmt.Sprintf("/roles/%d", roleID),
		Token:          token,
		Body:           dto.RoleRequest{Name: newName},
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate: func(t *testing.T, data json.RawMessage) *dto.RoleResponse {
			var resp dto.RoleResponse
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			require.Equal(t, roleID, resp.ID)
			require.Equal(t, newName, resp.Name)
			return &resp
		},
	})
}

func DeleteRole(t *testing.T, token string, roleID int) {
	DoRequest(t, RequestConfig[any]{
		Method:         "DELETE",
		Path:           fmt.Sprintf("/roles/%d", roleID),
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate:       nil,
	})
}

func RoleExists(t *testing.T, token, name string) bool {
	roles := ListRoles(t, token)
	for _, r := range roles {
		if r.Name == name {
			return true
		}
	}
	return false
}
