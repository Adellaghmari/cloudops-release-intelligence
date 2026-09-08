package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/config"
	appevents "github.com/Adellaghmari/cloudops-release-intelligence/internal/events"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/httpapi"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/platform/eventbridge"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/dynamo"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", slog.String("err", err.Error()))
		os.Exit(1)
	}
	logger := httpapi.NewLogger(parseLevel(cfg.LogLevel)).With(
		slog.String("service", "cloudops-worker"),
		slog.String("env", cfg.Env),
		slog.String("git_sha", cfg.GitSHA),
	)
	store, err := openStore(context.Background(), cfg, logger)
	if err != nil {
		logger.Error("store", slog.String("err", err.Error()))
		os.Exit(1)
	}
	catalog := service.NewCatalog(store)
	proc := appevents.NewProcessor(store, 3)
	proc.SetAnalyzer(catalog)
	h := &handler{proc: proc, logger: logger}
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		lambda.Start(h.Handle)
		return
	}
	logger.Info("worker idle outside lambda; SQS event source mapping is the production trigger")
}

type handler struct {
	proc   *appevents.Processor
	logger *slog.Logger
}

func (h *handler) Handle(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, rec := range sqsEvent.Records {
		env, err := eventbridge.UnwrapSQSBody(rec.Body)
		if err != nil {
			h.logger.ErrorContext(ctx, "sqs_unwrap_failed")
			return err
		}
		h.logger.InfoContext(ctx, "worker_event",
			slog.String("event_id", env.EventID.String()),
			slog.String("event_type", string(env.EventType)),
			slog.String("correlation_id", env.CorrelationID),
		)
		if _, err := h.proc.Handle(ctx, env, time.Now().UTC()); err != nil {
			return err
		}
	}
	return nil
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
