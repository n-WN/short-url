package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	App         AppConfig         `mapstructure:"app"`
	BloomFilter BloomFilterConfig `mapstructure:"bloom_filter"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit"`
	Cache       CacheConfig       `mapstructure:"cache"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AppConfig struct {
	Port    int    `mapstructure:"port"`
	Env     string `mapstructure:"env"`
	BaseURL string `mapstructure:"base_url"`
}

type BloomFilterConfig struct {
	Key       string  `mapstructure:"key"`
	Capacity  int     `mapstructure:"capacity"`
	ErrorRate float64 `mapstructure:"error_rate"`
}

type RateLimitConfig struct {
	Requests int           `mapstructure:"requests"`
	Window   time.Duration `mapstructure:"window"`
}

type CacheConfig struct {
	TTL time.Duration `mapstructure:"ttl"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/app")

	// Environment variable mappings
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		// If config file not found, we'll rely on environment variables
		fmt.Printf("Config file not found, using environment variables: %v\n", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_NAME", "shorturl")
	viper.SetDefault("DB_SSLMODE", "disable")

	// Redis defaults
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	// App defaults
	viper.SetDefault("APP_PORT", 8080)
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("BASE_URL", "http://localhost:8080")

	// Bloom filter defaults
	viper.SetDefault("BLOOM_FILTER_KEY", "used_short_codes")
	viper.SetDefault("BLOOM_FILTER_CAPACITY", 1000000)
	viper.SetDefault("BLOOM_FILTER_ERROR_RATE", 0.001)

	// Rate limit defaults
	viper.SetDefault("RATE_LIMIT_REQUESTS", 100)
	viper.SetDefault("RATE_LIMIT_WINDOW", "60s")

	// Cache defaults
	viper.SetDefault("CACHE_TTL", "3600s")

	// Bind environment variables
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")

	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	viper.BindEnv("app.port", "APP_PORT")
	viper.BindEnv("app.env", "APP_ENV")
	viper.BindEnv("app.base_url", "BASE_URL")

	viper.BindEnv("bloom_filter.key", "BLOOM_FILTER_KEY")
	viper.BindEnv("bloom_filter.capacity", "BLOOM_FILTER_CAPACITY")
	viper.BindEnv("bloom_filter.error_rate", "BLOOM_FILTER_ERROR_RATE")

	viper.BindEnv("rate_limit.requests", "RATE_LIMIT_REQUESTS")
	viper.BindEnv("rate_limit.window", "RATE_LIMIT_WINDOW")

	viper.BindEnv("cache.ttl", "CACHE_TTL")
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
