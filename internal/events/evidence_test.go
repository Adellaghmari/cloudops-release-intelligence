package events

import (
	"context"
	"testing"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository/memory"
)

func TestLiveEnvelopeBecomesOperationalEvidence(t *testing.T) {
	store := memory.New()
	p := NewProcessor(store, 3)
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	env := Envelope{
		EventID: "evt_gha_1", EventType: domain.EventTypeCIRunCompleted, OccurredAt: now,
		Source: domain.EventProducerGitHub, SchemaVersion: CurrentSchema,
		Payload: map[string]string{
			"git_sha": "abcdef1", "workflow_run_id": "77", "workflow_result": "success",
		},
	}
	if _, err := p.Handle(context.Background(), env, now); err != nil {
		t.Fatal(err)
	}
	list, err := store.ListOperationalEvidence(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	if list[0].Source != domain.DataSourceLive || list[0].WorkflowRunID != "77" {
		t.Fatalf("%+v", list[0])
	}
}

func TestSyntheticEnvelopeIsNotLiveEvidence(t *testing.T) {
	store := memory.New()
	p := NewProcessor(store, 3)
	now := time.Now().UTC()
	env := Envelope{
		EventID: "evt_syn_1", EventType: domain.EventTypeCIRunCompleted, OccurredAt: now,
		Source: domain.EventProducerSynthetic, SchemaVersion: CurrentSchema,
	}
	if _, err := p.Handle(context.Background(), env, now); err != nil {
		t.Fatal(err)
	}
	list, err := store.ListOperationalEvidence(context.Background())
	if err != nil || len(list) != 0 {
		t.Fatalf("synthetic must not become live evidence: %v", list)
	}
}
