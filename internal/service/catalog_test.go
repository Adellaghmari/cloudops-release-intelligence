package service

import (
	"context"
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/repository/memory"
)

func TestCatalogFiltersSourceAndNotFound(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	now := time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC)
	_ = store.CreateService(ctx, domain.Service{
		ID: "cloudops-api", Name: "CloudOps API", Criticality: domain.CriticalityHigh,
		Source: domain.DataSourceLive, CreatedAt: now, UpdatedAt: now,
	})
	_ = store.CreateService(ctx, domain.Service{
		ID: "payments-service", Name: "Payments", Criticality: domain.CriticalityCritical,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	})
	cat := NewCatalog(store)
	src := domain.DataSourceSynthetic
	list, err := cat.ListServices(ctx, &src)
	if err != nil || len(list) != 1 || list[0].ID != "payments-service" {
		t.Fatalf("filter=%v err=%v", list, err)
	}
	_, err = cat.GetService(ctx, "no-such-service")
	if !domain.IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetReleaseIncludesRelated(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	now := time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC)
	done := now.Add(5 * time.Minute)
	_ = store.CreateService(ctx, domain.Service{
		ID: "payments-service", Name: "Payments", Criticality: domain.CriticalityCritical,
		Source: domain.DataSourceSynthetic, CreatedAt: now, UpdatedAt: now,
	})
	_ = store.CreateRelease(ctx, domain.Release{
		ID: "rel_pay", ServiceID: "payments-service", Version: "1.4.2", GitSHA: "abcdef1",
		Environment: domain.EnvironmentProd, Status: domain.ReleaseStatusDeployed,
		Source: domain.DataSourceSynthetic, CreatedAt: now,
	})
	_ = store.CreateCommit(ctx, domain.Commit{
		SHA: "abcdef1", ReleaseID: "rel_pay", ServiceID: "payments-service",
		Message: "tighten auth", Author: "northstar", FilesChanged: 12,
		CommittedAt: now,
	})
	_ = store.CreateDeployment(ctx, domain.Deployment{
		ID: "dep_pay", ReleaseID: "rel_pay", ServiceID: "payments-service",
		Environment: domain.EnvironmentProd, Status: domain.DeploymentStatusSucceeded,
		Target: "lambda:payments-service", StartedAt: now, CompletedAt: &done,
	})
	cat := NewCatalog(store)
	detail, err := cat.GetRelease(ctx, "rel_pay")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Commit == nil || detail.Deployment == nil {
		t.Fatal("expected related commit and deployment")
	}
	if detail.CIRun != nil {
		t.Fatal("missing ci run should be omitted, not invented")
	}
	if detail.Service.Source != domain.DataSourceSynthetic {
		t.Fatal("release detail must preserve source")
	}
}

func TestParseSourceQuery(t *testing.T) {
	got, err := ParseSourceQuery("")
	if err != nil || got != nil {
		t.Fatalf("empty should be nil, got %v %v", got, err)
	}
	got, err = ParseSourceQuery("live")
	if err != nil || got == nil || *got != domain.DataSourceLive {
		t.Fatalf("live: %v %v", got, err)
	}
	if _, err := ParseSourceQuery("prod"); err == nil {
		t.Fatal("expected invalid source")
	}
}
