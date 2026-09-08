package events

import (
	"strconv"
	"strings"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
)

func EvidenceFromEnvelope(e Envelope) (domain.OperationalEvidence, bool) {
	if e.Source == domain.EventProducerSynthetic {
		return domain.OperationalEvidence{}, false
	}
	kind := ""
	switch e.EventType {
	case domain.EventTypeCIRunCompleted:
		kind = "pipeline"
	case domain.EventTypeSecurityScanCompleted:
		kind = "security"
	case domain.EventTypeArtifactPublished:
		kind = "artifact"
	case domain.EventTypeDeploymentSucceeded:
		kind = "deploy"
	default:
		return domain.OperationalEvidence{}, false
	}
	payload := e.Payload
	if payload == nil {
		payload = map[string]string{}
	}
	ev := domain.OperationalEvidence{
		ID:                 e.EventID,
		Kind:               kind,
		GitSHA:             firstNonEmpty(payload["git_sha"], payload["sha"]),
		Branch:             payload["branch"],
		WorkflowName:       firstNonEmpty(payload["workflow_name"], payload["workflow"]),
		WorkflowRunID:      firstNonEmpty(payload["workflow_run_id"], payload["run_id"]),
		WorkflowResult:     firstNonEmpty(payload["workflow_result"], payload["result"]),
		TestResult:         payload["test_result"],
		ImageDigest:        payload["image_digest"],
		ArtifactID:         payload["artifact_id"],
		SecurityScanResult: payload["security_scan_result"],
		RecordedAt:         e.OccurredAt,
		Source:             domain.DataSourceLive,
	}
	if raw := payload["build_duration_ms"]; raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			ev.BuildDurationMS = &n
		}
	}
	if raw := firstNonEmpty(payload["deployment_timestamp"], payload["deployed_at"]); raw != "" {
		if ts, err := time.Parse(time.RFC3339, raw); err == nil {
			ev.DeploymentTimestamp = &ts
		}
	}
	return ev.Normalized(), true
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
