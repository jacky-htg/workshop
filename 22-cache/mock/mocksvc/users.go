package mocksvc

import (
	"context"
	"workshop/internal/model"
	"workshop/pkg/errors"
)

type MockUsers struct {
	ListFunc     func(ctx context.Context, search, order, sort string, limit, page int) ([]model.User, model.Pagination, *errors.BusinessError)
	CreateFunc   func(ctx context.Context, user *model.User) *errors.BusinessError
	FindByIDFunc func(ctx context.Context, id string) (*model.User, *errors.BusinessError)
	UpdateFunc   func(ctx context.Context, user *model.User) *errors.BusinessError
	DeletFunc    func(ctx context.Context, id string) *errors.BusinessError
}

func (m *MockUsers) List(ctx context.Context, search, order, sort string, limit, page int) ([]model.User, model.Pagination, *errors.BusinessError) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, search, order, sort, limit, page)
	}
	return nil, model.Pagination{}, nil
}

func (m *MockUsers) Create(ctx context.Context, user *model.User) *errors.BusinessError {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil
}

func (m *MockUsers) FindByID(ctx context.Context, id string) (*model.User, *errors.BusinessError) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUsers) Update(ctx context.Context, user *model.User) *errors.BusinessError {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockUsers) Delete(ctx context.Context, id string) *errors.BusinessError {
	if m.DeletFunc != nil {
		return m.DeletFunc(ctx, id)
	}
	return nil
}
