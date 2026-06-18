package cache

import (
	"context"
	"time"
	"workshop/config"
)

type CacheClient interface {
	Ping(ctx context.Context) (string, error)
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
	SetWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error
	Del(ctx context.Context, keys []string) (int64, error)
	Close() error

	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}) error
	SetJSONWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error

	SAdd(ctx context.Context, key string, value string) (bool, error)
	SMembers(ctx context.Context, key string) ([]string, error)
	SCard(ctx context.Context, key string) (int64, error)
	Exists(ctx context.Context, key string) (bool, error)
	SRem(ctx context.Context, key string, value ...string) error
	SScan(ctx context.Context, key string, cursor string, defaultBatchSize int) ([]string, string, error)
}

func NewCache(cfg config.CacheConfig) (CacheClient, error) {
	if cfg.ClusterMode {
		var cluster clusterCache
		return cluster.open(cfg)
	}

	var standalone standaloneCache
	return standalone.open(cfg)
}
