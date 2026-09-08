# Architecture

## Decision summary

| Decision | Choice | Why |
| --- | --- | --- |
| Frontend | Angular standalone + RxJS + SCSS | Distinct from React/Next; real operations-console SPA |
| Backend | Go + Gin | Distinct from Python/FastAPI; typed domain, testable engines |
| Primary datastore | DynamoDB on-demand, single table | Access-pattern first, serverless, no RDS cost |
| Raw evidence | S3 | Immutable webhook/CI/scan payloads |
| Async path | EventBridge → SQS → Lambda worker → DLQ | Real event-driven processing, bounded retries |
| Compute | Lambda (container image from ECR) | No always-on compute; image digest supports rollback evidence |
| Edge | CloudFront + S3 (web), API Gateway HTTP API (API) | TLS, low idle cost |
| Region | `eu-west-1` | User timezone UTC+2; full service coverage; single region |
| Policy | OPA embedded in the worker/API process | Policy as code without a standing OPA cluster |
| Auth for public demo | Unauthenticated read + rate-limited write for demo reset | Recruiter usable; no login theatre |
| CI auth to AWS | GitHub OIDC | No long-lived access keys |
| Kubernetes | Not used | Would add cost and complexity without serving this architecture |
| LLM | Not used | Deterministic operational analysis is the point |

## High-level flow

```text
Producer (GitHub webhook, demo seed, or internal API)
    → API Gateway HTTP API
    → Lambda cloudops-api (Go/Gin)
        → validate + normalize
        → S3 put (raw payload, if present)
        → DynamoDB conditional put (idempotent canonical event)
        → EventBridge PutEvents
            → rule: analysis events
            → SQS cloudops-analysis
                → Lambda cloudops-worker
                    → risk / health / graph / policy / rollback
                    → DynamoDB assessments
                → SQS DLQ after 3 failures

Angular console
    → CloudFront → S3 static origin
    → CloudFront /api/* → API Gateway
```

The analysis engines are plain Go packages. The worker is the production caller. Demo seed may invoke the same packages in-process after writing inputs so the recruiter does not wait on the queue, then still emit EventBridge events. Duplicate worker delivery is a no-op via idempotent assessment keys.

Do not compute "the product" entirely inside one HTTP handler with a fake queue.

## Repository layout (target)

```text
cmd/api/                 local HTTP server (Lambda adapter later)
internal/domain/         entities, typed IDs, enums, errors
internal/httpapi/        Gin routers, middleware, DTOs (named to avoid colliding with net/http)
internal/service/        application use cases
internal/repository/     persistence ports
internal/repository/memory/  in-memory adapter
internal/repository/dynamo/  AWS SDK v2 adapter (local fake tests)
internal/events/         envelope, Processor, MemoryBus, local DLQ
internal/risk/           deterministic risk engine
internal/health/         pre/post comparator + correlation
internal/config/         env/config
internal/localseed/      local synthetic + live-identity catalog
frontend/                Angular 21 operations console
```

Reserved for later phases (do not create empty):

```text
cmd/worker/
internal/aws/
internal/graph/
internal/policy/
internal/rollback/
internal/replay/
internal/demo/
policies/
infra/
.github/workflows/
```

Go module: `github.com/adell/cloudops-release-intelligence` (placeholder until the public GitHub remote exists). Toolchain: Go 1.27.

HTTP handlers must not contain scoring, graph, or policy logic. AWS SDK calls must not leak into `internal/domain` or the engine packages.

Phase 1 list filters: `GET /api/v1/services?source=live|synthetic` and `GET /api/v1/releases?source=&service_id=`. These are query conveniences, not DynamoDB key design leaked into the domain.

## Compute packaging

One Go module, two binaries, one container image with distinct commands:

- `api` — Gin behind `aws-lambda-go-api-proxy` for API Gateway HTTP API; also `http.ListenAndServe` when `APP_ENV=local`
- `worker` — SQS event handler

Image identity is `${ECR_REPO}@sha256:...`. That digest is a first-class rollback signal and a first-class live demo field.

Lambda is **not** placed in a VPC. VPC Lambda would require NAT Gateway cost for AWS API access.

