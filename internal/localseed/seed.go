package localseed

import (
	"context"
	"fmt"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/repository"
)

// Load writes a small local catalog so GET endpoints are inspectable without AWS.
// Live rows are this product's own service identities only. They do not invent
// production company releases. Northstar Commerce rows are explicitly synthetic.
func Load(ctx context.Context, store repository.Store, now time.Time) error {
	now = now.UTC()
	if err := seedServices(ctx, store, now); err != nil {
		return err
	}
	if err := seedDependencies(ctx, store, now); err != nil {
		return err
	}
	if err := seedSyntheticRelease(ctx, store, now); err != nil {
		return err
	}
	return seedScenarios(ctx, store, now)
}

func seedServices(ctx context.Context, store repository.Store, now time.Time) error {
	services := []domain.Service{
		svc("cloudops-api", "CloudOps API", "This project's Go API identity. Live releases appear when dogfood exists.", domain.CriticalityHigh, domain.DataSourceLive, now),
		svc("cloudops-web", "CloudOps Web", "This project's Angular console identity. No live release is seeded in Phase 1.", domain.CriticalityModerate, domain.DataSourceLive, now),
		svc("web-storefront", "Web Storefront", "Public storefront BFF", domain.CriticalityHigh, domain.DataSourceSynthetic, now),
		svc("checkout-api", "Checkout API", "Checkout orchestration", domain.CriticalityCritical, domain.DataSourceSynthetic, now),
		svc("payments-service", "Payments Service", "Capture and authorize", domain.CriticalityCritical, domain.DataSourceSynthetic, now),
		svc("inventory-service", "Inventory Service", "Availability and reservations", domain.CriticalityHigh, domain.DataSourceSynthetic, now),
		svc("customer-api", "Customer API", "Accounts and profiles", domain.CriticalityHigh, domain.DataSourceSynthetic, now),
		svc("fulfillment-api", "Fulfillment API", "Post-order fulfillment", domain.CriticalityHigh, domain.DataSourceSynthetic, now),
		svc("notification-worker", "Notification Worker", "Email/SMS/push worker", domain.CriticalityModerate, domain.DataSourceSynthetic, now),
	}
	for _, s := range services {
		if err := store.CreateService(ctx, s); err != nil {
			return fmt.Errorf("seed service %s: %w", s.ID, err)
		}
	}
	return nil
}

func seedDependencies(ctx context.Context, store repository.Store, now time.Time) error {
	edges := [][2]string{
		{"web-storefront", "checkout-api"},
		{"web-storefront", "customer-api"},
		{"web-storefront", "inventory-service"},
		{"checkout-api", "payments-service"},
		{"checkout-api", "customer-api"},
		{"checkout-api", "inventory-service"},
		{"checkout-api", "fulfillment-api"},
		{"payments-service", "notification-worker"},
		{"fulfillment-api", "inventory-service"},
		{"fulfillment-api", "notification-worker"},
	}
	for _, e := range edges {
		d := domain.Dependency{
			From:      domain.ServiceID(e[0]),
			To:        domain.ServiceID(e[1]),
			Kind:      domain.DependencyKindRuntime,
			CreatedAt: now,
		}
		if e[0] == "payments-service" {
			d.Kind = domain.DependencyKindAsync
		}
		if err := store.CreateDependency(ctx, d); err != nil {
			return fmt.Errorf("seed dependency %s: %w", d.Key(), err)
		}
	}
	return nil
}

func seedSyntheticRelease(ctx context.Context, store repository.Store, now time.Time) error {
	if err := seedPriorPayments(ctx, store, now); err != nil {
		return err
	}
	relID := domain.ReleaseID("rel_northstar_payments_demo")
	sha := domain.CommitSHA("c0ffee1")
	started := now.Add(-40 * time.Minute)
	done := now.Add(-38 * time.Minute)
	rel := domain.Release{
		ID:          relID,
		ServiceID:   "payments-service",
		Version:     "0.1.0-demo",
		GitSHA:      sha,
		Environment: domain.EnvironmentLocal,
		Status:      domain.ReleaseStatusDeployed,
		Source:      domain.DataSourceSynthetic,
		Scenario:    "SAFE_RELEASE",
		CreatedAt:   started,
	}
	if err := store.CreateRelease(ctx, rel); err != nil {
		return err
	}
	if err := store.CreateCommit(ctx, domain.Commit{
		SHA: sha, ReleaseID: relID, ServiceID: "payments-service",
		Message: "demo: small notification template tweak",
		Author:  "northstar-demo", FilesChanged: 3, LinesAdded: 18, LinesDeleted: 4,
		CommittedAt: started.Add(-20 * time.Minute),
	}); err != nil {
		return err
	}
	if err := store.CreateCIRun(ctx, domain.CIRun{
		ID: "ci_northstar_payments_demo", ReleaseID: relID,
		WorkflowName: "northstar-ci", Status: domain.CIRunStatusSucceeded,
		ExternalRunID: "demo-run-1", StartedAt: started.Add(-15 * time.Minute), CompletedAt: &started,
	}); err != nil {
		return err
	}
	if err := store.CreateDeployment(ctx, domain.Deployment{
		ID: "dep_northstar_payments_demo", ReleaseID: relID, ServiceID: "payments-service",
		Environment: domain.EnvironmentLocal, Status: domain.DeploymentStatusSucceeded,
		Target: "local:payments-service", ArtifactURI: "s3://northstar-demo/payments/0.1.0.zip",
		ImageDigest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		StartedAt:   started, CompletedAt: &done,
	}); err != nil {
		return err
	}
	svcID := domain.ServiceID("payments-service")
	if err := seedHealthPair(ctx, store, relID, svcID, done, now, 400, 0.007, 0.9996, 182); err != nil {
		return err
	}
	return seedTimeline(ctx, store, now, relID, svcID, started, done, true, false)
}

