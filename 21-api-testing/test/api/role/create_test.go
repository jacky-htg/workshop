//go:build integration

package role_test

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

type RoleCreateScenario struct {
	Name                 string
	Request              dto.RoleRequest
	ExpectHttpStatusCode int
	ExpectBusinessCode   string
	ExpectMessage        string
	ValidateResponse     func(t *testing.T, data json.RawMessage)
}

var roleCreateScenarios = []RoleCreateScenario{
	{
		Name:                 "success",
		Request:              dto.RoleRequest{Name: "kasir"},
		ExpectHttpStatusCode: http.StatusCreated,
		ExpectBusinessCode:   "B1",
		ExpectMessage:        "Created",
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var resp dto.RoleResponse
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)

			assert.Equal(t, "kasir", resp.Name)
			assert.NotEmpty(t, resp.ID)
		},
	},
	{
		Name:                 "internal error - duplicate",
		Request:              dto.RoleRequest{Name: "kasir"},
		ExpectHttpStatusCode: http.StatusInternalServerError,
		ExpectBusinessCode:   errors.InternalServerErrorCode,
		ExpectMessage:        "error creating role",
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			assert.Equal(t, "{}", string(data))
		},
	},
	{
		Name:                 "invalid input - required",
		Request:              dto.RoleRequest{Name: ""},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        errors.InvalidInputMessage,
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var resp map[string]string
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			assert.Contains(t, resp["name"], "Name is required")
		},
	},
	{
		Name:                 "invalid input - too short",
		Request:              dto.RoleRequest{Name: "ka"},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        errors.InvalidInputMessage,
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var resp map[string]string
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			assert.Contains(t, resp["name"], "Name is too short")
		},
	},
	{
		Name:                 "invalid input - too long",
		Request:              dto.RoleRequest{Name: "kasir melebihi 25 karakter"},
		ExpectHttpStatusCode: http.StatusBadRequest,
		ExpectBusinessCode:   errors.InvalidInputCode,
		ExpectMessage:        errors.InvalidInputMessage,
		ValidateResponse: func(t *testing.T, data json.RawMessage) {
			var resp map[string]string
			err := json.Unmarshal(data, &resp)
			require.NoError(t, err)
			assert.Contains(t, resp["name"], "Name is too long")
		},
	},
}

func TestRole_Create(t *testing.T) {
	token := setup.GetToken(t, "admin@example.com", "1234")
	for _, sc := range roleCreateScenarios {
		t.Run(sc.Name, func(t *testing.T) {
			w := setup.CallAPI(t, "POST", "/roles", token, sc.Request)
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