## API shape

Versioned REST under `/api/v1`. JSON only.

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/api/v1/health` | Liveness |
| GET | `/api/v1/ready` | Readiness of **current** dependencies only (Phase 1: `in_memory_store`) |
| GET | `/api/v1/releases` | Recent releases |
| GET | `/api/v1/releases/{id}` | Aggregate release decision surface |
| GET | `/api/v1/releases/{id}/risk` | Risk assessment |
| GET | `/api/v1/releases/{id}/health` | Health comparison |
| GET | `/api/v1/releases/{id}/impact` | Blast radius |
| GET | `/api/v1/releases/{id}/policies` | Policy evaluations |
| GET | `/api/v1/releases/{id}/rollback` | Rollback readiness |
| GET | `/api/v1/releases/{id}/timeline` | Evidence events |
| GET | `/api/v1/releases/compare?a=&b=` | Release Replay |
| GET | `/api/v1/services` | Service catalog |
| GET | `/api/v1/services/{id}` | Service + edges |
| GET | `/api/v1/incidents` | Incidents |
| POST | `/api/v1/events` | Authenticated/internal ingestion (signature or IAM) |
| POST | `/api/v1/demo/reset` | Rate-limited synthetic reseeding |
| GET | `/api/v1/status` | System status for the Architecture/Status page |

Public demo reads are open. `POST /events` is not a public recruiter toy: GitHub signature or IAM. `POST /demo/reset` is public but strictly rate-limited and synthetic-only.

Error bodies are stable:

```json
{
  "error": {
    "code": "RELEASE_NOT_FOUND",
    "message": "release not found",
    "request_id": "..."
  }
}
```

No stack traces, no AWS error internals, no webhook signatures.

## Frontend architecture

Angular current stable, standalone components, typed models, `HttpClient`, route-level loading/error states, SCSS with a structured theme.

Reactive rules:

- HTTP in services, not components
- RxJS for streams (polling, combineLatest for release detail)
- Angular signals for local view state where they fit current Angular
- Do not recreate React Query or Redux theatre

Visual identity: dark engineering operations console. Dense signal, tabular numbers, monospace IDs. Amber for risk, red for block/degraded, green for pass/stable. Not a pastel SaaS marketing dashboard and not a copy of an Azure/AI product.

## Environments

| Name | Purpose |
| --- | --- |
| local | In-memory or DynamoDB local, no AWS cost |
| staging | Optional same-account isolated resource prefix `cloudops-stg-*` if budget allows; otherwise skipped |
| prod | Public recruiter environment `cloudops-prod-*` |

Portfolio default: **one production environment** plus local. A second full AWS environment doubles cost. If staging is added, it must share the same Terraform and use a workspace or `environment` variable, not a copy-paste stack.

## Reliability defaults

- Timeouts on all outbound AWS calls via `context.Context`
- SQS `maxReceiveCount = 3`, then DLQ
- Idempotency on `event_id` via DynamoDB `attribute_not_exists`
- Assessments keyed so recomputation is overwrite-safe
- Rate limit demo reset and public GETs at API Gateway
- Worker visibility timeout > worst-case analysis duration

## Observability

Structured JSON logs + CloudWatch metrics + X-Ray on API and worker. Correlation fields: `request_id`, `correlation_id`, `event_id`, `release_id`, `service_id`, `deployment_id`. Details in [OBSERVABILITY.md](OBSERVABILITY.md).

## What we refused

| Temptation | Why not |
| --- | --- |
| EKS | Idle cost, ops burden, unused for this traffic shape |
| RDS/PostgreSQL | Project A already covers SQL; DynamoDB is the point |
| NAT Gateway | Avoided by keeping Lambda out of VPC |
| Always-on EC2/ECS service | Idle cost |
| Separate OPA sidecar/cluster | Embed the library |
| Microservices split (risk-service, graph-service, ...) | Modular monolith in one Go module is the correct size |
| Vercel / Azure / React | Explicitly out of scope |

## Next implementation slice

Phase 1 builds `internal/domain` + Gin API with an in-memory repository so engines and HTTP contracts can be tested before AWS.
