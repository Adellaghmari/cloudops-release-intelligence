package localseed

import (
	"context"
	"testing"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
)

func TestLoadDistinguishesLiveAndSynthetic(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	now := time.Date(2026, 9, 8, 21, 0, 0, 0, time.UTC)
	if err := Load(ctx, store, now); err != nil {
		t.Fatal(err)
	}
	live, err := store.GetService(ctx, "cloudops-api")
	if err != nil {
		t.Fatal(err)
	}
	if live.Source != domain.DataSourceLive {
		t.Fatalf("cloudops-api must be live, got %s", live.Source)
	}
	pay, err := store.GetService(ctx, "payments-service")
	if err != nil {
		t.Fatal(err)
	}
	if pay.Source != domain.DataSourceSynthetic {
		t.Fatalf("payments-service must be synthetic, got %s", pay.Source)
	}
	releases, err := store.ListReleases(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range releases {
		if r.Source == domain.DataSourceLive {
			t.Fatal("phase 1 seed must not invent live releases")
		}
	}
	if _, err := store.GetRelease(ctx, "rel_northstar_payments_demo"); err != nil {
		t.Fatal(err)
	}
	for _, id := range []domain.ReleaseID{
		"rel_northstar_risky_db", "rel_northstar_regression", "rel_northstar_blast",
		"rel_northstar_security", "rel_northstar_rollback",
	} {
		rel, err := store.GetRelease(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if rel.Source != domain.DataSourceSynthetic || rel.Scenario == "" {
			t.Fatalf("%s must be labeled synthetic scenario, got %+v", id, rel)
		}
	}
}

func TestReloadSyntheticIsIdempotentAndPreservesLive(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	now := time.Date(2026, 9, 8, 21, 0, 0, 0, time.UTC)
	if err := Load(ctx, store, now); err != nil {
		t.Fatal(err)
	}
	if err := store.PutOperationalEvidence(ctx, domain.OperationalEvidence{
		ID: "evt_live_keep", Kind: "pipeline", GitSHA: "abc1234", WorkflowRunID: "12",
		RecordedAt: now, Source: domain.DataSourceLive,
	}); err != nil {
		t.Fatal(err)
	}
	if err := ReloadSynthetic(ctx, store, now); err != nil {
		t.Fatal(err)
	}
	if err := ReloadSynthetic(ctx, store, now); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetService(ctx, "cloudops-api"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetRelease(ctx, "rel_northstar_payments_demo"); err != nil {
		t.Fatal(err)
	}
	ops, err := store.ListOperationalEvidence(ctx)
	if err != nil || len(ops) != 1 {
		t.Fatalf("live evidence=%v err=%v", ops, err)
	}
}

func TestLoadIdempotentFailureOnSecondCall(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	now := time.Now().UTC()
	if err := Load(ctx, store, now); err != nil {
		t.Fatal(err)
	}
	if err := Load(ctx, store, now); !domain.IsAlreadyExists(err) {
		t.Fatalf("second seed should hit duplicate identity, got %v", err)
	}
}
