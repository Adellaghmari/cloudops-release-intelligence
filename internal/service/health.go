package service

import (
	"context"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/health"
)

func (c *Catalog) CompareHealth(ctx context.Context, id domain.ReleaseID, now time.Time) (domain.HealthAssessment, error) {
	detail, err := c.GetRelease(ctx, id)
	if err != nil {
		return domain.HealthAssessment{}, err
	}
	in := health.Input{ChangedService: detail.Service.ID, DegradedService: detail.Service.ID}
	if detail.Deployment == nil || detail.Deployment.CompletedAt == nil {
		cmp := health.Compare(in)
		return persistHealth(c, ctx, id, cmp, now)
	}
	t := detail.Deployment.CompletedAt.UTC()
	in.DeployedAt = t

	snaps, err := c.store.ListHealthSnapshotsByService(ctx, detail.Service.ID)
	if err != nil {
		return domain.HealthAssessment{}, err
	}
	in.Baseline = snapshotWindow(snaps, domain.HealthWindowBaseline, id, t.Add(-60*time.Minute), t, t)
	in.Post = snapshotWindow(snaps, domain.HealthWindowPost, id, t.Add(2*time.Minute), t.Add(32*time.Minute), t)
	in.BaselineAlreadyBad = baselineBreached(in.Baseline)
	in.DegradationBeforeDeploy = degradationBefore(snaps, t)
	in.IncidentInWindow = incidentInPost(ctx, c, detail.Service.ID, t)
	overlap, err := c.overlappingNeighborDeploy(ctx, detail, t)
	if err != nil {
		return domain.HealthAssessment{}, err
	}
	in.OverlappingDeployment = overlap

	cmp := health.Compare(in)
	return persistHealth(c, ctx, id, cmp, now)
}

func persistHealth(c *Catalog, ctx context.Context, id domain.ReleaseID, cmp health.Comparison, now time.Time) (domain.HealthAssessment, error) {
	a := domain.HealthAssessment{
		ReleaseID: id, Overall: string(cmp.Overall), Correlation: string(cmp.Correlation),
		Reasons: cmp.Reasons, BaselineFrom: cmp.BaselineFrom, BaselineTo: cmp.BaselineTo,
		PostFrom: cmp.PostFrom, PostTo: cmp.PostTo, ModelVersion: cmp.ModelVersion,
		ComparedAt: now.UTC(),
		Disclaimer: "Health comparison is an operational signal, not proven causation.",
	}
	for _, m := range cmp.Metrics {
		a.Metrics = append(a.Metrics, domain.HealthMetricResult{
			Name: m.Name, Baseline: m.Baseline, Post: m.Post, AbsDelta: m.AbsDelta, PctDelta: m.PctDelta,
			Threshold: m.Threshold, Verdict: string(m.Verdict), Available: m.Available, Reason: m.Reason,
		})
	}
	if err := c.store.PutHealthComparison(ctx, a); err != nil {
		return domain.HealthAssessment{}, err
	}
	return a, nil
}

func snapshotWindow(snaps []domain.HealthSnapshot, kind domain.HealthWindowKind, rel domain.ReleaseID, start, end, deploy time.Time) *health.Window {
	for _, h := range snaps {
		if h.WindowKind != kind {
			continue
		}
		if warmupOnly(h, deploy) {
			continue
		}
		if h.ReleaseID != nil && *h.ReleaseID != rel {
			continue
		}
		w := &health.Window{
			Start: h.WindowStart, End: h.WindowEnd, RequestCount: h.RequestCount,
			ErrorRate: h.ErrorRate, Availability: h.Availability,
			P50: h.LatencyP50MS, P95: h.LatencyP95MS, P99: h.LatencyP99MS,
			HasError: true, HasAvail: true, HasP95: h.LatencyP95MS > 0,
		}
		return w
	}
	_ = start
	_ = end
	return nil
}

func warmupOnly(h domain.HealthSnapshot, deploy time.Time) bool {
	warmEnd := deploy.Add(2 * time.Minute)
	return !h.WindowStart.Before(deploy) && !h.WindowEnd.After(warmEnd)
}

func baselineBreached(w *health.Window) bool {
	if w == nil {
		return false
	}
	return (w.HasError && w.ErrorRate > 0.01) || (w.HasAvail && w.Availability < 0.999)
}

func degradationBefore(snaps []domain.HealthSnapshot, deploy time.Time) bool {
	for _, h := range snaps {
		if h.WindowKind == domain.HealthWindowPost && h.WindowStart.Before(deploy) {
			return true
		}
	}
	return false
}

func incidentInPost(ctx context.Context, c *Catalog, svc domain.ServiceID, t time.Time) bool {
	list, err := c.store.ListIncidentsByService(ctx, svc)
	if err != nil {
		return false
	}
	end := t.Add(32 * time.Minute)
	for _, inc := range list {
		if inc.OpenedAt.After(t) && !inc.OpenedAt.After(end) {
			return true
		}
	}
	return false
}

func (c *Catalog) overlappingNeighborDeploy(ctx context.Context, detail ReleaseDetail, t time.Time) (bool, error) {
	ids := map[domain.ServiceID]struct{}{detail.Service.ID: {}}
	from, err := c.store.ListDependenciesFrom(ctx, detail.Service.ID)
	if err != nil {
		return false, err
	}
	to, err := c.store.ListDependenciesTo(ctx, detail.Service.ID)
	if err != nil {
		return false, err
	}
	for _, d := range from {
		ids[d.To] = struct{}{}
	}
	for _, d := range to {
		ids[d.From] = struct{}{}
	}
	lo, hi := t.Add(-30*time.Minute), t.Add(32*time.Minute)
	for sid := range ids {
		rels, err := c.store.ListReleasesByService(ctx, sid)
		if err != nil {
			return false, err
		}
		for _, rel := range rels {
			if rel.ID == detail.Release.ID {
				continue
			}
			dep, err := optional(c.store.GetDeploymentByRelease(ctx, rel.ID))
			if err != nil {
				return false, err
			}
			if dep == nil || dep.CompletedAt == nil {
				continue
			}
			at := dep.CompletedAt.UTC()
			if !at.Before(lo) && !at.After(hi) {
				return true, nil
			}
		}
	}
	return false, nil
}
