package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
)

func sampleService(id, name string, source domain.DataSource) domain.Service {
	now := time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC)
	return domain.Service{
		ID:          domain.ServiceID(id),
		Name:        name,
		Criticality: domain.CriticalityHigh,
		Source:      source,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestCreateGetListAndDuplicate(t *testing.T) {
	ctx := context.Background()
	s := New()
	live := sampleService("cloudops-api", "CloudOps API", domain.DataSourceLive)
	syn := sampleService("payments-service", "Payments Service", domain.DataSourceSynthetic)
	if err := s.CreateService(ctx, live); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateService(ctx, syn); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetService(ctx, "payments-service")
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != domain.DataSourceSynthetic {
		t.Fatalf("source=%s", got.Source)
	}
	list, err := s.ListServices(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("list=%d err=%v", len(list), err)
	}
	if list[0].Name > list[1].Name {
		t.Fatal("expected name sort")
	}
	err = s.CreateService(ctx, syn)
	if !domain.IsAlreadyExists(err) {
		t.Fatalf("expected already exists, got %v", err)
	}
}

func TestNotFound(t *testing.T) {
	_, err := New().GetService(context.Background(), "missing-service")
	if !domain.IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
	var nf domain.NotFoundError
	if !errors.As(err, &nf) || nf.Resource != "service" {
		t.Fatalf("unexpected error payload: %#v", err)
	}
}

func TestEventIdempotencyKey(t *testing.T) {
	ctx := context.Background()
	s := New()
	rel := domain.ReleaseID("rel_one")
	svc := domain.ServiceID("payments-service")
	ev := domain.ReleaseEvent{
		ID:            "evt_stable_producer_1",
		Type:          domain.EventTypeDeploymentSucceeded,
		SchemaVersion: "1.0",
		OccurredAt:    time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC),
		IngestedAt:    time.Date(2026, 9, 8, 10, 0, 1, 0, time.UTC),
		Producer:      domain.EventProducerSynthetic,
		CorrelationID: "req_abc",
		ReleaseID:     &rel,
		ServiceID:     &svc,
	}
	if err := s.CreateEvent(ctx, ev); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateEvent(ctx, ev); !domain.IsAlreadyExists(err) {
		t.Fatalf("duplicate event_id must be rejected: %v", err)
	}
	got, err := s.GetEvent(ctx, ev.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != ev.ID {
		t.Fatalf("got %s", got.ID)
	}
}

func TestRejectsInvalidBeforeStore(t *testing.T) {
	err := New().CreateService(context.Background(), domain.Service{ID: "payments-service"})
	if !domain.IsInvalid(err) {
		t.Fatalf("expected invalid, got %v", err)
	}
}

func TestReleaseByService(t *testing.T) {
	ctx := context.Background()
	s := New()
	now := time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC)
	r1 := domain.Release{
		ID: "rel_1", ServiceID: "payments-service", Version: "1.0.0", GitSHA: "abc1234",
		Environment: domain.EnvironmentProd, Status: domain.ReleaseStatusPending,
		Source: domain.DataSourceSynthetic, CreatedAt: now,
	}
	r2 := domain.Release{
		ID: "rel_2", ServiceID: "checkout-api", Version: "1.0.0", GitSHA: "abc1235",
		Environment: domain.EnvironmentProd, Status: domain.ReleaseStatusPending,
		Source: domain.DataSourceSynthetic, CreatedAt: now.Add(time.Hour),
	}
	_ = s.CreateRelease(ctx, r1)
	_ = s.CreateRelease(ctx, r2)
	list, err := s.ListReleasesByService(ctx, "payments-service")
	if err != nil || len(list) != 1 || list[0].ID != "rel_1" {
		t.Fatalf("list=%v err=%v", list, err)
	}
}
