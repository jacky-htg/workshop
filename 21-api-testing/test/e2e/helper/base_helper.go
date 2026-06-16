// helper/api_helper.go
package helper

import (
	"encoding/json"
	"testing"
	"workshop/test/setup"

	"github.com/stretchr/testify/require"
)

type RequestConfig[T any] struct {
	Method         string
	Path           string
	Token          string
	Body           interface{}
	ExpectedStatus int
	ExpectedCode   string
	ExpectedMsg    string
	Validate       func(t *testing.T, data json.RawMessage) *T
}

func DoRequest[T any](t *testing.T, cfg RequestConfig[T]) *T {
	w := setup.CallAPI(t, cfg.Method, cfg.Path, cfg.Token, cfg.Body)
	defer w.Body.Close()

	require.Equal(t, cfg.ExpectedStatus, w.StatusCode)

	var resp struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	require.Equal(t, cfg.ExpectedCode, resp.Status)
	require.Equal(t, cfg.ExpectedMsg, resp.Message)

	if cfg.Validate != nil {
		return cfg.Validate(t, resp.Data)
	}
	return nil
}
