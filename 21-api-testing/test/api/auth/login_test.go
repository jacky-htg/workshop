//go:build integration

package auth_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"workshop/internal/dto"
	"workshop/pkg/errors"
	"workshop/test/setup"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LoginScenario struct {
	Name                 string
	Request              dto.LoginRequest
	ExpectHttpStatusCode int
	ExpectBusinessCode   string
	ExpectMessage        string
	ValidateResponse     func(t *testing.T, data json.RawMessage)
}

var loginScenarios = []LoginScenario{
	{
		Name:                 "success login - admin",
		Request:              dto.LoginRequest{Username: "admin@example.com", Password: "1234"},
		ExpectHttpStatusCode: http.StatusOK,
		ExpectBusinessCode:   "B1",
		ExpectMessage:        "Success",
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var loginData dto.LoginResponse
			err := json.Unmarshal(data, &loginData)
			require.NoError(t, err)

			assert.NotEmpty(t, loginData.Token)
			assert.Equal(t, "019eb960-a27d-73c8-9703-b23a9f50dc83", loginData.User.ID)
			assert.Equal(t, "Admin", loginData.User.Name)
			assert.Contains(t, loginData.Accesses, "root")
		},
	},
	{
		Name:                 "invalid input - username required",
		Request:              dto.LoginRequest{Username: "", Password: "1234"},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        errors.InvalidInputMessage,
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var validateData map[string]string
			err := json.Unmarshal(data, &validateData)
			require.NoError(t, err)
			assert.Contains(t, validateData["username"], "Username is required")
		},
	},
	{
		Name:                 "invalid input - username not email",
		Request:              dto.LoginRequest{Username: "admin", Password: "1234"},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        errors.InvalidInputMessage,
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var validateData map[string]string
			err := json.Unmarshal(data, &validateData)
			require.NoError(t, err)
			assert.Contains(t, validateData["username"], "Username must be a valid email address")
		},
	},
	{
		Name:                 "invalid input - password required",
		Request:              dto.LoginRequest{Username: "admin@example.com", Password: ""},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        errors.InvalidInputMessage,
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var validateData map[string]string
			err := json.Unmarshal(data, &validateData)
			require.NoError(t, err)
			assert.Contains(t, validateData["password"], "Password is required")
		},
	},
	{
		Name:                 "wrong password",
		Request:              dto.LoginRequest{Username: "admin@example.com", Password: "4321"},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        "Invalid username/password",
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			assert.Equal(t, "{}", string(data))
		},
	},
	{
		Name:                 "user not found",
		Request:              dto.LoginRequest{Username: "notfound@example.com", Password: "1234"},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        "Invalid username/password",
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			assert.Equal(t, "{}", string(data))
		},
	},
}

func TestAuth_Login(t *testing.T) {
	for _, sc := range loginScenarios {
		t.Run(sc.Name, func(t *testing.T) {
			w := setup.CallAPI(t, "POST", "/login", "", sc.Request)
			defer w.Body.Close()

			assert.Equal(t, sc.ExpectHttpStatusCode, w.StatusCode)

			var resp struct {
				Status  string          `json:"status"`
				Message string          `json:"message"`
				Data    json.RawMessage `json:"data"`
			}
			err := json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, sc.ExpectBusinessCode, resp.Status)
			assert.Equal(t, sc.ExpectMessage, resp.Message)

			if sc.ValidateResponse != nil {
				sc.ValidateResponse(t, resp.Data)
			}
		})
	}
}
