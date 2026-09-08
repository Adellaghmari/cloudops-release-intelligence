package service

import (
	"context"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/risk"
)

func (c *Catalog) AssessRisk(ctx context.Context, id domain.ReleaseID, now time.Time) (domain.RiskAssessment, error) {
	detail, err := c.GetRelease(ctx, id)
	if err != nil {
		return domain.RiskAssessment{}, err
	}
	in := risk.Input{ServiceCriticality: detail.Service.Criticality, Now: now}
	if detail.Commit != nil {
		files := detail.Commit.FilesChanged
		lines := detail.Commit.LinesAdded + detail.Commit.LinesDeleted
		mig := detail.Commit.MigrationPresent
		cfg := detail.Commit.ConfigChangePresent
		in.FilesChanged = &files
		in.LinesChanged = &lines
		in.MigrationPresent = &mig
		in.ConfigChangePresent = &cfg
	}
	if detail.CIRun != nil {
		in.FailedCIAttempts = &detail.CIRun.FailedAttempts
		in.FailedTests = &detail.CIRun.FailedTests
	}
	deps := 0
	seen := map[domain.ServiceID]struct{}{detail.Service.ID: {}}
	queue := []domain.ServiceID{detail.Service.ID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		inbound, err := c.store.ListDependenciesTo(ctx, cur)
		if err != nil {
			return domain.RiskAssessment{}, err
		}
		for _, d := range inbound {
			if _, ok := seen[d.From]; ok {
				continue
			}
			seen[d.From] = struct{}{}
			deps++
			queue = append(queue, d.From)
		}
	}
	in.DependentCount = &deps

	incidents, err := c.store.ListIncidentsByService(ctx, detail.Service.ID)
	if err != nil {
		return domain.RiskAssessment{}, err
	}
	count := 0
	cutoff := now.Add(-30 * 24 * time.Hour)
	for _, inc := range incidents {
		if inc.OpenedAt.After(cutoff) {
			count++
		}
	}
	in.RecentIncidents = &count

	if scan, err := optional(c.store.GetSecurityScanByRelease(ctx, id)); err != nil {
		return domain.RiskAssessment{}, err
	} else if scan != nil {
		sev := scan.HighestSeverity
		in.HighestSeverity = &sev
	}

	releases, err := c.store.ListReleasesByService(ctx, detail.Service.ID)
	if err != nil {
		return domain.RiskAssessment{}, err
	}
	failed := 0
	completed := 0
	for _, rel := range releases {
		if rel.ID == id {
			continue
		}
		if rel.Status == domain.ReleaseStatusFailed {
			failed++
			completed++
		}
		if rel.Status == domain.ReleaseStatusDeployed || rel.Status == domain.ReleaseStatusRolledBack {
			completed++
		}
	}
	in.RecentDeployFailures = &failed
	if completed >= 3 {
		rate := float64(failed) / float64(completed)
		in.HistoricalFailureRate = &rate
	}

	assessment := risk.Calculate(id, in)
	if err := c.store.PutRiskAssessment(ctx, assessment); err != nil {
		return domain.RiskAssessment{}, err
	}
	return assessment, nil
}
