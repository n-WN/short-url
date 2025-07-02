package models

import (
	"time"
)

// ShortLink 短链接数据模型
type ShortLink struct {
	ID          int64      `json:"id" db:"id"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	AccessCount int64      `json:"access_count" db:"access_count"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// CreateShortLinkRequest 创建短链接请求
type CreateShortLinkRequest struct {
	URL        string     `json:"url" binding:"required,url"`
	CustomCode string     `json:"custom_code,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// CreateShortLinkResponse 创建短链接响应
type CreateShortLinkResponse struct {
	ShortURL    string     `json:"short_url"`
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ShortLinkInfo 短链接信息响应
type ShortLinkInfo struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	AccessCount int64      `json:"access_count"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// IsExpired 检查短链接是否已过期
func (s *ShortLink) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}
