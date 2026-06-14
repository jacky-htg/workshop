package service_test

import (
	"context"
	"database/sql"
	"testing"
	"workshop/config"
	"workshop/internal/model"
	"workshop/internal/service"
	"workshop/mock/mockpkg"
	"workshop/mock/mockrepo"
	"workshop/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuths_Login_Success(t *testing.T) {
	pass, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var expectedError *errors.BusinessError

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:       "id-user",
				Password: string(pass),
				IsActive: true,
				Roles: []model.Role{
					{ID: 1, Name: "admin"},
				},
			}, nil
		},
	}
	roleRepo := &mockrepo.MockRoleRepo{
		GetAccessesByRolesFunc: func(ctx context.Context, roleIDs []int) ([]model.Access, error) {
			return []model.Access{
				{ID: 1, Name: "root", Alias: "root"},
			}, nil
		},
	}
	cfgToken := config.TokenConfig{}

	svc := service.NewAuths(log, cfgToken, repo, roleRepo)
	_, user, permissions, err := svc.Login(context.Background(), "admin@example.com", "secret")

	assert.Equal(t, expectedError, err)
	assert.Equal(t, "id-user", user.ID)
	assert.Equal(t, 1, len(permissions))
	assert.Equal(t, "root", permissions[0])
}

func TestAuths_Login_InActive(t *testing.T) {
	pass, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var expectedError *errors.BusinessError = errors.Forbidden("user inavtive")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:       "id-user",
				Password: string(pass),
				IsActive: false,
				Roles: []model.Role{
					{ID: 1, Name: "admin"},
				},
			}, nil
		},
	}
	roleRepo := &mockrepo.MockRoleRepo{
		GetAccessesByRolesFunc: func(ctx context.Context, roleIDs []int) ([]model.Access, error) {
			return []model.Access{
				{ID: 1, Name: "root", Alias: "root"},
			}, nil
		},
	}
	cfgToken := config.TokenConfig{}

	svc := service.NewAuths(log, cfgToken, repo, roleRepo)
	_, user, permissions, err := svc.Login(context.Background(), "admin@example.com", "secret")

	assert.Equal(t, expectedError, err)
	assert.Nil(t, user)
	assert.Equal(t, 1, len(permissions))
	assert.Equal(t, "root", permissions[0])
}

func TestAuths_Login_PasswordError(t *testing.T) {
	pass, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var expectedError *errors.BusinessError = errors.InvalidInput("Invalid username/password")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:       "id-user",
				Password: string(pass),
				IsActive: false,
				Roles: []model.Role{
					{ID: 1, Name: "admin"},
				},
			}, nil
		},
	}
	roleRepo := &mockrepo.MockRoleRepo{
		GetAccessesByRolesFunc: func(ctx context.Context, roleIDs []int) ([]model.Access, error) {
			return []model.Access{
				{ID: 1, Name: "root", Alias: "root"},
			}, nil
		},
	}
	cfgToken := config.TokenConfig{}

	svc := service.NewAuths(log, cfgToken, repo, roleRepo)
	_, user, permissions, err := svc.Login(context.Background(), "admin@example.com", "1234")

	assert.Equal(t, expectedError, err)
	assert.Nil(t, user)
	assert.Equal(t, 1, len(permissions))
	assert.Equal(t, "root", permissions[0])
}

func TestAuths_Login_GetPermissionError(t *testing.T) {
	pass, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	require.NoError(t, err)

	var expectedError *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error finding user")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return &model.User{
				ID:       "id-user",
				Password: string(pass),
				IsActive: false,
				Roles: []model.Role{
					{ID: 1, Name: "admin"},
				},
			}, nil
		},
	}
	roleRepo := &mockrepo.MockRoleRepo{
		GetAccessesByRolesFunc: func(ctx context.Context, roleIDs []int) ([]model.Access, error) {
			return nil, sql.ErrConnDone
		},
	}
	cfgToken := config.TokenConfig{}

	svc := service.NewAuths(log, cfgToken, repo, roleRepo)
	_, user, permissions, err := svc.Login(context.Background(), "admin@example.com", "1234")

	assert.Equal(t, expectedError, err)
	assert.Nil(t, user)
	assert.Equal(t, 0, len(permissions))
}

func TestAuths_Login_UserNotFound(t *testing.T) {
	var expectedError *errors.BusinessError = errors.InvalidInput("Invalid username/password")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return nil, nil
		},
	}
	roleRepo := &mockrepo.MockRoleRepo{}
	cfgToken := config.TokenConfig{}

	svc := service.NewAuths(log, cfgToken, repo, roleRepo)
	_, user, permissions, err := svc.Login(context.Background(), "admin@example.com", "1234")

	assert.Equal(t, expectedError, err)
	assert.Nil(t, user)
	assert.Equal(t, 0, len(permissions))
}

func TestAuths_Login_FindByIDError(t *testing.T) {
	var expectedError *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error finding user")

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockUserRepo{
		FindByEmailFunc: func(ctx context.Context, email string) (*model.User, error) {
			return nil, sql.ErrConnDone
		},
	}
	roleRepo := &mockrepo.MockRoleRepo{}
	cfgToken := config.TokenConfig{}

	svc := service.NewAuths(log, cfgToken, repo, roleRepo)
	_, user, permissions, err := svc.Login(context.Background(), "admin@example.com", "1234")

	assert.Equal(t, expectedError, err)
	assert.Nil(t, user)
	assert.Equal(t, 0, len(permissions))
}
