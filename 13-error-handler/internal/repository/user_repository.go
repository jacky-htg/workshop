package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"workshop/internal/model"

	"github.com/jacky-htg/go-libs/logger"
)

type UserRepository interface {
	List() ([]model.User, error)
	Create(*model.User) error
	FindById(id string) (*model.User, error)
	Update(*model.User) error
	Delete(id string) error
}

type userRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewUserRepository(db *sql.DB, log logger.Logger) UserRepository {
	return &userRepository{db: db, log: log}
}

// List : http handler for returning list of users
func (u *userRepository) List() ([]model.User, error) {
	query := `SELECT id, name, username, password, email, is_active FROM users WHERE deleted_at IS NULL`
	rows, err := u.db.Query(query)
	if err != nil {
		u.log.Error(context.Background(), "error: querying users", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive); err != nil {

			u.log.Error(context.Background(), "error: scanning user row", slog.Any("error", err))
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		u.log.Error(context.Background(), "error: iterating user rows", slog.Any("error", err))
		return nil, err
	}

	return users, nil
}

func (u *userRepository) Create(user *model.User) error {
	query := `INSERT INTO users (id, name, username, password, email, is_active) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := u.db.Exec(query, user.ID, user.Name, user.Username, user.Password, user.Email, user.IsActive)
	if err != nil {
		u.log.Error(context.Background(), "error: inserting user", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) FindById(id string) (*model.User, error) {
	query := `SELECT id, name, username, password, email, is_active FROM users WHERE id = $1 AND deleted_at IS NULL`
	row := u.db.QueryRow(query, id)

	var user model.User
	if err := row.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		u.log.Error(context.Background(), "error: scanning user row", slog.Any("error", err))
		return nil, err
	}

	return &user, nil
}

func (u *userRepository) Update(user *model.User) error {
	query := `UPDATE users SET name = $1, is_active = $2 WHERE id = $3 RETURNING username, email`
	err := u.db.QueryRow(query, user.Name, user.IsActive, user.ID).Scan(&user.Username, &user.Email)
	if err != nil {
		u.log.Error(context.Background(), "error: updating user", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) Delete(id string) error {
	query := `UPDATE users SET deleted_at = timezone('utc', now()) WHERE id = $1`
	_, err := u.db.Exec(query, id)
	if err != nil {
		u.log.Error(context.Background(), "error: deleting user", slog.Any("error", err))
		return err
	}

	return nil
}
