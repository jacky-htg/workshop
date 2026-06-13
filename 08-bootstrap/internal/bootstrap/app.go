package bootstrap

import (
	"database/sql"
	"fmt"
	"workshop/config"
	"workshop/pkg/database"

	_ "github.com/lib/pq"
)

type App struct {
	Config   config.Config
	Database *sql.DB

	Cleanup func()
}

func NewApp() (App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return App{}, fmt.Errorf("error: loading config: %w", err)
	}

	db, err := database.OpenDB(cfg)
	if err != nil {
		return App{}, fmt.Errorf("error: opening database: %w", err)
	}

	return App{
		Config:   cfg,
		Database: db,
		Cleanup: func() {
			if err := db.Close(); err != nil {
				fmt.Printf("error: closing database: %s\n", err)
			}
		},
	}, nil
}
