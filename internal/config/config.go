package config

import (
	"fmt"
	"strings"

	"github.com/nil-go/konf"
	"github.com/nil-go/konf/provider/env"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Env string
}

type ServerConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	DSN string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Load builds a Config by first applying defaults then overriding with
// environment variables prefixed with APP_ (e.g. APP_SERVER_PORT=9090).
func Load() (Config, error) {
	k := konf.New()

	if err := k.Load(defaults()); err != nil {
		return Config{}, fmt.Errorf("load defaults: %w", err)
	}

	if err := k.Load(env.New(
		env.WithPrefix("APP_"),
		env.WithNameSplitter(func(s string) []string {
			return strings.Split(strings.TrimPrefix(s, "APP_"), "_")
		}),
	)); err != nil {
		return Config{}, fmt.Errorf("load env: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("app", &cfg.App); err != nil {
		return Config{}, fmt.Errorf("unmarshal app config: %w", err)
	}
	if err := k.Unmarshal("server", &cfg.Server); err != nil {
		return Config{}, fmt.Errorf("unmarshal server config: %w", err)
	}
	if err := k.Unmarshal("database", &cfg.Database); err != nil {
		return Config{}, fmt.Errorf("unmarshal database config: %w", err)
	}
	if err := k.Unmarshal("redis", &cfg.Redis); err != nil {
		return Config{}, fmt.Errorf("unmarshal redis config: %w", err)
	}

	return cfg, nil
}

// defaults returns hardcoded baseline values loaded first so that env vars
// only need to specify values that differ from the defaults.
type defaultLoader struct{ values map[string]any }

func (d defaultLoader) Load() (map[string]any, error) { return d.values, nil }

func defaults() defaultLoader {
	return defaultLoader{values: map[string]any{
		"app": map[string]any{
			"env": "development",
		},
		"server": map[string]any{
			"host": "",
			"port": 8080,
		},
		"database": map[string]any{
			"dsn": "postgres://localhost:5432/taskmanagement?sslmode=disable",
		},
		"redis": map[string]any{
			"host":     "localhost",
			"port":     6379,
			"password": "",
			"db":       0,
		},
	}}
}
