package config

import (
	"testing"
	"time"
)

func TestParseEnvironmentUsesDefaults(t *testing.T) {
	cfg, err := parseEnvironment(map[string]string{
		"DATABASE_DSN": "host=localhost dbname=test",
	})
	if err != nil {
		t.Fatalf("parseEnvironment() error = %v", err)
	}
	if cfg.App.Name != "common-svr" {
		t.Fatalf("app name = %q, want common-svr", cfg.App.Name)
	}
	if cfg.App.Environment != "dev" {
		t.Fatalf("app environment = %q, want dev", cfg.App.Environment)
	}
	if cfg.Server.Address() != "0.0.0.0:8080" {
		t.Fatalf("server address = %q, want 0.0.0.0:8080", cfg.Server.Address())
	}
	if cfg.Server.ReadTimeout != 10*time.Second {
		t.Fatalf("read timeout = %s, want 10s", cfg.Server.ReadTimeout)
	}
	if cfg.Database.Driver != "postgres" {
		t.Fatalf("database driver = %q, want postgres", cfg.Database.Driver)
	}
	if cfg.Database.MaxOpenConnections != 20 {
		t.Fatalf("max open connections = %d, want 20", cfg.Database.MaxOpenConnections)
	}
	if !cfg.Database.AutoMigrate {
		t.Fatal("database auto migrate = false, want true")
	}
	if cfg.Telemetry.Endpoint != "" {
		t.Fatalf("telemetry endpoint = %q, want empty", cfg.Telemetry.Endpoint)
	}
	if cfg.Telemetry.Protocol != "grpc" {
		t.Fatalf("telemetry protocol = %q, want grpc", cfg.Telemetry.Protocol)
	}
	if cfg.Telemetry.SampleRatio != 1 {
		t.Fatalf("telemetry sample ratio = %v, want 1", cfg.Telemetry.SampleRatio)
	}
}

func TestParseEnvironmentLoadsAllOverrides(t *testing.T) {
	cfg, err := parseEnvironment(map[string]string{
		"SERVICE_NAME":                     "env-service",
		"APP_ENV":                          "prod",
		"HTTP_ADDR":                        "127.0.0.1",
		"HTTP_PORT":                        "9090",
		"HTTP_READ_TIMEOUT":                "2s",
		"HTTP_WRITE_TIMEOUT":               "3s",
		"HTTP_IDLE_TIMEOUT":                "4s",
		"HTTP_SHUTDOWN_TIMEOUT":            "5s",
		"LOG_LEVEL":                        "warn",
		"DATABASE_DRIVER":                  "postgres",
		"DATABASE_DSN":                     "host=localhost dbname=test",
		"DATABASE_MAX_OPEN_CONNECTIONS":    "42",
		"DATABASE_MAX_IDLE_CONNECTIONS":    "7",
		"DATABASE_CONNECTION_MAX_LIFETIME": "45m",
		"DATABASE_AUTO_MIGRATE":            "false",
		"OTEL_EXPORTER_OTLP_ENDPOINT":      "collector:4317",
		"OTEL_EXPORTER_OTLP_PROTOCOL":      "grpc",
		"OTEL_EXPORTER_OTLP_INSECURE":      "false",
		"OTEL_TRACES_SAMPLER_ARG":          "0.25",
	})
	if err != nil {
		t.Fatalf("parseEnvironment() error = %v", err)
	}
	if cfg.App.Name != "env-service" || cfg.App.Environment != "prod" {
		t.Fatalf("app config = %#v, want environment overrides", cfg.App)
	}
	if cfg.Server.Address() != "127.0.0.1:9090" {
		t.Fatalf("server address = %q, want 127.0.0.1:9090", cfg.Server.Address())
	}
	if cfg.Server.ReadTimeout != 2*time.Second || cfg.Server.ShutdownTimeout != 5*time.Second {
		t.Fatalf("server durations = %#v, want environment overrides", cfg.Server)
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("log level = %q, want warn", cfg.Log.Level)
	}
	if cfg.Database.MaxOpenConnections != 42 || cfg.Database.MaxIdleConnections != 7 {
		t.Fatalf("database pool config = %#v, want environment overrides", cfg.Database)
	}
	if cfg.Database.ConnectionMaxLifetime != 45*time.Minute || cfg.Database.AutoMigrate {
		t.Fatalf("database runtime config = %#v, want environment overrides", cfg.Database)
	}
	if cfg.Telemetry.Endpoint != "collector:4317" || cfg.Telemetry.Insecure {
		t.Fatalf("telemetry config = %#v, want environment overrides", cfg.Telemetry)
	}
	if cfg.Telemetry.SampleRatio != 0.25 {
		t.Fatalf("telemetry sample ratio = %v, want 0.25", cfg.Telemetry.SampleRatio)
	}
}

func TestParseEnvironmentRejectsInvalidTypedValue(t *testing.T) {
	_, err := parseEnvironment(map[string]string{
		"DATABASE_DSN": "host=localhost dbname=test",
		"HTTP_PORT":    "not-a-number",
	})
	if err == nil {
		t.Fatal("parseEnvironment() error = nil, want typed environment parse error")
	}
}
