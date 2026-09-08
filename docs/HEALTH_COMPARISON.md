# Deployment Health Comparator v1

Compares a **baseline window** before deployment with a **post-deployment window** after a warmup gap. Verdicts are thresholded operational judgments, not a claim that the release caused the change.

Persisted result includes raw window aggregates, per-metric deltas, per-metric verdicts, overall verdict, and an optional correlation object.

## Windows

Let `T` be `deployment.succeeded_at` (UTC).

| Window | Interval | Why |
| --- | --- | --- |
| Baseline | `[T-60m, T)` | Recent pre-change behavior |
| Warmup (excluded) | `[T, T+2m)` | Avoid deploy-time blip |
| Post | `[T+2m, T+32m]` | 30-minute observation |

If either window has insufficient traffic, overall verdict is `INSUFFICIENT_DATA` and correlation is `INSUFFICIENT_EVIDENCE`.

Demo snapshots are **pre-aggregated window records**, not raw CloudWatch datapoints. Live Project B health may later be filled from CloudWatch metric math for `cloudops-api` / `cloudops-web` only.

## Metrics

| Key | Unit | Direction of harm |
| --- | --- | --- |
| `request_count` | count | context only (sample size) |
| `error_rate` | 0–1 ratio | increase |
| `availability` | 0–1 ratio | decrease |
| `latency_p50_ms` | milliseconds | increase |
| `latency_p95_ms` | milliseconds | increase |
| `latency_p99_ms` | milliseconds | increase |
| `cpu_pct` | 0–100 | increase (optional) |
| `memory_pct` | 0–100 | increase (optional) |
| `queue_backlog` | count | increase (optional; workers) |

v1 always evaluates error rate, availability, p95. p50/p99 are shown. CPU/memory/queue participate when present on the snapshot.

## Minimum sample

`INSUFFICIENT_DATA` when:

- `request_count` in baseline **or** post `< 50`, or
- snapshot missing a required metric, or
- deployment success timestamp missing

## Per-metric verdicts

Defaults (override per service SLO when recorded on the service):

| Metric | SLO default | DEGRADED | SEVERELY_DEGRADED |
| --- | --- | --- | --- |
| error_rate | 1% | post > max(baseline×2, baseline+0.5pp) **or** post > SLO | post ≥ baseline×5 **and** post ≥ 2% , or post ≥ 5% |
| availability | 99.9% | drop > 0.20pp from baseline **or** post < SLO | drop > 1.00pp **or** post < 99.0% |
| latency_p95_ms | service SLO or 500ms | post > baseline×1.5 **and** +50ms absolute | post > baseline×3 **and** +200ms absolute |

`pp` = percentage points on the displayed percent scale (0.5pp means +0.005 on a 0–1 ratio).

STABLE if neither degraded rule fires.

No z-score in v1. Window aggregates are single points; a z-score against n=1 is theatre. If later we store actual timeseries inside a window, z-score can be revisited.

## Overall verdict

Worst of evaluated core metrics (error_rate, availability, p95):

`SEVERELY_DEGRADED > DEGRADED > INSUFFICIENT_DATA > STABLE`

`INSUFFICIENT_DATA` beats `STABLE` so the product does not call an empty window healthy.

## Release correlation v1

Computed only when overall is `DEGRADED` or `SEVERELY_DEGRADED`.

Evidence flags (all stored, never hidden):

| Flag | Rule |
| --- | --- |
| `temporal_proximity` | Degradation window starts within 15 minutes after `T` (true by construction of the post window) |
| `baseline_clear` | Baseline overall would have been STABLE using the same rules against a synthetic "pre-baseline" is **not** used. Instead: baseline error_rate ≤ SLO and availability ≥ SLO |
| `service_match` | Degraded service is the changed service |
| `graph_upstream` | Changed service is upstream of the degraded service (the degraded service depends on the changed service, transitively) |
| `incident_aligned` | An incident for that service opened in `(T, T+32m]` |

Result:

| Result | When |
| --- | --- |
| `LIKELY_RELEASE_CORRELATION` | `baseline_clear` AND (`service_match` OR `graph_upstream`) |
| `UNLIKELY` | degraded but baseline already breached the same SLO |
| `INSUFFICIENT_EVIDENCE` | degraded but neither service match nor graph relationship, or missing graph |

UI label examples:

- LIKELY RELEASE CORRELATION
- POST DEPLOYMENT REGRESSION
- INSUFFICIENT EVIDENCE

Forbidden labels: "root cause confirmed", "proven causation", "the release caused this."

## Limitations

- Fixed windows miss slow burns after 32 minutes.
- Percentage deltas on tiny baselines (error_rate 0.0 → 0.4%) are gated by absolute floors.
- Synthetic snapshots can be internally consistent and still not represent production physics.
- Multi-release overlap (two deploys in the same hour) is flagged `overlapping_deployments` when another successful deploy for a graph-neighbor occurs in `[T-30m, T+32m]`. Correlation then becomes `INSUFFICIENT_EVIDENCE` even if other flags match.

## Test obligations

- insufficient sample
- stable within noise
- error-rate degraded vs severely
- availability drop
- p95 relative+absolute gate
- overlapping deployments suppress correlation
- unknown service health → INSUFFICIENT_DATA
