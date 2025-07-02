package utils

import (
	"net/url"
	"strings"
)

// IsValidURL 检查 URL 是否有效
func IsValidURL(rawURL string) bool {
	if rawURL == "" {
		return false
	}

	// 如果没有协议前缀，添加 http://
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// 检查是否有 host
	if parsedURL.Host == "" {
		return false
	}

	// 检查协议是否为 http 或 https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	return true
}

// NormalizeURL 标准化 URL
func NormalizeURL(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	// 如果没有协议前缀，添加 http://
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// 移除尾随的斜杠（除非是根路径）
	if len(parsedURL.Path) > 1 && strings.HasSuffix(parsedURL.Path, "/") {
		parsedURL.Path = strings.TrimSuffix(parsedURL.Path, "/")
	}

	return parsedURL.String()
}

// IsValidShortCode 验证短码格式
func IsValidShortCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}

	encoder := NewBase62Encoder()
	return encoder.IsValidCode(code)
}
