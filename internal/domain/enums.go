package domain

import "strings"

// DataSource distinguishes this project's own operational metadata from
// Northstar Commerce demo data. Synthetic records must never be presented
// as live production-company evidence.
type DataSource string

const (
	DataSourceLive      DataSource = "live"
	DataSourceSynthetic DataSource = "synthetic"
)

type Criticality string

const (
	CriticalityLow      Criticality = "LOW"
	CriticalityModerate Criticality = "MODERATE"
	CriticalityHigh     Criticality = "HIGH"
	CriticalityCritical Criticality = "CRITICAL"
)

type Environment string

const (
	EnvironmentLocal   Environment = "local"
	EnvironmentStaging Environment = "staging"
	EnvironmentProd    Environment = "prod"
)

type ReleaseStatus string

const (
	ReleaseStatusPending    ReleaseStatus = "pending"
	ReleaseStatusDeployed   ReleaseStatus = "deployed"
	ReleaseStatusFailed     ReleaseStatus = "failed"
	ReleaseStatusRolledBack ReleaseStatus = "rolled_back"
)

type DeploymentStatus string

const (
	DeploymentStatusStarted   DeploymentStatus = "started"
	DeploymentStatusSucceeded DeploymentStatus = "succeeded"
	DeploymentStatusFailed    DeploymentStatus = "failed"
)

type CIRunStatus string

const (
	CIRunStatusStarted   CIRunStatus = "started"
	CIRunStatusSucceeded CIRunStatus = "succeeded"
	CIRunStatusFailed    CIRunStatus = "failed"
)

type IncidentStatus string

const (
	IncidentStatusOpen     IncidentStatus = "open"
	IncidentStatusResolved IncidentStatus = "resolved"
)

type DependencyKind string

const (
	DependencyKindRuntime DependencyKind = "runtime"
	DependencyKindAsync   DependencyKind = "async"
)

type HealthWindowKind string

const (
	HealthWindowBaseline HealthWindowKind = "baseline"
	HealthWindowPost     HealthWindowKind = "post"
)

type DecisionKind string

const (
	DecisionProceed         DecisionKind = "PROCEED"
	DecisionHold            DecisionKind = "HOLD"
	DecisionManualApprove   DecisionKind = "MANUAL_APPROVE"
	DecisionRollbackPrepare DecisionKind = "ROLLBACK_PREPARE"
)

type EventProducer string

const (
	EventProducerGitHub    EventProducer = "github-actions"
	EventProducerSynthetic EventProducer = "synthetic-demo"
	EventProducerAPI       EventProducer = "cloudops-api"
)

type EventType string

const (
	EventTypeChangeCommitRecorded   EventType = "change.commit.recorded"
	EventTypeCIRunStarted           EventType = "ci.run.started"
	EventTypeCIRunCompleted         EventType = "ci.run.completed"
	EventTypeSecurityScanCompleted  EventType = "security.scan.completed"
	EventTypeArtifactPublished      EventType = "artifact.published"
	EventTypeDeploymentStarted      EventType = "deployment.started"
	EventTypeDeploymentSucceeded    EventType = "deployment.succeeded"
	EventTypeDeploymentFailed       EventType = "deployment.failed"
	EventTypeHealthSnapshotRecorded EventType = "health.snapshot.recorded"
	EventTypeIncidentOpened         EventType = "incident.opened"
	EventTypeIncidentResolved       EventType = "incident.resolved"
	EventTypeDecisionRecorded       EventType = "decision.recorded"
	EventTypePolicyEvaluated        EventType = "policy.evaluated"
)

func ParseDataSource(raw string) (DataSource, error) {
	v := DataSource(strings.TrimSpace(raw))
	switch v {
	case DataSourceLive, DataSourceSynthetic:
		return v, nil
	default:
		return "", ValidationError{Field: "source", Message: "must be live or synthetic"}
	}
}

func ParseCriticality(raw string) (Criticality, error) {
	v := Criticality(strings.ToUpper(strings.TrimSpace(raw)))
	switch v {
	case CriticalityLow, CriticalityModerate, CriticalityHigh, CriticalityCritical:
		return v, nil
	default:
		return "", ValidationError{Field: "criticality", Message: "must be LOW, MODERATE, HIGH, or CRITICAL"}
	}
}

func ParseEnvironment(raw string) (Environment, error) {
	v := Environment(strings.TrimSpace(raw))
	switch v {
	case EnvironmentLocal, EnvironmentStaging, EnvironmentProd:
		return v, nil
	default:
		return "", ValidationError{Field: "environment", Message: "must be local, staging, or prod"}
	}
}

