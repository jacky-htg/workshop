package router

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/migration"
)

func Cli(
	db *sql.DB,
	log logger.Logger,
	command string,
	args []string) error {

	switch command {
	case "migrate":
		err := migration.Migrate(db, "migration")
		if err != nil {
			log.Error(context.Background(), "Migration failed", "error", err)
			return err
		}
		log.Info(context.Background(), "Migration completed successfully")
	default:
		return fmt.Errorf("Error: perintah tidak valid")
	}

	return nil
}
