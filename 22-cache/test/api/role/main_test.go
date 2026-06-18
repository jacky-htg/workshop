//go:build integration

package role_test

import (
	"os"
	"testing"
	"workshop/test/setup"
)

func TestMain(m *testing.M) {
	if err := setup.InitServer(); err != nil {
		panic(err)
	}
	defer setup.CloseServer()
	code := m.Run()
	os.Exit(code)
}
