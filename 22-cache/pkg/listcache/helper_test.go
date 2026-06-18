package listcache_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"workshop/mock/mockpkg"
	"workshop/pkg/listcache"

	"github.com/stretchr/testify/assert"
)

func TestGenerateListCacheKey(t *testing.T) {
	prefixKey := "users:"
	order := "name"
	sort := "asc"
	search := "john"
	limit := 10
	page := 1

	result := listcache.GenerateListCacheKey(prefixKey, order, sort, search, limit, page)
	expected := "users:order:name::sort:asc::search:john::limit:10::page:1"

	assert.Equal(t, expected, result)
}

func TestGenerateListCacheKey_EmptySearch(t *testing.T) {
	result := listcache.GenerateListCacheKey("roles:", "id", "desc", "", 20, 2)
	expected := "roles:order:id::sort:desc::search::limit:20::page:2"

	assert.Equal(t, expected, result)
}

func TestInvalidateListCache_Success(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SMembersFunc: func(ctx context.Context, key string) ([]string, error) {
			return []string{"key1", "key2", "key3"}, nil
		},
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return int64(len(keys)), nil
		},
	}

	err := listcache.InvalidateListCache(ctx, log, mockClient, "index:key")
	assert.NoError(t, err)
}

func TestInvalidateListCache_Success_WithDelError(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SMembersFunc: func(ctx context.Context, key string) ([]string, error) {
			return []string{"key1", "key2", "key3"}, nil
		},
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 0, fmt.Errorf("error del cachce")
		},
	}

	err := listcache.InvalidateListCache(ctx, log, mockClient, "index:key")
	assert.NoError(t, err)
}

func TestInvalidateListCache_EmptyKeys(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SMembersFunc: func(ctx context.Context, key string) ([]string, error) {
			return []string{}, nil
		},
	}

	err := listcache.InvalidateListCache(ctx, log, mockClient, "index:key")
	assert.NoError(t, err)
}

func TestInvalidateListCache_SMembersError(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SMembersFunc: func(ctx context.Context, key string) ([]string, error) {
			return nil, errors.New("redis error")
		},
	}

	err := listcache.InvalidateListCache(ctx, log, mockClient, "index:key")
	assert.Error(t, err)
}

func TestAddKeyToIndex_Success(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SAddFunc: func(ctx context.Context, key string, value string) (bool, error) {
			return true, nil
		},
		SCardFunc: func(ctx context.Context, key string) (int64, error) {
			return 5, nil
		},
	}

	err := listcache.AddKeyToIndex(ctx, log, mockClient, "cache:key", "index:key", 100)
	assert.NoError(t, err)
}

func TestAddKeyToIndex_KeyAlreadyExists(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SAddFunc: func(ctx context.Context, key string, value string) (bool, error) {
			return false, nil
		},
	}

	err := listcache.AddKeyToIndex(ctx, log, mockClient, "cache:key", "index:key", 100)
	assert.NoError(t, err)
}

func TestAddKeyToIndex_SAddError(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SAddFunc: func(ctx context.Context, key string, value string) (bool, error) {
			return false, errors.New("sadd error")
		},
	}

	err := listcache.AddKeyToIndex(ctx, log, mockClient, "cache:key", "index:key", 100)
	assert.Error(t, err)
}

func TestAddKeyToIndex_SCardError(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SAddFunc: func(ctx context.Context, key string, value string) (bool, error) {
			return true, nil
		},
		SCardFunc: func(ctx context.Context, key string) (int64, error) {
			return 0, fmt.Errorf("scard error")
		},
	}

	err := listcache.AddKeyToIndex(ctx, log, mockClient, "cache:key", "index:key", 100)
	assert.Error(t, err)
}

func TestAddKeyToIndex_TriggersCleanup(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SAddFunc: func(ctx context.Context, key string, value string) (bool, error) {
			return true, nil
		},
		SCardFunc: func(ctx context.Context, key string) (int64, error) {
			return 150, nil
		},
		SScanFunc: func(ctx context.Context, key, cursor string, count int) ([]string, string, error) {
			return []string{"key1", "key2"}, "0", nil
		},
		ExistsFunc: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
		SRemFunc: func(ctx context.Context, key string, members ...string) error {
			return nil
		},
	}

	err := listcache.AddKeyToIndex(ctx, log, mockClient, "cache:key", "index:key", 100)
	assert.NoError(t, err)
}

func TestCleanup_Success(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SScanFunc: func(ctx context.Context, key, cursor string, count int) ([]string, string, error) {
			return []string{"key1", "key2"}, "0", nil
		},
		ExistsFunc: func(ctx context.Context, key string) (bool, error) {
			return false, nil
		},
		SRemFunc: func(ctx context.Context, key string, members ...string) error {
			return fmt.Errorf("error srem")
		},
		SCardFunc: func(ctx context.Context, key string) (int64, error) {
			return 0, nil
		},
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 1, nil
		},
	}

	listcache.Cleanup(ctx, log, mockClient, "index:key")
}

func TestCleanup_Success_ExistError_totalChecked(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	var mycursor int = 1
	mockClient := &mockpkg.MockCache{
		SScanFunc: func(ctx context.Context, key, cursor string, count int) ([]string, string, error) {
			var keys []string
			for i := 1; i <= 50; i++ {
				keys = append(keys, fmt.Sprintf("key%d", i))
			}

			currCursor := mycursor
			if mycursor > 0 {
				mycursor = mycursor - 1
			}

			return keys, fmt.Sprintf("%d", currCursor), nil
		},
		ExistsFunc: func(ctx context.Context, key string) (bool, error) {
			return false, fmt.Errorf("error exist")
		},
		SRemFunc: func(ctx context.Context, key string, members ...string) error {
			return nil
		},
		SCardFunc: func(ctx context.Context, key string) (int64, error) {
			return 0, nil
		},
		DelFunc: func(ctx context.Context, keys []string) (int64, error) {
			return 1, nil
		},
	}

	listcache.Cleanup(ctx, log, mockClient, "index:key")
}

func TestCleanup_SScanError(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SScanFunc: func(ctx context.Context, key, cursor string, count int) ([]string, string, error) {
			return nil, "0", errors.New("sscan error")
		},
	}

	listcache.Cleanup(ctx, log, mockClient, "index:key")
}

func TestCleanup_WithExistingKeys(t *testing.T) {
	ctx := context.Background()
	log := &mockpkg.MockLogger{}
	mockClient := &mockpkg.MockCache{
		SScanFunc: func(ctx context.Context, key, cursor string, count int) ([]string, string, error) {
			return []string{"key1", "key2", "key3"}, "0", nil
		},
		ExistsFunc: func(ctx context.Context, key string) (bool, error) {
			return key == "key1", nil
		},
		SRemFunc: func(ctx context.Context, key string, members ...string) error {
			return nil
		},
		SCardFunc: func(ctx context.Context, key string) (int64, error) {
			return 1, nil
		},
	}

	listcache.Cleanup(ctx, log, mockClient, "index:key")
}
