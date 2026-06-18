package router

import (
	"context"
	"database/sql"
	"fmt"
	"workshop/config"
	"workshop/internal/repository"
	"workshop/internal/service"
	"workshop/pkg/cache"
	"workshop/pkg/listcache"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/migration"
)

func Cli(
	db *sql.DB,
	cache cache.CacheClient,
	log logger.Logger,
	ttl config.TTLConfig,
	command string,
	args []string) error {

	accessRepository := repository.NewAccessRepository(db, log)
	accessService := service.NewAccesses(db, cache, log, ttl, accessRepository)

	switch command {
	case "migrate":
		err := migration.Migrate(db, "migration")
		if err != nil {
			log.Error(context.Background(), "Migration failed", "error", err)
			return err
		}
		log.Info(context.Background(), "Migration completed successfully")
	case "scan-access":
		err := accessService.ScanAccess(context.Background(), "internal/router/api.go")
		if err != nil {
			log.Error(context.Background(), "Scan Access failed", "error", err)
			return err
		}
		log.Info(context.Background(), "Scan access completed successfully")
	case "cleanup-stale-cache":
		listcache.Cleanup(context.Background(), log, cache, args[0])
	default:
		return fmt.Errorf("Error: perintah tidak valid")
	}

	return nil
}
