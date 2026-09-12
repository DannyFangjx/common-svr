package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"common-svr/internal/common/config"
	"common-svr/internal/common/db/postgres"
	userrepository "common-svr/internal/common/db/user"
	healthhandler "common-svr/internal/handler/health"
	userhandler "common-svr/internal/handler/user"
	"common-svr/internal/model"
	"common-svr/internal/router"
	userservice "common-svr/internal/service/user"
)

type App struct {
	server *http.Server
	db     *sql.DB
	logger *slog.Logger
	config *config.Config
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	gormDB, sqlDB, err := postgres.Open(cfg.Database)
	if err != nil {
		return nil, err
	}

	if cfg.Database.AutoMigrate {
		if err := gormDB.AutoMigrate(&model.User{}); err != nil {
			_ = sqlDB.Close()
			return nil, fmt.Errorf("auto migrate users table: %w", err)
		}
	}

	userRepository := userrepository.NewRepository(gormDB)
	userService := userservice.NewService(userRepository)
	userHandler := userhandler.NewHandler(userService)
	healthHandler := healthhandler.NewHandler(sqlDB)
	engine := router.New(logger, healthHandler, userHandler)

	return &App{
		server: &http.Server{
			Addr:         cfg.Server.Address(),
			Handler:      engine,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
		db:     sqlDB,
		logger: logger,
		config: cfg,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	serverError := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverError <- err
		}
		close(serverError)
	}()

	select {
	case err := <-serverError:
		if err != nil {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.config.Server.ShutdownTimeout)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return nil
	}
}

func (a *App) Close() {
	if a.db == nil {
		return
	}
	if err := a.db.Close(); err != nil {
		a.logger.Error("close postgres", "error", err)
	}
}
