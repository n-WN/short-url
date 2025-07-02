package service

import (
	"context"
	"database/sql"
	"fmt"
	"short-url/internal/database"
	"short-url/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// CreateShortLink 创建短链接
func (r *Repository) CreateShortLink(ctx context.Context, shortLink *models.ShortLink) error {
	query := `
		INSERT INTO short_links (short_code, original_url, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.Pool.QueryRow(ctx, query, shortLink.ShortCode, shortLink.OriginalURL, shortLink.ExpiresAt).
		Scan(&shortLink.ID, &shortLink.CreatedAt, &shortLink.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create short link: %w", err)
	}

	return nil
}

// GetShortLinkByCode 根据短码获取短链接
func (r *Repository) GetShortLinkByCode(ctx context.Context, shortCode string) (*models.ShortLink, error) {
	query := `
		SELECT id, short_code, original_url, access_count, created_at, updated_at, expires_at
		FROM short_links
		WHERE short_code = $1
	`

	shortLink := &models.ShortLink{}
	err := r.db.Pool.QueryRow(ctx, query, shortCode).Scan(
		&shortLink.ID,
		&shortLink.ShortCode,
		&shortLink.OriginalURL,
		&shortLink.AccessCount,
		&shortLink.CreatedAt,
		&shortLink.UpdatedAt,
		&shortLink.ExpiresAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("short link not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get short link: %w", err)
	}

	return shortLink, nil
}

// ShortCodeExists 检查短码是否存在
func (r *Repository) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM short_links WHERE short_code = $1)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, shortCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if short code exists: %w", err)
	}

	return exists, nil
}

// IncrementAccessCount 增加访问次数
func (r *Repository) IncrementAccessCount(ctx context.Context, shortCode string) error {
	query := `
		UPDATE short_links 
		SET access_count = access_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE short_code = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to increment access count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("short link not found")
	}

	return nil
}

// GetShortLinksByTimeRange 根据时间范围获取短链接列表
func (r *Repository) GetShortLinksByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*models.ShortLink, error) {
	query := `
		SELECT id, short_code, original_url, access_count, created_at, updated_at, expires_at
		FROM short_links
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Pool.Query(ctx, query, start, end, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get short links: %w", err)
	}
	defer rows.Close()

	var shortLinks []*models.ShortLink
	for rows.Next() {
		shortLink := &models.ShortLink{}
		err := rows.Scan(
			&shortLink.ID,
			&shortLink.ShortCode,
			&shortLink.OriginalURL,
			&shortLink.AccessCount,
			&shortLink.CreatedAt,
			&shortLink.UpdatedAt,
			&shortLink.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan short link: %w", err)
		}
		shortLinks = append(shortLinks, shortLink)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return shortLinks, nil
}

// DeleteExpiredLinks 删除过期的短链接
func (r *Repository) DeleteExpiredLinks(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM short_links
		WHERE expires_at IS NOT NULL AND expires_at < CURRENT_TIMESTAMP
	`

	result, err := r.db.Pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired links: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetStats 获取统计信息
func (r *Repository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_links,
			SUM(access_count) as total_accesses,
			COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at > CURRENT_TIMESTAMP) as active_links,
			COUNT(*) FILTER (WHERE expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP) as expired_links
		FROM short_links
	`

	var totalLinks, totalAccesses int64
	var activeLinks, expiredLinks sql.NullInt64

	err := r.db.Pool.QueryRow(ctx, query).Scan(&totalLinks, &totalAccesses, &activeLinks, &expiredLinks)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_links":     totalLinks,
		"total_accesses":  totalAccesses,
		"active_links":    activeLinks.Int64,
		"expired_links":   expiredLinks.Int64,
		"permanent_links": totalLinks - activeLinks.Int64 - expiredLinks.Int64,
	}

	return stats, nil
}
