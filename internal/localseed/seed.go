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
	return seedSyntheticRelease(ctx, store, now)
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
		Target: "local:payments-service", StartedAt: started, CompletedAt: &done,
	}); err != nil {
		return err
	}
	svc := domain.ServiceID("payments-service")
	return store.CreateEvent(ctx, domain.ReleaseEvent{
		ID:            "evt_northstar_payments_demo_deploy",
		Type:          domain.EventTypeDeploymentSucceeded,
		SchemaVersion: "1.0",
		OccurredAt:    done,
		IngestedAt:    now,
		Producer:      domain.EventProducerSynthetic,
		CorrelationID: "req_local_seed",
		ReleaseID:     &relID,
		ServiceID:     &svc,
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
