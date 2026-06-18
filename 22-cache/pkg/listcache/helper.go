package listcache

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"workshop/pkg/cache"

	"github.com/jacky-htg/go-libs/logger"
)

const (
	defaultBatchSize = 100 // Jumlah key yang diproses per siklus SSCAN
)

func GenerateListCacheKey(prefixKey, order, sort, search string, limit, page int) string {
	return fmt.Sprintf("%sorder:%s::sort:%s::search:%s::limit:%d::page:%d",
		prefixKey, order, sort, search, limit, page)
}

func InvalidateListCache(ctx context.Context, log logger.Logger, client cache.CacheClient, indexKey string) error {
	// 1. Get semua key dari index
	keys, err := client.SMembers(ctx, indexKey)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// 2. Delete semua key (batch)
	batchSize := 100
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if _, err := client.Del(ctx, batch); err != nil {
			log.Error(ctx, "failed to delete cache batch",
				slog.Any("error", err),
				slog.Int("batch", i/batchSize))
		}
	}

	// 3. Hapus index (reset)
	if _, err := client.Del(ctx, []string{indexKey}); err != nil {
		log.Error(ctx, "failed to delete index", slog.Any("error", err))
	}

	log.Info(ctx, "invalidated list cache", slog.Int("count", len(keys)))
	return nil
}

func AddKeyToIndex(ctx context.Context, log logger.Logger, client cache.CacheClient, cacheKey, indexKey string, maxIndexSize int) error {
	added, err := client.SAdd(ctx, indexKey, cacheKey)
	if err != nil {
		return err
	}

	if !added {
		return nil
	}

	size, err := client.SCard(ctx, indexKey)
	if err != nil {
		return err
	}

	// Jika index terlalu besar, hapus yang paling tua
	if size > int64(maxIndexSize) {
		cleanupStaleIndexBatch(ctx, log, client, indexKey)
	}

	return nil
}

func Cleanup(ctx context.Context, log logger.Logger, client cache.CacheClient, indexKey string) {
	cleanupStaleIndexBatch(ctx, log, client, indexKey)
}

func cleanupStaleIndexBatch(ctx context.Context, log logger.Logger, client cache.CacheClient, indexKey string) {
	startTime := time.Now()
	totalRemoved := 0
	totalChecked := 0
	cursor := "0" // SSCAN menggunakan string cursor

	log.Debug(ctx, "starting cleanup batch", slog.String("index_key", indexKey))

	for {
		// 1. Gunakan SSCAN untuk iterasi bertahap
		keys, nextCursor, err := client.SScan(ctx, indexKey, cursor, defaultBatchSize)
		if err != nil {
			log.Error(ctx, "failed to scan index",
				slog.String("index_key", indexKey),
				slog.Any("error", err))
			return
		}

		totalChecked += len(keys)

		// 2. Proses keys dalam batch kecil
		if len(keys) > 0 {
			staleKeys := make([]string, 0)

			for _, key := range keys {
				exists, err := client.Exists(ctx, key)
				if err != nil {
					log.Warn(ctx, "failed to check existence",
						slog.String("key", key),
						slog.Any("error", err))
					continue
				}

				if !exists {
					staleKeys = append(staleKeys, key)
				}
			}

			// 3. Hapus stale keys dari index (batch kecil)
			if len(staleKeys) > 0 {
				// Hapus dalam batch kecil (10 keys per batch)
				for i := 0; i < len(staleKeys); i += 10 {
					end := i + 10
					if end > len(staleKeys) {
						end = len(staleKeys)
					}
					err = client.SRem(ctx, indexKey, staleKeys[i:end]...)
					if err != nil {
						log.Debug(ctx, "error srem", slog.Any("error", err))
					}
				}
				totalRemoved += len(staleKeys)
			}

			// 4. Log progress setiap 50 keys
			if totalChecked%50 == 0 && totalChecked > 0 {
				log.Debug(ctx, "cleanup progress",
					slog.Int("checked", totalChecked),
					slog.Int("removed", totalRemoved))
			}
		}

		// 5. Cek cursor - jika "0" berarti selesai
		if nextCursor == "0" {
			break
		}
		cursor = nextCursor
	}

	// 6. Jika semua stale sudah dibersihkan dan index kosong, hapus index key
	if totalChecked > 0 && totalRemoved > 0 {
		remaining, err := client.SCard(ctx, indexKey)
		if err == nil && remaining == 0 {
			client.Del(ctx, []string{indexKey})
			log.Info(ctx, "index emptied, removed key",
				slog.String("index_key", indexKey))
		}
	}

	elapsed := time.Since(startTime)
	if totalChecked > 0 {
		log.Info(ctx, "cleanup batch completed",
			slog.Int("checked", totalChecked),
			slog.Int("removed", totalRemoved),
			slog.Duration("duration", elapsed))
	}
}
