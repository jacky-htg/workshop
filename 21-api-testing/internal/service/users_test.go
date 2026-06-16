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

func TestUsers_List_Success(t *testing.T) {
	var expectedErr *errors.BusinessError

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		ListFunc: func(ctx context.Context, search, order, sort string, limit, offset int) ([]model.User, int, error) {
			return []model.User{
				{ID: "user-1", Name: "John"},
				{ID: "user-2", Name: "Smith"},
			}, 2, nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	users, pagination, err := svc.List(context.Background(), "", "", "", 10, 1)

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 2, len(users))
	assert.Equal(t, "John", users[0].Name)
	assert.Equal(t, 2, pagination.Count)
}

func TestUsers_List_Error(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error listing users")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		ListFunc: func(ctx context.Context, search, order, sort string, limit, offset int) ([]model.User, int, error) {
			return nil, 0, sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	users, pagination, err := svc.List(context.Background(), "", "", "", 10, 1)

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, users)
	assert.Equal(t, 0, pagination.Count)
}

func TestUsers_Create_Success(t *testing.T) {
	var expectedErr *errors.BusinessError

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{ID: "user-id", Roles: []model.Role{{ID: 1}}}
	mock.ExpectBegin()
	mock.ExpectCommit()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			return nil
		},
		AssignRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Create(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Create_CommitError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{ID: "user-id", Roles: []model.Role{{ID: 1}}}
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			return nil
		},
		AssignRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Create(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Create_AssignRoleError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error assign role")

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{ID: "user-id", Roles: []model.Role{{ID: 1}}}
	mock.ExpectBegin()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			return nil
		},
		AssignRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Create(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Create_Error(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error creating user")

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{}
	mock.ExpectBegin()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		CreateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Create(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Create_BeginTxError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{}
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{}

	svc := service.NewUsers(db, log, repo)
	err = svc.Create(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_FindByID_Success(t *testing.T) {
	var expectedErr *errors.BusinessError

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id}, nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	user, err := svc.FindByID(context.Background(), "user-id")

	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "user-id", user.ID)
}

func TestUsers_FindByID_NotFound(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.NotFound("user not found")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	user, err := svc.FindByID(context.Background(), "user-id")

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, user)
}

func TestUsers_FindByID_Error(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error finding user")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	user, err := svc.FindByID(context.Background(), "user-id")

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, user)
}

func TestUsers_Update_Success(t *testing.T) {
	var expectedErr *errors.BusinessError

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{
		ID: "user-id",
		Roles: []model.Role{
			{ID: 1, Name: "kasir"},
			{ID: 3, Name: "gudang"},
		},
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID: id,
				Roles: []model.Role{
					{ID: 1, Name: "kasir"},
					{ID: 2, Name: "finance"},
				},
			}, nil
		},
		UpdateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			user.Username = "admin"
			user.Email = "admin@example.com"
			return nil
		},
		AssignRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return nil
		},
		RemoveRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Update_CommitError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{
		ID: "user-id",
		Roles: []model.Role{
			{ID: 1, Name: "kasir"},
			{ID: 3, Name: "gudang"},
		},
	}
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID: id,
				Roles: []model.Role{
					{ID: 1, Name: "kasir"},
					{ID: 2, Name: "finance"},
				},
			}, nil
		},
		UpdateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			user.Username = "admin"
			user.Email = "admin@example.com"
			return nil
		},
		AssignRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return nil
		},
		RemoveRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Update_RemoveRoleError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error update assign role")

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{
		ID: "user-id",
		Roles: []model.Role{
			{ID: 1, Name: "kasir"},
			{ID: 3, Name: "gudang"},
		},
	}
	mock.ExpectBegin()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID: id,
				Roles: []model.Role{
					{ID: 1, Name: "kasir"},
					{ID: 2, Name: "finance"},
				},
			}, nil
		},
		UpdateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			user.Username = "admin"
			user.Email = "admin@example.com"
			return nil
		},
		AssignRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return nil
		},
		RemoveRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Update_AssignRoleError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error update assign role")

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{
		ID: "user-id",
		Roles: []model.Role{
			{ID: 1, Name: "kasir"},
			{ID: 3, Name: "gudang"},
		},
	}
	mock.ExpectBegin()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{
				ID: id,
				Roles: []model.Role{
					{ID: 1, Name: "kasir"},
					{ID: 2, Name: "finance"},
				},
			}, nil
		},
		UpdateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			user.Username = "admin"
			user.Email = "admin@example.com"
			return nil
		},
		AssignRoleFunc: func(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Update_Error(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error updating user")

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{ID: "user-id"}
	mock.ExpectBegin()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id}, nil
		},
		UpdateFunc: func(ctx context.Context, tx *sql.Tx, user *model.User) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Update_BeginTxError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{ID: "user-id"}
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id}, nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Update_NotFound(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.NotFound("user not found")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{ID: "user-id"}
	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Update_FindByIDError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error finding user")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	user := &model.User{ID: "user-id"}
	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Update(context.Background(), user)

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Delete_Success(t *testing.T) {
	var expectedErr *errors.BusinessError

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id}, nil
		},
		DeleteFunc: func(ctx context.Context, id string) error {
			return nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Delete(context.Background(), "user-id")

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Delete_Error(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error deleting user")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return &model.User{ID: id}, nil
		},
		DeleteFunc: func(ctx context.Context, id string) error {
			return sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Delete(context.Background(), "user-id")

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Delete_NotFound(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.NotFound("user not found")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, nil
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Delete(context.Background(), "user-id")

	assert.Equal(t, expectedErr, err)
}

func TestUsers_Delete_FindByIDError(t *testing.T) {
	var expectedErr *errors.BusinessError = errors.InternalServerErrorWrap(sql.ErrConnDone, "error finding user")

	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := &mockpkg.MockLogger{}
	repo := &mockrepo.MockUserRepo{
		FindByIDFunc: func(ctx context.Context, id string) (*model.User, error) {
			return nil, sql.ErrConnDone
		},
	}

	svc := service.NewUsers(db, log, repo)
	err = svc.Delete(context.Background(), "user-id")

	assert.Equal(t, expectedErr, err)
}
