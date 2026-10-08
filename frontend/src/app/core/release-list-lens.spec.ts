import { describe, expect, it } from 'vitest';
import { ReleaseSummary } from './models';
import { parseReleaseListLens, sortReleasesByLens } from './release-list-lens';
import { SyntheticDemoClock } from './synthetic-demo-clock';

function summary(partial: Partial<ReleaseSummary> & Pick<ReleaseSummary, 'release_id'>): ReleaseSummary {
  return {
    service_id: 'svc',
    service_name: 'Service',
    version: '1.0.0',
    source: 'synthetic',
    status: 'deployed',
    risk_score: 10,
    risk_category: 'LOW',
    health_overall: 'STABLE',
    health_correlation: 'NO_CLEAR_CORRELATION',
    policy_result: 'PASS',
    rollback_status: 'READY',
    direct_dependents: 0,
    transitive_impact: 0,
    incident_count: 0,
    attention_level: 'CLEAR',
    attention_reasons: [],
    created_at: '2026-10-01T08:00:00.000Z',
    ...partial,
  };
}

describe('parseReleaseListLens', () => {
  it('defaults missing or unknown values to overview', () => {
    expect(parseReleaseListLens(null)).toBe('overview');
    expect(parseReleaseListLens('nope')).toBe('overview');
    expect(parseReleaseListLens('risk')).toBe('risk');
  });
});

describe('sortReleasesByLens', () => {
  const clock = new SyntheticDemoClock();

  it('preserves API order for overview', () => {
    const releases = [
      summary({ release_id: 'a', risk_score: 5 }),
      summary({ release_id: 'b', risk_score: 90 }),
    ];
    const sorted = sortReleasesByLens(releases, 'overview', clock);
    expect(sorted.map((r) => r.release_id)).toEqual(['a', 'b']);
  });

  it('orders by highest risk score first', () => {
    const releases = [
      summary({ release_id: 'low', risk_score: 12 }),
      summary({ release_id: 'high', risk_score: 88 }),
    ];
    const sorted = sortReleasesByLens(releases, 'risk', clock);
    expect(sorted[0].release_id).toBe('high');
  });

  it('orders blocked policy results before pass', () => {
    const releases = [
      summary({ release_id: 'pass', policy_result: 'PASS' }),
      summary({ release_id: 'block', policy_result: 'BLOCK' }),
    ];
    const sorted = sortReleasesByLens(releases, 'policy', clock);
    expect(sorted[0].release_id).toBe('block');
  });
});
