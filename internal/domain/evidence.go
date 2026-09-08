package domain

import "time"

// OperationalEvidence is real Project B pipeline/deploy metadata.
// It is never synthesized. Demo reset must not delete it.
type OperationalEvidence struct {
	ID                  EventID
	Kind                string
	GitSHA              string
	Branch              string
	WorkflowName        string
	WorkflowRunID       string
	WorkflowResult      string
	TestResult          string
	BuildDurationMS     *int
	ImageDigest         string
	ArtifactID          string
	SecurityScanResult  string
	DeploymentTimestamp *time.Time
	RecordedAt          time.Time
	Source              DataSource
}

func (e OperationalEvidence) Validate() error {
	if err := e.ID.Validate(); err != nil {
		return err
	}
	switch e.Kind {
	case "pipeline", "security", "artifact", "deploy":
	default:
		return ValidationError{Field: "kind", Message: "must be pipeline, security, artifact, or deploy"}
	}
	if e.Source != DataSourceLive {
		return ValidationError{Field: "source", Message: "operational evidence must be live"}
	}
	return requireTime("recorded_at", e.RecordedAt)
}

func (e OperationalEvidence) Normalized() OperationalEvidence {
	e.RecordedAt = UTC(e.RecordedAt)
	if e.DeploymentTimestamp != nil {
		t := UTC(*e.DeploymentTimestamp)
		e.DeploymentTimestamp = &t
	}
	e.BuildDurationMS = clonePtr(e.BuildDurationMS)
	return e
}
