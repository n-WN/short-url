package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"short-url/internal/cache"
	"short-url/internal/config"
	"short-url/internal/database"
	"short-url/internal/handler"
	"short-url/internal/service"
	"short-url/pkg/logger"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志器
	zapLogger, err := logger.New(&cfg.App)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Starting short-url service",
		zap.String("env", cfg.App.Env),
		zap.Int("port", cfg.App.Port),
	)

	// 初始化数据库
	db, err := database.New(&cfg.Database)
	if err != nil {
		zapLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	zapLogger.Info("Database connected successfully")

	// 初始化Redis客户端
	redisClient := cache.NewRedisClient(&cfg.Redis, &cfg.Cache)
	defer redisClient.Close()

	// 测试Redis连接
	if err := redisClient.Ping(context.Background()); err != nil {
		zapLogger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	zapLogger.Info("Redis connected successfully")

	// 初始化布隆过滤器
	bloomFilter := cache.NewBloomFilter(redisClient, &cfg.BloomFilter)
	if err := bloomFilter.Initialize(context.Background()); err != nil {
		zapLogger.Fatal("Failed to initialize bloom filter", zap.Error(err))
	}

	zapLogger.Info("Bloom filter initialized successfully")

	// 初始化服务层
	repo := service.NewRepository(db)
	shortLinkService := service.NewShortLinkService(repo, redisClient, bloomFilter, cfg, zapLogger)

	// 初始化HTTP处理器
	httpHandler := handler.NewHandler(shortLinkService, zapLogger)

	// 设置路由
	router := handler.SetupRoutes(httpHandler, zapLogger)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		zapLogger.Info("Server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Server shutting down...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server shutdown complete")
}
