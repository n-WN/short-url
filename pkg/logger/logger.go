package logger

import (
	"short-url/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 创建新的日志器
func New(cfg *config.AppConfig) (*zap.Logger, error) {
	var zapConfig zap.Config

	if cfg.Env == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	logger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return logger, nil
}

// NewNop 创建一个无操作的日志器（用于测试）
func NewNop() *zap.Logger {
	return zap.NewNop()
}
