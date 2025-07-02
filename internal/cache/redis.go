package cache

import (
	"context"
	"short-url/internal/config"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
	config *config.CacheConfig
}

func NewRedisClient(redisConfig *config.RedisConfig, cacheConfig *config.CacheConfig) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Address(),
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	return &RedisClient{
		client: rdb,
		config: cacheConfig,
	}
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) SetWithDefaultTTL(ctx context.Context, key, value string) error {
	return r.Set(ctx, key, value, r.config.TTL)
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	return result == 1, err
}

func (r *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

// GetClient 返回原始的 Redis 客户端，用于执行原生命令
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
