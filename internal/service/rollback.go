package service

import (
	"context"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/rollback"
)

func (c *Catalog) AssessRollback(ctx context.Context, id domain.ReleaseID, now time.Time) (domain.RollbackAssessment, error) {
	detail, err := c.GetRelease(ctx, id)
	if err != nil {
		return domain.RollbackAssessment{}, err
	}
	prev, prevDep, err := c.previousSuccessful(ctx, detail)
	if err != nil {
		return domain.RollbackAssessment{}, err
	}
	in := rollback.Input{
		PreviousSuccessfulRelease: prev != nil,
		ImageDigestApplicable:     imageDigestApplicable(detail.Service.ID),
	}
	if prev != nil {
		in.PreviousVersionRecorded = prev.GitSHA != "" || prev.Version != ""
	}
	if prevDep != nil {
		in.PreviousArtifact = prevDep.ArtifactURI != "" || prevDep.ImageDigest != ""
		in.PreviousImageDigest = prevDep.ImageDigest != ""
		in.DeploymentTargetKnown = prevDep.Target != ""
	}
	if detail.Commit != nil {
		in.MigrationPresent = detail.Commit.MigrationPresent
		in.MigrationReversible = detail.Commit.MigrationReversible
	}
	out := rollback.Assess(in)
	a := domain.RollbackAssessment{
		ReleaseID: id, Status: out.Status, Missing: out.Missing, ModelVersion: out.ModelVersion,
		AssessedAt: now.UTC(),
		Disclaimer: "Decision support only. This product does not execute production rollback.",
	}
	for _, s := range out.Signals {
		a.Signals = append(a.Signals, domain.RollbackSignal{ID: s.ID, OK: s.OK, NA: s.NA, Detail: s.Detail, Missing: s.Missing})
	}
	if err := c.store.PutRollbackAssessment(ctx, a); err != nil {
		return domain.RollbackAssessment{}, err
	}
	return a, nil
}

func imageDigestApplicable(id domain.ServiceID) bool {
	return id != "web-storefront" && id != "cloudops-web"
}

func (c *Catalog) previousSuccessful(ctx context.Context, detail ReleaseDetail) (*domain.Release, *domain.Deployment, error) {
	list, err := c.store.ListReleasesByService(ctx, detail.Service.ID)
	if err != nil {
		return nil, nil, err
	}
	var best *domain.Release
	for i := range list {
		rel := list[i]
		if rel.ID == detail.Release.ID || rel.Environment != detail.Release.Environment {
			continue
		}
		if rel.Status != domain.ReleaseStatusDeployed {
			continue
		}
		if !rel.CreatedAt.Before(detail.Release.CreatedAt) {
			continue
		}
		if best == nil || rel.CreatedAt.After(best.CreatedAt) {
			best = &list[i]
		}
	}
	if best == nil {
		return nil, nil, nil
	}
	dep, err := optional(c.store.GetDeploymentByRelease(ctx, best.ID))
	if err != nil {
		return nil, nil, err
	}
	if dep == nil || dep.Status != domain.DeploymentStatusSucceeded {
		return nil, nil, nil
	}
	return best, dep, nil
}
