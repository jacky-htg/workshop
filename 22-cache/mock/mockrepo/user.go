package mockrepo

import (
	"context"
	"database/sql"
	"workshop/internal/model"
)

type MockUserRepo struct {
	ListFunc          func(ctx context.Context, search, order, sort string, limit, offset int) ([]model.User, int, error)
	CreateFunc        func(ctx context.Context, tx *sql.Tx, user *model.User) error
	FindByIDFunc      func(ctx context.Context, id string) (*model.User, error)
	FindByEmailFunc   func(ctx context.Context, email string) (*model.User, error)
	UpdateFunc        func(ctx context.Context, tx *sql.Tx, user *model.User) error
	DeleteFunc        func(ctx context.Context, id string) error
	AssignRoleFunc    func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error
	RemoveRoleFunc    func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error
	HasPermissionFunc func(ctx context.Context, email, routePath, routeGroup string) bool
}

func (m *MockUserRepo) List(ctx context.Context, search, order, sort string, limit, offset int) ([]model.User, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, search, order, sort, limit, offset)
	}
	return nil, 0, nil
}

func (m *MockUserRepo) Create(ctx context.Context, tx *sql.Tx, user *model.User) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, tx, user)
	}
	return nil
}

func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.FindByEmailFunc != nil {
		return m.FindByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepo) Update(ctx context.Context, tx *sql.Tx, user *model.User) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, tx, user)
	}
	return nil
}

func (m *MockUserRepo) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUserRepo) AssignRole(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
	if m.AssignRoleFunc != nil {
		return m.AssignRoleFunc(ctx, tx, userID, roleID)
	}
	return nil
}

func (m *MockUserRepo) RemoveRole(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
	if m.RemoveRoleFunc != nil {
		return m.RemoveRoleFunc(ctx, tx, userID, roleID)
	}
	return nil
}

func (m *MockUserRepo) HasPermission(ctx context.Context, email, routePath, routeGroup string) bool {
	if m.HasPermissionFunc != nil {
		return m.HasPermissionFunc(ctx, email, routePath, routeGroup)
	}
	return false
}
