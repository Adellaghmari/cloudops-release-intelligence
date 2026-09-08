package timeline

import (
	"sort"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
)

type Entry struct {
	EventID    domain.EventID
	Type       domain.EventType
	OccurredAt time.Time
	Producer   domain.EventProducer
	Summary    string
	ReleaseID  *domain.ReleaseID
	ServiceID  *domain.ServiceID
}

func FromEvents(events []domain.ReleaseEvent) []Entry {
	out := make([]Entry, 0, len(events))
	for _, e := range events {
		out = append(out, Entry{
			EventID: e.ID, Type: e.Type, OccurredAt: e.OccurredAt, Producer: e.Producer,
			Summary: string(e.Type), ReleaseID: e.ReleaseID, ServiceID: e.ServiceID,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].EventID.String() < out[j].EventID.String()
		}
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out
}
