package mocksvc

import (
	"context"
	"workshop/internal/model"
	"workshop/pkg/errors"
)

type MockAccessess struct {
	ListFunc       func(ctx context.Context) (map[int]*model.AccessTree, *errors.BusinessError)
	ScanAccessFunc func(ctx context.Context, path string) error
}

func (m *MockAccessess) List(ctx context.Context) (map[int]*model.AccessTree, *errors.BusinessError) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockAccessess) ScanAccess(ctx context.Context, path string) error {
	if m.ScanAccessFunc != nil {
		return m.ScanAccessFunc(ctx, path)
	}
	return nil
}
