# Policy as code

Release Policy Gate evaluates versioned Rego policies with Open Policy Agent **embedded in the Go process**. There is no OPA server, sidecar, or cluster.

Results are not AI. They are deterministic rule matches against an explicit JSON input.

## Bundle

```text
policies/
  VERSION          # semver, e.g. 1.0.0
  data.json        # optional static data (SLO defaults, critical services)
  release_gate.rego
```

`policy_version` stored on every evaluation is `VERSION` plus git SHA when running in CI/prod (`1.0.0+abc1234`). Locally, SHA may be `dev`.

## Decision vocabulary

Each rule emits one of:

| Result | Meaning |
| --- | --- |
| PASS | Rule inspected input and did not fire |
| WARN | Proceed allowed; evidence of concern |
| BLOCK | Promotion / continue is not allowed by policy |
| MANUAL_APPROVAL_REQUIRED | Proceed only with an explicit ReleaseDecision |

Aggregate (worst wins):

`BLOCK > MANUAL_APPROVAL_REQUIRED > WARN > PASS`

A release with both WARN and MANUAL_APPROVAL_REQUIRED is `MANUAL_APPROVAL_REQUIRED`.

## v1 rules

| Rule ID | Result | Condition |
| --- | --- | --- |
| `tests.must_pass` | BLOCK | Attached CI run has failing tests |
| `security.critical_vuln` | BLOCK | Scan highest severity is CRITICAL |
| `security.high_vuln` | WARN | Highest severity is HIGH |
| `risk.high_threshold` | WARN | Risk score ≥ 50 |
| `risk.critical_threshold` | MANUAL_APPROVAL_REQUIRED | Risk score ≥ 75 |
| `migration.rollback_plan` | WARN | `migration_present` and rollback is not READY |
| `critical_service.change` | MANUAL_APPROVAL_REQUIRED | Changed service criticality is CRITICAL |
| `health.availability_slo` | BLOCK | Post-deploy availability verdict is DEGRADED or SEVERELY_DEGRADED **and** comparison exists |
| `health.severe_regression` | BLOCK | Overall health is SEVERELY_DEGRADED |

Pre-deploy evaluations omit health rules (`skipped`, with reason `no_post_deploy_window`). Post-deploy re-evaluation is a new `PolicyEvaluation` row, not a silent mutation.

## Input document

The Go layer builds a stable input. Policies must not call AWS or HTTP.

```json
{
  "release": { "id": "...", "service_id": "payments-service" },
  "service": { "id": "payments-service", "criticality": "CRITICAL" },
  "change": {
    "files_changed": 42,
    "migration_present": true,
    "config_change_present": false
  },
  "ci": { "failed_tests": 0, "status": "succeeded" },
  "security": { "highest_severity": "NONE" },
  "risk": { "score": 72, "category": "HIGH" },
  "health": { "overall": "STABLE", "available": true },
  "rollback": { "status": "PARTIAL" }
}
```

## Persisted evaluation

Each evaluation stores:

- `evaluation_id`
- `policy_version`
- `result` (aggregate)
- `rules[]` with `id`, `result`, `message`, `input_excerpt`
- `evaluated_at`
- `phase` (`PRE_DEPLOY` | `POST_DEPLOY`)

UI must show **which policy fired, why, what input, what version**.

## Failure of the engine

If OPA returns an error, the product records `result: BLOCK` with rule `engine.evaluation_failure` and a sanitized error. Failing open would be worse than a blocked demo.

## Test obligations

- each rule fires on a fixture input
- aggregate precedence
- skipped health rules pre-deploy
- evaluation failure → BLOCK
- version string recorded
