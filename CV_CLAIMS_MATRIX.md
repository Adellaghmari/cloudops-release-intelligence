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
| AWS Lambda | IMPLEMENTED | Terraform-applied functions serving traffic | `cmd/api` Lambda adapter + `cmd/worker` + Terraform image functions. Not applied. |
| Amazon API Gateway | IMPLEMENTED | Public HTTPS API | HTTP API Terraform. Not applied. |
| Amazon DynamoDB | TESTED | Production table with real items | Adapter + single-table mapping + conditional EventID writes + reset/evidence tests against a local DynamoDB-compatible fake. No production table. |
| Amazon S3 | IMPLEMENTED | Frontend origin and/or raw event objects | Terraform web + raw buckets; Go raw-evidence writer. Not applied. |
| Amazon CloudFront | IMPLEMENTED | Public site URL | Terraform distribution + OAC. Not applied. |
| Amazon EventBridge | IMPLEMENTED | Custom bus receiving real events | Go PutEvents adapter + Terraform bus/rule. Not AWS verified. |
| Amazon SQS | IMPLEMENTED | Worker consuming analysis queue + DLQ | Worker Lambda + Terraform queue/DLQ (`maxReceiveCount=3`). Not AWS verified. |
| Amazon ECR | IMPLEMENTED | Immutable image digest deployed to Lambda | Terraform repo (`IMMUTABLE` tags). No pushed digest. |
| Amazon CloudWatch | IMPLEMENTED | Production logs/metrics in the account | 14-day log groups + DLQ alarm in Terraform. No live logs. |
| AWS X-Ray | IMPLEMENTED | Traces for API and worker | Lambda `tracing_config = Active`. No real trace yet. |
| AWS IAM least privilege | IMPLEMENTED | Applied roles with scoped policies | Terraform IAM for Lambda + GitHub OIDC. Not applied. |
| GitHub OIDC to AWS | IMPLEMENTED | Workflows assume role without static keys | Trust policies + `cd.yml`. Not LIVE VERIFIED until `sts get-caller-identity` succeeds in Actions. |
| AWS Secrets Manager | PLANNED / MAY OMIT | Only if a real secret is required | Intentionally avoided unless necessary |

## Infrastructure, CI, DevSecOps

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Terraform | IMPLEMENTED | `plan`/`apply` of the intended stack | `infra/` written, fmt/validate in CI. Not applied. |
| GitHub Actions CI | IMPLEMENTED | Successful workflow runs on the repo | `.github/workflows/ci.yml` on Ubuntu. No green run until the GitHub remote exists. |
| GitHub Actions CD | IMPLEMENTED | Staging/prod deploy from main | `cd.yml` OIDC + digest push. Gated on `AWS_DEPLOY_ROLE_ARN`. |
| Trivy scanning | IMPLEMENTED | Workflow step that can fail the build | fs + config + image in CI. No recorded result yet. |
| Syft SBOM | IMPLEMENTED | Generated artifact attached to release/build | `anchore/sbom-action` in CI. Not generated locally. |
| Cosign keyless signing | IMPLEMENTED | Signed ECR image verified in CI | `cosign sign` in CD after push. Not verified. |
| Linux | IMPLEMENTED | Real Ubuntu CI + Linux OCI/Lambda runtime | Dockerfile, scripts, Actions `ubuntu-latest`. Not LIVE VERIFIED until Lambda serves traffic. Workstation has no Docker. |
| Docker / OCI containers | IMPLEMENTED | Multi-stage image, non-root, digest identity | Dockerfile + CI inspect. Not pushed to ECR. |
| Open Policy Agent / Rego | TESTED | Policies execute against real release input | `policies/release_gate.rego` evaluated in-process via opa/v1/rego. Local only. |

## Product capabilities

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Release Risk Engine | TESTED | Persisted score + contributing signals from real logic | `internal/risk` + GET /releases/:id/risk. Local only. |
| Deployment Health Comparator | TESTED | Pre/post windows, raw metrics, verdict | `internal/health` + GET /releases/:id/health. Local only. Not CloudWatch verified. |
| Release correlation | TESTED | Explicit evidence object, not "causation" | LIKELY/POSSIBLE/NO_CLEAR/INSUFFICIENT_DATA with reasons. Never PROVEN_CAUSE. Local only. |
| Change Impact Graph | TESTED | BFS/DFS blast radius from persisted edges | `internal/graph` + GET /releases/:id/impact + SVG. Language is potential impact. Local only. |
| Release Policy Gate | TESTED | Versioned policy eval with PASS/WARN/BLOCK/MANUAL | Real Rego tests + GET /releases/:id/policy. Fail closed. Local only. |
| Rollback Readiness | TESTED | READY/PARTIAL/NOT READY/UNKNOWN + missing prereqs | `internal/rollback` + GET /releases/:id/rollback. No auto rollback. Local only. |
| Release Replay | TESTED | Deterministic diff of two persisted releases | GET /replay + Angular Replay page. Local only. |
| Release evidence timeline | TESTED | Chronological events from storage | GET /releases/:id/timeline from persisted events. Local only. |
| Event-driven analysis | TESTED | Ingest → EventBridge → SQS → worker, not in-request theatre | Local processor still TESTED. AWS path IMPLEMENTED (adapters + Terraform), not AWS verified. |
| Idempotent ingestion | TESTED | Duplicate event_id does not double-apply | Processor + memory/DynamoDB conditional CreateEvent + HTTP 200 duplicate. Local only. |
| Public recruiter demo | IMPLEMENTED | Unauthenticated synthetic scenarios through real logic | Local console + `POST /demo/reset`. Not publicly deployed. |
| Self-dogfooding CI evidence | IMPLEMENTED | This repo's real SHA/run/deploy metadata visible in-product | `GET /status` + live evidence ingest. No live rows until a real pipeline event is posted. |

## Testing and reliability

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Go table-driven unit tests | TESTED | Engines covered | Domain/repo/API tests passing locally. Engine packages not started. |
| Integration tests | TESTED | Repository/event boundaries | DynamoDB-compatible fake + event processor tests. Not DynamoDB Local / LocalStack. |
| Cypress E2E | TESTED | Recruiter demo paths | `frontend/cypress/e2e/recruiter.cy.ts` against local API+Angular. Not public. |
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
