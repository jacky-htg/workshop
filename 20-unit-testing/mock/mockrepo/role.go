package mockrepo

import (
	"context"
	"workshop/internal/model"
)

type MockRoleRepo struct {
	CreateFunc   func(ctx context.Context, role *model.Role) error
	FindByIDFunc func(ctx context.Context, id int) (*model.Role, error)
	ListFunc     func(ctx context.Context) ([]model.Role, error)
	UpdateFunc   func(ctx context.Context, role *model.Role) error
	DeleteFunc   func(ctx context.Context, id int) error

	// Many-to-many dengan Access
	GrantAccessFunc        func(ctx context.Context, roleID, accessID int) error
	RevokeAccessFunc       func(ctx context.Context, roleID, accessID int) error
	GetAccessesByRolesFunc func(ctx context.Context, roleIDs []int) ([]model.Access, error)

	// Helper
	HasAccessFunc func(ctx context.Context, roleID, accessID int) (bool, error)
}

func (m *MockRoleRepo) Create(ctx context.Context, role *model.Role) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, role)
	}
	return nil
}

func (m *MockRoleRepo) FindByID(ctx context.Context, id int) (*model.Role, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRoleRepo) List(ctx context.Context) ([]model.Role, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockRoleRepo) Update(ctx context.Context, role *model.Role) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, role)
	}
	return nil
}

func (m *MockRoleRepo) Delete(ctx context.Context, id int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRoleRepo) GrantAccess(ctx context.Context, roleID, accessID int) error {
	if m.GrantAccessFunc != nil {
		return m.GrantAccessFunc(ctx, roleID, accessID)
	}
	return nil
}

func (m *MockRoleRepo) RevokeAccess(ctx context.Context, roleID, accessID int) error {
	if m.RevokeAccessFunc != nil {
		return m.RevokeAccessFunc(ctx, roleID, accessID)
	}
	return nil
}

func (m *MockRoleRepo) GetAccessesByRoles(ctx context.Context, roleIDs []int) ([]model.Access, error) {
	if m.GetAccessesByRolesFunc != nil {
		return m.GetAccessesByRolesFunc(ctx, roleIDs)
	}
	return nil, nil
}

func (m *MockRoleRepo) HasAccess(ctx context.Context, roleID, accessID int) (bool, error) {
	if m.HasAccessFunc != nil {
		return m.HasAccessFunc(ctx, roleID, accessID)
	}
	return false, nil
}
