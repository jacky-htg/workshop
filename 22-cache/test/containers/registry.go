package containers

import (
	"context"
	"sync"
)

type ContainerRegistry struct {
	mu       sync.RWMutex
	postgres *PostgreSQLContainer
	valkey   *ValkeyContainer
}

var (
	registry *ContainerRegistry
	once     sync.Once
)

func GetRegistry() *ContainerRegistry {
	once.Do(func() {
		registry = &ContainerRegistry{}
	})
	return registry
}

func (r *ContainerRegistry) StartPostgres(ctx context.Context) (*PostgreSQLContainer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.postgres == nil {
		pg, err := NewPostgreSQLContainer(ctx)
		if err != nil {
			return nil, err
		}
		r.postgres = pg
	}
	return r.postgres, nil
}

func (r *ContainerRegistry) StartValkey(ctx context.Context) (*ValkeyContainer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.valkey == nil {
		vk, err := NewValkeyContainer(ctx)
		if err != nil {
			return nil, err
		}
		r.valkey = vk
	}
	return r.valkey, nil
}

func (r *ContainerRegistry) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.postgres != nil {
		r.postgres.Close()
	}

	if r.valkey != nil {
		r.valkey.Close()
	}
}
