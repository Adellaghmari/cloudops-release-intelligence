import { ReleaseSummary } from './models';
import { SyntheticDemoClock } from './synthetic-demo-clock';

export const RELEASE_LIST_LENSES = [
  'overview',
  'risk',
  'health',
  'impact',
  'policy',
  'rollback',
  'timeline',
] as const;

export type ReleaseListLens = (typeof RELEASE_LIST_LENSES)[number];

export function parseReleaseListLens(value: string | null | undefined): ReleaseListLens {
  if (!value) {
    return 'overview';
  }
  return RELEASE_LIST_LENSES.includes(value as ReleaseListLens)
    ? (value as ReleaseListLens)
    : 'overview';
}

export function sortReleasesByLens(
  releases: ReleaseSummary[],
  lens: ReleaseListLens,
  demoClock: SyntheticDemoClock,
  reference = new Date(),
): ReleaseSummary[] {
  if (lens === 'overview') {
    return releases;
  }

  const indexed = releases.map((release, index) => ({ release, index }));
  indexed.sort((left, right) => {
    const cmp = compareByLens(left.release, right.release, lens, demoClock, reference);
    if (cmp !== 0) {
      return cmp;
    }
    return left.index - right.index;
  });
  return indexed.map((entry) => entry.release);
}

function compareByLens(
  left: ReleaseSummary,
  right: ReleaseSummary,
  lens: ReleaseListLens,
  demoClock: SyntheticDemoClock,
  reference: Date,
): number {
  switch (lens) {
    case 'risk':
      return compareDesc(left.risk_score, right.risk_score) || compareDesc(riskCategoryRank(left.risk_category), riskCategoryRank(right.risk_category));
    case 'health':
      return compareDesc(healthRank(left.health_overall), healthRank(right.health_overall));
    case 'impact':
      return (
        compareDesc(impactTotal(left), impactTotal(right)) ||
        compareDesc(left.direct_dependents, right.direct_dependents)
      );
    case 'policy':
      return compareDesc(policyRank(left.policy_result), policyRank(right.policy_result));
    case 'rollback':
      return compareDesc(rollbackRank(left.rollback_status), rollbackRank(right.rollback_status));
    case 'timeline':
      return compareDesc(
        displayCreatedAt(left, demoClock, reference),
        displayCreatedAt(right, demoClock, reference),
      );
    default:
      return 0;
  }
}

function impactTotal(release: ReleaseSummary): number {
  return release.direct_dependents + release.transitive_impact;
}

function displayCreatedAt(
  release: ReleaseSummary,
  demoClock: SyntheticDemoClock,
  reference: Date,
): number {
  const instant =
    release.source === 'synthetic'
      ? demoClock.displayInstant(release.created_at, reference)
      : new Date(release.created_at);
  return instant.getTime();
}

function riskCategoryRank(category: string): number {
  switch (category) {
    case 'CRITICAL':
      return 4;
    case 'HIGH':
      return 3;
    case 'MODERATE':
      return 2;
    case 'LOW':
      return 1;
    default:
      return 0;
  }
}

function healthRank(overall: string): number {
  switch (overall) {
    case 'SEVERELY_DEGRADED':
      return 5;
    case 'DEGRADED':
      return 4;
    case 'INSUFFICIENT_DATA':
      return 3;
    case 'STABLE':
      return 2;
    default:
      return 1;
  }
}

function policyRank(result: string): number {
  switch (result) {
    case 'BLOCK':
      return 4;
    case 'MANUAL_APPROVAL_REQUIRED':
      return 3;
    case 'WARN':
      return 2;
    case 'PASS':
      return 1;
    default:
      return 0;
  }
}

function rollbackRank(status: string): number {
  switch (status) {
    case 'NOT_READY':
      return 4;
    case 'UNKNOWN':
      return 3;
    case 'PARTIAL':
      return 2;
    case 'READY':
      return 1;
    default:
      return 0;
  }
}

function compareDesc(left: number, right: number): number {
  if (left === right) {
    return 0;
  }
  return left > right ? -1 : 1;
}

export function releaseListLensFocus(lens: ReleaseListLens, column: ReleaseListLens): boolean {
  return lens !== 'overview' && lens === column;
}
