package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	appevents "github.com/Adellaghmari/cloudops-release-intelligence/internal/events"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
)

func TestHandleReportsOnlyPoisonMessage(t *testing.T) {
	h := &handler{
		proc:   appevents.NewProcessor(memory.New(), 3),
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	good := `{"event_id":"evt_worker_good","event_type":"ci.run.completed","occurred_at":"2026-09-09T12:00:00Z","source":"cloudops-api","schema_version":"1.0"}`
	resp, err := h.Handle(context.Background(), events.SQSEvent{Records: []events.SQSMessage{
		{MessageId: "good-1", Body: good},
		{MessageId: "poison-1", Body: "not-json"},
	}})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}
	if len(resp.BatchItemFailures) != 1 || resp.BatchItemFailures[0].ItemIdentifier != "poison-1" {
		t.Fatalf("failures = %+v", resp.BatchItemFailures)
	}
}

func TestHandleDuplicateEventIDIsSuccess(t *testing.T) {
	h := &handler{
		proc:   appevents.NewProcessor(memory.New(), 3),
		logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	body := `{"event_id":"evt_worker_dup","event_type":"ci.run.completed","occurred_at":"2026-09-09T12:00:00Z","source":"cloudops-api","schema_version":"1.0"}`
	first, err := h.Handle(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{MessageId: "m1", Body: body}}})
	if err != nil || len(first.BatchItemFailures) != 0 {
		t.Fatalf("first: err=%v failures=%+v", err, first.BatchItemFailures)
	}
	second, err := h.Handle(context.Background(), events.SQSEvent{Records: []events.SQSMessage{{MessageId: "m2", Body: body}}})
	if err != nil || len(second.BatchItemFailures) != 0 {
		t.Fatalf("duplicate should ack, err=%v failures=%+v", err, second.BatchItemFailures)
	}
}
