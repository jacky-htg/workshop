package bootstrap

import (
	"database/sql"
	"fmt"
	"workshop/config"
	"workshop/pkg/database"

	"github.com/jacky-htg/go-libs/logger"
	_ "github.com/lib/pq"
)

type App struct {
	Config   config.Config
	Database *sql.DB
	Log      logger.Logger

	Cleanup func()
}

func NewApp() (App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return App{}, fmt.Errorf("error: loading config: %s", err)
	}

	db, err := database.OpenDB(cfg)
	if err != nil {
		return App{}, fmt.Errorf("error: opening database: %s", err)
	}

	log := logger.InitLogger(nil)

	return App{
		Config:   cfg,
		Database: db,
		Log:      log,
		Cleanup: func() {
			if err := db.Close(); err != nil {
				fmt.Printf("error: closing database: %s\n", err)
			}
		},
	}, nil
}
