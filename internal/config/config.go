package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/database"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig    `mapstructure:"server"`
	Database database.Config `mapstructure:"database"`
	Redis    RedisConfig     `mapstructure:"redis"`
	Logging  LoggingConfig   `mapstructure:"logging"`
	OpenAI   OpenAIConfig    `mapstructure:"openai"`
	Scraper  ScraperConfig   `mapstructure:"scraper"`
	Worker   WorkerConfig    `mapstructure:"worker"`
	CORS     CORSConfig      `mapstructure:"cors"`
	Auth     AuthConfig      `mapstructure:"auth"`
	App      AppConfig       `mapstructure:"app"`
	Email    EmailConfig     `mapstructure:"email"`
	Storage  StorageConfig   `mapstructure:"storage"`
}

type ServerConfig struct {
	Port                    int           `mapstructure:"port"`
	Host                    string        `mapstructure:"host"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
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
	BaseURL     string        `mapstructure:"base_url"`
	Model       string        `mapstructure:"model"`
	MaxTokens   int           `mapstructure:"max_tokens"`
	Temperature float64       `mapstructure:"temperature"`
	Timeout     time.Duration `mapstructure:"timeout"`
	MaxRetries  int           `mapstructure:"max_retries"`
}

type ScraperConfig struct {
	UserAgent             string        `mapstructure:"user_agent"`
	RequestTimeout        time.Duration `mapstructure:"request_timeout"`
	RequestDelay          time.Duration `mapstructure:"request_delay"`
	MaxConcurrentRequests int           `mapstructure:"max_concurrent_requests"`
	RateLimitPerMinute    int           `mapstructure:"rate_limit_per_minute"`
	MaxRetries            int           `mapstructure:"max_retries"`
	RetryAttempts         int           `mapstructure:"retry_attempts"`
	RetryDelay            time.Duration `mapstructure:"retry_delay"`
	RespectRobotsTxt      bool          `mapstructure:"respect_robots_txt"`
}

type WorkerConfig struct {
	NumWorkers         int           `mapstructure:"num_workers"`
	MaxWorkers         int           `mapstructure:"max_workers"`
	QueueCheckInterval time.Duration `mapstructure:"queue_check_interval"`
	PollInterval       time.Duration `mapstructure:"poll_interval"`
	JobTimeout         time.Duration `mapstructure:"job_timeout"`
	MaxRetryAttempts   int           `mapstructure:"max_retry_attempts"`
	MaxRetries         int           `mapstructure:"max_retries"`
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
	JWTSecret      string        `mapstructure:"jwt_secret"`
	JWTExpiresIn   time.Duration `mapstructure:"jwt_expires_in"`
	BCryptCost     int           `mapstructure:"bcrypt_cost"`
	SessionTimeout time.Duration `mapstructure:"session_timeout"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	Debug       bool   `mapstructure:"debug"`
	BaseURL     string `mapstructure:"base_url"`
}

