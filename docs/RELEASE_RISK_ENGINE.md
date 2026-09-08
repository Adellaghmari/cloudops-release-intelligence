# Release Risk Engine v1

Deterministic pre-deployment risk scoring. The score is a **weighted operational signal**, not a probability of failure and not a prediction.

Every assessment persists:

- `score` (0–100, clipped)
- `score_raw` (sum before clip)
- `category` (`LOW` | `MODERATE` | `HIGH` | `CRITICAL`)
- `factors[]` (`code`, `label`, `points`, `rationale`, `input`)
- `model_version` (`risk-v1`)
- `assessed_at`

UI copy must say "Release risk signal" or "Release risk", never "probability of outage."

## Category bounds

| Score | Category |
| --- | --- |
| 0–24 | LOW |
| 25–49 | MODERATE |
| 50–74 | HIGH |
| 75–100 | CRITICAL |

## Factor model

Points are additive. Missing optional data contributes `0` and is listed as `insufficient_input` when the factor would otherwise apply. The engine never invents a factor to fill the score.

### 1. Change size — max 10 — `change_size`

Uses the larger of file-band and line-band.

| Files changed | Points |
| --- | --- |
| 0–5 | 0 |
| 6–20 | 3 |
| 21–50 | 6 |
| 51+ | 10 |

| Lines changed (added+deleted) | Points |
| --- | --- |
| 0–49 | 0 |
| 50–300 | 4 |
| 301–1000 | 7 |
| 1001+ | 10 |

Why: large diffs are harder to review and more likely to include unintended coupling. This is a size heuristic, not a quality judgment.

### 2. Service criticality — max 20 — `service_criticality`

| Criticality | Points |
| --- | --- |
| LOW | 0 |
| MODERATE | 4 |
| HIGH | 10 |
| CRITICAL | 20 |

Why: the same change size is not the same operational bet on `payments-service` vs `notification-worker`.

### 3. Database migration — max 15 — `database_migration`

`+15` if the change set is flagged `migration_present`.

Why: schema changes have rollback asymmetry. Expand/contract discipline is not inferred; presence alone is the signal.

### 4. Configuration change — max 8 — `config_change`

`+8` if flagged `config_change_present` (runtime config, feature flags, IAM-adjacent manifests, secret references).

Why: config errors skip compile-time feedback and often present as "the code didn't change much."

### 5. Failed CI attempts — max 10 — `ci_failed_attempts`

For the release branch/SHA before the successful run: `min(failed_attempts * 4, 10)`.

Why: flakiness or repeated breakage is a process-risk signal. It is not proof the final artifact is bad.

### 6. Test failures in the artifact run — max 15 — `test_failures`

`+15` if the run attached to the release has failing tests. Policy will typically BLOCK; risk still records the factor so replay and timeline stay coherent.

### 7. Recent deployment failures — max 10 — `recent_deploy_failures`

Failed deployments for the same service in the trailing 14 days:

| Count | Points |
| --- | --- |
| 0 | 0 |
| 1 | 5 |
| 2+ | 10 |

### 8. Dependency fan-out — max 12 — `dependency_fanout`

Count of distinct **dependents** of the changed service (direct + transitive), from the graph engine, excluding the changed service itself.

| Dependents | Points |
| --- | --- |
| 0–1 | 0 |
| 2–3 | 6 |
| 4–6 | 9 |
| 7+ | 12 |

Why: more reverse-dependencies means a larger potential blast radius. This is potential impact, not observed impact.

### 9. Recent incident frequency — max 8 — `recent_incidents`

Incidents on the changed service in the trailing 30 days:

| Count | Points |
| --- | --- |
| 0 | 0 |
| 1 | 4 |
| 2+ | 8 |

### 10. Rollback gap — max 7 — `rollback_gap`

| Rollback readiness | Points |
| --- | --- |
| READY | 0 |
| PARTIAL | 3 |
| UNKNOWN | 4 |
| NOT_READY | 7 |

Circular evaluation: risk may be computed before rollback if rollback inputs exist; otherwise this factor is `0` with `insufficient_input` and a later reassessment updates it. Implementation order: compute rollback first, then risk, in the worker.

### 11. Security findings — max 15 — `security_findings`

Highest severity on the attached scan of the artifact:

| Highest | Points |
| --- | --- |
| none / informational | 0 |
| MEDIUM | 4 |
| HIGH | 10 |
| CRITICAL | 15 |

### 12. Historical release failure rate — max 8 — `historical_failure_rate`

Among the last 10 **completed** deployments of the same service:

| Failure rate | Points |
| --- | --- |
| < 10% | 0 |
| 10–29% | 4 |
| ≥ 30% | 8 |

Fewer than 3 historical deployments: factor omitted (`insufficient_input`), not guessed.

## Normalization and bounds

- Sum `score_raw`.
- `score = min(100, score_raw)`.
- Never negative. Factors cannot subtract in v1 (no "good vibes" discounting). A clean CI run simply contributes 0.
- Unknown services: `service_criticality` treated as HIGH (10) with rationale "unknown service defaulted to HIGH" rather than silent 0.

## Limitations (must appear in UI help and docs)

- Not calibrated against a large failure corpus.
- Weights are engineering judgment for a commerce platform.
- Does not know code semantic risk (a 3-line auth change can be worse than a 400-line CSS change). Critical-path flags and migration flags exist to compensate partially.
- Fan-out counts reachable dependents, not actual runtime traffic.
- Historical rate is unstable on tiny samples and is suppressed below 3 deploys.

## Test obligations

Table-driven tests for: each factor band, clipping at 100, omitted insufficient inputs, unknown service default, category boundaries (24/25, 49/50, 74/75), and "no mysterious leftover points" (sum of factor points equals `score_raw`).
