package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/config"
	"github.com/adell/cloudops-release-intelligence/internal/httpapi"
	"github.com/adell/cloudops-release-intelligence/internal/localseed"
	"github.com/adell/cloudops-release-intelligence/internal/repository/memory"
	"github.com/adell/cloudops-release-intelligence/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", slog.String("err", err.Error()))
		os.Exit(1)
	}

	logger := httpapi.NewLogger(parseLevel(cfg.LogLevel)).With(
		slog.String("service", cfg.ServiceName),
		slog.String("env", cfg.Env),
		slog.String("version", cfg.Version),
	)

	store := memory.New()
	if cfg.SeedLocalData {
		if err := localseed.Load(context.Background(), store, time.Now().UTC()); err != nil {
			logger.Error("local seed failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
		logger.Info("local synthetic catalog seeded")
	}

	engine := httpapi.NewEngine(cfg, service.NewCatalog(store), logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", slog.String("addr", cfg.HTTPAddr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownWait)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown", slog.String("err", err.Error()))
			os.Exit(1)
		}
		logger.Info("stopped")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("listen", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}
}

func parseLevel(v string) slog.Level {
	switch v {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
