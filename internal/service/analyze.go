package service

import (
	"context"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
)

// AnalyzeRelease runs the decision engines for a persisted release.
// Puts are idempotent. GET handlers still compute on demand if nothing is stored.
func (c *Catalog) AnalyzeRelease(ctx context.Context, id domain.ReleaseID) error {
	now := time.Now().UTC()
	if _, err := c.store.GetRelease(ctx, id); err != nil {
		return err
	}
	if _, err := c.AssessRisk(ctx, id, now); err != nil {
		return err
	}
	if _, err := c.CompareHealth(ctx, id, now); err != nil {
		return err
	}
	if _, err := c.Impact(ctx, id); err != nil {
		return err
	}
	if _, err := c.EvaluatePolicy(ctx, id, now); err != nil {
		return err
	}
	if _, err := c.AssessRollback(ctx, id, now); err != nil {
		return err
	}
	return nil
}
