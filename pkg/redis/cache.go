package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// GetCached retrieves data from Redis cache and unmarshals it into the provided interface
func GetCached[T any](ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (*T, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Key does not exist
			return nil, nil
		}
		log.Printf("Redis GetCached error for key %s: %v", key, err)
		return nil, err
	}

	var result T
	err = json.Unmarshal([]byte(val), &result)
	if err != nil {
		log.Printf("Redis unmarshal error for key %s: %v", key, err)
		return nil, err
	}

	return &result, nil
}

// SetCached marshals data and stores it in Redis with TTL
func SetCached[T any](ctx context.Context, rdb *redis.Client, key string, data T, ttl time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Redis marshal error for key %s: %v", key, err)
		return err
	}

	err = rdb.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		log.Printf("Redis SetCached error for key %s: %v", key, err)
		return err
	}

	return nil
}

// DeleteCached removes a key from Redis
func DeleteCached(ctx context.Context, rdb *redis.Client, key string) error {
	return rdb.Del(ctx, key).Err()
}

// GetTrendingCacheKey generates cache key for trending data
func GetTrendingCacheKey(entityType string, page, limit int) string {
	return fmt.Sprintf("trending:%s:page:%d:limit:%d", entityType, page, limit)
}

// PaginationCacheData is a helper struct to cache pagination along with data
type PaginationCacheData[T any] struct {
	Data       T                `json:"data"`
	Pagination map[string]int64 `json:"pagination"`
}
