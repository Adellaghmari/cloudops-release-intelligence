package policy

import (
	"context"
	"testing"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
)

func TestRulesFireAndPrecedence(t *testing.T) {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	block := Evaluate(context.Background(), Input{
		ReleaseID: "rel_a", ServiceID: "payments-service", Criticality: "HIGH",
		FailedTests: 2, CIStatus: "failed", Severity: "NONE", RiskScore: 10, RiskCategory: "LOW",
		HealthAvail: false, Rollback: "READY", Phase: "PRE_DEPLOY",
	}, now)
	if block.Result != "BLOCK" || !hasRule(block, "tests.must_pass", "BLOCK") {
		t.Fatalf("tests must block: %+v", block)
	}

	sec := Evaluate(context.Background(), Input{
		ReleaseID: "rel_a", ServiceID: "payments-service", Criticality: "HIGH",
		Severity: "CRITICAL", HealthAvail: false, Rollback: "READY", Phase: "PRE_DEPLOY",
	}, now)
	if sec.Result != "BLOCK" || !hasRule(sec, "security.critical_vuln", "BLOCK") {
		t.Fatalf("%+v", sec)
	}

	crit := Evaluate(context.Background(), Input{
		ReleaseID: "rel_a", ServiceID: "checkout-api", Criticality: "CRITICAL",
		Severity: "NONE", RiskScore: 80, RiskCategory: "CRITICAL",
		HealthAvail: false, Rollback: "READY", Phase: "PRE_DEPLOY",
	}, now)
	if crit.Result != "MANUAL_APPROVAL_REQUIRED" {
		t.Fatalf("precedence: %s %+v", crit.Result, crit.Rules)
	}

	warn := Evaluate(context.Background(), Input{
		ReleaseID: "rel_a", ServiceID: "notification-worker", Criticality: "MODERATE",
		Severity: "HIGH", RiskScore: 55, RiskCategory: "HIGH", Migration: true, Rollback: "PARTIAL",
		HealthAvail: false, Phase: "PRE_DEPLOY",
	}, now)
	if warn.Result != "WARN" {
		t.Fatalf("%s %+v", warn.Result, warn.Rules)
	}

	health := Evaluate(context.Background(), Input{
		ReleaseID: "rel_a", ServiceID: "payments-service", Criticality: "HIGH",
		Severity: "NONE", HealthAvail: true, HealthOverall: "SEVERELY_DEGRADED",
		Rollback: "READY", Phase: "POST_DEPLOY",
	}, now)
	if health.Result != "BLOCK" || !hasRule(health, "health.severe_regression", "BLOCK") {
		t.Fatalf("%+v", health)
	}
}

func TestHealthSkippedPreDeploy(t *testing.T) {
	eval := Evaluate(context.Background(), Input{
		ReleaseID: "rel_a", ServiceID: "payments-service", Criticality: "HIGH",
		Severity: "NONE", HealthAvail: false, Rollback: "READY", Phase: "PRE_DEPLOY",
	}, time.Now().UTC())
	for _, r := range eval.Rules {
		if r.ID == "health.severe_regression" && !r.Skipped {
			t.Fatal("health rules must be skipped without a post window")
		}
	}
	if eval.PolicyVersion == "" {
		t.Fatal("version required")
	}
}

func TestFailClosed(t *testing.T) {
	eval := failClosed(domain.PolicyEvaluation{ReleaseID: "rel_a"}, context.Canceled)
	if eval.Result != "BLOCK" || eval.Rules[0].ID != engineRule {
		t.Fatalf("%+v", eval)
	}
}

func hasRule(eval domain.PolicyEvaluation, id, result string) bool {
	for _, r := range eval.Rules {
		if r.ID == id && r.Result == result && !r.Skipped {
			return true
		}
	}
	return false
}
