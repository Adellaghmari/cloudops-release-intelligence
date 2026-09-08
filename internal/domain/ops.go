package domain

import "time"

type HealthSnapshot struct {
	ID           HealthSnapshotID
	ServiceID    ServiceID
	ReleaseID    *ReleaseID
	WindowKind   HealthWindowKind
	WindowStart  time.Time
	WindowEnd    time.Time
	RequestCount int
	ErrorRate    float64
	Availability float64
	LatencyP50MS float64
	LatencyP95MS float64
	LatencyP99MS float64
	CPUPct       *float64
	MemoryPct    *float64
	QueueBacklog *int
	CapturedAt   time.Time
	Source       DataSource
}

func (h HealthSnapshot) Validate() error {
	if err := h.ID.Validate(); err != nil {
		return err
	}
	if err := h.ServiceID.Validate(); err != nil {
		return err
	}
	if h.ReleaseID != nil {
		if err := h.ReleaseID.Validate(); err != nil {
			return err
		}
	}
	if err := h.WindowKind.Validate(); err != nil {
		return err
	}
	if err := requireTime("window_start", h.WindowStart); err != nil {
		return err
	}
	if err := requireTime("window_end", h.WindowEnd); err != nil {
		return err
	}
	if !h.WindowEnd.After(h.WindowStart) {
		return ValidationError{Field: "window_end", Message: "must be after window_start"}
	}
	if h.RequestCount < 0 {
		return ValidationError{Field: "request_count", Message: "must not be negative"}
	}
	if h.ErrorRate < 0 || h.ErrorRate > 1 {
		return ValidationError{Field: "error_rate", Message: "must be between 0 and 1"}
	}
	if h.Availability < 0 || h.Availability > 1 {
		return ValidationError{Field: "availability", Message: "must be between 0 and 1"}
	}
	if h.LatencyP50MS < 0 || h.LatencyP95MS < 0 || h.LatencyP99MS < 0 {
		return ValidationError{Field: "latency", Message: "must not be negative"}
	}
	if err := h.Source.Validate(); err != nil {
		return err
	}
	return requireTime("captured_at", h.CapturedAt)
}

func (h HealthSnapshot) Normalized() HealthSnapshot {
	h.WindowStart = UTC(h.WindowStart)
	h.WindowEnd = UTC(h.WindowEnd)
	h.CapturedAt = UTC(h.CapturedAt)
	h.ReleaseID = clonePtr(h.ReleaseID)
	h.CPUPct = clonePtr(h.CPUPct)
	h.MemoryPct = clonePtr(h.MemoryPct)
	h.QueueBacklog = clonePtr(h.QueueBacklog)
	return h
}

type Incident struct {
	ID         IncidentID
	ServiceID  ServiceID
	ReleaseID  *ReleaseID
	Title      string
	Status     IncidentStatus
	OpenedAt   time.Time
	ResolvedAt *time.Time
	Source     DataSource
}

func (i Incident) Validate() error {
	if err := i.ID.Validate(); err != nil {
		return err
	}
	if err := i.ServiceID.Validate(); err != nil {
		return err
	}
	if i.ReleaseID != nil {
		if err := i.ReleaseID.Validate(); err != nil {
			return err
		}
	}
	if err := requireName("title", i.Title, 200); err != nil {
		return err
	}
	if err := i.Status.Validate(); err != nil {
		return err
	}
	if err := requireTime("opened_at", i.OpenedAt); err != nil {
		return err
	}
	if i.Status == IncidentStatusResolved && i.ResolvedAt == nil {
		return ValidationError{Field: "resolved_at", Message: "is required when incident is resolved"}
	}
	if i.Status == IncidentStatusOpen && i.ResolvedAt != nil {
		return ValidationError{Field: "resolved_at", Message: "must be empty while open"}
	}
	if err := i.Source.Validate(); err != nil {
		return err
	}
	return nil
}

func (i Incident) Normalized() Incident {
	i.OpenedAt = UTC(i.OpenedAt)
	if i.ResolvedAt != nil {
		t := UTC(*i.ResolvedAt)
		i.ResolvedAt = &t
	}
	i.ReleaseID = clonePtr(i.ReleaseID)
	return i
}

type ReleaseEvent struct {
	ID            EventID
	Type          EventType
	SchemaVersion string
	OccurredAt    time.Time
	IngestedAt    time.Time
	Producer      EventProducer
	CorrelationID string
	ReleaseID     *ReleaseID
	ServiceID     *ServiceID
	DeploymentID  *DeploymentID
}

func (e ReleaseEvent) Validate() error {
	if err := e.ID.Validate(); err != nil {
		return err
	}
	if err := e.Type.Validate(); err != nil {
		return err
	}
	if e.SchemaVersion != "1.0" {
		return ValidationError{Field: "schema_version", Message: "must be 1.0"}
	}
	if err := requireTime("occurred_at", e.OccurredAt); err != nil {
		return err
	}
	if err := requireTime("ingested_at", e.IngestedAt); err != nil {
		return err
	}
	if err := e.Producer.Validate(); err != nil {
		return err
	}
	if err := validateID("correlation_id", e.CorrelationID); err != nil {
		return err
	}
	if e.ReleaseID != nil {
		if err := e.ReleaseID.Validate(); err != nil {
			return err
		}
	}
	if e.ServiceID != nil {
		if err := e.ServiceID.Validate(); err != nil {
			return err
		}
	}
	if e.DeploymentID != nil {
		if err := e.DeploymentID.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (e ReleaseEvent) Normalized() ReleaseEvent {
	e.OccurredAt = UTC(e.OccurredAt)
	e.IngestedAt = UTC(e.IngestedAt)
	e.ReleaseID = clonePtr(e.ReleaseID)
	e.ServiceID = clonePtr(e.ServiceID)
	e.DeploymentID = clonePtr(e.DeploymentID)
	return e
}
