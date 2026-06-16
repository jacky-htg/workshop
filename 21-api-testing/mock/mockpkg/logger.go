package mockpkg

import (
	"context"

	"github.com/jacky-htg/go-libs/logger"
)

type MockLogger struct{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Debug(ctx context.Context, msg string, args ...any) {}
func (m *MockLogger) Info(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Warn(ctx context.Context, msg string, args ...any)  {}
func (m *MockLogger) Error(ctx context.Context, msg string, args ...any) {}
func (m *MockLogger) With(args ...any) logger.Logger {
	return &MockLogger{}
}
