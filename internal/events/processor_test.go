package events

import (
	"context"
	"testing"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
)

func TestNormalizeAndIdempotency(t *testing.T) {
	store := memory.New()
	p := NewProcessor(store, 3)
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	rel := domain.ReleaseID("rel_evt")
	_ = store.CreateService(context.Background(), domain.Service{
		ID: "payments-service", Name: "P", Criticality: domain.CriticalityHigh,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	})
	_ = store.CreateRelease(context.Background(), domain.Release{
		ID: rel, ServiceID: "payments-service", Version: "1", GitSHA: "abc1234",
		Environment: domain.EnvironmentLocal, Status: domain.ReleaseStatusPending,
		Source: domain.DataSourceSynthetic, CreatedAt: now,
	})
	env := Envelope{
		EventID: "evt_1", EventType: domain.EventTypeDeploymentSucceeded, OccurredAt: now,
		Source: domain.EventProducerSynthetic, SchemaVersion: CurrentSchema, ReleaseID: &rel,
		CorrelationID: "corr_1",
	}
	r1, err := p.Handle(context.Background(), env, now)
	if err != nil || r1.Duplicate {
		t.Fatalf("%v %+v", err, r1)
	}
	r2, err := p.Handle(context.Background(), env, now)
	if err != nil || !r2.Duplicate {
		t.Fatalf("expected duplicate, got %+v %v", r2, err)
	}
}

func TestMalformedAndUnsupportedSchema(t *testing.T) {
	p := NewProcessor(memory.New(), 3)
	now := time.Now().UTC()
	bad := Envelope{EventID: "bad id", EventType: domain.EventTypeCIRunCompleted, OccurredAt: now, Source: domain.EventProducerAPI, SchemaVersion: CurrentSchema}
	r, _ := p.Handle(context.Background(), bad, now)
	if !r.Dropped {
		t.Fatal("expected drop")
	}
	old := Envelope{EventID: "evt_old", EventType: domain.EventTypeCIRunCompleted, OccurredAt: now, Source: domain.EventProducerAPI, SchemaVersion: "9.9"}
	r, _ = p.Handle(context.Background(), old, now)
	if !r.Dropped || r.Reason != "unsupported schema version" {
		t.Fatalf("%+v", r)
	}
	if len(p.DLQ()) != 2 {
		t.Fatalf("dlq=%d", len(p.DLQ()))
	}
}

func TestUnknownTypeAndReprocess(t *testing.T) {
	store := memory.New()
	p := NewProcessor(store, 3)
	now := time.Now().UTC()
	unknown := Envelope{
		EventID: "evt_unknown", EventType: domain.EventType("not.a.type"), OccurredAt: now,
		Source: domain.EventProducerAPI, SchemaVersion: CurrentSchema,
	}
	r, _ := p.Handle(context.Background(), unknown, now)
	if !r.Dropped || r.Reason != "unsupported event type" {
		t.Fatalf("%+v", r)
	}
	rel := domain.ReleaseID("rel_re")
	_ = store.CreateService(context.Background(), domain.Service{
		ID: "payments-service", Name: "P", Criticality: domain.CriticalityHigh,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	})
	_ = store.CreateRelease(context.Background(), domain.Release{
		ID: rel, ServiceID: "payments-service", Version: "1", GitSHA: "abc1234",
		Environment: domain.EnvironmentLocal, Status: domain.ReleaseStatusPending,
		Source: domain.DataSourceSynthetic, CreatedAt: now,
	})
	env := Envelope{
		EventID: "evt_re", EventType: domain.EventTypeCIRunCompleted, OccurredAt: now,
		Source: domain.EventProducerAPI, SchemaVersion: CurrentSchema, ReleaseID: &rel,
	}
	if _, err := p.Handle(context.Background(), env, now); err != nil {
		t.Fatal(err)
	}
	again, err := p.Handle(context.Background(), env, now)
	if err != nil || !again.Duplicate {
		t.Fatalf("reprocess must be idempotent: %+v %v", again, err)
	}
}

func TestPolicyEvaluatedDoesNotDispatch(t *testing.T) {
	if TriggersAnalysis(domain.EventTypePolicyEvaluated) {
		t.Fatal("policy.evaluated must not re-enter analysis")
	}
	p := NewProcessor(memory.New(), 3)
	now := time.Now().UTC()
	r, err := p.Handle(context.Background(), Envelope{
		EventID: "evt_pol", EventType: domain.EventTypePolicyEvaluated, OccurredAt: now,
		Source: domain.EventProducerAPI, SchemaVersion: CurrentSchema,
	}, now)
	if err != nil || r.Dropped || r.Duplicate {
		t.Fatalf("%v %+v", err, r)
	}
	if r.Reason != "timeline only; analysis not dispatched" {
		t.Fatalf("reason=%s", r.Reason)
	}
}

func TestPoisonBoundedRetries(t *testing.T) {
	inner := memory.New()
	store := &failingStore{Store: inner, failTimes: 3}
	p := NewProcessor(store, 3)
	now := time.Now().UTC()
	env := Envelope{
		EventID: "evt_poison", EventType: domain.EventTypeCIRunCompleted, OccurredAt: now,
		Source: domain.EventProducerAPI, SchemaVersion: CurrentSchema,
	}
	var last ProcessResult
	var err error
	for i := 0; i < 3; i++ {
		last, err = p.Handle(context.Background(), env, now)
	}
	if err == nil || !last.Dropped || last.Reason != "poison after bounded retries" {
		t.Fatalf("expected poison DLQ, got %+v %v", last, err)
	}
	if len(p.DLQ()) != 1 {
		t.Fatalf("dlq=%d", len(p.DLQ()))
	}
}

type failingStore struct {
	repository.Store
	failTimes int
	n         int
}

func (f *failingStore) CreateEvent(ctx context.Context, e domain.ReleaseEvent) error {
	f.n++
	if f.n <= f.failTimes {
		return context.DeadlineExceeded
	}
	return f.Store.CreateEvent(ctx, e)
}

func TestOutOfOrderRelease(t *testing.T) {
	store := memory.New()
	p := NewProcessor(store, 3)
	now := time.Now().UTC()
	rel := domain.ReleaseID("rel_missing")
	env := Envelope{
		EventID: "evt_oo", EventType: domain.EventTypeDeploymentStarted, OccurredAt: now,
		Source: domain.EventProducerSynthetic, SchemaVersion: CurrentSchema, ReleaseID: &rel,
		CorrelationID: "c",
	}
	r, err := p.Handle(context.Background(), env, now)
	if err != nil || !r.Pending {
		t.Fatalf("%v %+v", err, r)
	}
}
