package httpapi

import "time"

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type readyDependency struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type readyResponse struct {
	Status       string            `json:"status"`
	Service      string            `json:"service"`
	Version      string            `json:"version"`
	Dependencies []readyDependency `json:"dependencies"`
}

type serviceJSON struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Criticality string    `json:"criticality"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type dependencyJSON struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

type releaseJSON struct {
	ID          string    `json:"id"`
	ServiceID   string    `json:"service_id"`
	Version     string    `json:"version"`
	GitSHA      string    `json:"git_sha"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
}

type commitJSON struct {
	SHA                 string    `json:"sha"`
	Message             string    `json:"message"`
	Author              string    `json:"author"`
	FilesChanged        int       `json:"files_changed"`
	LinesAdded          int       `json:"lines_added"`
	LinesDeleted        int       `json:"lines_deleted"`
	MigrationPresent    bool      `json:"migration_present"`
	ConfigChangePresent bool      `json:"config_change_present"`
	CommittedAt         time.Time `json:"committed_at"`
}

type deploymentJSON struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	Environment string     `json:"environment"`
	Target      string     `json:"target"`
	ImageDigest string     `json:"image_digest,omitempty"`
	ArtifactURI string     `json:"artifact_uri,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ciRunJSON struct {
	ID           string     `json:"id"`
	WorkflowName string     `json:"workflow_name"`
	Status       string     `json:"status"`
	FailedTests  int        `json:"failed_tests"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type serviceListResponse struct {
	Services []serviceJSON `json:"services"`
}

type serviceDetailResponse struct {
	Service    serviceJSON      `json:"service"`
	DependsOn  []dependencyJSON `json:"depends_on"`
	DependedBy []dependencyJSON `json:"depended_by"`
}

type releaseListResponse struct {
	Releases []releaseJSON `json:"releases"`
}

type riskFactorJSON struct {
	Code      string `json:"code"`
	Label     string `json:"label"`
	Points    int    `json:"points"`
	Rationale string `json:"rationale"`
	Input     string `json:"input"`
	Omitted   bool   `json:"omitted"`
}

type riskResponse struct {
	ReleaseID    string           `json:"release_id"`
	Score        int              `json:"score"`
	ScoreRaw     int              `json:"score_raw"`
	Category     string           `json:"category"`
	ModelVersion string           `json:"model_version"`
	AssessedAt   time.Time        `json:"assessed_at"`
	Disclaimer   string           `json:"disclaimer"`
	Factors      []riskFactorJSON `json:"factors"`
}

type releaseDetailResponse struct {
	Release    releaseJSON     `json:"release"`
	Service    serviceJSON     `json:"service"`
	Commit     *commitJSON     `json:"commit,omitempty"`
	Deployment *deploymentJSON `json:"deployment,omitempty"`
	CIRun      *ciRunJSON      `json:"ci_run,omitempty"`
}

type eventIngestRequest struct {
	EventID       string            `json:"event_id"`
	EventType     string            `json:"event_type"`
	OccurredAt    time.Time         `json:"occurred_at"`
	CorrelationID string            `json:"correlation_id"`
	ReleaseID     string            `json:"release_id"`
	ServiceID     string            `json:"service_id"`
	Source        string            `json:"source"`
	SchemaVersion string            `json:"schema_version"`
	Payload       map[string]string `json:"payload"`
}

type eventIngestResponse struct {
	EventID   string `json:"event_id"`
	Accepted  bool   `json:"accepted"`
	Duplicate bool   `json:"duplicate"`
	Pending   bool   `json:"pending,omitempty"`
	Reason    string `json:"reason,omitempty"`
}