func ParseReleaseStatus(raw string) (ReleaseStatus, error) {
	v := ReleaseStatus(strings.TrimSpace(raw))
	switch v {
	case ReleaseStatusPending, ReleaseStatusDeployed, ReleaseStatusFailed, ReleaseStatusRolledBack:
		return v, nil
	default:
		return "", ValidationError{Field: "status", Message: "invalid release status"}
	}
}

func ParseDeploymentStatus(raw string) (DeploymentStatus, error) {
	v := DeploymentStatus(strings.TrimSpace(raw))
	switch v {
	case DeploymentStatusStarted, DeploymentStatusSucceeded, DeploymentStatusFailed:
		return v, nil
	default:
		return "", ValidationError{Field: "status", Message: "invalid deployment status"}
	}
}

func ParseCIRunStatus(raw string) (CIRunStatus, error) {
	v := CIRunStatus(strings.TrimSpace(raw))
	switch v {
	case CIRunStatusStarted, CIRunStatusSucceeded, CIRunStatusFailed:
		return v, nil
	default:
		return "", ValidationError{Field: "status", Message: "invalid ci run status"}
	}
}

func ParseIncidentStatus(raw string) (IncidentStatus, error) {
	v := IncidentStatus(strings.TrimSpace(raw))
	switch v {
	case IncidentStatusOpen, IncidentStatusResolved:
		return v, nil
	default:
		return "", ValidationError{Field: "status", Message: "invalid incident status"}
	}
}

func ParseDependencyKind(raw string) (DependencyKind, error) {
	v := DependencyKind(strings.TrimSpace(raw))
	switch v {
	case DependencyKindRuntime, DependencyKindAsync:
		return v, nil
	default:
		return "", ValidationError{Field: "kind", Message: "must be runtime or async"}
	}
}

func ParseHealthWindowKind(raw string) (HealthWindowKind, error) {
	v := HealthWindowKind(strings.TrimSpace(raw))
	switch v {
	case HealthWindowBaseline, HealthWindowPost:
		return v, nil
	default:
		return "", ValidationError{Field: "window_kind", Message: "must be baseline or post"}
	}
}

func ParseDecisionKind(raw string) (DecisionKind, error) {
	v := DecisionKind(strings.ToUpper(strings.TrimSpace(raw)))
	switch v {
	case DecisionProceed, DecisionHold, DecisionManualApprove, DecisionRollbackPrepare:
		return v, nil
	default:
		return "", ValidationError{Field: "decision", Message: "invalid release decision"}
	}
}

func ParseEventProducer(raw string) (EventProducer, error) {
	v := EventProducer(strings.TrimSpace(raw))
	switch v {
	case EventProducerGitHub, EventProducerSynthetic, EventProducerAPI:
		return v, nil
	default:
		return "", ValidationError{Field: "producer", Message: "invalid event producer"}
	}
}

func ParseEventType(raw string) (EventType, error) {
	v := EventType(strings.TrimSpace(raw))
	switch v {
	case EventTypeChangeCommitRecorded, EventTypeCIRunStarted, EventTypeCIRunCompleted,
		EventTypeSecurityScanCompleted, EventTypeArtifactPublished, EventTypeDeploymentStarted,
		EventTypeDeploymentSucceeded, EventTypeDeploymentFailed, EventTypeHealthSnapshotRecorded,
		EventTypeIncidentOpened, EventTypeIncidentResolved, EventTypeDecisionRecorded,
		EventTypePolicyEvaluated:
		return v, nil
	default:
		return "", ValidationError{Field: "event_type", Message: "unknown event type"}
	}
}

func (v DataSource) Validate() error {
	_, err := ParseDataSource(string(v))
	return err
}

func (v Criticality) Validate() error {
	_, err := ParseCriticality(string(v))
	return err
}

func (v Environment) Validate() error {
	_, err := ParseEnvironment(string(v))
	return err
}

func (v ReleaseStatus) Validate() error {
	_, err := ParseReleaseStatus(string(v))
	return err
}

func (v DeploymentStatus) Validate() error {
	_, err := ParseDeploymentStatus(string(v))
	return err
}

func (v CIRunStatus) Validate() error {
	_, err := ParseCIRunStatus(string(v))
	return err
}

func (v IncidentStatus) Validate() error {
	_, err := ParseIncidentStatus(string(v))
	return err
}

func (v DependencyKind) Validate() error {
	_, err := ParseDependencyKind(string(v))
	return err
}

func (v HealthWindowKind) Validate() error {
	_, err := ParseHealthWindowKind(string(v))
	return err
}

func (v DecisionKind) Validate() error {
	_, err := ParseDecisionKind(string(v))
	return err
}

func (v EventProducer) Validate() error {
	_, err := ParseEventProducer(string(v))
	return err
}

func (v EventType) Validate() error {
	_, err := ParseEventType(string(v))
	return err
}