func seedPriorPayments(ctx context.Context, store repository.Store, now time.Time) error {
	relID := domain.ReleaseID("rel_northstar_payments_prior")
	sha := domain.CommitSHA("bada111")
	started := now.Add(-7 * 24 * time.Hour)
	done := started.Add(5 * time.Minute)
	if err := store.CreateRelease(ctx, domain.Release{
		ID: relID, ServiceID: "payments-service", Version: "0.0.9-demo", GitSHA: sha,
		Environment: domain.EnvironmentLocal, Status: domain.ReleaseStatusDeployed,
		Source: domain.DataSourceSynthetic, CreatedAt: started,
	}); err != nil {
		return err
	}
	if err := store.CreateCommit(ctx, domain.Commit{
		SHA: sha, ReleaseID: relID, ServiceID: "payments-service",
		Message: "demo: prior stable payments build", Author: "northstar-demo",
		FilesChanged: 2, LinesAdded: 8, LinesDeleted: 1, CommittedAt: started.Add(-30 * time.Minute),
	}); err != nil {
		return err
	}
	if err := store.CreateDeployment(ctx, domain.Deployment{
		ID: "dep_northstar_payments_prior", ReleaseID: relID, ServiceID: "payments-service",
		Environment: domain.EnvironmentLocal, Status: domain.DeploymentStatusSucceeded,
		Target: "local:payments-service", ArtifactURI: "s3://northstar-demo/payments/0.0.9.zip",
		ImageDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		StartedAt:   started, CompletedAt: &done,
	}); err != nil {
		return err
	}
	return seedHealthPair(ctx, store, relID, "payments-service", done, now, 300, 0.006, 0.9997, 170)
}

func seedHealthPair(ctx context.Context, store repository.Store, rel domain.ReleaseID, svc domain.ServiceID, deployed, now time.Time, req int, errRate, avail, p95 float64) error {
	baseStart := deployed.Add(-60 * time.Minute)
	postStart := deployed.Add(2 * time.Minute)
	postEnd := deployed.Add(32 * time.Minute)
	if err := store.CreateHealthSnapshot(ctx, domain.HealthSnapshot{
		ID: domain.HealthSnapshotID("hlt_" + rel.String() + "_base"), ServiceID: svc, ReleaseID: &rel,
		WindowKind: domain.HealthWindowBaseline, WindowStart: baseStart, WindowEnd: deployed,
		RequestCount: req, ErrorRate: errRate, Availability: avail, LatencyP50MS: p95 * 0.7, LatencyP95MS: p95, LatencyP99MS: p95 * 1.2,
		CapturedAt: deployed, Source: domain.DataSourceSynthetic,
	}); err != nil {
		return err
	}
	return store.CreateHealthSnapshot(ctx, domain.HealthSnapshot{
		ID: domain.HealthSnapshotID("hlt_" + rel.String() + "_post"), ServiceID: svc, ReleaseID: &rel,
		WindowKind: domain.HealthWindowPost, WindowStart: postStart, WindowEnd: postEnd,
		RequestCount: req, ErrorRate: errRate, Availability: avail, LatencyP50MS: p95 * 0.7, LatencyP95MS: p95, LatencyP99MS: p95 * 1.2,
		CapturedAt: now, Source: domain.DataSourceSynthetic,
	})
}

func svc(id, name, desc string, crit domain.Criticality, src domain.DataSource, now time.Time) domain.Service {
	return domain.Service{
		ID:          domain.ServiceID(id),
		Name:        name,
		Description: desc,
		Criticality: crit,
		Source:      src,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
