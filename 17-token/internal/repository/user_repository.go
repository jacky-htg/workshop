package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"workshop/internal/model"

	"github.com/jacky-htg/go-libs/logger"
)

type UserRepository interface {
	List(ctx context.Context) ([]model.User, error)
	Create(ctx context.Context, user *model.User) error
	FindById(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewUserRepository(db *sql.DB, log logger.Logger) UserRepository {
	return &userRepository{db: db, log: log}
}

// List : http handler for returning list of users
func (u *userRepository) List(ctx context.Context) ([]model.User, error) {
	query := `SELECT id, name, username, password, email, is_active FROM users WHERE deleted_at IS NULL`
	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		u.log.Error(ctx, "error: querying users", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive); err != nil {

			u.log.Error(ctx, "error: scanning user row", slog.Any("error", err))
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		u.log.Error(ctx, "error: iterating user rows", slog.Any("error", err))
		return nil, err
	}

	return users, nil
}

func (u *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, name, username, password, email, is_active) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := u.db.ExecContext(ctx, query, user.ID, user.Name, user.Username, user.Password, user.Email, user.IsActive)
	if err != nil {
		u.log.Error(ctx, "error: inserting user", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) FindById(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, name, username, password, email, is_active FROM users WHERE id = $1 AND deleted_at IS NULL`
	row := u.db.QueryRowContext(ctx, query, id)

	var user model.User
	if err := row.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		u.log.Error(ctx, "error: scanning user row", slog.Any("error", err))
		return nil, err
	}

	return &user, nil
}

func (u *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, name, username, password, email, is_active FROM users WHERE email = $1 AND deleted_at IS NULL`
	row := u.db.QueryRowContext(ctx, query, email)

	var user model.User
	if err := row.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		u.log.Error(ctx, "error: scanning user row", slog.Any("error", err))
		return nil, err
	}

	return &user, nil
}

func (u *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET name = $1, is_active = $2 WHERE id = $3 RETURNING username, email`
	err := u.db.QueryRowContext(ctx, query, user.Name, user.IsActive, user.ID).Scan(&user.Username, &user.Email)
	if err != nil {
		u.log.Error(ctx, "error: updating user", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = timezone('utc', now()) WHERE id = $1`
	_, err := u.db.ExecContext(ctx, query, id)
	if err != nil {
		u.log.Error(ctx, "error: deleting user", slog.Any("error", err))
		return err
	}

	return nil
}
