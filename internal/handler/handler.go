package handler

import (
	"errors"
	"net/http"
	"runtime"
	"short-url/internal/models"
	"short-url/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	shortLinkService *service.ShortLinkService
	logger           *zap.Logger
}

func NewHandler(shortLinkService *service.ShortLinkService, logger *zap.Logger) *Handler {
	return &Handler{
		shortLinkService: shortLinkService,
		logger:           logger,
	}
}

// CreateShortLink 创建短链接
func (h *Handler) CreateShortLink(c *gin.Context) {
	var req models.CreateShortLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		respondWithError(c, http.StatusBadRequest, "invalid request format")
		return
	}

	response, err := h.shortLinkService.CreateShortLink(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create short link", zap.Error(err))

		switch {
		case errors.Is(err, service.ErrInvalidURL):
			respondWithError(c, http.StatusBadRequest, "invalid URL")
		case errors.Is(err, service.ErrShortCodeExists):
			respondWithError(c, http.StatusConflict, "short code already exists")
		default:
			respondWithError(c, http.StatusInternalServerError, "failed to create short link")
		}
		return
	}

	respondWithSuccess(c, http.StatusCreated, response, "short link created successfully")
}

// RedirectToOriginal 重定向到原始URL
func (h *Handler) RedirectToOriginal(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		respondWithError(c, http.StatusBadRequest, "short code is required")
		return
	}

	originalURL, err := h.shortLinkService.GetOriginalURL(c.Request.Context(), shortCode)
	if err != nil {
		h.logger.Error("failed to get original URL", zap.Error(err), zap.String("short_code", shortCode))

		switch {
		case errors.Is(err, service.ErrShortCodeNotFound):
			respondWithError(c, http.StatusNotFound, "short link not found")
		case errors.Is(err, service.ErrExpiredLink):
			respondWithError(c, http.StatusGone, "short link has expired")
		default:
			respondWithError(c, http.StatusInternalServerError, "failed to resolve short link")
		}
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}

// GetShortLinkInfo 获取短链接信息
func (h *Handler) GetShortLinkInfo(c *gin.Context) {
	shortCode := c.Param("code")
	if shortCode == "" {
		respondWithError(c, http.StatusBadRequest, "short code is required")
		return
	}

	info, err := h.shortLinkService.GetShortLinkInfo(c.Request.Context(), shortCode)
	if err != nil {
		h.logger.Error("failed to get short link info", zap.Error(err), zap.String("short_code", shortCode))

		switch {
		case errors.Is(err, service.ErrShortCodeNotFound):
			respondWithError(c, http.StatusNotFound, "short link not found")
		default:
			respondWithError(c, http.StatusInternalServerError, "failed to get short link info")
		}
		return
	}

	respondWithSuccess(c, http.StatusOK, info)
}

// GetStats 获取统计信息
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.shortLinkService.GetStats(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to get stats", zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, "failed to get statistics")
		return
	}

	respondWithSuccess(c, http.StatusOK, stats)
}

// Health 健康检查
func (h *Handler) Health(c *gin.Context) {
	health := Health{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]interface{}),
	}

	// 这里可以添加对各个服务的健康检查
	// 例如数据库连接、Redis连接等

	respondWithSuccess(c, http.StatusOK, health)
}

// MemoryStats 内存统计接口
func (h *Handler) MemoryStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 手动触发GC
	runtime.GC()

	// 再次读取内存统计
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	stats := map[string]interface{}{
		"before_gc": map[string]interface{}{
			"alloc_mb":       bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":         bToMb(m.Sys),
			"heap_alloc_mb":  bToMb(m.HeapAlloc),
			"heap_sys_mb":    bToMb(m.HeapSys),
			"heap_idle_mb":   bToMb(m.HeapIdle),
			"heap_inuse_mb":  bToMb(m.HeapInuse),
			"stack_inuse_mb": bToMb(m.StackInuse),
			"stack_sys_mb":   bToMb(m.StackSys),
			"num_gc":         m.NumGC,
			"num_goroutine":  runtime.NumGoroutine(),
		},
		"after_gc": map[string]interface{}{
			"alloc_mb":       bToMb(m2.Alloc),
			"total_alloc_mb": bToMb(m2.TotalAlloc),
			"sys_mb":         bToMb(m2.Sys),
			"heap_alloc_mb":  bToMb(m2.HeapAlloc),
			"heap_sys_mb":    bToMb(m2.HeapSys),
			"heap_idle_mb":   bToMb(m2.HeapIdle),
			"heap_inuse_mb":  bToMb(m2.HeapInuse),
			"stack_inuse_mb": bToMb(m2.StackInuse),
			"stack_sys_mb":   bToMb(m2.StackSys),
			"num_gc":         m2.NumGC,
			"num_goroutine":  runtime.NumGoroutine(),
		},
		"gc_triggered": m2.NumGC > m.NumGC,
	}

	respondWithSuccess(c, http.StatusOK, stats)
}

// bToMb 字节转MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// CleanExpiredLinks 清理过期链接（管理员接口）
func (h *Handler) CleanExpiredLinks(c *gin.Context) {
	deletedCount, err := h.shortLinkService.CleanExpiredLinks(c.Request.Context())
	if err != nil {
		h.logger.Error("failed to clean expired links", zap.Error(err))
		respondWithError(c, http.StatusInternalServerError, "failed to clean expired links")
		return
	}

	result := map[string]interface{}{
		"deleted_count": deletedCount,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	respondWithSuccess(c, http.StatusOK, result, "expired links cleaned successfully")
}
