package events

import (
	"strings"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
)

const CurrentSchema = "1.0"

type Envelope struct {
	EventID       domain.EventID
	EventType     domain.EventType
	OccurredAt    time.Time
	ReceivedAt    time.Time
	CorrelationID string
	ReleaseID     *domain.ReleaseID
	ServiceID     *domain.ServiceID
	Source        domain.EventProducer
	SchemaVersion string
	Payload       map[string]string
}

type ValidationResult struct {
	Envelope Envelope
	Err      error
	Drop     bool
	Reason   string
}

func Normalize(e Envelope, now time.Time) ValidationResult {
	e.OccurredAt = e.OccurredAt.UTC()
	e.ReceivedAt = now.UTC()
	if e.SchemaVersion == "" {
		e.SchemaVersion = CurrentSchema
	}
	if e.SchemaVersion != CurrentSchema {
		return ValidationResult{Envelope: e, Drop: true, Reason: "unsupported schema version"}
	}
	if err := e.EventID.Validate(); err != nil {
		return ValidationResult{Envelope: e, Err: err, Drop: true, Reason: "malformed event_id"}
	}
	if err := e.EventType.Validate(); err != nil {
		return ValidationResult{Envelope: e, Drop: true, Reason: "unsupported event type"}
	}
	if err := e.Source.Validate(); err != nil {
		return ValidationResult{Envelope: e, Err: err, Drop: true, Reason: "invalid source"}
	}
	if strings.TrimSpace(e.CorrelationID) == "" {
		e.CorrelationID = "corr_" + e.EventID.String()
	}
	if e.OccurredAt.IsZero() {
		return ValidationResult{Envelope: e, Drop: true, Reason: "missing occurred_at"}
	}
	return ValidationResult{Envelope: e}
}

func ToDomain(e Envelope) domain.ReleaseEvent {
	return domain.ReleaseEvent{
		ID: e.EventID, Type: e.EventType, SchemaVersion: e.SchemaVersion,
		OccurredAt: e.OccurredAt, IngestedAt: e.ReceivedAt, Producer: e.Source,
		CorrelationID: e.CorrelationID, ReleaseID: e.ReleaseID, ServiceID: e.ServiceID,
	}
}
