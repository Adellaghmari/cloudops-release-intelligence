package domain

import "time"

type Release struct {
	ID          ReleaseID
	ServiceID   ServiceID
	Version     string
	GitSHA      CommitSHA
	Environment Environment
	Status      ReleaseStatus
	Source      DataSource
	CreatedAt   time.Time
}

func (r Release) Validate() error {
	if err := r.ID.Validate(); err != nil {
		return err
	}
	if err := r.ServiceID.Validate(); err != nil {
		return err
	}
	if err := requireName("version", r.Version, 64); err != nil {
		return err
	}
	if err := r.GitSHA.Validate(); err != nil {
		return err
	}
	if err := r.Environment.Validate(); err != nil {
		return err
	}
	if err := r.Status.Validate(); err != nil {
		return err
	}
	if err := r.Source.Validate(); err != nil {
		return err
	}
	return requireTime("created_at", r.CreatedAt)
}

func (r Release) Normalized() Release {
	r.CreatedAt = UTC(r.CreatedAt)
	return r
}

type Deployment struct {
	ID          DeploymentID
	ReleaseID   ReleaseID
	ServiceID   ServiceID
	Environment Environment
	Status      DeploymentStatus
	Target      string
	ImageDigest string
	ArtifactURI string
	StartedAt   time.Time
	CompletedAt *time.Time
}

func (d Deployment) Validate() error {
	if err := d.ID.Validate(); err != nil {
		return err
	}
	if err := d.ReleaseID.Validate(); err != nil {
		return err
	}
	if err := d.ServiceID.Validate(); err != nil {
		return err
	}
	if err := d.Environment.Validate(); err != nil {
		return err
	}
	if err := d.Status.Validate(); err != nil {
		return err
	}
	if err := requireName("target", d.Target, 200); err != nil {
		return err
	}
	if err := requireTime("started_at", d.StartedAt); err != nil {
		return err
	}
	if d.Status != DeploymentStatusStarted && d.CompletedAt == nil {
		return ValidationError{Field: "completed_at", Message: "is required when deployment is finished"}
	}
	if d.Status == DeploymentStatusStarted && d.CompletedAt != nil {
		return ValidationError{Field: "completed_at", Message: "must be empty while started"}
	}
	if d.CompletedAt != nil && d.CompletedAt.Before(d.StartedAt) {
		return ValidationError{Field: "completed_at", Message: "must not precede started_at"}
	}
	if d.ImageDigest != "" && !validImageDigest(d.ImageDigest) {
		return ValidationError{Field: "image_digest", Message: "must be sha256:<hex> when set"}
	}
	return nil
}

func (d Deployment) Normalized() Deployment {
	d.StartedAt = UTC(d.StartedAt)
	if d.CompletedAt != nil {
		t := UTC(*d.CompletedAt)
		d.CompletedAt = &t
	}
	return d
}

type Commit struct {
	SHA                 CommitSHA
	ReleaseID           ReleaseID
	ServiceID           ServiceID
	Message             string
	Author              string
	FilesChanged        int
	LinesAdded          int
	LinesDeleted        int
	MigrationPresent    bool
	MigrationReversible *bool
	ConfigChangePresent bool
	CommittedAt         time.Time
}

func (c Commit) Validate() error {
	if err := c.SHA.Validate(); err != nil {
		return err
	}
	if err := c.ReleaseID.Validate(); err != nil {
		return err
	}
	if err := c.ServiceID.Validate(); err != nil {
		return err
	}
	if err := requireName("message", c.Message, 500); err != nil {
		return err
	}
	if err := requireName("author", c.Author, 120); err != nil {
		return err
	}
	if c.FilesChanged < 0 || c.LinesAdded < 0 || c.LinesDeleted < 0 {
		return ValidationError{Field: "change_size", Message: "must not be negative"}
	}
	return requireTime("committed_at", c.CommittedAt)
}

func (c Commit) Normalized() Commit {
	c.CommittedAt = UTC(c.CommittedAt)
	return c
}

type CIRun struct {
	ID             CIRunID
	ReleaseID      ReleaseID
	WorkflowName   string
	Status         CIRunStatus
	FailedTests    int
	FailedAttempts int
	ExternalRunID  string
	StartedAt      time.Time
	CompletedAt    *time.Time
}

func (r CIRun) Validate() error {
	if err := r.ID.Validate(); err != nil {
		return err
	}
	if err := r.ReleaseID.Validate(); err != nil {
		return err
	}
	if err := requireName("workflow_name", r.WorkflowName, 120); err != nil {
		return err
	}
	if err := r.Status.Validate(); err != nil {
		return err
	}
	if r.FailedTests < 0 || r.FailedAttempts < 0 {
		return ValidationError{Field: "failures", Message: "must not be negative"}
	}
	if err := requireTime("started_at", r.StartedAt); err != nil {
		return err
	}
	if r.Status != CIRunStatusStarted && r.CompletedAt == nil {
		return ValidationError{Field: "completed_at", Message: "is required when ci run is finished"}
	}
	if r.ExternalRunID != "" {
		if err := validateID("external_run_id", r.ExternalRunID); err != nil {
			return err
		}
	}
	return nil
}

func (r CIRun) Normalized() CIRun {
	r.StartedAt = UTC(r.StartedAt)
	if r.CompletedAt != nil {
		t := UTC(*r.CompletedAt)
		r.CompletedAt = &t
	}
	return r
}

type ReleaseDecision struct {
	ID        DecisionID
	ReleaseID ReleaseID
	Decision  DecisionKind
	Actor     string
	Reason    string
	DecidedAt time.Time
}

func (d ReleaseDecision) Validate() error {
	if err := d.ID.Validate(); err != nil {
		return err
	}
	if err := d.ReleaseID.Validate(); err != nil {
		return err
	}
	if err := d.Decision.Validate(); err != nil {
		return err
	}
	if err := requireName("actor", d.Actor, 120); err != nil {
		return err
	}
	if len(d.Reason) > 500 {
		return ValidationError{Field: "reason", Message: "is too long"}
	}
	return requireTime("decided_at", d.DecidedAt)
}

func (d ReleaseDecision) Normalized() ReleaseDecision {
	d.DecidedAt = UTC(d.DecidedAt)
	return d
}

func validImageDigest(v string) bool {
	const prefix = "sha256:"
	if len(v) != len(prefix)+64 || v[:len(prefix)] != prefix {
		return false
	}
	for _, r := range v[len(prefix):] {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}
