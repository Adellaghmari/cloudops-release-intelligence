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

X-Ray **Active** tracing on the API Lambda and worker Lambda is LIVE VERIFIED.

API Gateway is an **HTTP API** (`apigatewayv2`). HTTP APIs do not provide a standalone API Gateway X-Ray segment (unlike REST API). That is an intentional cost/simplicity tradeoff, not a missing defect to "fix" by migrating API types.

Pass `correlation_id` as an annotation when present. Additional DynamoDB subsegments are optional polish only — do not rebuild the backend solely for a prettier service map.

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
