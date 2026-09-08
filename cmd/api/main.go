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
	"github.com/adell/cloudops-release-intelligence/internal/events"
	"github.com/adell/cloudops-release-intelligence/internal/httpapi"
	"github.com/adell/cloudops-release-intelligence/internal/localseed"
	"github.com/adell/cloudops-release-intelligence/internal/repository"
	"github.com/adell/cloudops-release-intelligence/internal/repository/dynamo"
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

	store, err := openStore(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("store", slog.String("err", err.Error()))
		os.Exit(1)
	}
	if cfg.SeedLocalData {
		if err := localseed.Load(context.Background(), store, time.Now().UTC()); err != nil {
			logger.Error("local seed failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
		logger.Info("local synthetic catalog seeded")
	}

	engine := httpapi.NewEngine(cfg, service.NewCatalog(store), logger, events.NewProcessor(store, 3))

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

func openStore(ctx context.Context, cfg config.Config, logger *slog.Logger) (repository.Store, error) {
	if cfg.StoreDriver != "dynamodb" {
		logger.Info("using memory store")
		return memory.New(), nil
	}
	client, err := dynamo.NewClient(ctx, dynamo.ClientOptions{
		Region: cfg.AWSRegion, Endpoint: cfg.DynamoEndpoint, Table: cfg.DynamoTable,
	})
	if err != nil {
		return nil, err
	}
	if err := dynamo.EnsureTable(ctx, client, cfg.DynamoTable); err != nil {
		return nil, err
	}
	logger.Info("using dynamodb store", slog.String("table", cfg.DynamoTable))
	return dynamo.New(client, cfg.DynamoTable), nil
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
