package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	wcfg "workshop/config"

	glide "github.com/valkey-io/valkey-glide/go/v2"
	"github.com/valkey-io/valkey-glide/go/v2/config"
	"github.com/valkey-io/valkey-glide/go/v2/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

type clusterCache struct {
	client *glide.ClusterClient
}

func (c *clusterCache) open(cfg wcfg.CacheConfig) (CacheClient, error) {
	addresses, err := c.parseAddresses(cfg.ClusterNodes)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, fmt.Errorf("no valid cluster nodes found")
	}

	// Buat config dengan NewClusterClientConfiguration()
	clusterConfig := config.NewClusterClientConfiguration().
		WithRequestTimeout(cfg.DialTimeout)

	// Tambahkan setiap node satu per satu dengan WithAddress()
	for _, addr := range addresses {
		clusterConfig = clusterConfig.WithAddress(&config.NodeAddress{
			Host: addr.Host,
			Port: addr.Port,
		})
	}

	// Set credentials jika ada
	if cfg.Password != "" {
		var creds *config.ServerCredentials
		if cfg.Username != "" {
			creds = config.NewServerCredentials(cfg.Username, cfg.Password)
		} else {
			creds = config.NewServerCredentialsWithDefaultUsername(cfg.Password)
		}
		clusterConfig = clusterConfig.WithCredentials(creds)
	}

	// Optional: set read from strategy
	if cfg.ReadFrom != "" {
		switch cfg.ReadFrom {
		case "PRIMARY":
			clusterConfig = clusterConfig.WithReadFrom(config.Primary)
		case "REPLICA":
			clusterConfig = clusterConfig.WithReadFrom(config.PreferReplica)
		case "ANY":
			clusterConfig = clusterConfig.WithReadFrom(config.AzAffinityReplicaAndPrimary)
		}
	}

	clusterClient, err := glide.NewClusterClient(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster client: %w", err)
	}

	// Test koneksi
	ctx := context.Background()
	if _, err := clusterClient.Ping(ctx); err != nil {
		clusterClient.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return &clusterCache{client: clusterClient}, nil
}

func (c *clusterCache) Ping(ctx context.Context) (string, error) {
	return c.client.Ping(ctx)
}

func (c *clusterCache) Set(ctx context.Context, key string, value string) error {
	_, err := c.client.Set(ctx, key, value)
	return err
}

func (c *clusterCache) SetWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	_, err := c.client.SetWithOptions(ctx, key, value, options.SetOptions{
		Expiry: options.NewExpiryIn(expiry),
	})
	return err
}

func (c *clusterCache) Get(ctx context.Context, key string) (string, error) {
	result, err := c.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if result.IsNil() {
		return "", nil
	}

	return result.Value(), nil
}

func (c *clusterCache) Del(ctx context.Context, keys []string) (int64, error) {
	return c.client.Del(ctx, keys)
}

func (c *clusterCache) Close() error {
	c.client.Close()
	return nil
}

func (c *clusterCache) parseAddresses(nodes []string) ([]config.NodeAddress, error) {
	addresses := make([]config.NodeAddress, 0, len(nodes))
	for _, node := range nodes {
		parts := strings.Split(node, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid node format: %s (expected host:port)", node)
		}
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid port in node: %s", node)
		}
		addresses = append(addresses, config.NodeAddress{
			Host: parts[0],
			Port: port,
		})
	}
	return addresses, nil
}

func (c *clusterCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON for key %s: %w", key, err)
	}

	return nil
}

func (c *clusterCache) SetJSON(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	if err := c.Set(ctx, key, string(data)); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

func (c *clusterCache) SetJSONWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON for key %s: %w", key, err)
	}

	if err := c.SetWithExpiry(ctx, key, string(data), expiry); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

func (c *clusterCache) SAdd(ctx context.Context, key, value string) (bool, error) {
	count, err := c.client.SAdd(ctx, key, []string{value})
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (c *clusterCache) SMembers(ctx context.Context, key string) ([]string, error) {
	mapStr, err := c.client.SMembers(ctx, key)
	if err != nil {
		return nil, err
	}

	members := make([]string, 0, len(mapStr))
	for member := range mapStr {
		members = append(members, member)
	}
	return members, nil
}

func (c *clusterCache) SCard(ctx context.Context, key string) (int64, error) {
	return c.client.SCard(ctx, key)
}

func (c *clusterCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.Exists(ctx, []string{key})
	if err != nil {
		return false, err
	}

	if count > 0 {
		return false, nil
	}

	return true, nil
}

func (c *clusterCache) SRem(ctx context.Context, key string, value ...string) error {
	_, err := c.client.SRem(ctx, key, value)
	return err
}

func (c *clusterCache) SScan(ctx context.Context, key string, cursor string, defaultBatchSize int) ([]string, string, error) {
	cursorModel := models.NewCursorFromString(cursor)
	result, err := c.client.SScan(ctx, key, cursorModel)
	if err != nil {
		return nil, "", fmt.Errorf("failed to SScan key %s: %w", key, err)
	}
	return result.Data, result.Cursor.String(), nil
}
