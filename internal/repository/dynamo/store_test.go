package dynamo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/aws/smithy-go"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	fake := NewFake()
	if err := EnsureTable(context.Background(), fake, "cloudops-main-test"); err != nil {
		t.Fatal(err)
	}
	return New(fake, "cloudops-main-test")
}

func TestKeysStayOutOfDomain(t *testing.T) {
	pk := servicePK("payments-service")
	if pk != "SERVICE#payments-service" {
		t.Fatalf("pk=%s", pk)
	}
}

func TestRoundTripAndQueries(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	pay := domain.Service{
		ID: "payments-service", Name: "Payments Service", Criticality: domain.CriticalityCritical,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	}
	chk := domain.Service{
		ID: "checkout-api", Name: "Checkout API", Criticality: domain.CriticalityCritical,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.CreateService(ctx, pay); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateService(ctx, chk); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateService(ctx, pay); !domain.IsAlreadyExists(err) {
		t.Fatalf("dup service: %v", err)
	}
	got, err := s.GetService(ctx, "payments-service")
	if err != nil || got.Name != "Payments Service" {
		t.Fatalf("get=%v err=%v", got, err)
	}
	list, err := s.ListServices(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("list=%d err=%v", len(list), err)
	}
	if err := s.CreateDependency(ctx, domain.Dependency{
		From: "checkout-api", To: "payments-service", Kind: domain.DependencyKindRuntime, CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	from, err := s.ListDependenciesFrom(ctx, "checkout-api")
	if err != nil || len(from) != 1 || from[0].To != "payments-service" {
		t.Fatalf("from=%v err=%v", from, err)
	}
	to, err := s.ListDependenciesTo(ctx, "payments-service")
	if err != nil || len(to) != 1 || to[0].From != "checkout-api" {
		t.Fatalf("to=%v err=%v", to, err)
	}

	rel := domain.Release{
		ID: "rel_pay_1", ServiceID: "payments-service", Version: "1.0.0", GitSHA: "abc1234",
		Environment: domain.EnvironmentLocal, Status: domain.ReleaseStatusDeployed,
		Source: domain.DataSourceSynthetic, CreatedAt: now,
	}
	if err := s.CreateRelease(ctx, rel); err != nil {
		t.Fatal(err)
	}
	bySvc, err := s.ListReleasesByService(ctx, "payments-service")
	if err != nil || len(bySvc) != 1 {
		t.Fatalf("bySvc=%v err=%v", bySvc, err)
	}
	all, err := s.ListReleases(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("all=%v err=%v", all, err)
	}
}

func TestEventIdempotencyAndTimeline(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	relID := domain.ReleaseID("rel_evt")
	svcID := domain.ServiceID("payments-service")
	_ = s.CreateService(ctx, domain.Service{
		ID: svcID, Name: "Payments", Criticality: domain.CriticalityHigh,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	})
	_ = s.CreateRelease(ctx, domain.Release{
		ID: relID, ServiceID: svcID, Version: "1.0.0", GitSHA: "abc1234",
		Environment: domain.EnvironmentLocal, Status: domain.ReleaseStatusPending,
		Source: domain.DataSourceSynthetic, CreatedAt: now,
	})
	ev := domain.ReleaseEvent{
		ID: "evt_stable", Type: domain.EventTypeDeploymentSucceeded, SchemaVersion: "1.0",
		OccurredAt: now, IngestedAt: now.Add(time.Second), Producer: domain.EventProducerSynthetic,
		CorrelationID: "req_1", ReleaseID: &relID, ServiceID: &svcID,
	}
	if err := s.CreateEvent(ctx, ev); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateEvent(ctx, ev); !domain.IsAlreadyExists(err) {
		t.Fatalf("duplicate event: %v", err)
	}
	got, err := s.GetEvent(ctx, "evt_stable")
	if err != nil || got.Type != domain.EventTypeDeploymentSucceeded {
		t.Fatalf("get event %v %v", got, err)
	}
	tl, err := s.ListEventsByRelease(ctx, relID)
	if err != nil || len(tl) != 1 || tl[0].ID != "evt_stable" {
		t.Fatalf("timeline=%v err=%v", tl, err)
	}
}

func TestNotFoundAndCanceledContext(t *testing.T) {
	s := testStore(t)
	_, err := s.GetService(context.Background(), "missing-service")
	if !domain.IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = s.CreateService(ctx, domain.Service{
		ID: "x-service", Name: "X", Criticality: domain.CriticalityLow,
		Source: domain.DataSourceSynthetic, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled, got %v", err)
	}
}

func TestPublicErrorsDoNotLeakAWS(t *testing.T) {
	err := wrapErr("put_service", errors.New("AccessDeniedException: not authorized secret=AKIATEST"))
	if err == nil || err.Error() != "storage put_service failed" {
		t.Fatalf("wrapped=%v", err)
	}
	if strings.Contains(err.Error(), "AKIATEST") || strings.Contains(err.Error(), "not authorized") {
		t.Fatalf("leaked aws detail: %v", err)
	}

	coded := wrapErr("put_service", &fakeAPIError{code: "UnrecognizedClientException", msg: "secret=SESSION"})
	if coded == nil || coded.Error() != "storage put_service failed (UnrecognizedClientException)" {
		t.Fatalf("coded wrapped=%v", coded)
	}
	if strings.Contains(coded.Error(), "SESSION") {
		t.Fatalf("leaked aws message: %v", coded)
	}
}

type fakeAPIError struct {
	code string
	msg  string
}

func (e *fakeAPIError) Error() string                 { return e.code + ": " + e.msg }
func (e *fakeAPIError) ErrorCode() string             { return e.code }
func (e *fakeAPIError) ErrorMessage() string          { return e.msg }
func (e *fakeAPIError) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

func TestOperationalEvidenceAndResetPreserveLive(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	if err := s.CreateService(ctx, domain.Service{
		ID: "cloudops-api", Name: "CloudOps API", Criticality: domain.CriticalityHigh,
		Source: domain.DataSourceLive, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateService(ctx, domain.Service{
		ID: "checkout-api", Name: "Checkout", Criticality: domain.CriticalityCritical,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	rel := domain.Release{
		ID: "rel_northstar_reset", ServiceID: "checkout-api", Version: "1.0.0", GitSHA: "abc1234",
		Environment: domain.EnvironmentLocal, Status: domain.ReleaseStatusDeployed,
		Source: domain.DataSourceSynthetic, CreatedAt: now,
	}
	if err := s.CreateRelease(ctx, rel); err != nil {
		t.Fatal(err)
	}
	evID := domain.EventID("evt_reset_1")
	if err := s.CreateEvent(ctx, domain.ReleaseEvent{
		ID: evID, Type: domain.EventTypeDeploymentSucceeded, SchemaVersion: "1.0",
		OccurredAt: now, IngestedAt: now, Producer: domain.EventProducerSynthetic,
		CorrelationID: "corr_reset", ReleaseID: &rel.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.PutOperationalEvidence(ctx, domain.OperationalEvidence{
		ID: "evt_live_pipeline", Kind: "pipeline", GitSHA: "abcdef1", WorkflowRunID: "99",
		WorkflowResult: "success", RecordedAt: now, Source: domain.DataSourceLive,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.PutOperationalEvidence(ctx, domain.OperationalEvidence{
		ID: "evt_live_pipeline", Kind: "pipeline", GitSHA: "abcdef1", WorkflowRunID: "99",
		WorkflowResult: "success", RecordedAt: now, Source: domain.DataSourceLive,
	}); !domain.IsAlreadyExists(err) {
		t.Fatalf("expected duplicate evidence, got %v", err)
	}
	if err := s.ResetSynthetic(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetService(ctx, "checkout-api"); !domain.IsNotFound(err) {
		t.Fatalf("synthetic service should be gone: %v", err)
	}
	if _, err := s.GetRelease(ctx, rel.ID); !domain.IsNotFound(err) {
		t.Fatalf("synthetic release should be gone: %v", err)
	}
	live, err := s.GetService(ctx, "cloudops-api")
	if err != nil || live.Source != domain.DataSourceLive {
		t.Fatalf("live service=%v err=%v", live, err)
	}
	ops, err := s.ListOperationalEvidence(ctx)
	if err != nil || len(ops) != 1 || ops[0].WorkflowRunID != "99" {
		t.Fatalf("ops=%v err=%v", ops, err)
	}
}
