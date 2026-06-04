package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/relay/backend/internal/config"
	"github.com/relay/backend/internal/logger"
	"github.com/relay/backend/internal/server"
)

func main() {
	cfg := config.Load()
	appLogger := logger.New(cfg)

	appLogger.Info("starting relay api server",
		"env", cfg.App.Env,
		"version", "0.0.1",
	)

	httpServer := server.New(cfg, appLogger)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		if err := httpServer.Run(); err != nil {
			errChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		appLogger.Info("received shutdown signal",
			"signal", sig.String(),
		)

		appLogger.Info("shutting down server...")
		if err := httpServer.Shutdown(30 * time.Second); err != nil {
			appLogger.Error("error during shutdown",
				"error", err.Error(),
			)
			os.Exit(1)
		}

		appLogger.Info("server shut down successfully")
	case err := <-errChan:
		appLogger.Error("server error",
			"error", err.Error(),
		)
		os.Exit(1)
	}
}
