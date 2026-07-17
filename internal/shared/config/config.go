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
