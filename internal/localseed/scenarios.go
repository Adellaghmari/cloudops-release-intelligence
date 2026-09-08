package localseed

import (
	"context"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/internal/repository"
)

func seedScenarios(ctx context.Context, store repository.Store, now time.Time) error {
	if err := seedRiskyDB(ctx, store, now); err != nil {
		return err
	}
	if err := seedRegression(ctx, store, now); err != nil {
		return err
	}
	if err := seedBlast(ctx, store, now); err != nil {
		return err
	}
	if err := seedSecurity(ctx, store, now); err != nil {
		return err
	}
	return seedRollbackGap(ctx, store, now)
}

func seedRiskyDB(ctx context.Context, store repository.Store, now time.Time) error {
	rel := domain.ReleaseID("rel_northstar_risky_db")
	svc := domain.ServiceID("checkout-api")
	t := now.Add(-3 * time.Hour)
	done := t.Add(6 * time.Minute)
	if err := putRelease(ctx, store, rel, svc, "1.4.0-RISKY_DATABASE_RELEASE", "d0db000", "RISKY_DATABASE_RELEASE", t); err != nil {
		return err
	}
	rev := false
	if err := store.CreateCommit(ctx, domain.Commit{
		SHA: "d0db000", ReleaseID: rel, ServiceID: svc,
		Message: "SYNTHETIC DEMO: checkout ledger migration", Author: "northstar-demo",
		FilesChanged: 86, LinesAdded: 1400, LinesDeleted: 220, MigrationPresent: true, MigrationReversible: &rev,
		CommittedAt: t.Add(-40 * time.Minute),
	}); err != nil {
		return err
	}
	if err := putCI(ctx, store, "ci_risky_db", rel, domain.CIRunStatusSucceeded, 0, 1, t); err != nil {
		return err
	}
	if err := putDeploy(ctx, store, "dep_risky_db", rel, svc, t, done, "local:checkout-api", "", ""); err != nil {
		return err
	}
	if err := seedHealthPair(ctx, store, rel, svc, done, now, 220, 0.008, 0.9994, 210); err != nil {
		return err
	}
	return seedTimeline(ctx, store, now, rel, svc, t, done, true, true)
}

func seedRegression(ctx context.Context, store repository.Store, now time.Time) error {
	rel := domain.ReleaseID("rel_northstar_regression")
	svc := domain.ServiceID("checkout-api")
	t := now.Add(-90 * time.Minute)
	done := t.Add(4 * time.Minute)
	if err := putRelease(ctx, store, rel, svc, "1.4.1-POST_DEPLOY_REGRESSION", "defaced", "POST_DEPLOY_REGRESSION", t); err != nil {
		return err
	}
	if err := store.CreateCommit(ctx, domain.Commit{
		SHA: "defaced", ReleaseID: rel, ServiceID: svc,
		Message: "SYNTHETIC DEMO: checkout retry budget change", Author: "northstar-demo",
		FilesChanged: 12, LinesAdded: 80, LinesDeleted: 20, ConfigChangePresent: true,
		CommittedAt: t.Add(-25 * time.Minute),
	}); err != nil {
		return err
	}
	if err := putCI(ctx, store, "ci_regression", rel, domain.CIRunStatusSucceeded, 0, 0, t); err != nil {
		return err
	}
	if err := putDeploy(ctx, store, "dep_regression", rel, svc, t, done, "local:checkout-api", "s3://northstar-demo/checkout/1.4.1.zip", "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"); err != nil {
		return err
	}
	if err := store.CreateHealthSnapshot(ctx, domain.HealthSnapshot{
		ID: "hlt_rel_northstar_regression_base", ServiceID: svc, ReleaseID: &rel,
		WindowKind: domain.HealthWindowBaseline, WindowStart: done.Add(-60 * time.Minute), WindowEnd: done,
		RequestCount: 500, ErrorRate: 0.006, Availability: 0.9995, LatencyP95MS: 190,
		LatencyP50MS: 120, LatencyP99MS: 240, CapturedAt: done, Source: domain.DataSourceSynthetic,
	}); err != nil {
		return err
	}
	if err := store.CreateHealthSnapshot(ctx, domain.HealthSnapshot{
		ID: "hlt_rel_northstar_regression_post", ServiceID: svc, ReleaseID: &rel,
		WindowKind: domain.HealthWindowPost, WindowStart: done.Add(2 * time.Minute), WindowEnd: done.Add(32 * time.Minute),
		RequestCount: 480, ErrorRate: 0.062, Availability: 0.981, LatencyP95MS: 640,
		LatencyP50MS: 300, LatencyP99MS: 900, CapturedAt: now, Source: domain.DataSourceSynthetic,
	}); err != nil {
		return err
	}
	opened := done.Add(8 * time.Minute)
	if err := store.CreateIncident(ctx, domain.Incident{
		ID: "inc_northstar_regression", ServiceID: svc, ReleaseID: &rel,
		Title: "SYNTHETIC DEMO: checkout error budget burn", Status: domain.IncidentStatusOpen,
		OpenedAt: opened, Source: domain.DataSourceSynthetic,
	}); err != nil {
		return err
	}
	if err := seedTimeline(ctx, store, now, rel, svc, t, done, true, false); err != nil {
		return err
	}
	return emit(ctx, store, now, "evt_"+rel.String()+"_inc", domain.EventTypeIncidentOpened, opened, rel, svc)
}

