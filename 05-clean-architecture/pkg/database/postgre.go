package database

import "database/sql"

func OpenDB() (*sql.DB, error) {
	return sql.Open("postgres", "postgres://postgres:1234@localhost:5432/workshop?sslmode=disable")
}
