package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"workshop/internal/dto"

	"github.com/stretchr/testify/require"
)

func CreateUser(t *testing.T, token, name, email, password string, roles []int) *dto.UserResponse {
	return DoRequest(t, RequestConfig[dto.UserResponse]{
		Method: "POST",
		Path:   "/users",
		Token:  token,
		Body: dto.UserRequest{
			Name:     name,
			Username: email,
			Email:    email,
			Password: password,
			IsActive: true,
			Roles:    roles,
		},
		ExpectedStatus: http.StatusCreated,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Created",
		Validate: func(t *testing.T, data json.RawMessage) *dto.UserResponse {
			var resp dto.UserResponse
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			require.NotEmpty(t, resp.ID)
			require.Equal(t, email, resp.Email)
			return &resp
		},
	})
}

func GetUser(t *testing.T, token, userID string) *dto.UserResponse {
	return DoRequest(t, RequestConfig[dto.UserResponse]{
		Method:         "GET",
		Path:           fmt.Sprintf("/users/%s", userID),
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate: func(t *testing.T, data json.RawMessage) *dto.UserResponse {
			var resp dto.UserResponse
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			require.Equal(t, userID, resp.ID)
			return &resp
		},
	})
}

func DeleteUser(t *testing.T, token, userID string) {
	DoRequest(t, RequestConfig[any]{
		Method:         "DELETE",
		Path:           fmt.Sprintf("/users/%s", userID),
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate:       nil,
	})
}
