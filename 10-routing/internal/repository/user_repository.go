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
	query := `SELECT id, name, username, password, email, is_active FROM users`
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
