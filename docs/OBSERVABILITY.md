# Observability

Project B uses **Amazon CloudWatch** and **AWS X-Ray**, not an Azure-centric telemetry stack.

## Logs

Structured JSON on stdout from Lambda.

Required fields when known:

- `level`
- `msg`
- `request_id`
- `correlation_id`
- `event_id`
- `release_id`
- `service_id`
- `deployment_id`

Never log secrets, signatures, or full raw webhook bodies (S3 pointer only).

Retention: 14 days.

## Traces

X-Ray on API Lambda and worker Lambda. Pass `correlation_id` as an annotation. Subsegments for DynamoDB, EventBridge, S3, OPA eval, risk compute.

## Metrics (custom namespace `CloudOps`)

| Metric | Why |
| --- | --- |
| `ApiRequests` | traffic |
| `ApiLatencyMs` | p50/p95 via log metric or EMF |
| `EventsIngested` | ingest volume |
| `EventsDuplicate` | idempotency working |
| `WorkerInvocations` | async path alive |
| `WorkerFailures` | worker bugs / poison |
| `AnalysisDurationMs` | risk+health+graph+policy |
| `DynamoErrors` | storage health |
| `DlqVisible` | alarm immediately |
| `DemoResets` | abuse detection |

## Alarms (prod)

- DLQ visible messages > 0
- Worker error rate
- API 5xx
- Budget (see [COST.md](COST.md))

No paging integration in v1 (portfolio). Alarms exist and are documented with screenshots/URLs when live.

## Health endpoints

- `GET /api/v1/health` — process up
- `GET /api/v1/ready` — DynamoDB `DescribeTable` or a cheap consistent get
- `GET /api/v1/status` — composed view for the System Status page (API, queue depth if permitted, last successful analysis)

Frontend should show degraded API as an explicit error state, not an empty dashboard pretending to be healthy.