type EmailConfig struct {
	Provider  string     `mapstructure:"provider"` // "smtp" or "mock"
	SMTP      SMTPConfig `mapstructure:"smtp"`
	FromEmail string     `mapstructure:"from_email"`
	FromName  string     `mapstructure:"from_name"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

type StorageConfig struct {
	Type         string `mapstructure:"type"`           // "filesystem" or "s3"
	BasePath     string `mapstructure:"base_path"`      // Local filesystem path
	PublicURL    string `mapstructure:"public_url"`     // Public URL base
	FlyerBaseURL string `mapstructure:"flyer_base_url"` // Base URL for flyer images (can be changed per environment)
	MaxRetries   int    `mapstructure:"max_retries"`    // Retry attempts for file operations
}

func Load(env string) (*Config, error) {
	v := viper.New()

	// First, load YAML config for application settings
	v.SetConfigName(env)
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	// Read YAML config (application settings like timeouts, retry counts)
	if err := v.ReadInConfig(); err != nil {
		// YAML config is optional, continue without it
		v.SetConfigType("env")
	}

	// Then, load .env file for infrastructure settings (this will override YAML where applicable)
	v.SetConfigFile(".env")
	v.SetConfigType("env")

	// Merge .env into existing config
	if err := v.MergeInConfig(); err != nil {
		// Try environment-specific .env file
		envFile := fmt.Sprintf(".env.%s", env)
		v.SetConfigFile(envFile)
		v.MergeInConfig() // Ignore error - .env files are optional
	}

	// Environment variable support with automatic binding (highest priority)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind all environment variables to config structure
	bindEnvironmentVariables(v)

	// Set defaults (lowest priority)
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
	v.BindEnv("openai.base_url", "OPENAI_BASE_URL")
	v.BindEnv("openai.model", "OPENAI_MODEL")
	v.BindEnv("openai.max_tokens", "OPENAI_MAX_TOKENS")
	v.BindEnv("openai.temperature", "OPENAI_TEMPERATURE")
	v.BindEnv("openai.timeout", "OPENAI_TIMEOUT")
	v.BindEnv("openai.max_retries", "OPENAI_MAX_RETRIES")

	// Logging configuration
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.format", "LOG_FORMAT")

	// Auth configuration
	v.BindEnv("auth.jwt_secret", "JWT_SECRET")
	v.BindEnv("auth.session_secret", "SESSION_SECRET")

	// App configuration
	v.BindEnv("app.environment", "APP_ENV")
	v.BindEnv("app.debug", "APP_DEBUG")
	v.BindEnv("app.base_url", "APP_BASE_URL")

	// Email configuration
	v.BindEnv("email.provider", "EMAIL_PROVIDER")
	v.BindEnv("email.from_email", "EMAIL_FROM")
	v.BindEnv("email.from_name", "EMAIL_FROM_NAME")
	v.BindEnv("email.smtp.host", "SMTP_HOST")
	v.BindEnv("email.smtp.port", "SMTP_PORT")
	v.BindEnv("email.smtp.username", "SMTP_USERNAME")
	v.BindEnv("email.smtp.password", "SMTP_PASSWORD")
	v.BindEnv("email.smtp.use_tls", "SMTP_USE_TLS")

	// Storage configuration
	v.BindEnv("storage.type", "STORAGE_TYPE")
	v.BindEnv("storage.base_path", "STORAGE_BASE_PATH")
	v.BindEnv("storage.public_url", "STORAGE_PUBLIC_URL")
	v.BindEnv("storage.flyer_base_url", "FLYER_BASE_URL")
	v.BindEnv("storage.max_retries", "STORAGE_MAX_RETRIES")
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
	v.SetDefault("cors.allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"})
	v.SetDefault("cors.exposed_headers", []string{"X-Request-ID"})
	v.SetDefault("cors.allow_credentials", true)
	v.SetDefault("cors.max_age", 86400)

	// Scraper defaults
	v.SetDefault("scraper.rate_limit_per_minute", 60)
	v.SetDefault("scraper.user_agent", "Kainuguru Bot 1.0")

	// OpenAI defaults
	v.SetDefault("openai.base_url", "https://api.openai.com/v1")
	v.SetDefault("openai.model", "gpt-4o")
	v.SetDefault("openai.max_tokens", 4000)
	v.SetDefault("openai.temperature", 0.1)
	v.SetDefault("openai.timeout", "120s")
	v.SetDefault("openai.max_retries", 3)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// App defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.base_url", "http://localhost:3000")

	// Email defaults
	v.SetDefault("email.provider", "mock") // Use mock by default in development
	v.SetDefault("email.from_name", "Kainuguru")
	v.SetDefault("email.from_email", "noreply@kainuguru.lt")
	v.SetDefault("email.smtp.port", 587)
	v.SetDefault("email.smtp.use_tls", true)

	// Storage defaults
	v.SetDefault("storage.type", "filesystem")
	v.SetDefault("storage.base_path", "../kainuguru-public")
	v.SetDefault("storage.public_url", "http://localhost:8080")
	v.SetDefault("storage.flyer_base_url", "http://localhost:8080")
	v.SetDefault("storage.max_retries", 3)
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

	// Redis validation - only for main API server, not for commands
	// Commands may not need Redis, so skip validation

	// Auth validation - only required for non-test environments
	if cfg.App.Environment != "test" && cfg.App.Environment != "testing" {
		if cfg.Auth.JWTSecret == "" {
			// Auth is only required for main API server, not commands
			// Skip validation
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
