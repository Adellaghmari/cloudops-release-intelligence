package service

import (
	"context"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/replay"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/timeline"
)

func (c *Catalog) Timeline(ctx context.Context, id domain.ReleaseID) ([]timeline.Entry, error) {
	if _, err := c.store.GetRelease(ctx, id); err != nil {
		return nil, err
	}
	events, err := c.store.ListEventsByRelease(ctx, id)
	if err != nil {
		return nil, err
	}
	return timeline.FromEvents(events), nil
}

type ReplayResult struct {
	AID    domain.ReleaseID
	BID    domain.ReleaseID
	Fields []replay.Field
}

func (c *Catalog) Replay(ctx context.Context, aID, bID domain.ReleaseID, now time.Time) (ReplayResult, error) {
	a, err := c.snapshot(ctx, aID, now)
	if err != nil {
		return ReplayResult{}, err
	}
	b, err := c.snapshot(ctx, bID, now)
	if err != nil {
		return ReplayResult{}, err
	}
	fields := []replay.Field{
		replay.CompareStrings("release.service_id", a.service, b.service),
		replay.CompareStrings("release.version", a.version, b.version),
		replay.CompareStrings("release.status", a.status, b.status),
		replay.CompareInts("change.files_changed", a.files, b.files),
		replay.CompareBools("change.migration_present", a.migration, b.migration),
		replay.CompareBools("change.config_change_present", a.config, b.config),
		replay.CompareStrings("ci.status", a.ci, b.ci),
		replay.CompareInts("ci.failed_tests", a.failedTests, b.failedTests),
		replay.CompareInts("risk.score", a.risk, b.risk),
		replay.CompareStrings("risk.category", a.riskCat, b.riskCat),
		replay.CompareStrings("health.overall", a.health, b.health),
		replay.CompareStrings("health.correlation", a.corr, b.corr),
		replay.CompareInts("impact.direct_dependents", a.impact, b.impact),
		replay.CompareStrings("policy.result", a.policy, b.policy),
		replay.CompareStrings("rollback.status", a.rollback, b.rollback),
		replay.CompareInts("incidents.open_in_window", a.incidents, b.incidents),
	}
	return ReplayResult{AID: aID, BID: bID, Fields: fields}, nil
}

type releaseSnap struct {
	service, version, status, ci, riskCat, health, corr, policy, rollback string
	files, failedTests, risk, impact, incidents                           int
	migration, config                                                     bool
}

func (c *Catalog) snapshot(ctx context.Context, id domain.ReleaseID, now time.Time) (releaseSnap, error) {
	d, err := c.GetRelease(ctx, id)
	if err != nil {
		return releaseSnap{}, err
	}
	s := releaseSnap{service: d.Service.ID.String(), version: d.Release.Version, status: string(d.Release.Status)}
	if d.Commit != nil {
		s.files = d.Commit.FilesChanged
		s.migration = d.Commit.MigrationPresent
		s.config = d.Commit.ConfigChangePresent
	}
	if d.CIRun != nil {
		s.ci = string(d.CIRun.Status)
		s.failedTests = d.CIRun.FailedTests
	}
	riskA, err := c.AssessRisk(ctx, id, now)
	if err != nil {
		return releaseSnap{}, err
	}
	s.risk, s.riskCat = riskA.Score, string(riskA.Category)
	h, err := c.CompareHealth(ctx, id, now)
	if err != nil {
		return releaseSnap{}, err
	}
	s.health, s.corr = h.Overall, h.Correlation
	imp, err := c.Impact(ctx, id)
	if err != nil {
		return releaseSnap{}, err
	}
	s.impact = len(imp.DirectDependents)
	pol, err := c.EvaluatePolicy(ctx, id, now)
	if err != nil {
		return releaseSnap{}, err
	}
	s.policy = pol.Result
	rb, err := c.AssessRollback(ctx, id, now)
	if err != nil {
		return releaseSnap{}, err
	}
	s.rollback = string(rb.Status)
	incs, err := c.store.ListIncidentsByService(ctx, d.Service.ID)
	if err != nil {
		return releaseSnap{}, err
	}
	for _, inc := range incs {
		if inc.ReleaseID != nil && *inc.ReleaseID == id {
			s.incidents++
		}
	}
	return s, nil
}
