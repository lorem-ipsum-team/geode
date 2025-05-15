package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	http_server "github.com/lorem-ipsum-team/geode/internal/http"
	postgres_repo "github.com/lorem-ipsum-team/geode/internal/postgres"
	"github.com/lorem-ipsum-team/swipe/pkg/logger"
)

const (
	gracefulTimeout = 30 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logLevel := os.Getenv("LOG_LEVEL")
	logFormat := os.Getenv("LOG_FORMAT")

	log := logger.Init(logFormat, logLevel)

	db_url := os.Getenv("DB_URL")

	postgresRepo, err := postgres_repo.NewRepo(ctx, log, db_url)
	if err != nil {
		log.ErrorContext(ctx, "failed to create postgres_repo", slog.Any("error", err))

		return
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	server := http_server.New(log, listenAddr, postgresRepo)

	go func() {
		slog.InfoContext(ctx, "Start server", slog.String("addr", listenAddr))

		if err := server.Server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.ErrorContext(ctx, "Server crashed", slog.Any("error", err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	defer cancel()
	log.InfoContext(shutdownCtx, "Shutting down...")

	if err := server.Server.Shutdown(shutdownCtx); err != nil {
		log.ErrorContext(shutdownCtx, "Graceful shutdown failed", slog.Any("error", err))
		log.Info("Shutting down forcefully...")

		if err := server.Server.Close(); err != nil {
			log.Error("Forceful shutdown failed", slog.Any("error", err))
		}
	}

	log.Info("Server stopped")
}
