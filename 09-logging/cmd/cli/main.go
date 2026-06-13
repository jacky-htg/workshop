package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"workshop/config"
	"workshop/pkg/database"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/migration"
	_ "github.com/lib/pq"
)

func main() {
	log := logger.InitLogger(nil)
	if err := run(log); err != nil {
		log.Debug(context.Background(), "application error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(log logger.Logger) error {

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error: loading config: %w", err)
	}

	db, err := database.OpenDB(cfg)
	if err != nil {
		return fmt.Errorf("error: opening database: %w", err)
	}
	defer db.Close()

	flag.Parse()

	if len(flag.Args()) > 0 && flag.Arg(0) == "migrate" {
		if err := migration.Migrate(db, "migration"); err != nil {
			return fmt.Errorf("error: running migrations: %w", err)
		}
		log.Info(context.Background(), "migrations completed successfully")
	}

	return nil
}
