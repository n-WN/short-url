package service

import (
	"context"
	"errors"
	"fmt"
	"short-url/internal/cache"
	"short-url/internal/config"
	"short-url/internal/models"
	"short-url/internal/utils"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	ErrShortCodeNotFound = errors.New("short code not found")
	ErrShortCodeExists   = errors.New("short code already exists")
	ErrExpiredLink       = errors.New("short link has expired")
	ErrInvalidURL        = errors.New("invalid URL")
)

type ShortLinkService struct {
	repo        *Repository
	cache       *cache.RedisClient
	bloomFilter *cache.BloomFilter
	encoder     *utils.Base62Encoder
	config      *config.Config
	logger      *zap.Logger
}

func NewShortLinkService(
	repo *Repository,
	cache *cache.RedisClient,
	bloomFilter *cache.BloomFilter,
	config *config.Config,
	logger *zap.Logger,
) *ShortLinkService {
	encoder := utils.NewBase62Encoder()
	encoder.SetCodeLength(6) // 设置短码长度为6

	return &ShortLinkService{
		repo:        repo,
		cache:       cache,
		bloomFilter: bloomFilter,
		encoder:     encoder,
		config:      config,
		logger:      logger,
	}
}

// CreateShortLink 创建短链接
func (s *ShortLinkService) CreateShortLink(ctx context.Context, req *models.CreateShortLinkRequest) (*models.CreateShortLinkResponse, error) {
	// 验证URL
	if !utils.IsValidURL(req.URL) {
		return nil, ErrInvalidURL
	}

	// 标准化URL
	normalizedURL := utils.NormalizeURL(req.URL)

	// 生成短码
	var shortCode string
	var err error

	if req.CustomCode != "" {
		// 使用自定义短码
		if !utils.IsValidShortCode(req.CustomCode) {
			return nil, fmt.Errorf("invalid custom code format")
		}
		shortCode = req.CustomCode
	} else {
		// 生成随机短码
		shortCode, err = s.generateUniqueShortCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	// 检查短码是否已存在
	exists, err := s.isShortCodeExists(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to check short code existence: %w", err)
	}
	if exists {
		return nil, ErrShortCodeExists
	}

	// 创建短链接对象
	shortLink := &models.ShortLink{
		ShortCode:   shortCode,
		OriginalURL: normalizedURL,
		ExpiresAt:   req.ExpiresAt,
	}

	// 保存到数据库
	if err := s.repo.CreateShortLink(ctx, shortLink); err != nil {
		return nil, fmt.Errorf("failed to save short link: %w", err)
	}

	// 添加到布隆过滤器
	if err := s.bloomFilter.Add(ctx, shortCode); err != nil {
		s.logger.Warn("failed to add to bloom filter", zap.Error(err))
	}

	// 缓存到Redis
	if err := s.cache.SetWithDefaultTTL(ctx, s.cacheKey(shortCode), normalizedURL); err != nil {
		s.logger.Warn("failed to cache short link", zap.Error(err))
	}

	// 构建响应
	response := &models.CreateShortLinkResponse{
		ShortURL:    s.buildShortURL(shortCode),
		ShortCode:   shortCode,
		OriginalURL: normalizedURL,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   shortLink.CreatedAt,
	}

	return response, nil
}

// GetOriginalURL 获取原始URL并重定向
func (s *ShortLinkService) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	// 首先检查缓存
	originalURL, err := s.cache.Get(ctx, s.cacheKey(shortCode))
	if err == nil {
		// 异步增加访问计数
		go func() {
			if err := s.repo.IncrementAccessCount(context.Background(), shortCode); err != nil {
				s.logger.Error("failed to increment access count", zap.Error(err))
			}
		}()
		return originalURL, nil
	}

	// 缓存未命中，查询数据库
	if err != redis.Nil {
		s.logger.Warn("cache lookup error", zap.Error(err))
	}

	shortLink, err := s.repo.GetShortLinkByCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrShortCodeNotFound
		}
		return "", fmt.Errorf("failed to get short link: %w", err)
	}

	// 检查是否过期
	if shortLink.IsExpired() {
		return "", ErrExpiredLink
	}

	// 更新缓存
	if err := s.cache.SetWithDefaultTTL(ctx, s.cacheKey(shortCode), shortLink.OriginalURL); err != nil {
		s.logger.Warn("failed to update cache", zap.Error(err))
	}

	// 增加访问计数
	if err := s.repo.IncrementAccessCount(ctx, shortCode); err != nil {
		s.logger.Error("failed to increment access count", zap.Error(err))
	}

	return shortLink.OriginalURL, nil
}

