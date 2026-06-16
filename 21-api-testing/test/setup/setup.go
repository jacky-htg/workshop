package setup

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	"workshop/config"
	"workshop/internal/dto"
	"workshop/internal/repository"
	"workshop/internal/router"
	"workshop/internal/service"
	"workshop/test/containers"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
	"github.com/stretchr/testify/require"
)

var (
	testServer *httptest.Server
	once       sync.Once
	initErr    error
)

// InitServer inisialisasi server sekali untuk semua test
func InitServer() error {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		registry := containers.GetRegistry()

		pg, err := registry.StartPostgres(ctx)
		if err != nil {
			initErr = err
			return
		}

		if err := pg.RunMigrations("../../../migration"); err != nil {
			initErr = err
			return
		}

		cfg := config.Config{
			Server: config.ServerConfig{GatewayTimeout: 30 * time.Second},
			Token:  config.TokenConfig{TokenSalt: "test-secret-key", TokenExp: 1},
		}

		log := logger.InitLogger(nil)
		validate := validator.New()

		repo := repository.NewAccessRepository(pg.DB, log)
		accessSvc := service.NewAccesses(pg.DB, log, repo)
		if err := accessSvc.ScanAccess(context.Background(), "../../data/route.go"); err != nil {
			initErr = err
			return
		}

		router := router.Api(cfg, pg.DB, log, validate)
		testServer = httptest.NewServer(router)
	})
	return initErr
}

// CloseServer cleanup
func CloseServer() {
	if testServer != nil {
		testServer.Close()
	}
	registry := containers.GetRegistry()
	registry.CloseAll()
}

// GetServerURL returns test server URL
func GetServerURL() string {
	return testServer.URL
}

func CallAPI(t *testing.T, method, path, token string, body interface{}) *http.Response {
	t.Helper()

	var reqBody *bytes.Buffer = &bytes.Buffer{}
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, GetServerURL()+path, reqBody)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func GetToken(t *testing.T, username, password string) string {
	bodyReq := dto.LoginRequest{
		Username: username,
		Password: password,
	}

	w := CallAPI(t, "POST", "/login", "", bodyReq)
	defer w.Body.Close()

	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	return resp.Data.Token
}
