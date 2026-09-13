//go:build integration

package integration_test

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"common-svr/internal/common/config"
	"common-svr/internal/common/db/postgres"
	userrepository "common-svr/internal/common/db/user"
	healthhandler "common-svr/internal/handler/health"
	userhandler "common-svr/internal/handler/user"
	"common-svr/internal/model"
	"common-svr/internal/router"
	userservice "common-svr/internal/service/user"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("common_server_test"),
		tcpostgres.WithUsername("common_server_test"),
		tcpostgres.WithPassword("common_server_test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Errorf("terminate postgres container: %v", err)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get postgres connection string: %v", err)
	}

	gormDB, sqlDB, err := postgres.Open(config.Database{
		Driver:                "postgres",
		DSN:                   dsn,
		MaxOpenConnections:    5,
		MaxIdleConnections:    2,
		ConnectionMaxLifetime: time.Minute,
	})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close postgres: %v", err)
		}
	})

	if err := gormDB.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("migrate users table: %v", err)
	}

	repository := userrepository.NewRepository(gormDB)
	service := userservice.NewService(repository)
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	engine := router.New(
		"common-svr-integration-test",
		logger,
		healthhandler.NewHandler(sqlDB),
		userhandler.NewHandler(service),
	)

	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)
	return server
}
