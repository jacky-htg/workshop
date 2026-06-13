package router

import (
	"context"
	"database/sql"
	"fmt"
	"workshop/internal/repository"
	"workshop/internal/service"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/migration"
)

func Cli(
	db *sql.DB,
	log logger.Logger,
	command string,
	args []string) error {

	accessRepository := repository.NewAccessRepository(db, log)
	accessService := service.NewAccesses(db, log, accessRepository)

	switch command {
	case "migrate":
		err := migration.Migrate(db, "migration")
		if err != nil {
			log.Error(context.Background(), "Migration failed", "error", err)
			return err
		}
		log.Info(context.Background(), "Migration completed successfully")
	case "scan-access":
		err := accessService.ScanAccess(context.Background())
		if err != nil {
			log.Error(context.Background(), "Scan Access failed", "error", err)
			return err
		}
		log.Info(context.Background(), "Scan access completed successfully")
	default:
		return fmt.Errorf("Error: perintah tidak valid")
	}

	return nil
}
