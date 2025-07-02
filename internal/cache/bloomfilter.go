package cache

import (
	"context"
	"fmt"
	"short-url/internal/config"
)

type BloomFilter struct {
	redisClient *RedisClient
	config      *config.BloomFilterConfig
}

func NewBloomFilter(redisClient *RedisClient, config *config.BloomFilterConfig) *BloomFilter {
	return &BloomFilter{
		redisClient: redisClient,
		config:      config,
	}
}

// Initialize 初始化布隆过滤器
func (bf *BloomFilter) Initialize(ctx context.Context) error {
	// 检查布隆过滤器是否已存在
	exists, err := bf.redisClient.Exists(ctx, bf.config.Key)
	if err != nil {
		return fmt.Errorf("failed to check bloom filter existence: %w", err)
	}

	if !exists {
		// 创建布隆过滤器
		// BF.RESERVE key error_rate capacity
		cmd := bf.redisClient.GetClient().Do(ctx, "BF.RESERVE", bf.config.Key, bf.config.ErrorRate, bf.config.Capacity)
		if err := cmd.Err(); err != nil {
			return fmt.Errorf("failed to create bloom filter: %w", err)
		}
	}

	return nil
}

// Add 向布隆过滤器添加元素
func (bf *BloomFilter) Add(ctx context.Context, item string) error {
	cmd := bf.redisClient.GetClient().Do(ctx, "BF.ADD", bf.config.Key, item)
	return cmd.Err()
}

// Exists 检查元素是否可能存在于布隆过滤器中
func (bf *BloomFilter) Exists(ctx context.Context, item string) (bool, error) {
	cmd := bf.redisClient.GetClient().Do(ctx, "BF.EXISTS", bf.config.Key, item)
	if err := cmd.Err(); err != nil {
		return false, err
	}

	result, err := cmd.Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// MAdd 批量添加元素
func (bf *BloomFilter) MAdd(ctx context.Context, items []string) ([]bool, error) {
	if len(items) == 0 {
		return []bool{}, nil
	}

	args := make([]interface{}, len(items)+2)
	args[0] = "BF.MADD"
	args[1] = bf.config.Key
	for i, item := range items {
		args[i+2] = item
	}

	cmd := bf.redisClient.GetClient().Do(ctx, args...)
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	results, err := cmd.Slice()
	if err != nil {
		return nil, err
	}

	boolResults := make([]bool, len(results))
	for i, result := range results {
		if val, ok := result.(int64); ok {
			boolResults[i] = val == 1
		}
	}

	return boolResults, nil
}

// MExists 批量检查元素是否存在
func (bf *BloomFilter) MExists(ctx context.Context, items []string) ([]bool, error) {
	if len(items) == 0 {
		return []bool{}, nil
	}

	args := make([]interface{}, len(items)+2)
	args[0] = "BF.MEXISTS"
	args[1] = bf.config.Key
	for i, item := range items {
		args[i+2] = item
	}

	cmd := bf.redisClient.GetClient().Do(ctx, args...)
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	results, err := cmd.Slice()
	if err != nil {
		return nil, err
	}

	boolResults := make([]bool, len(results))
	for i, result := range results {
		if val, ok := result.(int64); ok {
			boolResults[i] = val == 1
		}
	}

	return boolResults, nil
}

// Info 获取布隆过滤器信息
func (bf *BloomFilter) Info(ctx context.Context) (map[string]interface{}, error) {
	cmd := bf.redisClient.GetClient().Do(ctx, "BF.INFO", bf.config.Key)
	if err := cmd.Err(); err != nil {
		return nil, err
	}

	result, err := cmd.Slice()
	if err != nil {
		return nil, err
	}

	info := make(map[string]interface{})
	for i := 0; i < len(result); i += 2 {
		if i+1 < len(result) {
			key := fmt.Sprintf("%v", result[i])
			info[key] = result[i+1]
		}
	}

	return info, nil
}
