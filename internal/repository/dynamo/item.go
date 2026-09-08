package dynamo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type record struct {
	PK         string `dynamodbav:"PK"`
	SK         string `dynamodbav:"SK"`
	GSI1PK     string `dynamodbav:"GSI1PK,omitempty"`
	GSI1SK     string `dynamodbav:"GSI1SK,omitempty"`
	GSI2PK     string `dynamodbav:"GSI2PK,omitempty"`
	GSI2SK     string `dynamodbav:"GSI2SK,omitempty"`
	EntityType string `dynamodbav:"EntityType"`
	Payload    string `dynamodbav:"Payload"`
}

func marshalRecord(r record) (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(r)
}

func unmarshalRecord(av map[string]types.AttributeValue) (record, error) {
	var r record
	if err := attributevalue.UnmarshalMap(av, &r); err != nil {
		return record{}, fmt.Errorf("decode item: %w", err)
	}
	return r, nil
}

func encodePayload(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode payload: %w", err)
	}
	return string(b), nil
}

func decodePayload(raw string, dest any) error {
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	return nil
}

type servicePayload struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Criticality string    `json:"criticality"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type dependencyPayload struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"created_at"`
}

type releasePayload struct {
	ID          string    `json:"id"`
	ServiceID   string    `json:"service_id"`
	Version     string    `json:"version"`
	GitSHA      string    `json:"git_sha"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"created_at"`
}

type deploymentPayload struct {
	ID          string     `json:"id"`
	ReleaseID   string     `json:"release_id"`
	ServiceID   string     `json:"service_id"`
	Environment string     `json:"environment"`
	Status      string     `json:"status"`
	Target      string     `json:"target"`
	ImageDigest string     `json:"image_digest"`
	ArtifactURI string     `json:"artifact_uri"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type commitPayload struct {
	SHA                 string    `json:"sha"`
	ReleaseID           string    `json:"release_id"`
	ServiceID           string    `json:"service_id"`
	Message             string    `json:"message"`
	Author              string    `json:"author"`
	FilesChanged        int       `json:"files_changed"`
	LinesAdded          int       `json:"lines_added"`
	LinesDeleted        int       `json:"lines_deleted"`
	MigrationPresent    bool      `json:"migration_present"`
	ConfigChangePresent bool      `json:"config_change_present"`
	CommittedAt         time.Time `json:"committed_at"`
}

type ciPayload struct {
	ID             string     `json:"id"`
	ReleaseID      string     `json:"release_id"`
	WorkflowName   string     `json:"workflow_name"`
	Status         string     `json:"status"`
	FailedTests    int        `json:"failed_tests"`
	FailedAttempts int        `json:"failed_attempts"`
	ExternalRunID  string     `json:"external_run_id"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type healthPayload struct {
	ID           string    `json:"id"`
	ServiceID    string    `json:"service_id"`
	ReleaseID    *string   `json:"release_id,omitempty"`
	WindowKind   string    `json:"window_kind"`
	WindowStart  time.Time `json:"window_start"`
	WindowEnd    time.Time `json:"window_end"`
	RequestCount int       `json:"request_count"`
	ErrorRate    float64   `json:"error_rate"`
	Availability float64   `json:"availability"`
	LatencyP50MS float64   `json:"latency_p50_ms"`
	LatencyP95MS float64   `json:"latency_p95_ms"`
	LatencyP99MS float64   `json:"latency_p99_ms"`
	CPUPct       *float64  `json:"cpu_pct,omitempty"`
	MemoryPct    *float64  `json:"memory_pct,omitempty"`
	QueueBacklog *int      `json:"queue_backlog,omitempty"`
	CapturedAt   time.Time `json:"captured_at"`
	Source       string    `json:"source"`
}

type incidentPayload struct {
	ID         string     `json:"id"`
	ServiceID  string     `json:"service_id"`
	ReleaseID  *string    `json:"release_id,omitempty"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	OpenedAt   time.Time  `json:"opened_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	Source     string     `json:"source"`
}

type eventPayload struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	SchemaVersion string    `json:"schema_version"`
	OccurredAt    time.Time `json:"occurred_at"`
	IngestedAt    time.Time `json:"ingested_at"`
	Producer      string    `json:"producer"`
	CorrelationID string    `json:"correlation_id"`
	ReleaseID     *string   `json:"release_id,omitempty"`
	ServiceID     *string   `json:"service_id,omitempty"`
	DeploymentID  *string   `json:"deployment_id,omitempty"`
}

type decisionPayload struct {
	ID        string    `json:"id"`
	ReleaseID string    `json:"release_id"`
	Decision  string    `json:"decision"`
	Actor     string    `json:"actor"`
	Reason    string    `json:"reason"`
	DecidedAt time.Time `json:"decided_at"`
}

func serviceFrom(p servicePayload) domain.Service {
	return domain.Service{
		ID: domain.ServiceID(p.ID), Name: p.Name, Description: p.Description,
		Criticality: domain.Criticality(p.Criticality), Source: domain.DataSource(p.Source),
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func releaseFrom(p releasePayload) domain.Release {
	return domain.Release{
		ID: domain.ReleaseID(p.ID), ServiceID: domain.ServiceID(p.ServiceID),
		Version: p.Version, GitSHA: domain.CommitSHA(p.GitSHA),
		Environment: domain.Environment(p.Environment), Status: domain.ReleaseStatus(p.Status),
		Source: domain.DataSource(p.Source), CreatedAt: p.CreatedAt,
	}
}

func eventFrom(p eventPayload) domain.ReleaseEvent {
	var rel *domain.ReleaseID
	var svc *domain.ServiceID
	var dep *domain.DeploymentID
	if p.ReleaseID != nil {
		id := domain.ReleaseID(*p.ReleaseID)
		rel = &id
	}
	if p.ServiceID != nil {
		id := domain.ServiceID(*p.ServiceID)
		svc = &id
	}
	if p.DeploymentID != nil {
		id := domain.DeploymentID(*p.DeploymentID)
		dep = &id
	}
	return domain.ReleaseEvent{
		ID: domain.EventID(p.ID), Type: domain.EventType(p.Type), SchemaVersion: p.SchemaVersion,
		OccurredAt: p.OccurredAt, IngestedAt: p.IngestedAt, Producer: domain.EventProducer(p.Producer),
		CorrelationID: p.CorrelationID, ReleaseID: rel, ServiceID: svc, DeploymentID: dep,
	}
}