// GetShortLinkInfo 获取短链接信息
func (s *ShortLinkService) GetShortLinkInfo(ctx context.Context, shortCode string) (*models.ShortLinkInfo, error) {
	shortLink, err := s.repo.GetShortLinkByCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShortCodeNotFound
		}
		return nil, fmt.Errorf("failed to get short link: %w", err)
	}

	return &models.ShortLinkInfo{
		ShortCode:   shortLink.ShortCode,
		OriginalURL: shortLink.OriginalURL,
		AccessCount: shortLink.AccessCount,
		CreatedAt:   shortLink.CreatedAt,
		ExpiresAt:   shortLink.ExpiresAt,
	}, nil
}

// generateUniqueShortCode 生成唯一的短码
func (s *ShortLinkService) generateUniqueShortCode(ctx context.Context) (string, error) {
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		// 生成随机短码
		shortCode, err := s.encoder.GenerateRandomCode()
		if err != nil {
			return "", err
		}

		// 使用布隆过滤器快速检查
		exists, err := s.bloomFilter.Exists(ctx, shortCode)
		if err != nil {
			s.logger.Warn("bloom filter check failed", zap.Error(err))
			// 如果布隆过滤器失败，直接检查数据库
			dbExists, dbErr := s.repo.ShortCodeExists(ctx, shortCode)
			if dbErr != nil {
				return "", dbErr
			}
			if !dbExists {
				return shortCode, nil
			}
		} else if !exists {
			// 布隆过滤器说不存在，那就不存在
			return shortCode, nil
		}

		// 布隆过滤器说可能存在，需要进一步确认
		dbExists, err := s.repo.ShortCodeExists(ctx, shortCode)
		if err != nil {
			return "", err
		}
		if !dbExists {
			return shortCode, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique short code after %d retries", maxRetries)
}

// isShortCodeExists 检查短码是否存在
func (s *ShortLinkService) isShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	// 首先检查布隆过滤器
	exists, err := s.bloomFilter.Exists(ctx, shortCode)
	if err != nil {
		s.logger.Warn("bloom filter check failed", zap.Error(err))
		// 布隆过滤器失败，直接查数据库
		return s.repo.ShortCodeExists(ctx, shortCode)
	}

	if !exists {
		// 布隆过滤器说不存在，那就确定不存在
		return false, nil
	}

	// 布隆过滤器说可能存在，需要查数据库确认
	return s.repo.ShortCodeExists(ctx, shortCode)
}

// buildShortURL 构建完整的短链接URL
func (s *ShortLinkService) buildShortURL(shortCode string) string {
	return fmt.Sprintf("%s/%s", s.config.App.BaseURL, shortCode)
}

// cacheKey 生成缓存键
func (s *ShortLinkService) cacheKey(shortCode string) string {
	return fmt.Sprintf("shorturl:%s", shortCode)
}

// GetStats 获取统计信息
func (s *ShortLinkService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetStats(ctx)
}

// CleanExpiredLinks 清理过期链接
func (s *ShortLinkService) CleanExpiredLinks(ctx context.Context) (int64, error) {
	deletedCount, err := s.repo.DeleteExpiredLinks(ctx)
	if err != nil {
		return 0, err
	}

	s.logger.Info("cleaned expired links", zap.Int64("count", deletedCount))
	return deletedCount, nil
}
