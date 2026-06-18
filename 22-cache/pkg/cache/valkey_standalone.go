package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	wcfg "workshop/config"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
	"github.com/valkey-io/valkey-glide/go/v2/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

type standaloneCache struct {
	client *glide.Client
}

func (s *standaloneCache) open(cfg wcfg.CacheConfig) (CacheClient, error) {
	clientConfig := config.NewClientConfiguration().
		WithAddress(&config.NodeAddress{
			Host: cfg.Host,
			Port: cfg.Port,
		}).
		WithRequestTimeout(cfg.DialTimeout)

	if cfg.Password != "" {
		var creds *config.ServerCredentials
		if cfg.Username != "" {
			creds = config.NewServerCredentials(cfg.Username, cfg.Password)
		} else {
			creds = config.NewServerCredentialsWithDefaultUsername(cfg.Password)
		}
		clientConfig = clientConfig.WithCredentials(creds)
	}

	client, err := glide.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Test koneksi
	ctx := context.Background()
	if _, err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return &standaloneCache{client: client}, nil
}

func (s *standaloneCache) Ping(ctx context.Context) (string, error) {
	return s.client.Ping(ctx)
}

func (s *standaloneCache) Set(ctx context.Context, key string, value string) error {
	_, err := s.client.Set(ctx, key, value)
	return err
}

func (s *standaloneCache) SetWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	_, err := s.client.SetWithOptions(ctx, key, value, options.SetOptions{
		Expiry: options.NewExpiryIn(expiry),
	})
	return err
}

func (s *standaloneCache) Get(ctx context.Context, key string) (string, error) {
	result, err := s.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if result.IsNil() {
		return "", nil
	}

	return result.Value(), nil
}

func (s *standaloneCache) Del(ctx context.Context, keys []string) (int64, error) {
	return s.client.Del(ctx, keys)
}

func (s *standaloneCache) Close() error {
	s.client.Close()
	return nil
}

func (s *standaloneCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := s.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}

	return nil
}

func (s *standaloneCache) SetJSON(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	if err := s.Set(ctx, key, string(data)); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

func (s *standaloneCache) SetJSONWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	if err := s.SetWithExpiry(ctx, key, string(data), expiry); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

func (s *standaloneCache) SAdd(ctx context.Context, key, value string) (bool, error) {
	count, err := s.client.SAdd(ctx, key, []string{value})
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (s *standaloneCache) SMembers(ctx context.Context, key string) ([]string, error) {
	mapStr, err := s.client.SMembers(ctx, key)
	if err != nil {
		return nil, err
	}

	members := make([]string, 0, len(mapStr))
	for member := range mapStr {
		members = append(members, member)
	}
	return members, nil
}

func (s *standaloneCache) SCard(ctx context.Context, key string) (int64, error) {
	return s.client.SCard(ctx, key)
}

func (s *standaloneCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.client.Exists(ctx, []string{key})
	if err != nil {
		return false, err
	}

	if count > 0 {
		return false, nil
	}

	return true, nil
}

func (s *standaloneCache) SRem(ctx context.Context, key string, value ...string) error {
	_, err := s.client.SRem(ctx, key, value)
	return err
}

func (s *standaloneCache) SScan(ctx context.Context, key string, cursor string, defaultBatchSize int) ([]string, string, error) {
	cursorModel := models.NewCursorFromString(cursor)
	result, err := s.client.SScan(ctx, key, cursorModel)
	if err != nil {
		return nil, "", fmt.Errorf("failed to SScan key %s: %w", key, err)
	}
	return result.Data, result.Cursor.String(), nil
}