func seedBlast(ctx context.Context, store repository.Store, now time.Time) error {
	rel := domain.ReleaseID("rel_northstar_blast")
	svc := domain.ServiceID("inventory-service")
	t := now.Add(-5 * time.Hour)
	done := t.Add(5 * time.Minute)
	if err := putRelease(ctx, store, rel, svc, "2.0.0-DEPENDENCY_BLAST_RADIUS", "abc1234", "DEPENDENCY_BLAST_RADIUS", t); err != nil {
		return err
	}
	if err := store.CreateCommit(ctx, domain.Commit{
		SHA: "abc1234", ReleaseID: rel, ServiceID: svc,
		Message: "SYNTHETIC DEMO: reservation schema tweak", Author: "northstar-demo",
		FilesChanged: 9, LinesAdded: 40, LinesDeleted: 8, CommittedAt: t.Add(-20 * time.Minute),
	}); err != nil {
		return err
	}
	if err := putCI(ctx, store, "ci_blast", rel, domain.CIRunStatusSucceeded, 0, 0, t); err != nil {
		return err
	}
	if err := putDeploy(ctx, store, "dep_blast", rel, svc, t, done, "local:inventory-service", "s3://northstar-demo/inventory/2.0.0.zip", "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"); err != nil {
		return err
	}
	if err := seedHealthPair(ctx, store, rel, svc, done, now, 300, 0.005, 0.9998, 140); err != nil {
		return err
	}
	return seedTimeline(ctx, store, now, rel, svc, t, done, true, false)
}

func seedSecurity(ctx context.Context, store repository.Store, now time.Time) error {
	rel := domain.ReleaseID("rel_northstar_security")
	svc := domain.ServiceID("customer-api")
	t := now.Add(-6 * time.Hour)
	done := t.Add(4 * time.Minute)
	if err := putRelease(ctx, store, rel, svc, "0.9.0-SECURITY_BLOCK", "cafebad", "SECURITY_BLOCK", t); err != nil {
		return err
	}
	if err := store.CreateCommit(ctx, domain.Commit{
		SHA: "cafebad", ReleaseID: rel, ServiceID: svc,
		Message: "SYNTHETIC DEMO: session cookie change", Author: "northstar-demo",
		FilesChanged: 7, LinesAdded: 30, LinesDeleted: 4, CommittedAt: t.Add(-18 * time.Minute),
	}); err != nil {
		return err
	}
	if err := putCI(ctx, store, "ci_security", rel, domain.CIRunStatusSucceeded, 0, 0, t); err != nil {
		return err
	}
	if err := store.CreateSecurityScan(ctx, domain.SecurityScan{
		ID: "scan_northstar_security", ReleaseID: rel, HighestSeverity: domain.SeverityCritical, FindingCount: 1, ScannedAt: t.Add(-10 * time.Minute),
	}); err != nil {
		return err
	}
	if err := putDeploy(ctx, store, "dep_security", rel, svc, t, done, "local:customer-api", "s3://northstar-demo/customer/0.9.0.zip", ""); err != nil {
		return err
	}
	if err := seedHealthPair(ctx, store, rel, svc, done, now, 180, 0.004, 0.9999, 110); err != nil {
		return err
	}
	if err := seedTimeline(ctx, store, now, rel, svc, t, done, true, false); err != nil {
		return err
	}
	return emit(ctx, store, now, "evt_"+rel.String()+"_sec", domain.EventTypeSecurityScanCompleted, t.Add(-10*time.Minute), rel, svc)
}

