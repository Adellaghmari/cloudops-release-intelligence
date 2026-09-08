package domain

import (
	"testing"
	"time"
)

func ts(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func validService() Service {
	return Service{
		ID:          "payments-service",
		Name:        "Payments Service",
		Description: "Capture and authorize",
		Criticality: CriticalityCritical,
		Source:      DataSourceSynthetic,
		CreatedAt:   ts("2026-09-08T10:00:00Z"),
		UpdatedAt:   ts("2026-09-08T10:00:00Z"),
	}
}

func TestServiceValidate(t *testing.T) {
	ok := validService()
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := ok
	bad.Criticality = "NOPE"
	if err := bad.Validate(); err == nil || !IsInvalid(err) {
		t.Fatalf("expected invalid criticality, got %v", err)
	}
	bad = ok
	bad.Source = "production"
	if err := bad.Validate(); err == nil {
		t.Fatal("expected invalid source")
	}
	bad = ok
	bad.UpdatedAt = bad.CreatedAt.Add(-time.Hour)
	if err := bad.Validate(); err == nil {
		t.Fatal("expected updated_at error")
	}
}

func TestDependencyRejectsSelfEdge(t *testing.T) {
	d := Dependency{
		From:      "checkout-api",
		To:        "checkout-api",
		Kind:      DependencyKindRuntime,
		CreatedAt: ts("2026-09-08T10:00:00Z"),
	}
	if err := d.Validate(); err == nil {
		t.Fatal("expected self-dependency error")
	}
}

func TestReleaseAndRelated(t *testing.T) {
	rel := Release{
		ID:          "rel_abc",
		ServiceID:   "payments-service",
		Version:     "1.4.2",
		GitSHA:      "abc1234def",
		Environment: EnvironmentProd,
		Status:      ReleaseStatusDeployed,
		Source:      DataSourceSynthetic,
		CreatedAt:   ts("2026-09-08T10:00:00Z"),
	}
	if err := rel.Validate(); err != nil {
		t.Fatal(err)
	}
	done := ts("2026-09-08T10:05:00Z")
	dep := Deployment{
		ID:          "dep_abc",
		ReleaseID:   rel.ID,
		ServiceID:   rel.ServiceID,
		Environment: EnvironmentProd,
		Status:      DeploymentStatusSucceeded,
		Target:      "lambda:payments-service",
		ImageDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		StartedAt:   ts("2026-09-08T10:04:00Z"),
		CompletedAt: &done,
	}
	if err := dep.Validate(); err != nil {
		t.Fatal(err)
	}
	started := dep
	started.Status = DeploymentStatusStarted
	started.CompletedAt = nil
	if err := started.Validate(); err != nil {
		t.Fatal(err)
	}
	started.CompletedAt = &done
	if err := started.Validate(); err == nil {
		t.Fatal("started deployment cannot have completed_at")
	}
}

func TestHealthSnapshotBounds(t *testing.T) {
	rid := ReleaseID("rel_abc")
	h := HealthSnapshot{
		ID:           "hlt_abc",
		ServiceID:    "checkout-api",
		ReleaseID:    &rid,
		WindowKind:   HealthWindowPost,
		WindowStart:  ts("2026-09-08T10:02:00Z"),
		WindowEnd:    ts("2026-09-08T10:32:00Z"),
		RequestCount: 200,
		ErrorRate:    0.049,
		Availability: 0.988,
		LatencyP95MS: 463,
		CapturedAt:   ts("2026-09-08T10:32:00Z"),
		Source:       DataSourceSynthetic,
	}
	if err := h.Validate(); err != nil {
		t.Fatal(err)
	}
	h.ErrorRate = 1.2
	if err := h.Validate(); err == nil {
		t.Fatal("expected error_rate bound")
	}
}

func TestIncidentAndEvent(t *testing.T) {
	inc := Incident{
		ID:        "inc_abc",
		ServiceID: "checkout-api",
		Title:     "Error rate spike",
		Status:    IncidentStatusOpen,
		OpenedAt:  ts("2026-09-08T10:06:00Z"),
		Source:    DataSourceSynthetic,
	}
	if err := inc.Validate(); err != nil {
		t.Fatal(err)
	}
	inc.Status = IncidentStatusResolved
	if err := inc.Validate(); err == nil {
		t.Fatal("resolved requires resolved_at")
	}
	rel := ReleaseID("rel_abc")
	svc := ServiceID("checkout-api")
	ev := ReleaseEvent{
		ID:            "evt_abc",
		Type:          EventTypeDeploymentSucceeded,
		SchemaVersion: "1.0",
		OccurredAt:    ts("2026-09-08T10:05:00Z"),
		IngestedAt:    ts("2026-09-08T10:05:01Z"),
		Producer:      EventProducerSynthetic,
		CorrelationID: "req_abc",
		ReleaseID:     &rel,
		ServiceID:     &svc,
	}
	if err := ev.Validate(); err != nil {
		t.Fatal(err)
	}
	ev.SchemaVersion = "2.0"
	if err := ev.Validate(); err == nil {
		t.Fatal("expected schema version error")
	}
}

func TestNormalizeUTC(t *testing.T) {
	loc := time.FixedZone("CET", 2*3600)
	s := validService()
	s.CreatedAt = time.Date(2026, 9, 8, 12, 0, 0, 0, loc)
	s.UpdatedAt = s.CreatedAt
	n := s.Normalized()
	if n.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected UTC, got %v", n.CreatedAt.Location())
	}
}

func TestDecisionValidate(t *testing.T) {
	d := ReleaseDecision{
		ID:        "dec_abc",
		ReleaseID: "rel_abc",
		Decision:  DecisionHold,
		Actor:     "release-manager",
		Reason:    "migration without rollback plan",
		DecidedAt: ts("2026-09-08T10:10:00Z"),
	}
	if err := d.Validate(); err != nil {
		t.Fatal(err)
	}
}
