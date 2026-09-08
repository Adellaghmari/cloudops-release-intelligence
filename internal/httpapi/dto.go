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

type releaseDetailResponse struct {
	Release    releaseJSON     `json:"release"`
	Service    serviceJSON     `json:"service"`
	Commit     *commitJSON     `json:"commit,omitempty"`
	Deployment *deploymentJSON `json:"deployment,omitempty"`
	CIRun      *ciRunJSON      `json:"ci_run,omitempty"`
}
