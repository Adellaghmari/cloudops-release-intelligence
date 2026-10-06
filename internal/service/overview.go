package service

import (
	"context"
	"sort"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/graph"
)

// ReleaseSummary is a read model for recruiter-facing release triage. It only
// summarizes persisted evidence and deterministic engine outputs; it is not a
// recorded human decision.
type ReleaseSummary struct {
	Release         domain.Release
	Service         domain.Service
	Risk            domain.RiskAssessment
	Health          domain.HealthAssessment
	Impact          graph.Result
	Policy          domain.PolicyEvaluation
	Rollback        domain.RollbackAssessment
	IncidentCount   int
	AttentionLevel  string
	AttentionReason []string
}

type Overview struct {
	Releases           []ReleaseSummary
	AttentionReleaseID domain.ReleaseID
	LiveServices       int
	SyntheticServices  int
}

type ServiceSummary struct {
	Service         domain.Service
	DependsOnCount  int
	DependedByCount int
	ReleaseCount    int
	LatestReleaseID *domain.ReleaseID
}

func (c *Catalog) Overview(ctx context.Context, now time.Time) (Overview, error) {
	services, err := c.store.ListServices(ctx)
	if err != nil {
		return Overview{}, err
	}
	out := Overview{}
	for _, svc := range services {
		if svc.Source == domain.DataSourceLive {
			out.LiveServices++
		} else if svc.Source == domain.DataSourceSynthetic {
			out.SyntheticServices++
		}
	}

	releases, err := c.store.ListReleases(ctx)
	if err != nil {
		return Overview{}, err
	}
	out.Releases = make([]ReleaseSummary, 0, len(releases))
	for _, rel := range releases {
		summary, err := c.releaseSummary(ctx, rel, now)
		if err != nil {
			return Overview{}, err
		}
		out.Releases = append(out.Releases, summary)
	}
	sort.SliceStable(out.Releases, func(i, j int) bool {
		leftPriority, rightPriority := attentionPriority(out.Releases[i]), attentionPriority(out.Releases[j])
		if leftPriority != rightPriority {
			return leftPriority > rightPriority
		}
		left, right := attentionRank(out.Releases[i]), attentionRank(out.Releases[j])
		if left != right {
			return left > right
		}
		return out.Releases[i].Release.CreatedAt.After(out.Releases[j].Release.CreatedAt)
	})
	if len(out.Releases) > 0 && out.Releases[0].AttentionLevel != "CLEAR" {
		out.AttentionReleaseID = out.Releases[0].Release.ID
	}
	return out, nil
}

func (c *Catalog) ListServiceSummaries(ctx context.Context, source *domain.DataSource) ([]ServiceSummary, error) {
	services, err := c.ListServices(ctx, source)
	if err != nil {
		return nil, err
	}
	out := make([]ServiceSummary, 0, len(services))
	for _, svc := range services {
		dependsOn, err := c.store.ListDependenciesFrom(ctx, svc.ID)
		if err != nil {
			return nil, err
		}
		dependedBy, err := c.store.ListDependenciesTo(ctx, svc.ID)
		if err != nil {
			return nil, err
		}
		releases, err := c.store.ListReleasesByService(ctx, svc.ID)
		if err != nil {
			return nil, err
		}
		var latest *domain.ReleaseID
		if len(releases) > 0 {
			sort.SliceStable(releases, func(i, j int) bool {
				return releases[i].CreatedAt.After(releases[j].CreatedAt)
			})
			id := releases[0].ID
			latest = &id
		}
		out = append(out, ServiceSummary{
			Service: svc, DependsOnCount: len(dependsOn), DependedByCount: len(dependedBy),
			ReleaseCount: len(releases), LatestReleaseID: latest,
		})
	}
	return out, nil
}

func (c *Catalog) ListReleasesForService(ctx context.Context, id domain.ServiceID) ([]domain.Release, error) {
	releases, err := c.store.ListReleasesByService(ctx, id)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(releases, func(i, j int) bool {
		return releases[i].CreatedAt.After(releases[j].CreatedAt)
	})
	return releases, nil
}

