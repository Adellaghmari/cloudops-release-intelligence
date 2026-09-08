# Event architecture

At least one path is genuinely asynchronous: **ingest persists, then EventBridge fans out to SQS, then the worker analyzes.**

## Envelope

All events share a versioned envelope. `event_id` is stable for idempotency. Producers that have a natural ID (GitHub `workflow_run.id` + event name) must hash/map to a deterministic `event_id`. Otherwise use ULID.

```json
{
  "event_id": "01J...",
  "event_type": "deployment.succeeded",
  "schema_version": "1.0",
  "occurred_at": "2026-09-08T11:42:00Z",
  "ingested_at": "2026-09-08T11:42:01Z",
  "producer": "github-actions",
  "correlation_id": "rel_...",
  "release_id": "rel_...",
  "service_id": "payments-service",
  "deployment_id": "dep_...",
  "payload": {}
}
```

Unknown future fields in `payload` are stored raw in S3 and ignored by v1 workers. `schema_version` major bump is required for breaking envelope changes. v1 workers drop unsupported `event_type` to an `ignored` audit item, they do not crash-loop.

## Event types v1

| Type | Triggers analysis |
| --- | --- |
| `change.commit.recorded` | maybe risk (if CI already present) |
| `ci.run.started` | timeline only |
| `ci.run.completed` | risk, policy PRE_DEPLOY |
| `security.scan.completed` | risk, policy PRE_DEPLOY |
| `artifact.published` | rollback inputs, timeline |
| `deployment.started` | timeline |
| `deployment.succeeded` | rollback, policy PRE if missing, schedule/await health |
| `deployment.failed` | risk inputs for later releases, timeline |
| `health.snapshot.recorded` | health comparator + correlation + policy POST_DEPLOY |
| `incident.opened` | correlation refresh, timeline |
| `incident.resolved` | timeline |
| `decision.recorded` | timeline |
| `policy.evaluated` | timeline (emitted after eval; must not re-enter policy infinitely) |

`policy.evaluated` is **not** subscribed by the analysis rule that runs OPA. That prevents a loop.

## Ingest path

1. Validate envelope (types, RFC3339 timestamps, IDs, payload size cap 256 KB).
2. Put raw body to S3 key:
   `s3://{bucket}/raw/dt={yyyy-mm-dd}/{event_type}/{event_id}.json`
3. DynamoDB conditional put `PK=IDEM#{event_id}` — if item exists, return 200 with `duplicate: true` and do not PutEvents again.
4. Write canonical event under the release timeline key.
5. `PutEvents` to the custom bus with `event_id` as the EventBridge `Resources` / trace field.
6. Return `202 Accepted` for new events (`200` for duplicates).

Malformed events: `400` with `INVALID_EVENT`. They never reach SQS.

## Queue and worker

- EventBridge rule pattern: analysis types listed above (except `policy.evaluated` as a trigger).
- SQS: visibility timeout 60s, `maxReceiveCount=3`, DLQ `cloudops-analysis-dlq`.
- Worker uses `event_id` + `analysis_kind` idempotency (`IDEM#{event_id}#{kind}`) so partial retries do not double-write contradictory assessments. Latest assessment pointers are overwritten with the same computed bytes.
- Poison JSON after 3 receives lands on DLQ. CloudWatch alarm on DLQ `ApproximateNumberOfMessagesVisible > 0`.
- Bounded retries inside the worker for DynamoDB: 3 attempts, exponential backoff with jitter, respect context cancel. No infinite loop.

## Out of order

Events can arrive out of order (health before deploy metadata). Worker rules:

- If required parents are missing, write `ANALYSIS_PENDING` with reason and rely on the later event to recompute.
- Do not invent a deployment time.
- Duplicate delivery is success.

## GitHub webhooks

If/when enabled: verify `X-Hub-Signature-256` with the webhook secret from SSM/Secrets Manager. Reject unsigned bodies. Never log the signature or secret.

Until webhook auth is live, GitHub evidence enters via the project's own Actions job posting a signed-or-IAM-internal payload, plus the synthetic demo path.

## Producers

| Producer | Trust |
| --- | --- |
| `synthetic-demo` | Demo reset only |
| `github-actions` | OIDC-signed AWS call or webhook HMAC |
| `cloudops-api` | Internal |

## Local implementation (Phase 5)

Local async behavior is exercised without AWS:

| Port | Local adapter | Production adapter (later) |
| --- | --- | --- |
| `events.Bus` | `MemoryBus` (in-process `Processor.Handle`) | EventBridge `PutEvents` (not built) |
| Worker | `Processor` in the API process | Lambda consumer of SQS (not built) |
| DLQ | in-memory `Processor.DLQ()` | SQS DLQ `cloudops-analysis-dlq` (not built) |
| Idempotency | `CreateEvent` conditional on `event_id` | DynamoDB `IDEM#{event_id}` (adapter already designed) |

`POST /api/v1/events` accepts a versioned envelope (`schema_version` `1.0`). New events return `202 Accepted`. Duplicate `event_id` returns `200` with `duplicate: true` and does not create a second domain effect. Malformed IDs, unknown types, and unsupported schema versions return `400 INVALID_EVENT` and are written to the local DLQ.

`policy.evaluated` is persisted for the timeline and does **not** dispatch analysis (`TriggersAnalysis` is false). That prevents an evaluation loop.

Out-of-order: if `release_id` is set but the release row is missing, the event is stored and the result is `pending`. A later release create plus re-ingest of a **new** event_id can complete analysis; a duplicate of the same event_id remains a no-op.

Bounded retries: transient `CreateEvent` failures retry up to 3 times, then the envelope is dead-lettered as poison.

### Not AWS verified

Docker and Java are unavailable on this workstation, so LocalStack, EventBridge, and SQS were not executed. Do not claim real EventBridge/SQS delivery, visibility timeout, or `maxReceiveCount` behavior. Those remain PLANNED until Phase 12 infrastructure exists.

## Test obligations

- duplicate `event_id`
- out-of-order health vs deploy
- malformed envelope
- ignored unknown type
- worker idempotency per kind
- no policy.evaluated reprocessing loop
