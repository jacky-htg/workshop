package helper

import (
	"encoding/json"
	"net/http"
	"testing"
	"workshop/internal/dto"
	"workshop/test/setup"

	"github.com/stretchr/testify/require"
)

func Login(t *testing.T, email, password string) string {
	w := setup.CallAPI(t, "POST", "/login", "", dto.LoginRequest{Username: email, Password: password})
	defer w.Body.Close()

	require.Equal(t, http.StatusOK, w.StatusCode)

	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data.Token)

	return resp.Data.Token
}

func LoginExpectError(t *testing.T, email, password string, expectedStatus int, expectedCode, expectedMsg string) {
	DoRequest(t, RequestConfig[any]{
		Method:         "POST",
		Path:           "/login",
		Token:          "",
		Body:           dto.LoginRequest{Username: email, Password: password},
		ExpectedStatus: expectedStatus,
		ExpectedCode:   expectedCode,
		ExpectedMsg:    expectedMsg,
		Validate: func(t *testing.T, data json.RawMessage) *any {
			require.Equal(t, "{}", string(data))
			return nil
		},
	})
}
