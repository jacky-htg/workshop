package repository_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"workshop/internal/model"
	"workshop/internal/repository"
	"workshop/mock/mockpkg"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	qUserCreate   = `INSERT INTO users (id, name, username, password, email, is_active) VALUES ($1, $2, $3, $4, $5, $6)`
	qUserFindByID = `
						SELECT u.id, u.name, u.username, u.password, u.email, u.is_active, 
								COALESCE(
									json_agg(
										json_build_object(
											'id', r.id,
											'name', r.name
										)
									) FILTER (WHERE r.id IS NOT NULL),
									'[]'::json
								)  AS roles
						FROM users u
						LEFT JOIN roles_users ru ON (u.id = ru.user_id)
						LEFT JOIN roles r ON (ru.role_id = r.id) 
						WHERE u.id = $1 AND u.deleted_at IS NULL 
						GROUP BY u.id, u.name, u.username, u.password, u.email, u.is_active`
	qUserFindByEmail = `
						SELECT u.id, u.name, u.username, u.password, u.email, u.is_active, 
								COALESCE(
									json_agg(
										json_build_object(
											'id', r.id,
											'name', r.name
										)
									) FILTER (WHERE r.id IS NOT NULL),
									'[]'::json
								)  AS roles
						FROM users u
						LEFT JOIN roles_users ru ON (u.id = ru.user_id)
						LEFT JOIN roles r ON (ru.role_id = r.id) 
						WHERE u.email = $1 AND u.deleted_at IS NULL 
						GROUP BY u.id, u.name, u.username, u.password, u.email, u.is_active`
	qUserUpdate        = `UPDATE users SET name = $1, is_active = $2 WHERE id = $3 RETURNING username, email`
	qUserDelete        = `UPDATE users SET deleted_at = timezone('utc', now()) WHERE id = $1`
	qUserAssignRole    = `INSERT INTO roles_users (role_id, user_id) VALUES ($1, $2)`
	qUserRemoveRole    = `DELETE FROM roles_users WHERE role_id = $1 AND user_id = $2`
	qUserHasPermission = `
						SELECT true 
						FROM users u
						JOIN roles_users ru ON (u.id = ru.user_id)
						JOIN roles r ON (ru.role_id = r.id)
						JOIN access_roles ar ON (r.id = ar.role_id) 
						JOIN access a ON (ar.access_id = a.id)
						WHERE u.email = $1 AND (a.name = $2 OR a.name = $3 OR a.name = 'root') `
)

func TestUserRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()

	// 1. Mock COUNT query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)).
		WithArgs().
		WillReturnRows(countRows)

	// 2. Mock SELECT query
	rows := sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active"}).
		AddRow("user-1", "John Doe", "john", "password123", "john@example.com", true).
		AddRow("user-2", "Jane Doe", "jane", "password456", "jane@example.com", false)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, username, password, email, is_active FROM users WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 10 OFFSET 0`)).
		WithArgs().
		WillReturnRows(rows)

	// Execute
	users, count, err := repo.List(ctx, "", "id", "ASC", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, users, 2)
	assert.Equal(t, "John Doe", users[0].Name)
	assert.Equal(t, "Jane Doe", users[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_WithSearch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()
	search := "john"

	// 1. Mock COUNT query with search
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND (name ILIKE $1)`)).
		WithArgs("%john%").
		WillReturnRows(countRows)

	// 2. Mock SELECT query with search
	rows := sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active"}).
		AddRow("user-1", "John Doe", "john", "password123", "john@example.com", true)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, username, password, email, is_active FROM users WHERE deleted_at IS NULL AND (name ILIKE $1) ORDER BY id ASC LIMIT 10 OFFSET 0`)).
		WithArgs("%john%").
		WillReturnRows(rows)

	users, count, err := repo.List(ctx, search, "id", "ASC", 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Len(t, users, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_CountError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()

	// Mock COUNT query to return error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)).
		WithArgs().
		WillReturnError(sql.ErrConnDone)

	users, count, err := repo.List(ctx, "", "id", "ASC", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, 0, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_SelectError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()

	// 1. Mock COUNT success
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)).
		WithArgs().
		WillReturnRows(countRows)

	// 2. Mock SELECT error
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, username, password, email, is_active FROM users WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 10 OFFSET 0`)).
		WithArgs().
		WillReturnError(sql.ErrConnDone)

	users, count, err := repo.List(ctx, "", "id", "ASC", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, 2, count) // Count masih tetap 2 meskipun select error
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()

	// 1. Mock COUNT success
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)).
		WithArgs().
		WillReturnRows(countRows)

	// 2. Mock SELECT with wrong column type (ID as integer, should be string)
	rows := sqlmock.NewRows([]string{"name", "username", "password", "email", "is_active"}).
		AddRow("John Doe", "john", "pass", "john@example.com", true) // error jumlah kolom tidak sesuai dengan scan

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, username, password, email, is_active FROM users WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 10 OFFSET 0`)).
		WithArgs().
		WillReturnRows(rows)

	users, count, err := repo.List(ctx, "", "id", "ASC", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, 1, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_OrderError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()

	// 1. Mock COUNT success
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)).
		WithArgs().
		WillReturnRows(countRows)

	users, count, err := repo.List(ctx, "", "unknown-field", "ASC", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, 1, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_List_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()

	// 1. Mock COUNT success
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)).
		WithArgs().
		WillReturnRows(countRows)

	// 2. Mock SELECT with row iteration error
	rows := sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active"}).
		AddRow("user-1", "John Doe", "john", "pass", "john@example.com", true).
		AddRow("user-2", "Jane Doe", "jane", "pass", "jane@example.com", false).
		RowError(1, errors.New("iteration error"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, username, password, email, is_active FROM users WHERE deleted_at IS NULL ORDER BY id ASC LIMIT 10 OFFSET 0`)).
		WithArgs().
		WillReturnRows(rows)

	users, count, err := repo.List(ctx, "", "id", "ASC", 10, 0)

	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Equal(t, 2, count)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	user := &model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "secret",
		IsActive: true,
	}

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta(qUserCreate)).
		WithArgs(user.ID, user.Name, user.Username, user.Password, user.Email, user.IsActive).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, tx, user)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)
	user := &model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "secret",
		IsActive: true,
	}

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta(qUserCreate)).
		WithArgs(user.ID, user.Name, user.Username, user.Password, user.Email, user.IsActive).
		WillReturnError(errors.New("duplicate key value violates unique constraint"))

	err = repo.Create(ctx, tx, user)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	expectedRoles := []model.Role{
		{ID: 1, Name: "finance"},
		{ID: 1, Name: "kasir"},
	}

	expected := model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "secret",
		IsActive: true,
		Roles:    expectedRoles,
	}

	rolesJSON, err := json.Marshal(expectedRoles)
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByID)).
		WithArgs(expected.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active", "roles"}).
			AddRow(expected.ID, expected.Name, expected.Username, expected.Password, expected.Email, expected.IsActive, rolesJSON))

	ctx := context.Background()
	user, err := repo.FindByID(ctx, expected.ID)

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, user.ID)
	assert.Equal(t, expected.Name, user.Name)
	assert.Equal(t, len(expected.Roles), len(user.Roles))
	assert.Equal(t, expected.Roles[0].Name, user.Roles[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByID)).
		WithArgs("user-id").
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	user, err := repo.FindByID(ctx, "user-id")

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	expected := model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Email:    "admin@example.com",
		Username: "admin",
		Password: "secret",
	}

	expectedRoles := []model.Role{
		{ID: 1, Name: "finance"},
		{ID: 1, Name: "kasir"},
	}

	rolesJSON, err := json.Marshal(expectedRoles)
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByID)).
		WithArgs(expected.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active", "roles"}).
			AddRow(expected.ID, expected.Name, expected.Username, expected.Password, expected.Email, "active", rolesJSON))

	ctx := context.Background()
	user, err := repo.FindByID(ctx, expected.ID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "sql: Scan error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_UnmarshalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	expected := model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "secret",
		IsActive: true,
	}

	rolesJSON := []byte(`{"id":1)`) // invalid closing bracket

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByID)).
		WithArgs(expected.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active", "roles"}).
			AddRow(expected.ID, expected.Name, expected.Username, expected.Password, expected.Email, expected.IsActive, rolesJSON))

	ctx := context.Background()
	user, err := repo.FindByID(ctx, expected.ID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid character ')'")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	expectedRoles := []model.Role{
		{ID: 1, Name: "finance"},
		{ID: 1, Name: "kasir"},
	}

	expected := model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Username: "admin",
		Password: "secret",
		IsActive: true,
		Roles:    expectedRoles,
	}

	rolesJSON, err := json.Marshal(expectedRoles)
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByEmail)).
		WithArgs("admin@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active", "roles"}).
			AddRow(expected.ID, expected.Name, expected.Username, expected.Password, expected.Email, expected.IsActive, rolesJSON))

	ctx := context.Background()
	user, err := repo.FindByEmail(ctx, "admin@example.com")

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, user.ID)
	assert.Equal(t, expected.Name, user.Name)
	assert.Equal(t, len(expected.Roles), len(user.Roles))
	assert.Equal(t, expected.Roles[0].Name, user.Roles[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByEmail)).
		WithArgs("admin@example.com").
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	user, err := repo.FindByEmail(ctx, "admin@example.com")

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	expected := model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Username: "admin",
		Password: "secret",
	}

	expectedRoles := []model.Role{
		{ID: 1, Name: "finance"},
		{ID: 1, Name: "kasir"},
	}

	rolesJSON, err := json.Marshal(expectedRoles)
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByEmail)).
		WithArgs("admin@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active", "roles"}).
			AddRow(expected.ID, expected.Name, expected.Username, expected.Password, expected.Email, "active", rolesJSON))

	ctx := context.Background()
	user, err := repo.FindByEmail(ctx, "admin@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "sql: Scan error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_UnmarshalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	expected := model.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Name:     "admin",
		Username: "admin",
		Password: "secret",
		IsActive: true,
	}

	rolesJSON := []byte(`{"id":1)`) // invalid closing bracket

	mock.ExpectQuery(regexp.QuoteMeta(qUserFindByEmail)).
		WithArgs("admin@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "username", "password", "email", "is_active", "roles"}).
			AddRow(expected.ID, expected.Name, expected.Username, expected.Password, expected.Email, expected.IsActive, rolesJSON))

	ctx := context.Background()
	user, err := repo.FindByEmail(ctx, "admin@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid character ')'")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	user := &model.User{ID: "550e8400-e29b-41d4-a716-446655440000", Name: "admin", IsActive: true}

	mock.ExpectQuery(regexp.QuoteMeta(qUserUpdate)).
		WithArgs(user.Name, user.IsActive, user.ID).
		WillReturnRows(sqlmock.NewRows([]string{"username", "email"}).AddRow("admin", "admin@example.com"))

	err = repo.Update(ctx, tx, user)

	assert.NoError(t, err)
	assert.Equal(t, "admin", user.Username)
	assert.Equal(t, "admin@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Update_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	user := &model.User{ID: "550e8400-e29b-41d4-a716-446655440000", Name: "admin", IsActive: true}

	mock.ExpectQuery(regexp.QuoteMeta(qUserUpdate)).
		WithArgs(user.Name, user.IsActive, user.ID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("admin@example.com"))

	err = repo.Update(ctx, tx, user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql: expected 1 destination arguments in Scan")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qUserDelete)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(ctx, "user-id")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qUserDelete)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.Delete(ctx, "user-id")

	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_AssignRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qUserAssignRole)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.AssignRole(ctx, tx, "user-id", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_AssignRole_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qUserAssignRole)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.AssignRole(ctx, tx, "user-id", 1)

	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_RemoveRole_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qUserRemoveRole)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.RemoveRole(ctx, tx, "user-id", 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_RemoveRole_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qUserRemoveRole)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.RemoveRole(ctx, tx, "user-id", 1)

	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_HasPermission_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qUserHasPermission)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"has_permission"}).AddRow(true))

	ctx := context.Background()
	hasPermission := repo.HasPermission(ctx, "admin@example.com", "GET /users", "users")

	assert.True(t, hasPermission)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_HasPermission_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewUserRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qUserHasPermission)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"has_permission"}).AddRow("invalid-bool"))

	ctx := context.Background()
	hasPermission := repo.HasPermission(ctx, "admin@example.com", "GET /users", "users")

	assert.False(t, hasPermission)
	assert.NoError(t, mock.ExpectationsWereMet())
}
