package service

import (
	"context"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/policy"
)

func (c *Catalog) EvaluatePolicy(ctx context.Context, id domain.ReleaseID, now time.Time) (domain.PolicyEvaluation, error) {
	detail, err := c.GetRelease(ctx, id)
	if err != nil {
		return domain.PolicyEvaluation{}, err
	}
	riskA, err := c.AssessRisk(ctx, id, now)
	if err != nil {
		return domain.PolicyEvaluation{}, err
	}
	healthA, err := c.CompareHealth(ctx, id, now)
	if err != nil {
		return domain.PolicyEvaluation{}, err
	}
	in := policy.Input{
		ReleaseID: id.String(), ServiceID: detail.Service.ID.String(),
		Criticality: string(detail.Service.Criticality), RiskScore: riskA.Score,
		RiskCategory: string(riskA.Category), HealthOverall: healthA.Overall,
		HealthAvail: healthA.Overall != "" && healthA.Overall != "INSUFFICIENT_DATA",
		Phase:       "PRE_DEPLOY",
	}
	if rb, err := c.AssessRollback(ctx, id, now); err != nil {
		return domain.PolicyEvaluation{}, err
	} else {
		in.Rollback = string(rb.Status)
	}
	if in.HealthAvail {
		in.Phase = "POST_DEPLOY"
	}
	if detail.Commit != nil {
		in.FilesChanged = detail.Commit.FilesChanged
		in.Migration = detail.Commit.MigrationPresent
		in.ConfigChange = detail.Commit.ConfigChangePresent
	}
	if detail.CIRun != nil {
		in.CIStatus = string(detail.CIRun.Status)
		in.FailedTests = detail.CIRun.FailedTests
	}
	if scan, err := optional(c.store.GetSecurityScanByRelease(ctx, id)); err != nil {
		return domain.PolicyEvaluation{}, err
	} else if scan != nil {
		in.Severity = string(scan.HighestSeverity)
	} else {
		in.Severity = string(domain.SeverityNone)
	}
	eval := policy.Evaluate(ctx, in, now)
	if err := c.store.PutPolicyEvaluation(ctx, eval); err != nil {
		return domain.PolicyEvaluation{}, err
	}
	return eval, nil
}
