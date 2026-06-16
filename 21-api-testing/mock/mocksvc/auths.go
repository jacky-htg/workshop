package mocksvc

import (
	"context"
	"workshop/internal/model"
	"workshop/pkg/errors"
)

type MockAuths struct {
	LoginFunc func(ctx context.Context, email, password string) (string, *model.User, []string, *errors.BusinessError)
}

func (m *MockAuths) Login(ctx context.Context, email, password string) (string, *model.User, []string, *errors.BusinessError) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, email, password)
	}
	return "", nil, nil, nil
}
