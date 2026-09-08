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

type healthMetricJSON struct {
	Name      string   `json:"name"`
	Baseline  float64  `json:"baseline"`
	Post      float64  `json:"post"`
	AbsDelta  float64  `json:"abs_delta"`
	PctDelta  *float64 `json:"pct_delta,omitempty"`
	Threshold string   `json:"threshold"`
	Verdict   string   `json:"verdict"`
	Available bool     `json:"available"`
	Reason    string   `json:"reason"`
}

type healthCompareResponse struct {
	ReleaseID    string             `json:"release_id"`
	Overall      string             `json:"overall"`
	Correlation  string             `json:"correlation"`
	Reasons      []string           `json:"reasons"`
	Metrics      []healthMetricJSON `json:"metrics"`
	BaselineFrom time.Time          `json:"baseline_from"`
	BaselineTo   time.Time          `json:"baseline_to"`
	PostFrom     time.Time          `json:"post_from"`
	PostTo       time.Time          `json:"post_to"`
	ModelVersion string             `json:"model_version"`
	ComparedAt   time.Time          `json:"compared_at"`
	Disclaimer   string             `json:"disclaimer"`
}

type impactNodeJSON struct {
	ID    string `json:"id"`
	Role  string `json:"role"`
	Depth int    `json:"depth"`
}

type impactEdgeJSON struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type impactResponse struct {
	ChangedService       string           `json:"changed_service_id"`
	DirectDependents     []string         `json:"direct_dependents"`
	TransitiveDependents []string         `json:"transitive_dependents"`
	Upstream             []string         `json:"upstream_dependencies"`
	CriticalInRadius     []string         `json:"critical_in_radius"`
	Nodes                []impactNodeJSON `json:"nodes"`
	Edges                []impactEdgeJSON `json:"edges"`
	MaxDepth             int              `json:"max_dependent_depth"`
	Cycles               [][]string       `json:"cycles"`
	Unknown              bool             `json:"unknown"`
	Empty                bool             `json:"empty"`
	Algorithm            string           `json:"algorithm"`
	Disclaimer           string           `json:"disclaimer"`
}

type policyRuleJSON struct {
	ID           string `json:"id"`
	Result       string `json:"result"`
	Message      string `json:"message"`
	InputExcerpt string `json:"input_excerpt"`
	Skipped      bool   `json:"skipped"`
}

type policyResponse struct {
	ReleaseID     string           `json:"release_id"`
	PolicyVersion string           `json:"policy_version"`
	Result        string           `json:"result"`
	Phase         string           `json:"phase"`
	Rules         []policyRuleJSON `json:"rules"`
	EvaluatedAt   time.Time        `json:"evaluated_at"`
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
