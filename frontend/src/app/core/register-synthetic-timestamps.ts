import {
  HealthCompareResponse,
  PolicyResponse,
  ReleaseDetailResponse,
  RiskResponse,
  RollbackResponse,
  TimelineResponse,
} from './models';
import { SyntheticDemoClock, collectSyntheticTimestamps } from './synthetic-demo-clock';

export function registerSyntheticDemoTimestamps(
  clock: SyntheticDemoClock,
  detail: ReleaseDetailResponse | null,
  evidence: {
    risk: RiskResponse | null;
    health: HealthCompareResponse | null;
    policy: PolicyResponse | null;
    rollback: RollbackResponse | null;
    timeline: TimelineResponse | null;
  },
) {
  if (!detail || detail.release.source !== 'synthetic') {
    return;
  }

  const stamps = collectSyntheticTimestamps([
    { created_at: detail.release.created_at },
    detail.commit
      ? {
          committed_at: detail.commit.committed_at,
        }
      : {},
    detail.deployment
      ? {
          started_at: detail.deployment.started_at,
          completed_at: detail.deployment.completed_at,
        }
      : {},
    detail.ci_run
      ? {
          started_at: detail.ci_run.started_at,
          completed_at: detail.ci_run.completed_at,
        }
      : {},
    evidence.risk ? { assessed_at: evidence.risk.assessed_at } : {},
    evidence.health
      ? {
          baseline_from: evidence.health.baseline_from,
          baseline_to: evidence.health.baseline_to,
          post_from: evidence.health.post_from,
          post_to: evidence.health.post_to,
          compared_at: evidence.health.compared_at,
        }
      : {},
    evidence.policy ? { evaluated_at: evidence.policy.evaluated_at } : {},
    evidence.rollback ? { assessed_at: evidence.rollback.assessed_at } : {},
    ...(evidence.timeline?.entries.map((entry) => ({ occurred_at: entry.occurred_at })) ?? []),
  ]);

  clock.noteTimestamps(stamps);
}
