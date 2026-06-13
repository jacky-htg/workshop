package mocksvc

import (
	"context"
	"workshop/internal/model"
	"workshop/pkg/errors"
)

type MockRoles struct {
	ListFunc     func(ctx context.Context) ([]model.Role, *errors.BusinessError)
	FindByIDFunc func(ctx context.Context, id int) (*model.Role, *errors.BusinessError)
	CreateFunc   func(ctx context.Context, role *model.Role) *errors.BusinessError
	UpdateFunc   func(ctx context.Context, role *model.Role) *errors.BusinessError
	DeleteFunc   func(ctx context.Context, id int) *errors.BusinessError
	GrantFunc    func(ctx context.Context, roleID, accessID int) *errors.BusinessError
	RevokeFunc   func(ctx context.Context, roleID, accessID int) *errors.BusinessError
}

func (m *MockRoles) List(ctx context.Context) ([]model.Role, *errors.BusinessError) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockRoles) FindByID(ctx context.Context, id int) (*model.Role, *errors.BusinessError) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRoles) Create(ctx context.Context, role *model.Role) *errors.BusinessError {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, role)
	}
	return nil
}

func (m *MockRoles) Update(ctx context.Context, role *model.Role) *errors.BusinessError {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, role)
	}
	return nil
}

func (m *MockRoles) Delete(ctx context.Context, id int) *errors.BusinessError {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRoles) Grant(ctx context.Context, roleID, accessID int) *errors.BusinessError {
	if m.GrantFunc != nil {
		return m.GrantFunc(ctx, roleID, accessID)
	}
	return nil
}

func (m *MockRoles) Revoke(ctx context.Context, roleID, accessID int) *errors.BusinessError {
	if m.RevokeFunc != nil {
		return m.RevokeFunc(ctx, roleID, accessID)
	}
	return nil
}
