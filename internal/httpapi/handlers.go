package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/events"
	"github.com/adell/cloudops-release-intelligence/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	catalog     *service.Catalog
	processor   *events.Processor
	logger      *slog.Logger
	serviceName string
	version     string
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Service: h.serviceName,
		Version: h.version,
	})
}

func (h *Handler) Ready(c *gin.Context) {
	c.JSON(http.StatusOK, readyResponse{
		Status:  "ready",
		Service: h.serviceName,
		Version: h.version,
		Dependencies: []readyDependency{
			{Name: "in_memory_store", Status: "ok"},
		},
	})
}

func (h *Handler) ListServices(c *gin.Context) {
	src, err := service.ParseSourceQuery(c.Query("source"))
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	list, err := h.catalog.ListServices(c.Request.Context(), src)
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	out := make([]serviceJSON, 0, len(list))
	for _, s := range list {
		out = append(out, mapService(s))
	}
	c.JSON(http.StatusOK, serviceListResponse{Services: out})
}

func (h *Handler) GetService(c *gin.Context) {
	id, err := domain.ParseServiceID(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ID", "invalid service id")
		return
	}
	detail, err := h.catalog.GetService(c.Request.Context(), id)
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	c.JSON(http.StatusOK, serviceDetailResponse{
		Service:    mapService(detail.Service),
		DependsOn:  mapDeps(detail.DependsOn),
		DependedBy: mapDeps(detail.DependedBy),
	})
}

func (h *Handler) ListReleases(c *gin.Context) {
	src, err := service.ParseSourceQuery(c.Query("source"))
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	var serviceID *domain.ServiceID
	if raw := c.Query("service_id"); raw != "" {
		id, err := domain.ParseServiceID(raw)
		if err != nil {
			writeError(c, http.StatusBadRequest, "INVALID_ID", "invalid service id")
			return
		}
		serviceID = &id
	}
	list, err := h.catalog.ListReleases(c.Request.Context(), src, serviceID)
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	out := make([]releaseJSON, 0, len(list))
	for _, r := range list {
		out = append(out, mapRelease(r))
	}
	c.JSON(http.StatusOK, releaseListResponse{Releases: out})
}

