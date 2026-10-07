package service

import (
	"context"
	"testing"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
)

func TestOverviewCompletesMissingAnalysis(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	store := memory.New()
	if err := store.CreateService(ctx, domain.Service{
		ID: "checkout-api", Name: "Checkout API", Criticality: domain.CriticalityCritical,
		Source: domain.DataSourceLive, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateRelease(ctx, domain.Release{
		ID: "rel_pending", ServiceID: "checkout-api", Version: "1.0.0", GitSHA: "abcdef1",
		Environment: domain.EnvironmentProd, Status: domain.ReleaseStatusDeployed,
		Source: domain.DataSourceLive, CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}

	overview, err := NewCatalog(store).Overview(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Releases) != 1 {
		t.Fatalf("overview should include the analyzed release: %+v", overview)
	}
	if _, err := store.GetPolicyEvaluation(ctx, "rel_pending"); err != nil {
		t.Fatalf("policy evaluation was not persisted: %v", err)
	}
	if _, err := store.GetRiskAssessment(ctx, "rel_pending"); err != nil {
		t.Fatalf("risk assessment was not persisted: %v", err)
	}
	if _, err := store.GetHealthComparison(ctx, "rel_pending"); err != nil {
		t.Fatalf("health assessment was not persisted: %v", err)
	}
	if _, err := store.GetRollbackAssessment(ctx, "rel_pending"); err != nil {
		t.Fatalf("rollback assessment was not persisted: %v", err)
	}
}

func TestOverviewDoesNotInventAttentionForClearRelease(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	store := memory.New()
	if err := store.CreateService(ctx, domain.Service{
		ID: "payments-service", Name: "Payments Service", Criticality: domain.CriticalityHigh,
		Source: domain.DataSourceLive, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateRelease(ctx, domain.Release{
		ID: "rel_clear", ServiceID: "payments-service", Version: "1.0.0", GitSHA: "abcdef1",
		Environment: domain.EnvironmentProd, Status: domain.ReleaseStatusDeployed,
		Source: domain.DataSourceLive, CreatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutPolicyEvaluation(ctx, domain.PolicyEvaluation{ReleaseID: "rel_clear", Result: "PASS"}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutRiskAssessment(ctx, domain.RiskAssessment{ReleaseID: "rel_clear", Score: 8, Category: domain.RiskLow}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutHealthComparison(ctx, domain.HealthAssessment{ReleaseID: "rel_clear", Overall: "STABLE"}); err != nil {
		t.Fatal(err)
	}
	if err := store.PutRollbackAssessment(ctx, domain.RollbackAssessment{ReleaseID: "rel_clear", Status: domain.RollbackReady}); err != nil {
		t.Fatal(err)
	}

	overview, err := NewCatalog(store).Overview(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(overview.Releases) != 1 || overview.Releases[0].AttentionLevel != "CLEAR" {
		t.Fatalf("clear release summary=%+v", overview.Releases)
	}
	if overview.AttentionReleaseID != "" {
		t.Fatalf("clear release must not become attention: %s", overview.AttentionReleaseID)
	}
}

func TestAttentionPriorityKeepsCriticalAheadOfReview(t *testing.T) {
	critical := ReleaseSummary{AttentionLevel: "CRITICAL_ATTENTION"}
	review := ReleaseSummary{
		AttentionLevel: "REVIEW",
		Risk:           domain.RiskAssessment{Score: 100},
		IncidentCount:  20,
	}
	if attentionPriority(critical) <= attentionPriority(review) {
		t.Fatalf("critical priority=%d review priority=%d", attentionPriority(critical), attentionPriority(review))
	}
}

func TestAttentionExplainsManualApproval(t *testing.T) {
	level, reasons := attention(ReleaseSummary{
		Policy: domain.PolicyEvaluation{Result: "MANUAL_APPROVAL_REQUIRED"},
	})
	if level != "REVIEW" || len(reasons) != 1 || reasons[0] != "POLICY_MANUAL_APPROVAL" {
		t.Fatalf("level=%s reasons=%v", level, reasons)
	}
}
