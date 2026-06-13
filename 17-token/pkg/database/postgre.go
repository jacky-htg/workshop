package database

import (
	"database/sql"
	"fmt"

	"workshop/config"
)

func OpenDB(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s search_path=%s application_name=%s",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Username,
			cfg.Database.Password,
			cfg.Database.Database,
			cfg.Database.SslMode,
			cfg.Database.Schema,
			cfg.Database.ApplicationName,
		),
	)

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to ping %s DB: %v", cfg.Database.Database, err)
	}
	return db, nil
}