func (h *Handler) GetRelease(c *gin.Context) {
	id, err := domain.ParseReleaseID(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ID", "invalid release id")
		return
	}
	detail, err := h.catalog.GetRelease(c.Request.Context(), id)
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	resp := releaseDetailResponse{
		Release: mapRelease(detail.Release),
		Service: mapService(detail.Service),
	}
	if detail.Commit != nil {
		v := mapCommit(*detail.Commit)
		resp.Commit = &v
	}
	if detail.Deployment != nil {
		v := mapDeployment(*detail.Deployment)
		resp.Deployment = &v
	}
	if detail.CIRun != nil {
		v := mapCIRun(*detail.CIRun)
		resp.CIRun = &v
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetReleaseRisk(c *gin.Context) {
	id, err := domain.ParseReleaseID(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ID", "invalid release id")
		return
	}
	a, err := h.catalog.AssessRisk(c.Request.Context(), id, time.Now().UTC())
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	factors := make([]riskFactorJSON, 0, len(a.Factors))
	for _, f := range a.Factors {
		factors = append(factors, riskFactorJSON{
			Code: f.Code, Label: f.Label, Points: f.Points, Rationale: f.Rationale, Input: f.Input, Omitted: f.Omitted,
		})
	}
	c.JSON(http.StatusOK, riskResponse{
		ReleaseID: a.ReleaseID.String(), Score: a.Score, ScoreRaw: a.ScoreRaw,
		Category: string(a.Category), ModelVersion: a.ModelVersion, AssessedAt: a.AssessedAt,
		Disclaimer: "Release risk signal, not a probability of failure.",
		Factors:    factors,
	})
}

func (h *Handler) GetReleaseHealth(c *gin.Context) {
	id, err := domain.ParseReleaseID(c.Param("id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_ID", "invalid release id")
		return
	}
	a, err := h.catalog.CompareHealth(c.Request.Context(), id, time.Now().UTC())
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	metrics := make([]healthMetricJSON, 0, len(a.Metrics))
	for _, m := range a.Metrics {
		metrics = append(metrics, healthMetricJSON{
			Name: m.Name, Baseline: m.Baseline, Post: m.Post, AbsDelta: m.AbsDelta, PctDelta: m.PctDelta,
			Threshold: m.Threshold, Verdict: m.Verdict, Available: m.Available, Reason: m.Reason,
		})
	}
	c.JSON(http.StatusOK, healthCompareResponse{
		ReleaseID: a.ReleaseID.String(), Overall: a.Overall, Correlation: a.Correlation,
		Reasons: a.Reasons, Metrics: metrics, BaselineFrom: a.BaselineFrom, BaselineTo: a.BaselineTo,
		PostFrom: a.PostFrom, PostTo: a.PostTo, ModelVersion: a.ModelVersion, ComparedAt: a.ComparedAt,
		Disclaimer: a.Disclaimer,
	})
}

func (h *Handler) IngestEvent(c *gin.Context) {
	if h.processor == nil {
		writeError(c, http.StatusServiceUnavailable, "UNAVAILABLE", "event processor is not configured")
		return
	}
	var body eventIngestRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_EVENT", "malformed event envelope")
		return
	}
	env, err := envelopeFromRequest(body)
	if err != nil {
		writeError(c, http.StatusBadRequest, "INVALID_EVENT", err.Error())
		return
	}
	result, err := h.processor.Handle(c.Request.Context(), env, time.Now().UTC())
	if result.Dropped {
		writeError(c, http.StatusBadRequest, "INVALID_EVENT", result.Reason)
		return
	}
	if err != nil {
		writeDomainError(c, h.logger, err)
		return
	}
	status := http.StatusAccepted
	if result.Duplicate {
		status = http.StatusOK
	}
	c.JSON(status, eventIngestResponse{
		EventID:   body.EventID,
		Accepted:  true,
		Duplicate: result.Duplicate,
		Pending:   result.Pending,
		Reason:    result.Reason,
	})
}

func mapService(s domain.Service) serviceJSON {
	return serviceJSON{
		ID:          s.ID.String(),
		Name:        s.Name,
		Description: s.Description,
		Criticality: string(s.Criticality),
		Source:      string(s.Source),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func mapRelease(r domain.Release) releaseJSON {
	return releaseJSON{
		ID:          r.ID.String(),
		ServiceID:   r.ServiceID.String(),
		Version:     r.Version,
		GitSHA:      r.GitSHA.String(),
		Environment: string(r.Environment),
		Status:      string(r.Status),
		Source:      string(r.Source),
		CreatedAt:   r.CreatedAt,
	}
}

func mapDeps(deps []domain.Dependency) []dependencyJSON {
	out := make([]dependencyJSON, 0, len(deps))
	for _, d := range deps {
		out = append(out, dependencyJSON{From: d.From.String(), To: d.To.String(), Kind: string(d.Kind)})
	}
	return out
}

func mapCommit(c domain.Commit) commitJSON {
	return commitJSON{
		SHA: c.SHA.String(), Message: c.Message, Author: c.Author,
		FilesChanged: c.FilesChanged, LinesAdded: c.LinesAdded, LinesDeleted: c.LinesDeleted,
		MigrationPresent: c.MigrationPresent, ConfigChangePresent: c.ConfigChangePresent,
		CommittedAt: c.CommittedAt,
	}
}

func mapDeployment(d domain.Deployment) deploymentJSON {
	return deploymentJSON{
		ID: d.ID.String(), Status: string(d.Status), Environment: string(d.Environment),
		Target: d.Target, ImageDigest: d.ImageDigest, ArtifactURI: d.ArtifactURI,
		StartedAt: d.StartedAt, CompletedAt: d.CompletedAt,
	}
}

func envelopeFromRequest(body eventIngestRequest) (events.Envelope, error) {
	env := events.Envelope{
		EventID:       domain.EventID(body.EventID),
		EventType:     domain.EventType(body.EventType),
		OccurredAt:    body.OccurredAt,
		CorrelationID: body.CorrelationID,
		Source:        domain.EventProducer(body.Source),
		SchemaVersion: body.SchemaVersion,
		Payload:       body.Payload,
	}
	if body.ReleaseID != "" {
		id, err := domain.ParseReleaseID(body.ReleaseID)
		if err != nil {
			return events.Envelope{}, err
		}
		env.ReleaseID = &id
	}
	if body.ServiceID != "" {
		id, err := domain.ParseServiceID(body.ServiceID)
		if err != nil {
			return events.Envelope{}, err
		}
		env.ServiceID = &id
	}
	return env, nil
}

func mapCIRun(r domain.CIRun) ciRunJSON {
	return ciRunJSON{
		ID: r.ID.String(), WorkflowName: r.WorkflowName, Status: string(r.Status),
		FailedTests: r.FailedTests, StartedAt: r.StartedAt, CompletedAt: r.CompletedAt,
	}
}
