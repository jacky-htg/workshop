package service_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
	"workshop/config"
	"workshop/internal/model"
	"workshop/internal/service"
	"workshop/mock/mockpkg"
	"workshop/mock/mockrepo"
	"workshop/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoles_List_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			return fmt.Errorf("error faind cache")
		},
		SetJSONWithExpiryFunc: func(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
			return fmt.Errorf("error set cache")
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		ListFunc: func(ctx context.Context) ([]model.Role, error) {
			return []model.Role{{ID: 1, Name: "admin"}}, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	roles, err := svc.List(context.Background())

	assert.Equal(t, expectedError, err)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, "admin", roles[0].Name)
}

func TestRoles_List_Success_FromCache(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			results, ok := dest.(*[]model.Role)
			require.True(t, ok, "dest is not []model.Role")

			*results = []model.Role{{ID: 1, Name: "admin"}}
			return nil
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{}

	svc := service.NewRoles(cache, log, ttl, repo)

	roles, err := svc.List(context.Background())

	assert.Equal(t, expectedError, err)
	assert.Equal(t, 1, len(roles))
	assert.Equal(t, "admin", roles[0].Name)
}

func TestRoles_List_Error(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error listing roles")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			return fmt.Errorf("error faind cache")
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		ListFunc: func(ctx context.Context) ([]model.Role, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	roles, err := svc.List(context.Background())

	assert.Equal(t, expectedError, err)
	assert.Equal(t, 0, len(roles))
}

func TestRoles_FindByID_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			return fmt.Errorf("error get cache")
		},
		SetJSONWithExpiryFunc: func(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
			return fmt.Errorf("error set cache")
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: 1, Name: "admin"}, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role, err := svc.FindByID(context.Background(), 1)

	assert.Equal(t, expectedError, err)
	assert.Equal(t, 1, role.ID)
	assert.Equal(t, "admin", role.Name)
}

func TestRoles_FindByID_Success_FromCache(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			role, ok := dest.(**model.Role)
			require.True(t, ok)

			*role = &model.Role{ID: 1, Name: "admin"}
			return nil
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{}

	svc := service.NewRoles(cache, log, ttl, repo)

	role, err := svc.FindByID(context.Background(), 1)

	assert.Equal(t, expectedError, err)
	assert.Equal(t, 1, role.ID)
	assert.Equal(t, "admin", role.Name)
}

func TestRoles_FindByID_NotFound(t *testing.T) {
	expectedError := errors.NotFound("role not found")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			return fmt.Errorf("error get cache")
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role, err := svc.FindByID(context.Background(), 1)

	assert.Equal(t, expectedError, err)
	assert.Nil(t, role)
}

func TestRoles_FindByID_Error(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error finding role")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			return fmt.Errorf("error get cache")
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role, err := svc.FindByID(context.Background(), 1)

	assert.Equal(t, expectedError, err)
	assert.Nil(t, role)
}

func TestRoles_Create_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}
	role := &model.Role{Name: "admin"}
	repo := &mockrepo.MockRoleRepo{
		CreateFunc: func(ctx context.Context, role *model.Role) error {
			return nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Create(context.Background(), role)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Create_Error(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error creating role")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}
	role := &model.Role{Name: "admin"}
	repo := &mockrepo.MockRoleRepo{
		CreateFunc: func(ctx context.Context, role *model.Role) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Create(context.Background(), role)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Create_DelCacheError(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error creating role")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, sql.ErrConnDone
		},
	}
	ttl := config.TTLConfig{}
	role := &model.Role{Name: "admin"}
	repo := &mockrepo.MockRoleRepo{}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Create(context.Background(), role)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
		UpdateFunc: func(ctx context.Context, role *model.Role) error {
			return nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Error_FindByID(t *testing.T) {
	setupErr := fmt.Errorf("error db")
	expectedError := errors.InternalServerErrorWrap(setupErr, "error finding role")

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, setupErr
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Error_NotFound(t *testing.T) {
	expectedError := errors.NotFound("role not found")

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Error_Update(t *testing.T) {
	setupErr := fmt.Errorf("error db")
	expectedError := errors.InternalServerErrorWrap(setupErr, "error updating role")

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
		UpdateFunc: func(ctx context.Context, role *model.Role) error {
			return setupErr
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Update_Error_DelCache(t *testing.T) {
	setupErr := fmt.Errorf("error db")
	expectedError := errors.InternalServerErrorWrap(setupErr, "error updating role")

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, setupErr
		},
	}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)

	role := &model.Role{ID: 1, Name: "Super Admin"}
	err := svc.Update(context.Background(), role)

	assert.Equal(t, expectedError, err)
}

func TestRoles_Delete_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
		DeleteFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Delete(context.Background(), 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Delete_Error(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error delete role")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
		DeleteFunc: func(ctx context.Context, id int) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Delete(context.Background(), 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Delete_DelCacheError(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error delete role")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, sql.ErrConnDone
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return &model.Role{ID: id, Name: "admin"}, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Delete(context.Background(), 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Delete_NotFound(t *testing.T) {
	expectedError := errors.NotFound("role not found")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Delete(context.Background(), 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Delete_FindByIDError(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error finding role")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		FindByIDFunc: func(ctx context.Context, id int) (*model.Role, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Delete(context.Background(), 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Grant_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return false, nil
		},
		GrantAccessFunc: func(ctx context.Context, roleID, accessID int) error {
			return nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Grant(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Grant_Error(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error grant access")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return false, nil
		},
		GrantAccessFunc: func(ctx context.Context, roleID, accessID int) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Grant(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Grant_DelCacheError(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error grant access")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, sql.ErrConnDone
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return false, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Grant(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Grant_HasAccessError(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error grant access")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return false, sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Grant(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Revoke_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return true, nil
		},
		RevokeAccessFunc: func(ctx context.Context, roleID, accessID int) error {
			return nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Revoke(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Revoke_Error(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error revoke access")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return true, nil
		},
		RevokeAccessFunc: func(ctx context.Context, roleID, accessID int) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Revoke(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Revoke_DelCacheError(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error revoke access")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, sql.ErrConnDone
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return true, nil
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Revoke(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}

func TestRoles_Revoke_HasAccessError(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error revoke access")
	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockRoleRepo{
		HasAccessFunc: func(ctx context.Context, roleID, accessID int) (bool, error) {
			return false, sql.ErrConnDone
		},
	}

	svc := service.NewRoles(cache, log, ttl, repo)
	err := svc.Revoke(context.Background(), 1, 1)
	assert.Equal(t, expectedError, err)
}
