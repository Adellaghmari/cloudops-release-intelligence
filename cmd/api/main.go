package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awseb "github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/config"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/events"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/httpapi"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/localseed"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/platform/eventbridge"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/platform/s3evidence"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/dynamo"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/runtime"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/service"
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
		slog.String("git_sha", cfg.GitSHA),
	)

	ctx := context.Background()
	store, err := openStore(ctx, cfg, logger)
	if err != nil {
		logger.Error("store", slog.String("err", err.Error()))
		os.Exit(1)
	}
	if cfg.SeedLocalData {
		if err := localseed.Load(ctx, store, time.Now().UTC()); err != nil {
			if !domain.IsAlreadyExists(err) {
				logger.Error("local seed failed", slog.String("err", err.Error()))
				os.Exit(1)
			}
			logger.Info("catalog already present; seed skipped")
		} else {
			logger.Info("synthetic catalog seeded")
		}
	}

	catalog := service.NewCatalog(store)
	proc := events.NewProcessor(store, 3)
	proc.SetAnalyzer(catalog)

	var bus events.Bus
	var putRaw func(context.Context, events.Envelope) error
	if cfg.EventBusName != "" || cfg.RawEvidenceBucket != "" {
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWSRegion))
		if err != nil {
			logger.Error("aws config", slog.String("err", err.Error()))
			os.Exit(1)
		}
		if cfg.EventBusName != "" {
			bus = eventbridge.New(awseb.NewFromConfig(awsCfg), cfg.EventBusName, cfg.EventSource)
		}
		if cfg.RawEvidenceBucket != "" {
			w := s3evidence.New(s3.NewFromConfig(awsCfg), cfg.RawEvidenceBucket)
			putRaw = w.PutEvent
		}
	}

	engine := httpapi.NewEngineWith(httpapi.EngineConfig{
		Config:    cfg,
		Catalog:   catalog,
		Logger:    logger,
		Processor: proc,
		Store:     store,
		Bus:       bus,
		PutRaw:    putRaw,
	})

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		logger.Info("lambda http adapter starting")
		lambda.Start(httpadapter.NewV2(engine).ProxyWithContext)
		return
	}

	runCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := runtime.ServeHTTP(runCtx, srv, cfg.ShutdownWait); err != nil {
		logger.Error("listen", slog.String("err", err.Error()))
		os.Exit(1)
	}
	logger.Info("stopped")
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
	if cfg.EnsureDynamoTable {
		if err := dynamo.EnsureTable(ctx, client, cfg.DynamoTable); err != nil {
			return nil, err
		}
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
