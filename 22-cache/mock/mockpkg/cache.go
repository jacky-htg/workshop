package mockpkg

import (
	"context"
	"time"
)

type MockCache struct {
	SetFunc           func(ctx context.Context, key string, value string) error
	GetFunc           func(ctx context.Context, key string) (string, error)
	SetWithExpiryFunc func(ctx context.Context, key string, value string, expiry time.Duration) error
	DelFunc           func(ctx context.Context, keys []string) (int64, error)
	CloseFunc         func() error

	GetJSONFunc           func(ctx context.Context, key string, dest interface{}) error
	SetJSONFunc           func(ctx context.Context, key string, value interface{}) error
	SetJSONWithExpiryFunc func(ctx context.Context, key string, value interface{}, expiry time.Duration) error

	SAddFunc     func(ctx context.Context, key string, value string) (bool, error)
	SMembersFunc func(ctx context.Context, key string) ([]string, error)
	SCardFunc    func(ctx context.Context, key string) (int64, error)
	ExistsFunc   func(ctx context.Context, key string) (bool, error)
	SRemFunc     func(ctx context.Context, key string, value ...string) error
	SScanFunc    func(ctx context.Context, key string, cursor string, defaultBatchSize int) ([]string, string, error)
}

func (m *MockCache) Ping(ctx context.Context) (string, error) {
	return "pong", nil
}

func (m *MockCache) Close() error {
	return nil
}

func (m *MockCache) Set(ctx context.Context, key string, value string) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value)
	}
	return nil
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", nil
}

func (m *MockCache) SetWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	if m.SetWithExpiryFunc != nil {
		return m.SetWithExpiryFunc(ctx, key, value, expiry)
	}
	return nil
}

func (m *MockCache) Del(ctx context.Context, keys []string) (int64, error) {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, keys)
	}
	return 0, nil
}

func (m *MockCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	if m.GetJSONFunc != nil {
		return m.GetJSONFunc(ctx, key, dest)
	}
	return nil
}

func (m *MockCache) SetJSON(ctx context.Context, key string, value interface{}) error {
	if m.SetJSONFunc != nil {
		return m.SetJSONFunc(ctx, key, value)
	}
	return nil
}

func (m *MockCache) SetJSONWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	if m.SetJSONWithExpiryFunc != nil {
		return m.SetJSONWithExpiryFunc(ctx, key, value, expiry)
	}
	return nil
}

func (m *MockCache) SAdd(ctx context.Context, key string, value string) (bool, error) {
	if m.SAddFunc != nil {
		return m.SAddFunc(ctx, key, value)
	}
	return false, nil
}

func (m *MockCache) SMembers(ctx context.Context, key string) ([]string, error) {
	if m.SMembersFunc != nil {
		return m.SMembersFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockCache) SCard(ctx context.Context, key string) (int64, error) {
	if m.SCardFunc != nil {
		return m.SCardFunc(ctx, key)
	}
	return 0, nil
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, key)
	}
	return false, nil
}

func (m *MockCache) SRem(ctx context.Context, key string, value ...string) error {
	if m.SRemFunc != nil {
		return m.SRemFunc(ctx, key, value...)
	}
	return nil
}

func (m *MockCache) SScan(ctx context.Context, key string, cursor string, defaultBatchSize int) ([]string, string, error) {
	if m.SScanFunc != nil {
		return m.SScanFunc(ctx, key, cursor, defaultBatchSize)
	}
	return nil, "", nil
}
