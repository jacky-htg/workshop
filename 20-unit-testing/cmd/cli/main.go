package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"workshop/config"
	"workshop/internal/router"
	"workshop/pkg/database"

	"github.com/jacky-htg/go-libs/logger"
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
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("Usage: program <command> [arguments...]")
	}

	command := args[0]
	commandArgs := args[1:]

	if err := router.Cli(db, log, command, commandArgs); err != nil {
		return fmt.Errorf("error: executing command: %w", err)
	}

	return nil
}
