package service_test

import (
	"context"
	"database/sql"
	"testing"
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

	svc := service.NewAccesses(db, log, repo)
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

	repo := &mockrepo.MockAccessRepo{
		ListFunc: func(ctx context.Context) ([]model.Access, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewAccesses(db, log, repo)
	accessTree, err := svc.List(context.Background())

	assert.Equal(t, expectedError, err)
	assert.Nil(t, accessTree)
}

func TestAccesses_ScanAccess_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()

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
	svc := service.NewAccesses(db, log, repo)
	err = svc.ScanAccess(context.Background(), "../../mock/mockdata/route.go")

	assert.NoError(t, err)
}

func TestAccesses_ScanAccess_ParsedFileError(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()

	repo := &mockrepo.MockAccessRepo{}

	svc := service.NewAccesses(db, log, repo)
	err = svc.ScanAccess(context.Background(), "route.go")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse file")
}

func TestAccesses_ScanAccess_NoRouteDefinitionError(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()

	repo := &mockrepo.MockAccessRepo{}

	svc := service.NewAccesses(db, log, repo)
	err = svc.ScanAccess(context.Background(), "accesses.go")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no route definitions found")
}

func TestAccesses_ScanAccess_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := &mockrepo.MockAccessRepo{}
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	svc := service.NewAccesses(db, log, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}

func TestAccesses_ScanAccess_CommitTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()

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
	svc := service.NewAccesses(db, log, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}

func TestAccesses_ScanAccess_CreateGroupError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()

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
	svc := service.NewAccesses(db, log, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}

func TestAccesses_ScanAccess_CreateAccessError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()

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
	svc := service.NewAccesses(db, log, repo)
	err = svc.ScanAccess(context.Background(), "../router/api.go")

	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
}
