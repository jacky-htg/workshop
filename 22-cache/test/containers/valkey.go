package containers

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"workshop/config"
	"workshop/pkg/cache"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ValkeyContainer struct {
	Container testcontainers.Container
	Cache     cache.CacheClient
	Host      string
	Port      string
}

func NewValkeyContainer(ctx context.Context) (*ValkeyContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "valkey/valkey:9-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithOccurrence(1).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start valkey container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	portInt, err := strconv.Atoi(port.Port())
	if err != nil {
		return nil, err
	}

	cache, err := cache.NewCache(config.CacheConfig{
		Host: host,
		Port: portInt,
	})
	if err != nil {
		return nil, err
	}

	return &ValkeyContainer{
		Container: container,
		Cache:     cache,
		Host:      host,
		Port:      port.Port(),
	}, nil
}

func (p *ValkeyContainer) Close() error {
	if p.Cache != nil {
		p.Cache.Close()
	}
	return p.Container.Terminate(context.Background())
}
