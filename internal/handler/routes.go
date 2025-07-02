package handler

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes 设置路由
func SetupRoutes(handler *Handler, logger *zap.Logger) *gin.Engine {
	// 根据环境设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// 中间件
	r.Use(gin.Recovery())
	r.Use(LoggerMiddleware(logger))
	r.Use(CORSMiddleware())

	// 健康检查
	r.GET("/health", handler.Health)

	// 调试接口
	r.GET("/debug/memory", handler.MemoryStats)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		v1.POST("/shorten", handler.CreateShortLink)
		v1.GET("/info/:code", handler.GetShortLinkInfo)
		v1.GET("/stats", handler.GetStats)

		// 管理员接口
		admin := v1.Group("/admin")
		{
			admin.POST("/clean", handler.CleanExpiredLinks)
		}
	}

	// 短链接重定向（放在最后，避免与API路由冲突）
	r.GET("/:code", handler.RedirectToOriginal)

	return r
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			zap.String("client_ip", param.ClientIP),
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 限流中间件（可选实现）
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里可以实现基于IP或用户的限流逻辑
		// 例如使用Redis存储访问计数和时间窗口
		c.Next()
	}
}
