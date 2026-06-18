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

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccesses_List_Success(t *testing.T) {
	var expectedError *errors.BusinessError
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			return fmt.Errorf("error cache")
		},
		SetJSONWithExpiryFunc: func(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
			return fmt.Errorf("error cache")
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockAccessRepo{
		ListFunc: func(ctx context.Context) ([]model.Access, error) {
			rootID := 1
			userID := 2
			roleID := 3
			return []model.Access{
				{ID: userID, ParentID: &rootID, Name: "users", Alias: "users"},
				{ID: roleID, ParentID: &rootID, Name: "roles", Alias: "roles"},
				{ID: 4, ParentID: &userID, Name: "GET /users", Alias: "users:list"},
				{ID: 5, ParentID: &userID, Name: "POST /users", Alias: "users:create"},
				{ID: 6, ParentID: &roleID, Name: "GET /roles", Alias: "roles:list"},
				{ID: 7, ParentID: &roleID, Name: "GET /roles/{id}", Alias: "roles:view"},
			}, nil
		},
	}

	svc := service.NewAccesses(db, cache, log, ttl, repo)
	accessTree, err := svc.List(context.Background())

	assert.Equal(t, expectedError, err)
	assert.Equal(t, 2, len(accessTree))
}

func TestAccesses_List_Success_FromCache(t *testing.T) {
	var expectedError *errors.BusinessError
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			results, ok := dest.(*map[int]*model.AccessTree)
			if !ok {
				return fmt.Errorf("dest is not *map[int]*model.AccessTree")
			}
			*results = map[int]*model.AccessTree{
				0: {ID: 11, Name: "users", Alias: "users"},
				1: {ID: 12, Name: "roles", Alias: "roles"},
			}
			return nil
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockAccessRepo{}

	svc := service.NewAccesses(db, cache, log, ttl, repo)
	accessTree, err := svc.List(context.Background())

	assert.Equal(t, expectedError, err)
	assert.Equal(t, 2, len(accessTree))
}

func TestAccesses_List_Error(t *testing.T) {
	expectedError := errors.InternalServerErrorWrap(sql.ErrConnDone, "error listing access")
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		GetJSONFunc: func(ctx context.Context, key string, dest interface{}) error {
			return fmt.Errorf("error cache")
		},
	}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockAccessRepo{
		ListFunc: func(ctx context.Context) ([]model.Access, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewAccesses(db, cache, log, ttl, repo)
	accessTree, err := svc.List(context.Background())

	assert.Equal(t, expectedError, err)
	assert.Nil(t, accessTree)
}

func TestAccesses_ScanAccess_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, fmt.Errorf("error delete cache")
		},
	}
	ttl := config.TTLConfig{}

	var createdAccesses []model.Access
	repo := &mockrepo.MockAccessRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, access *model.Access) error {
			access.ID = len(createdAccesses) + 1
			createdAccesses = append(createdAccesses, *access)
			return nil
		},
	}

	mock.ExpectBegin()
	mock.ExpectCommit()
	svc := service.NewAccesses(db, cache, log, ttl, repo)
	err = svc.ScanAccess(context.Background(), "../../mock/mockdata/route.go")

	assert.NoError(t, err)
}

func TestAccesses_ScanAccess_ParsedFileError(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockAccessRepo{}

	svc := service.NewAccesses(db, cache, log, ttl, repo)
	err = svc.ScanAccess(context.Background(), "route.go")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse file")
}

func TestAccesses_ScanAccess_NoRouteDefinitionError(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}

	repo := &mockrepo.MockAccessRepo{}

	svc := service.NewAccesses(db, cache, log, ttl, repo)
	err = svc.ScanAccess(context.Background(), "accesses.go")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no route definitions found")
}

func TestAccesses_ScanAccess_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}
	repo := &mockrepo.MockAccessRepo{}
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	svc := service.NewAccesses(db, cache, log, ttl, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}

func TestAccesses_ScanAccess_CommitTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}

	var createdAccesses []model.Access
	repo := &mockrepo.MockAccessRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, access *model.Access) error {
			access.ID = len(createdAccesses) + 1
			createdAccesses = append(createdAccesses, *access)
			return nil
		},
	}

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)
	svc := service.NewAccesses(db, cache, log, ttl, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}

func TestAccesses_ScanAccess_CreateGroupError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}

	var createdAccesses []model.Access
	repo := &mockrepo.MockAccessRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, access *model.Access) error {
			if *access.ParentID == 1 {
				return sql.ErrConnDone
			}
			access.ID = len(createdAccesses) + 1
			createdAccesses = append(createdAccesses, *access)
			return nil
		},
	}

	mock.ExpectBegin()
	mock.ExpectRollback()
	svc := service.NewAccesses(db, cache, log, ttl, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}

func TestAccesses_ScanAccess_CreateAccessError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	cache := &mockpkg.MockCache{}
	ttl := config.TTLConfig{}

	var createdAccesses []model.Access
	repo := &mockrepo.MockAccessRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, access *model.Access) error {
			if *access.ParentID == 1 {
				return nil
			}
			access.ID = len(createdAccesses) + 1
			createdAccesses = append(createdAccesses, *access)
			return sql.ErrConnDone
		},
	}

	mock.ExpectBegin()
	mock.ExpectRollback()
	svc := service.NewAccesses(db, cache, log, ttl, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}
