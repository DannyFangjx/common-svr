package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

const defaultConfigFile = "config/app.yaml"

type Config struct {
	App      App      `yaml:"app"`
	Server   Server   `yaml:"server"`
	Log      Log      `yaml:"log"`
	Database Database `yaml:"database"`
}

type App struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
}

type Server struct {
	Host            string        `yaml:"address"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Log struct {
	Level string `yaml:"level"`
}

type Database struct {
	Driver                string        `yaml:"driver"`
	DSN                   string        `yaml:"-"`
	MaxOpenConnections    int           `yaml:"max_open_connections"`
	MaxIdleConnections    int           `yaml:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `yaml:"connection_max_lifetime"`
	AutoMigrate           bool          `yaml:"auto_migrate"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	configFile := envOrDefault("CONFIG_FILE", defaultConfigFile)
	raw, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", configFile, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", configFile, err)
	}

	if host := os.Getenv("HTTP_ADDR"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("HTTP_PORT"); port != "" {
		value, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("parse HTTP_PORT: %w", err)
		}
		cfg.Server.Port = value
	}
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Log.Level = level
	}
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		cfg.App.Name = name
	}
	if environment := os.Getenv("APP_ENV"); environment != "" {
		cfg.App.Environment = environment
	}
	if driver := os.Getenv("DATABASE_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
	}
	cfg.Database.DSN = os.Getenv("DATABASE_DSN")

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.App.Name == "" {
		return errors.New("service name is required")
	}
	if c.App.Environment == "" {
		return errors.New("application environment is required")
	}
	if c.Server.Host == "" {
		return errors.New("server address is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return errors.New("server port must be between 1 and 65535")
	}
	if c.Database.Driver != "postgres" {
		return fmt.Errorf("unsupported database driver %q", c.Database.Driver)
	}
	if c.Database.DSN == "" {
		return errors.New("DATABASE_DSN is required")
	}
	return nil
}

func envOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
