package database

import (
    "context"
    "os"
    "time"
    "github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis() error {
    redisAddr := os.Getenv("REDIS_HOST")
    if redisAddr == "" {
        redisAddr = "127.0.0.1:6379"
    }

    RDB = redis.NewClient(&redis.Options{
        Addr:         redisAddr,
        Password:     "", // Mặc định là trống
        DB:           0,  // Database index
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolSize:     10,
    })

    // Kiểm tra kết nối bằng Ping
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    return RDB.Ping(ctx).Err()
}