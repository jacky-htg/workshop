package mockrepo

import (
	"context"
	"database/sql"
	"workshop/internal/model"
)

type MockAccessRepo struct {
	ListFunc   func(ctx context.Context) ([]model.Access, error)
	CreateFunc func(ctx context.Context, tx *sql.Tx, access *model.Access) error
}

func (m *MockAccessRepo) List(ctx context.Context) ([]model.Access, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockAccessRepo) Create(ctx context.Context, tx *sql.Tx, access *model.Access) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, tx, access)
	}
	return nil
}