func seedRollbackGap(ctx context.Context, store repository.Store, now time.Time) error {
	rel := domain.ReleaseID("rel_northstar_rollback")
	svc := domain.ServiceID("fulfillment-api")
	t := now.Add(-2 * time.Hour)
	done := t.Add(3 * time.Minute)
	if err := putRelease(ctx, store, rel, svc, "0.1.0-ROLLBACK_NOT_READY", "f00ba12", "ROLLBACK_NOT_READY", t); err != nil {
		return err
	}
	if err := store.CreateCommit(ctx, domain.Commit{
		SHA: "f00ba12", ReleaseID: rel, ServiceID: svc,
		Message: "SYNTHETIC DEMO: first fulfillment cutover", Author: "northstar-demo",
		FilesChanged: 20, LinesAdded: 200, LinesDeleted: 15, CommittedAt: t.Add(-15 * time.Minute),
	}); err != nil {
		return err
	}
	if err := putCI(ctx, store, "ci_rollback", rel, domain.CIRunStatusSucceeded, 0, 0, t); err != nil {
		return err
	}
	if err := putDeploy(ctx, store, "dep_rollback", rel, svc, t, done, "", "", ""); err != nil {
		return err
	}
	if err := seedHealthPair(ctx, store, rel, svc, done, now, 90, 0.009, 0.9991, 250); err != nil {
		return err
	}
	return seedTimeline(ctx, store, now, rel, svc, t, done, false, false)
}

func putRelease(ctx context.Context, store repository.Store, id domain.ReleaseID, svc domain.ServiceID, version, sha, scenario string, at time.Time) error {
	return store.CreateRelease(ctx, domain.Release{
		ID: id, ServiceID: svc, Version: version, GitSHA: domain.CommitSHA(sha),
		Environment: domain.EnvironmentLocal, Status: domain.ReleaseStatusDeployed,
		Source: domain.DataSourceSynthetic, Scenario: scenario, CreatedAt: at,
	})
}

func putCI(ctx context.Context, store repository.Store, id domain.CIRunID, rel domain.ReleaseID, status domain.CIRunStatus, failed, attempts int, t time.Time) error {
	done := t
	return store.CreateCIRun(ctx, domain.CIRun{
		ID: id, ReleaseID: rel, WorkflowName: "northstar-ci", Status: status,
		FailedTests: failed, FailedAttempts: attempts, ExternalRunID: string(id),
		StartedAt: t.Add(-12 * time.Minute), CompletedAt: &done,
	})
}

func putDeploy(ctx context.Context, store repository.Store, id domain.DeploymentID, rel domain.ReleaseID, svc domain.ServiceID, start, done time.Time, target, artifact, digest string) error {
	if target == "" {
		target = "unknown"
	}
	return store.CreateDeployment(ctx, domain.Deployment{
		ID: id, ReleaseID: rel, ServiceID: svc, Environment: domain.EnvironmentLocal,
		Status: domain.DeploymentStatusSucceeded, Target: target, ArtifactURI: artifact, ImageDigest: digest,
		StartedAt: start, CompletedAt: &done,
	})
}

func seedTimeline(ctx context.Context, store repository.Store, now time.Time, rel domain.ReleaseID, svc domain.ServiceID, start, done time.Time, artifact, migration bool) error {
	if err := emit(ctx, store, now, "evt_"+rel.String()+"_commit", domain.EventTypeChangeCommitRecorded, start.Add(-20*time.Minute), rel, svc); err != nil {
		return err
	}
	if err := emit(ctx, store, now, "evt_"+rel.String()+"_ci", domain.EventTypeCIRunCompleted, start.Add(-5*time.Minute), rel, svc); err != nil {
		return err
	}
	if artifact {
		if err := emit(ctx, store, now, "evt_"+rel.String()+"_art", domain.EventTypeArtifactPublished, start.Add(-2*time.Minute), rel, svc); err != nil {
			return err
		}
	}
	if migration {
		_ = migration
	}
	if err := emit(ctx, store, now, "evt_"+rel.String()+"_dep", domain.EventTypeDeploymentStarted, start, rel, svc); err != nil {
		return err
	}
	if err := emit(ctx, store, now, "evt_"+rel.String()+"_ok", domain.EventTypeDeploymentSucceeded, done, rel, svc); err != nil {
		return err
	}
	return emit(ctx, store, now, "evt_"+rel.String()+"_hlt", domain.EventTypeHealthSnapshotRecorded, done.Add(10*time.Minute), rel, svc)
}

func emit(ctx context.Context, store repository.Store, now time.Time, id string, typ domain.EventType, at time.Time, rel domain.ReleaseID, svc domain.ServiceID) error {
	return store.CreateEvent(ctx, domain.ReleaseEvent{
		ID: domain.EventID(id), Type: typ, SchemaVersion: "1.0", OccurredAt: at, IngestedAt: now,
		Producer: domain.EventProducerSynthetic, CorrelationID: "corr_" + rel.String(),
		ReleaseID: &rel, ServiceID: &svc,
	})
}
