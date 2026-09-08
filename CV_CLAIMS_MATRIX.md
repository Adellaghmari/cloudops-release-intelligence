# CV claims matrix

Truthfulness rule: nothing goes on a CV because a package was mentioned, a Terraform resource exists locally, a mock test passed, or documentation describes the intent.

Allowed status values:

| Status | Meaning |
| --- | --- |
| PLANNED | Designed, not built |
| IMPLEMENTED | Code or config exists in this repository |
| TESTED | Automated tests cover the claim with meaningful assertions |
| LIVE VERIFIED | Proven in the public/production environment |

LIVE VERIFIED is the only status that supports a strong public claim such as "I deployed X on AWS."

This remains a portfolio project. It is not professional work experience.

## Platform and languages

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Go backend | TESTED | Public Go API serving real JSON | Local `cmd/api` + `go test ./...`. Not publicly deployed. |
| Gin HTTP API | TESTED | Versioned REST, tests, live `/api/v1` | Local `/api/v1` health/ready/services/releases + httptest. Not live. |
| Angular frontend | TESTED | Public CloudFront/S3 console | Angular 21 local console + production build + unit tests. Not publicly deployed. Angular 22 needs Node ≥22.22.3; this machine has 22.14.0. |
| TypeScript | TESTED | Angular app compiled and deployed | `ng build` succeeds locally. |
| RxJS | TESTED | HTTP/state streams in the console | Catalog/release pages use HttpClient + RxJS. |
| SCSS | TESTED | Structured styles, no Tailwind | Global + component SCSS, no Tailwind. |

## AWS

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| AWS Lambda | PLANNED | Terraform-applied functions serving traffic | None |
| Amazon API Gateway | PLANNED | Public HTTPS API | None |
| Amazon DynamoDB | TESTED | Production table with real items | Adapter + single-table mapping + conditional EventID writes tested against a local DynamoDB-compatible fake. No production table. Not DynamoDB Local (Docker/Java unavailable). |
| Amazon S3 | PLANNED | Frontend origin and/or raw event objects | None |
| Amazon CloudFront | PLANNED | Public site URL | None |
| Amazon EventBridge | PLANNED | Custom bus receiving real events | Local `events.Bus` port + MemoryBus only. No EventBridge resource. Not AWS verified. |
| Amazon SQS | PLANNED | Worker consuming analysis queue + DLQ | Local Processor + in-memory DLQ. No SQS queue. Not AWS verified. |
| Amazon ECR | PLANNED | Immutable image digest deployed to Lambda | None |
| Amazon CloudWatch | PLANNED | Production logs/metrics in the account | None |
| AWS X-Ray | PLANNED | Traces for API and worker | None |
| AWS IAM least privilege | PLANNED | Applied roles with scoped policies | None |
| GitHub OIDC to AWS | PLANNED | Workflows assume role without static keys | None |
| AWS Secrets Manager | PLANNED / MAY OMIT | Only if a real secret is required | Intentionally avoided unless necessary |

## Infrastructure, CI, DevSecOps

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Terraform | PLANNED | `plan`/`apply` of the intended stack | None |
| GitHub Actions CI | PLANNED | Successful workflow runs on the repo | None |
| GitHub Actions CD | PLANNED | Staging/prod deploy from main | None |
| Trivy scanning | PLANNED | Workflow step that can fail the build | None |
| Syft SBOM | PLANNED | Generated artifact attached to release/build | None |
| Cosign keyless signing | PLANNED | Signed ECR image verified in CI | None |
| Open Policy Agent / Rego | PLANNED | Policies execute against real release input | None |

## Product capabilities

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Release Risk Engine | TESTED | Persisted score + contributing signals from real logic | `internal/risk` + GET /releases/:id/risk. Local only. |
| Deployment Health Comparator | TESTED | Pre/post windows, raw metrics, verdict | `internal/health` + GET /releases/:id/health. Local only. Not CloudWatch verified. |
| Release correlation | TESTED | Explicit evidence object, not "causation" | LIKELY/POSSIBLE/NO_CLEAR/INSUFFICIENT_DATA with reasons. Never PROVEN_CAUSE. Local only. |
| Change Impact Graph | TESTED | BFS/DFS blast radius from persisted edges | `internal/graph` + GET /releases/:id/impact + SVG. Language is potential impact. Local only. |
| Release Policy Gate | PLANNED | Versioned policy eval with PASS/WARN/BLOCK/MANUAL | Spec only |
| Rollback Readiness | PLANNED | READY/PARTIAL/NOT READY/UNKNOWN + missing prereqs | Spec only |
| Release Replay | PLANNED | Deterministic diff of two persisted releases | Spec only |
| Release evidence timeline | PLANNED | Chronological events from storage | Spec only |
| Event-driven analysis | TESTED | Ingest → EventBridge → SQS → worker, not in-request theatre | Local envelope + Processor + POST /events + MemoryBus. EventBridge/SQS path not built. Not AWS verified. |
| Idempotent ingestion | TESTED | Duplicate event_id does not double-apply | Processor + memory/DynamoDB conditional CreateEvent + HTTP 200 duplicate. Local only. |
| Public recruiter demo | PLANNED | Unauthenticated synthetic scenarios through real logic | Spec only |
| Self-dogfooding CI evidence | PLANNED | This repo's real SHA/run/deploy metadata visible in-product | Spec only |

## Testing and reliability

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Go table-driven unit tests | TESTED | Engines covered | Domain/repo/API tests passing locally. Engine packages not started. |
| Integration tests | TESTED | Repository/event boundaries | DynamoDB-compatible fake + event processor tests. Not DynamoDB Local / LocalStack. |
| Cypress E2E | PLANNED | Recruiter demo paths | None |
| k6 performance test | PLANNED | Documented run + limitations | None |
| Dead letter handling | TESTED | SQS DLQ exists and is observable | In-process Processor DLQ for malformed/poison/unsupported schema. SQS DLQ not created. |

## Explicit non-claims

Do not list these on a CV for this project:

- Kubernetes / EKS
- Python / FastAPI / React / Next.js / Azure / Foundry / RAG / LLMs
- PostgreSQL as the primary datastore
- "Scientifically proven causality"
- "Supports millions of users"
- Professional production on-call experience
- NAT Gateway, RDS, or always-on clusters

## Update rule

Update this matrix at the end of every phase. Promote a row only when the evidence column can name a concrete artifact: test name, workflow run URL, Terraform apply, public URL, or CloudWatch log group.
