# Failure modes

The product must fail clearly. It must not hallucinate a neat conclusion from missing data.

## Ingestion and async

| Case | Behavior |
| --- | --- |
| Duplicate event | 200, `duplicate: true`, no second analysis emit |
| Out of order event | `ANALYSIS_PENDING` until parents exist, then recompute |
| Malformed event | 400, never queued |
| Unknown event type | persist raw, ignore analysis, no retry storm |
| Worker exception | SQS retry up to 3, then DLQ, alarm |
| Poison message | DLQ; operator/replay is manual; no infinite loop |
| EventBridge outage | ingest still persists DynamoDB+S3; return 503 if emit fails after persist? **Decision:** persist first, emit second; if emit fails, API returns 503 and relies on a compensation: a later worker catch-up is Phase 16. Until then, failed emit is logged and the idempotency record is **not** written so a client retry may PutEvents. Refine in implementation so we cannot double-persist timeline rows. See note below. |

**Idempotency vs emit failure (resolved rule):** write timeline + S3, then conditional idempotency item `pending_emit=true`, then PutEvents, then clear `pending_emit`. Retries resume emit if the flag is set. Worker is still idempotent per `event_id`+kind.

## Storage and compute

| Case | Behavior |
| --- | --- |
| DynamoDB timeout | bounded retry, then 503/worker fail |
| Missing deployment | health comparison INSUFFICIENT_DATA |
| Missing metrics | INSUFFICIENT_DATA, not STABLE |
| Partial health window | INSUFFICIENT_DATA |
| Policy engine error | BLOCK with `engine.evaluation_failure` |
| Missing rollback metadata | UNKNOWN or NOT_READY per [ROLLBACK_READINESS.md](ROLLBACK_READINESS.md) |
| Unknown service | 404 on resource routes; graph UNKNOWN_SERVICE; risk defaults criticality HIGH with explicit factor |
| Cyclic dependency graph | detected, BFS still terminates, cycles listed |

## Product honesty

| Temptation | Instead |
| --- | --- |
| Call empty metrics healthy | INSUFFICIENT_DATA |
| Call correlation causation | LIKELY_RELEASE_CORRELATION + evidence flags |
| Hide overlapping deploys | INSUFFICIENT_EVIDENCE |
| Invent blast radius | empty + unknown |
| Retry forever | cap 3 + DLQ |

## Chaos tests (required as they are built)

Go tests for the table above. Later: one worker test that a failing handler does not exceed 3 simulated receives. k6 is for happy-path load, not chaos.
