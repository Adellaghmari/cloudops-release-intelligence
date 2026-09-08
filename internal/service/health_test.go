package service

import (
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/localseed"
	"github.com/adell/cloudops-release-intelligence/internal/repository/memory"
)

func TestCompareHealthSeededRelease(t *testing.T) {
	store := memory.New()
	now := time.Date(2026, 9, 8, 21, 0, 0, 0, time.UTC)
	if err := localseed.Load(t.Context(), store, now); err != nil {
		t.Fatal(err)
	}
	c := NewCatalog(store)
	a, err := c.CompareHealth(t.Context(), "rel_northstar_payments_demo", now)
	if err != nil {
		t.Fatal(err)
	}
	if a.Overall == "" || a.Correlation == "" {
		t.Fatalf("%+v", a)
	}
	if a.Disclaimer == "" {
		t.Fatal("disclaimer required")
	}
}

func TestRegressionCorrelationIgnoresOtherReleaseWindows(t *testing.T) {
	store := memory.New()
	now := time.Date(2026, 9, 8, 21, 0, 0, 0, time.UTC)
	if err := localseed.Load(t.Context(), store, now); err != nil {
		t.Fatal(err)
	}
	c := NewCatalog(store)
	a, err := c.CompareHealth(t.Context(), "rel_northstar_regression", now)
	if err != nil {
		t.Fatal(err)
	}
	if a.Overall != "SEVERELY_DEGRADED" {
		t.Fatalf("overall=%s", a.Overall)
	}
	if a.Correlation != "LIKELY_RELEASE_CORRELATION" {
		t.Fatalf("corr=%s reasons=%v", a.Correlation, a.Reasons)
	}
}

func TestCompareHealthMissingRelease(t *testing.T) {
	c := NewCatalog(memory.New())
	_, err := c.CompareHealth(t.Context(), "rel_missing", time.Now().UTC())
	if err == nil || !domain.IsNotFound(err) {
		t.Fatalf("%v", err)
	}
}
