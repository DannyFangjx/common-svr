package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App       App
	Server    Server
	Log       Log
	Database  Database
	Telemetry Telemetry
}

type App struct {
	Name        string `env:"SERVICE_NAME" envDefault:"common-svr"`
	Environment string `env:"APP_ENV" envDefault:"dev"`
}

type Server struct {
	Host            string        `env:"HTTP_ADDR" envDefault:"0.0.0.0"`
	Port            int           `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Log struct {
	Level string `env:"LOG_LEVEL" envDefault:"info"`
}

type Database struct {
	Driver                string        `env:"DATABASE_DRIVER" envDefault:"postgres"`
	DSN                   string        `env:"DATABASE_DSN"`
	MaxOpenConnections    int           `env:"DATABASE_MAX_OPEN_CONNECTIONS" envDefault:"20"`
	MaxIdleConnections    int           `env:"DATABASE_MAX_IDLE_CONNECTIONS" envDefault:"5"`
	ConnectionMaxLifetime time.Duration `env:"DATABASE_CONNECTION_MAX_LIFETIME" envDefault:"30m"`
	AutoMigrate           bool          `env:"DATABASE_AUTO_MIGRATE" envDefault:"true"`
}

type Telemetry struct {
	Endpoint    string  `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	Protocol    string  `env:"OTEL_EXPORTER_OTLP_PROTOCOL" envDefault:"grpc"`
	Insecure    bool    `env:"OTEL_EXPORTER_OTLP_INSECURE" envDefault:"true"`
	SampleRatio float64 `env:"OTEL_TRACES_SAMPLER_ARG" envDefault:"1.0"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load .env: %w", err)
	}
	return parseEnvironment(nil)
}

func parseEnvironment(environment map[string]string) (*Config, error) {
	var cfg Config
	var err error
	if environment == nil {
		err = env.Parse(&cfg)
	} else {
		err = env.ParseWithOptions(&cfg, env.Options{Environment: environment})
	}
	if err != nil {
		return nil, fmt.Errorf("parse environment: %w", err)
	}

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
	if c.Telemetry.Protocol != "grpc" {
		return fmt.Errorf("unsupported telemetry protocol %q", c.Telemetry.Protocol)
	}
	if c.Telemetry.SampleRatio < 0 || c.Telemetry.SampleRatio > 1 {
		return errors.New("telemetry sample ratio must be between 0 and 1")
	}
	return nil
}
