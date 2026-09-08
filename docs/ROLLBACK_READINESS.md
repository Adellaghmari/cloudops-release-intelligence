# Rollback readiness

Assesses whether a **manual** rollback appears operationally possible. The product does not roll back production.

## Result

| Status | Meaning |
| --- | --- |
| READY | All required prerequisites present |
| PARTIAL | Some required, some missing/unknown |
| NOT_READY | A hard prerequisite is missing |
| UNKNOWN | Too little metadata to judge |

## Signals

| Signal | Required | Pass condition |
| --- | --- | --- |
| `previous_successful_release` | yes | A prior release for the same service + environment with deployment status succeeded and health not SEVERELY_DEGRADED if health exists |
| `previous_artifact` | yes | Artifact URI or filename recorded on that prior release |
| `previous_image_digest` | yes for container services | `sha256:` digest recorded |
| `deployment_target_known` | yes | environment + target (Lambda alias/function name, S3/CloudFront IDs, etc.) |
| `previous_version_recorded` | yes | immutable version (git SHA or version string) |
| `migration_reversibility` | yes if current change has `migration_present` | `reversible=true` or `rollback_plan_ref` present; else fail this signal |
| `config_rollback` | no | prior config hash recorded |

Container services in this product: Go API/worker. Frontend is an S3 artifact (object version / git SHA), not a digest; `previous_image_digest` is then `not_applicable` (counts as pass).

## Aggregation

- All required applicable signals pass → `READY`
- Any required signal is `missing` (known absent) → at least `PARTIAL`; if `previous_artifact` **or** `previous_successful_release` is missing → `NOT_READY`
- If required signals are `unknown` (metadata not loaded) and none are known-missing → `UNKNOWN`
- Mix of pass and unknown without known-missing → `PARTIAL`

## Output example

```json
{
  "status": "PARTIAL",
  "signals": [
    { "id": "previous_artifact", "ok": true },
    { "id": "previous_image_digest", "ok": true },
    { "id": "migration_reversibility", "ok": false, "detail": "migration present; reversibility unknown" },
    { "id": "previous_successful_release", "ok": true }
  ],
  "missing": ["migration_reversibility"],
  "model_version": "rollback-v1"
}
```

## Limitations

- Does not verify the prior artifact still exists in ECR/S3, only that metadata was recorded. A later live check against ECR can upgrade this; v1 is metadata-based.
- Does not execute rollback.
- Reversibility is an asserted flag from CI/demo metadata, not a database parser.

## Test obligations

- happy path READY
- missing previous artifact → NOT_READY
- migration without reversibility → PARTIAL
- frontend not_applicable digest
- empty metadata → UNKNOWN
