package timeline

import (
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
)

func TestSortStable(t *testing.T) {
	t0 := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	rel := domain.ReleaseID("rel_a")
	events := []domain.ReleaseEvent{
		{ID: "evt_b", Type: domain.EventTypeCIRunCompleted, OccurredAt: t0, ReleaseID: &rel},
		{ID: "evt_a", Type: domain.EventTypeChangeCommitRecorded, OccurredAt: t0, ReleaseID: &rel},
		{ID: "evt_c", Type: domain.EventTypeDeploymentSucceeded, OccurredAt: t0.Add(time.Minute), ReleaseID: &rel},
	}
	got := FromEvents(events)
	if got[0].EventID != "evt_a" || got[1].EventID != "evt_b" || got[2].EventID != "evt_c" {
		t.Fatalf("%v %v %v", got[0].EventID, got[1].EventID, got[2].EventID)
	}
}
