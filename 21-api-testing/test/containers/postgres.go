package containers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jacky-htg/go-libs/migration"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgreSQLContainer struct {
	Container testcontainers.Container
	DB        *sql.DB
	Host      string
	Port      string
}

func NewPostgreSQLContainer(ctx context.Context) (*PostgreSQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		host, port.Port())

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Ping database
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgreSQLContainer{
		Container: container,
		DB:        db,
		Host:      host,
		Port:      port.Port(),
	}, nil
}

func (p *PostgreSQLContainer) Close() error {
	if p.DB != nil {
		p.DB.Close()
	}
	return p.Container.Terminate(context.Background())
}

func (p *PostgreSQLContainer) RunMigrations(migrationsPath string) error {
	return migration.Migrate(p.DB, migrationsPath)
}
