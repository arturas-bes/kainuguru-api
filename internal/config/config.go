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

	// Load from .env file first if it exists
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	// Try to read .env file, but don't fail if it doesn't exist
	if err := v.ReadInConfig(); err != nil {
		// Try environment-specific .env file
		envFile := fmt.Sprintf(".env.%s", env)
		v.SetConfigFile(envFile)
		if err := v.ReadInConfig(); err != nil {
			// If no .env files, try YAML config files
			v.SetConfigName(env)
			v.SetConfigType("yaml")
			v.AddConfigPath("./configs")
			v.AddConfigPath("../configs")
			v.AddConfigPath("../../configs")

			if err := v.ReadInConfig(); err != nil {
				// If no config files found, continue with environment variables only
				v.SetConfigType("env")
			}
		}
	}

	// Environment variable support with automatic binding
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind all environment variables to config structure
	bindEnvironmentVariables(v)

	// Set defaults
	setDefaults(v)

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Handle comma-separated values for CORS
	if corsOrigins := v.GetString("cors.allowed_origins"); corsOrigins != "" {
		cfg.CORS.AllowedOrigins = strings.Split(corsOrigins, ",")
		for i := range cfg.CORS.AllowedOrigins {
			cfg.CORS.AllowedOrigins[i] = strings.TrimSpace(cfg.CORS.AllowedOrigins[i])
		}
	}
	if corsMethods := v.GetString("cors.allowed_methods"); corsMethods != "" {
		cfg.CORS.AllowedMethods = strings.Split(corsMethods, ",")
		for i := range cfg.CORS.AllowedMethods {
			cfg.CORS.AllowedMethods[i] = strings.TrimSpace(cfg.CORS.AllowedMethods[i])
		}
	}
	if corsHeaders := v.GetString("cors.allowed_headers"); corsHeaders != "" {
		cfg.CORS.AllowedHeaders = strings.Split(corsHeaders, ",")
		for i := range cfg.CORS.AllowedHeaders {
			cfg.CORS.AllowedHeaders[i] = strings.TrimSpace(cfg.CORS.AllowedHeaders[i])
		}
	}
	if corsExposed := v.GetString("cors.exposed_headers"); corsExposed != "" {
		cfg.CORS.ExposedHeaders = strings.Split(corsExposed, ",")
		for i := range cfg.CORS.ExposedHeaders {
			cfg.CORS.ExposedHeaders[i] = strings.TrimSpace(cfg.CORS.ExposedHeaders[i])
		}
	}

	// Validate required fields
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func bindEnvironmentVariables(v *viper.Viper) {
	// Server configuration
	v.BindEnv("server.host", "SERVER_HOST")
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

	// Database configuration
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.name", "DB_NAME")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.ssl_mode", "DB_SSLMODE")
	v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	v.BindEnv("database.max_idle_time", "DB_MAX_IDLE_TIME")

	// Redis configuration
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")
	v.BindEnv("redis.max_retries", "REDIS_MAX_RETRIES")
	v.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")

	// CORS configuration
	v.BindEnv("cors.allowed_origins", "CORS_ALLOWED_ORIGINS")
	v.BindEnv("cors.allowed_methods", "CORS_ALLOWED_METHODS")
	v.BindEnv("cors.allowed_headers", "CORS_ALLOWED_HEADERS")
	v.BindEnv("cors.exposed_headers", "CORS_EXPOSED_HEADERS")
	v.BindEnv("cors.allow_credentials", "CORS_ALLOW_CREDENTIALS")
	v.BindEnv("cors.max_age", "CORS_MAX_AGE")

	// Scraper configuration
	v.BindEnv("scraper.rate_limit_per_minute", "SCRAPER_RATE_LIMIT_PER_MINUTE")
	v.BindEnv("scraper.user_agent", "SCRAPER_USER_AGENT")

	// OpenAI configuration
	v.BindEnv("openai.api_key", "OPENAI_API_KEY")
	v.BindEnv("openai.model", "OPENAI_MODEL")
	v.BindEnv("openai.max_tokens", "OPENAI_MAX_TOKENS")
	v.BindEnv("openai.temperature", "OPENAI_TEMPERATURE")

	// Logging configuration
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.format", "LOG_FORMAT")

	// Auth configuration
	v.BindEnv("auth.jwt_secret", "JWT_SECRET")
	v.BindEnv("auth.session_secret", "SESSION_SECRET")

	// App configuration
	v.BindEnv("app.environment", "APP_ENV")
	v.BindEnv("app.debug", "APP_DEBUG")
}

func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "60s")

	// Database defaults
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 25)
	v.SetDefault("database.max_idle_time", "15m")

	// Redis defaults
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.max_retries", 3)
	v.SetDefault("redis.pool_size", 10)

	// CORS defaults
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"})
	v.SetDefault("cors.exposed_headers", []string{"X-Request-ID"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 86400)

	// Scraper defaults
	v.SetDefault("scraper.rate_limit_per_minute", 60)
	v.SetDefault("scraper.user_agent", "Kainuguru Bot 1.0")

	// OpenAI defaults
	v.SetDefault("openai.model", "gpt-4-vision-preview")
	v.SetDefault("openai.max_tokens", 4000)
	v.SetDefault("openai.temperature", 0.1)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// App defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
}

func validateConfig(cfg *Config) error {
	// Database validation
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	// Redis validation
	if cfg.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	// Auth validation - only required for non-test environments
	if cfg.App.Environment != "test" && cfg.App.Environment != "testing" {
		if cfg.Auth.JWTSecret == "" {
			return fmt.Errorf("JWT secret is required")
		}
	}

	// OpenAI validation - only required for production
	if cfg.App.Environment == "production" {
		if cfg.OpenAI.APIKey == "" {
			return fmt.Errorf("OpenAI API key is required in production")
		}
	}

	// Server validation
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	return nil
}