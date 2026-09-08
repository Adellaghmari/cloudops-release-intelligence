package service

import (
	"testing"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/localseed"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
)

func TestImpactPaymentsHasDependents(t *testing.T) {
	store := memory.New()
	if err := localseed.Load(t.Context(), store, time.Date(2026, 9, 8, 21, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	c := NewCatalog(store)
	r, err := c.Impact(t.Context(), "rel_northstar_payments_demo")
	if err != nil {
		t.Fatal(err)
	}
	if r.Unknown || len(r.DirectDependents) == 0 {
		t.Fatalf("%+v", r)
	}
	if r.Disclaimer == "" {
		t.Fatal("disclaimer required")
	}
}
