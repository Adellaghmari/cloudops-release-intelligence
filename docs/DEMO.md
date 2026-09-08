# Public demo

Unauthenticated. Synthetic data only for Northstar Commerce. Live strips labeled **LIVE** for this repository's own pipeline metadata.

## Dynamic time

Seed generates timestamps relative to seed time, stores UTC, renders local in the Angular client. Fixtures never contain the words "today" or "10 minutes ago" as stored values.

## Scenarios

Each scenario is a set of **inputs**. Verdicts come from engines.

### `SAFE_RELEASE`

Routine `notification-worker` change, small diff, CI green, no migration, rollback READY, post-deploy metrics within noise.

Expect: LOW risk, STABLE health, policy PASS, rollback READY, no correlation.

### `RISKY_DATABASE_RELEASE`

Large `payments-service` change, `migration_present`, fan-out high, rollback PARTIAL (reversibility unknown), CI green.

Expect: HIGH (or CRITICAL if other factors stack), MANUAL_APPROVAL_REQUIRED, rollback PARTIAL.

### `POST_DEPLOY_REGRESSION`

`checkout-api` deploy at T, baseline healthy, post window error_rate and p95 jump, incident opened at T+4m.

Expect: health DEGRADED or SEVERELY_DEGRADED, LIKELY RELEASE CORRELATION, policy BLOCK on health rules, rollback READY.

### `DEPENDENCY_BLAST_RADIUS`

Change to `payments-service` with the full graph present.

Expect: direct dependent `checkout-api`, transitive `web-storefront`, elevated `dependency_fanout` risk points. Graph payload must match BFS, not a UI constant.

### `SECURITY_BLOCK`

Scan on `inventory-service` artifact with CRITICAL finding.

Expect: policy BLOCK (`security.critical_vuln`), elevated security risk factor.

### `ROLLBACK_NOT_READY`

Release exists, previous artifact missing.

Expect: rollback NOT_READY, `rollback_gap` contributing to risk.

## Demo reset

`POST /api/v1/demo/reset` rebuilds services, edges, the six scenarios, and relative timestamps. Rate-limited. Always through real writers + real engines.

Frontend scenario switcher calls the API; it does not ship JSON verdicts.

## Recruiter path (ten seconds)

1. See name + tagline
2. Open `POST_DEPLOY_REGRESSION`
3. Read risk factors, health delta, graph, policy BLOCK, timeline
4. Replay against `SAFE_RELEASE`

## Cypress coverage (later)

Those four clicks plus reset, empty error state, and Architecture page listing only technologies actually used.
