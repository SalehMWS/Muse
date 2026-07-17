package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvTesting     Environment = "testing"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

type Config struct {
	App        App
	HTTP       HTTP
	Log        Log
	Postgres   Postgres
	Redis      Redis
	Migrations Migrations
	JWT        JWT
	Argon2     Argon2
	Instagram  Instagram
	AI         AI
	Scheduler  Scheduler
}

type App struct {
	Name string
	Env  Environment
}

func (a App) IsProduction() bool {
	return a.Env == EnvProduction
}

type HTTP struct {
	Port            int
	ShutdownTimeout time.Duration
}

type Log struct {
	Level string
}

type Postgres struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	ConnMaxLifetime time.Duration
}

func (p Postgres) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode,
	)
}

type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Migrations struct {
	Dir string
}

type JWT struct {
	Secret          string
	Issuer          string
	Audience        string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Argon2 struct {
	Memory      uint32
	Time        uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Instagram struct {
	ClientID           string
	ClientSecret       string
	RedirectURI        string
	Scopes             string
	AuthBaseURL        string
	APIBaseURL         string
	GraphBaseURL       string
	TokenEncryptionKey string
	StateSecret        string
	StateTTL           time.Duration
	HTTPTimeout        time.Duration
}

type AI struct {
	Provider    string
	BaseURL     string
	APIKey      string
	Model       string
	MaxTokens   int
	HTTPTimeout time.Duration
}

type Scheduler struct {
	PollInterval time.Duration
}

func (r Redis) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func Load() (*Config, error) {
	_ = godotenv.Load("configs/.env")

	env := Environment(getEnv("APP_ENV", string(EnvDevelopment)))

	httpPort, err := getEnvInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	shutdownTimeout, err := getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	pgPort, err := getEnvInt("POSTGRES_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	pgMaxConns, err := getEnvInt32("POSTGRES_MAX_CONNS", 10)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	pgMinConns, err := getEnvInt32("POSTGRES_MIN_CONNS", 2)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	pgConnMaxLifetime, err := getEnvDuration("POSTGRES_CONN_MAX_LIFETIME", time.Hour)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	redisPort, err := getEnvInt("REDIS_PORT", 6379)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	redisDB, err := getEnvInt("REDIS_DB", 0)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	jwtAccessTTL, err := getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	jwtRefreshTTL, err := getEnvDuration("JWT_REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	argon2Memory, err := getEnvUint32("ARGON2_MEMORY", 65536)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	argon2Time, err := getEnvUint32("ARGON2_TIME", 3)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	argon2Parallelism, err := getEnvUint8("ARGON2_PARALLELISM", 2)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	argon2SaltLength, err := getEnvUint32("ARGON2_SALT_LENGTH", 16)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	argon2KeyLength, err := getEnvUint32("ARGON2_KEY_LENGTH", 32)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	instagramStateTTL, err := getEnvDuration("INSTAGRAM_STATE_TTL", 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	instagramHTTPTimeout, err := getEnvDuration("INSTAGRAM_HTTP_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	aiMaxTokens, err := getEnvInt("AI_MAX_TOKENS", 1024)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	aiHTTPTimeout, err := getEnvDuration("AI_HTTP_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	schedulerPollInterval, err := getEnvDuration("SCHEDULER_POLL_INTERVAL", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	cfg := &Config{
		App: App{
			Name: getEnv("APP_NAME", "novaflow"),
			Env:  env,
		},
		HTTP: HTTP{
			Port:            httpPort,
			ShutdownTimeout: shutdownTimeout,
		},
		Log: Log{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Postgres: Postgres{
			Host:            getEnv("POSTGRES_HOST", "localhost"),
			Port:            pgPort,
			User:            getEnv("POSTGRES_USER", "novaflow"),
			Password:        getEnv("POSTGRES_PASSWORD", ""),
			Database:        getEnv("POSTGRES_DB", "novaflow"),
			SSLMode:         getEnv("POSTGRES_SSLMODE", "disable"),
			MaxConns:        pgMaxConns,
			MinConns:        pgMinConns,
			ConnMaxLifetime: pgConnMaxLifetime,
		},
		Redis: Redis{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     redisPort,
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
		Migrations: Migrations{
			Dir: getEnv("MIGRATIONS_DIR", "deployments/migrations"),
		},
		JWT: JWT{
			Secret:          getEnv("JWT_SECRET", "dev-only-insecure-secret-change-me-32chars"),
			Issuer:          getEnv("JWT_ISSUER", "novaflow"),
			Audience:        getEnv("JWT_AUDIENCE", "novaflow-api"),
			AccessTokenTTL:  jwtAccessTTL,
			RefreshTokenTTL: jwtRefreshTTL,
		},
		Argon2: Argon2{
			Memory:      argon2Memory,
			Time:        argon2Time,
			Parallelism: argon2Parallelism,
			SaltLength:  argon2SaltLength,
			KeyLength:   argon2KeyLength,
		},
		Instagram: Instagram{
			ClientID:           getEnv("INSTAGRAM_CLIENT_ID", ""),
			ClientSecret:       getEnv("INSTAGRAM_CLIENT_SECRET", ""),
			RedirectURI:        getEnv("INSTAGRAM_REDIRECT_URI", "http://localhost:8090/api/v1/instagram/callback"),
			Scopes:             getEnv("INSTAGRAM_SCOPES", "instagram_business_basic,instagram_business_content_publish"),
			AuthBaseURL:        getEnv("INSTAGRAM_AUTH_BASE_URL", "https://www.instagram.com"),
			APIBaseURL:         getEnv("INSTAGRAM_API_BASE_URL", "https://api.instagram.com"),
			GraphBaseURL:       getEnv("INSTAGRAM_GRAPH_BASE_URL", "https://graph.instagram.com"),
			TokenEncryptionKey: getEnv("INSTAGRAM_TOKEN_ENCRYPTION_KEY", "dev-only-insecure-instagram-token-key-change-me"),
			StateSecret:        getEnv("INSTAGRAM_STATE_SECRET", "dev-only-insecure-instagram-state-secret-change-me"),
			StateTTL:           instagramStateTTL,
			HTTPTimeout:        instagramHTTPTimeout,
		},
		AI: AI{
			Provider:    getEnv("AI_PROVIDER", "groq"),
			BaseURL:     getEnv("AI_BASE_URL", "https://api.groq.com/openai/v1"),
			APIKey:      getEnv("AI_API_KEY", ""),
			Model:       getEnv("AI_MODEL", "llama-3.1-8b-instant"),
			MaxTokens:   aiMaxTokens,
			HTTPTimeout: aiHTTPTimeout,
		},
		Scheduler: Scheduler{
			PollInterval: schedulerPollInterval,
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	switch c.App.Env {
	case EnvDevelopment, EnvTesting, EnvStaging, EnvProduction:
	default:
		return fmt.Errorf("invalid APP_ENV: %q", c.App.Env)
	}

	if c.App.IsProduction() && c.Postgres.Password == "" {
		return fmt.Errorf("POSTGRES_PASSWORD is required in production")
	}

	if c.App.IsProduction() && len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters in production")
	}

	if c.App.IsProduction() {
		if c.Instagram.ClientID == "" || c.Instagram.ClientSecret == "" {
			return fmt.Errorf("INSTAGRAM_CLIENT_ID and INSTAGRAM_CLIENT_SECRET are required in production")
		}
		if c.Instagram.RedirectURI == "" {
			return fmt.Errorf("INSTAGRAM_REDIRECT_URI is required in production")
		}
		if len(c.Instagram.TokenEncryptionKey) < 32 {
			return fmt.Errorf("INSTAGRAM_TOKEN_ENCRYPTION_KEY must be at least 32 characters in production")
		}
		if len(c.Instagram.StateSecret) < 32 {
			return fmt.Errorf("INSTAGRAM_STATE_SECRET must be at least 32 characters in production")
		}
		if c.AI.APIKey == "" {
			return fmt.Errorf("AI_API_KEY is required in production")
		}
		if c.AI.BaseURL == "" || c.AI.Model == "" {
			return fmt.Errorf("AI_BASE_URL and AI_MODEL are required in production")
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func getEnvInt32(key string, fallback int32) (int32, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return int32(parsed), nil
}

func getEnvUint32(key string, fallback uint32) (uint32, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return uint32(parsed), nil
}

func getEnvUint8(key string, fallback uint8) (uint8, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return uint8(parsed), nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}
