package service_test

import (
	"context"
	"fmt"
	"testing"
	"workshop/internal/model"
	"workshop/internal/service"
	"workshop/mock/mockpkg"
	"workshop/mock/mockrepo"
	"workshop/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestRoles_Update_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
		UpdateFunc: func(ctx context.Context, role *model.Role) error {
			return nil
		},
	}

	svc := service.NewRoles(log, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Error_FindByID(t *testing.T) {
	setupErr := fmt.Errorf("error db")
	expectedError := errors.InternalServerErrorWrap(setupErr, "error finding role")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, setupErr
		},
	}

	svc := service.NewRoles(log, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Error_NotFound(t *testing.T) {
	expectedError := errors.NotFound("role not found")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, nil
		},
	}

	svc := service.NewRoles(log, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Error_Update(t *testing.T) {
	setupErr := fmt.Errorf("error db")
	expectedError := errors.InternalServerErrorWrap(setupErr, "error updating role")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
		UpdateFunc: func(ctx context.Context, role *model.Role) error {
			return setupErr
		},
	}

	svc := service.NewRoles(log, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}
