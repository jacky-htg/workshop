package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"workshop/config"
	"workshop/pkg/cache"
	"workshop/pkg/database"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
	_ "github.com/lib/pq"
)

type App struct {
	Config   config.Config
	Database *sql.DB
	Log      logger.Logger
	Validate *validator.Validate
	Cache    cache.CacheClient

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

	log := logger.InitLogger(nil)
	validate := validator.New()

	cache, err := cache.NewCache(cfg.Cache)
	if err != nil {
		return App{}, fmt.Errorf("error: opening cache: %w", err)
	}

	return App{
		Config:   cfg,
		Database: db,
		Log:      log,
		Validate: validate,
		Cache:    cache,
		Cleanup: func() {
			if err := db.Close(); err != nil {
				log.Info(context.Background(), "error: closing database: %s\n", err)
			}

			if err := cache.Close(); err != nil {
				log.Info(context.Background(), "error: closing cache: %s\n", err)
			}
		},
	}, nil
}