func (c *Catalog) releaseSummary(ctx context.Context, rel domain.Release, now time.Time) (ReleaseSummary, error) {
	svc, err := c.store.GetService(ctx, rel.ServiceID)
	if err != nil {
		return ReleaseSummary{}, err
	}

	policyEval, riskAssessment, healthAssessment, rollbackAssessment, err := c.storedAssessments(ctx, rel.ID)
	if domain.IsNotFound(err) {
		policyEval, err = c.EvaluatePolicy(ctx, rel.ID, now)
		if err != nil {
			return ReleaseSummary{}, err
		}
		riskAssessment, err = c.store.GetRiskAssessment(ctx, rel.ID)
		if err != nil {
			return ReleaseSummary{}, err
		}
		healthAssessment, err = c.store.GetHealthComparison(ctx, rel.ID)
		if err != nil {
			return ReleaseSummary{}, err
		}
		rollbackAssessment, err = c.store.GetRollbackAssessment(ctx, rel.ID)
		if err != nil {
			return ReleaseSummary{}, err
		}
	} else if err != nil {
		return ReleaseSummary{}, err
	}

	impact, err := c.Impact(ctx, rel.ID)
	if err != nil {
		return ReleaseSummary{}, err
	}
	incidents, err := c.store.ListIncidentsByService(ctx, rel.ServiceID)
	if err != nil {
		return ReleaseSummary{}, err
	}
	incidentCount := 0
	for _, incident := range incidents {
		if incident.ReleaseID != nil && *incident.ReleaseID == rel.ID {
			incidentCount++
		}
	}
	summary := ReleaseSummary{
		Release: rel, Service: svc, Risk: riskAssessment, Health: healthAssessment,
		Impact: impact, Policy: policyEval, Rollback: rollbackAssessment, IncidentCount: incidentCount,
	}
	summary.AttentionLevel, summary.AttentionReason = attention(summary)
	return summary, nil
}

func (c *Catalog) storedAssessments(
	ctx context.Context,
	id domain.ReleaseID,
) (domain.PolicyEvaluation, domain.RiskAssessment, domain.HealthAssessment, domain.RollbackAssessment, error) {
	policyEval, err := c.store.GetPolicyEvaluation(ctx, id)
	if err != nil {
		return domain.PolicyEvaluation{}, domain.RiskAssessment{}, domain.HealthAssessment{}, domain.RollbackAssessment{}, err
	}
	riskAssessment, err := c.store.GetRiskAssessment(ctx, id)
	if err != nil {
		return domain.PolicyEvaluation{}, domain.RiskAssessment{}, domain.HealthAssessment{}, domain.RollbackAssessment{}, err
	}
	healthAssessment, err := c.store.GetHealthComparison(ctx, id)
	if err != nil {
		return domain.PolicyEvaluation{}, domain.RiskAssessment{}, domain.HealthAssessment{}, domain.RollbackAssessment{}, err
	}
	rollbackAssessment, err := c.store.GetRollbackAssessment(ctx, id)
	if err != nil {
		return domain.PolicyEvaluation{}, domain.RiskAssessment{}, domain.HealthAssessment{}, domain.RollbackAssessment{}, err
	}
	return policyEval, riskAssessment, healthAssessment, rollbackAssessment, nil
}

func attention(summary ReleaseSummary) (string, []string) {
	reasons := make([]string, 0, 5)
	if summary.Policy.Result == "BLOCK" {
		reasons = append(reasons, "POLICY_BLOCK")
	}
	if summary.Health.Overall == "SEVERELY_DEGRADED" {
		reasons = append(reasons, "HEALTH_SEVERELY_DEGRADED")
	} else if summary.Health.Overall == "DEGRADED" {
		reasons = append(reasons, "HEALTH_DEGRADED")
	}
	if summary.Policy.Result == "MANUAL_APPROVAL_REQUIRED" {
		reasons = append(reasons, "POLICY_MANUAL_APPROVAL")
	}
	if summary.Rollback.Status == domain.RollbackNotReady {
		reasons = append(reasons, "ROLLBACK_NOT_READY")
	}
	if summary.Risk.Category == domain.RiskCritical || summary.Risk.Category == domain.RiskHigh {
		reasons = append(reasons, "RISK_ELEVATED")
	}
	if summary.IncidentCount > 0 {
		reasons = append(reasons, "INCIDENT_IN_RELEASE_WINDOW")
	}
	switch {
	case summary.Policy.Result == "BLOCK" || summary.Health.Overall == "SEVERELY_DEGRADED":
		return "CRITICAL_ATTENTION", reasons
	case len(reasons) > 0 || summary.Policy.Result == "MANUAL_APPROVAL_REQUIRED":
		return "REVIEW", reasons
	default:
		return "CLEAR", reasons
	}
}

func attentionPriority(summary ReleaseSummary) int {
	switch summary.AttentionLevel {
	case "CRITICAL_ATTENTION":
		return 2
	case "REVIEW":
		return 1
	default:
		return 0
	}
}

func attentionRank(summary ReleaseSummary) int {
	rank := summary.Risk.Score
	if summary.Policy.Result == "BLOCK" {
		rank += 100
	} else if summary.Policy.Result == "MANUAL_APPROVAL_REQUIRED" {
		rank += 20
	}
	if summary.Health.Overall == "SEVERELY_DEGRADED" {
		rank += 80
	} else if summary.Health.Overall == "DEGRADED" {
		rank += 40
	}
	if summary.Rollback.Status == domain.RollbackNotReady {
		rank += 30
	}
	rank += summary.IncidentCount * 25
	return rank
}
