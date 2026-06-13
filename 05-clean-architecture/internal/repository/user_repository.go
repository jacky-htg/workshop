package repository

import (
	"database/sql"
	"log"
	"workshop/internal/model"
)

type UserRepository interface {
	List() ([]model.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// List : http handler for returning list of users
func (u *userRepository) List() ([]model.User, error) {
	query := `SELECT id, name, username, password, email, is_active FROM users`
	rows, err := u.db.Query(query)
	if err != nil {
		log.Printf("error: querying users: %s", err)
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive); err != nil {
			log.Printf("error: scanning user row: %s", err)
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error: iterating user rows: %s", err)
		return nil, err
	}

	return users, nil
}
