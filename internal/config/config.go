package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/kainuguru/kainuguru-api/internal/database"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database database.Config `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	OpenAI   OpenAIConfig   `mapstructure:"openai"`
	Scraper  ScraperConfig  `mapstructure:"scraper"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Auth     AuthConfig     `mapstructure:"auth"`
	App      AppConfig      `mapstructure:"app"`
}

type ServerConfig struct {
	Port                     int           `mapstructure:"port"`
	Host                     string        `mapstructure:"host"`
	ReadTimeout              time.Duration `mapstructure:"read_timeout"`
	WriteTimeout             time.Duration `mapstructure:"write_timeout"`
	IdleTimeout              time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout  time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type OpenAIConfig struct {
	APIKey      string        `mapstructure:"api_key"`
	Model       string        `mapstructure:"model"`
	MaxTokens   int           `mapstructure:"max_tokens"`
	Temperature float64       `mapstructure:"temperature"`
	Timeout     time.Duration `mapstructure:"timeout"`
	MaxRetries  int           `mapstructure:"max_retries"`
}

type ScraperConfig struct {
	UserAgent               string        `mapstructure:"user_agent"`
	RequestTimeout          time.Duration `mapstructure:"request_timeout"`
	MaxConcurrentRequests   int           `mapstructure:"max_concurrent_requests"`
	RateLimitPerMinute      int           `mapstructure:"rate_limit_per_minute"`
	RetryAttempts           int           `mapstructure:"retry_attempts"`
	RetryDelay              time.Duration `mapstructure:"retry_delay"`
}

type WorkerConfig struct {
	MaxWorkers   int           `mapstructure:"max_workers"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
	JobTimeout   time.Duration `mapstructure:"job_timeout"`
	MaxRetries   int           `mapstructure:"max_retries"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposedHeaders   []string `mapstructure:"exposed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

type AuthConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret"`
	JWTExpiresIn    time.Duration `mapstructure:"jwt_expires_in"`
	BCryptCost      int           `mapstructure:"bcrypt_cost"`
	SessionTimeout  time.Duration `mapstructure:"session_timeout"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
}

func Load(env string) (*Config, error) {
	v := viper.New()

	// Set config file
	v.SetConfigName(env)
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	// Environment variable support
	v.SetEnvPrefix("KAINUGURU")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Override with environment variables
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.name", "DB_NAME")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("openai.api_key", "OPENAI_API_KEY")
	v.BindEnv("auth.jwt_secret", "JWT_SECRET")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if cfg.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if cfg.OpenAI.APIKey == "" && cfg.App.Environment != "testing" {
		return fmt.Errorf("OpenAI API key is required")
	}

	return nil
}