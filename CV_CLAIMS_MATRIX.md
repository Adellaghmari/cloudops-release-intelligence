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
| AWS Lambda | IMPLEMENTED | Terraform-applied functions serving traffic | `cmd/api` uses `aws-lambda-go` + `httpadapter`; `cmd/worker` is a genuine SQS handler with `ReportBatchItemFailures`. Functions Active on digest `08ab5782…` but init fails: image lacked `/var/task/bootstrap` (`Runtime.InvalidEntrypoint`). Corrected bootstrap dispatcher is in source; not LIVE VERIFIED until new digest is applied. |
| Amazon API Gateway | IMPLEMENTED | Public HTTPS API | HTTP API exists (`8kci5uht3d` / `cloudops-prod-http`), but requests return `Internal Server Error` while Lambda init fails. Not LIVE VERIFIED. |
| Amazon DynamoDB | TESTED | Production table with real items | Adapter TESTED against a local DynamoDB-compatible fake. Table `cloudops-prod-main` exists in `eu-west-1` (`PAY_PER_REQUEST`, GSI1, GSI2, ACTIVE). No application PutItem yet. Not LIVE VERIFIED for data. |
| Amazon S3 | LIVE VERIFIED | Frontend origin and/or raw event objects | Buckets `cloudops-prod-web-7be25877` and `cloudops-prod-raw-7be25877`: BlockPublicAcls/IgnorePublicAcls/BlockPublicPolicy/RestrictPublicBuckets all true; AES256; raw lifecycle expire 90 days. SPA/raw objects not uploaded. |
| Amazon CloudFront | LIVE VERIFIED | Distribution + OAC in the account | Distribution `E2220NQVG6GU75` (`d34fwrlm14h6js.cloudfront.net`) enabled; origin `s3-web` uses OAC `E3TW0G8ABTQWD8`. Angular not uploaded. Not a public recruiter demo. |
| Amazon EventBridge | IMPLEMENTED | Custom bus receiving real events | Bus `cloudops-prod-release-events`, enabled rule `cloudops-prod-analysis`, target = `cloudops-prod-analysis` SQS. No successful processing yet. Not LIVE VERIFIED for processing. |
| Amazon SQS | IMPLEMENTED | Worker consuming analysis queue + DLQ | Queues exist; visibility 360; redrive `maxReceiveCount=3`; event source mapping Enabled with `ReportBatchItemFailures`. Worker has not successfully processed. Not LIVE VERIFIED for consume. |
| Amazon ECR | LIVE VERIFIED | Immutable image exists in the repository | Repo `cloudops-prod-api`: `IMMUTABLE` + scan on push. Historical broken image `sha-c22171c…` / `sha256:08ab5782…`. Corrected image pending push. |
| Amazon CloudWatch | LIVE VERIFIED | Intended log groups and DLQ alarm exist | Groups `/aws/lambda/cloudops-prod-api`, `/aws/lambda/cloudops-prod-worker`, `/aws/apigateway/cloudops-prod-http` retention 14 days; alarm `cloudops-prod-analysis-dlq`. Init logs show `Runtime.InvalidEntrypoint` on the old image. |
| AWS X-Ray | IMPLEMENTED | Traces for API and worker | Lambda `tracing_config = Active`. Failed-init trace IDs exist; successful app traces not LIVE VERIFIED. |
| AWS IAM least privilege | IMPLEMENTED | Applied roles with scoped policies | Roles applied: `cloudops-prod-api`, `cloudops-prod-worker`, `cloudops-prod-github-deploy`, `cloudops-prod-github-plan`. GitHub deploy policy still uses `Resource: "*"` for several APIs. Not LIVE VERIFIED as least privilege. |
| GitHub OIDC to AWS | LIVE VERIFIED | Workflows assume role without static keys | CD run [34370805611](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805611): `sts get-caller-identity` = `arn:aws:sts::912415493331:assumed-role/cloudops-prod-github-deploy/GitHubActions`. No `AWS_ACCESS_KEY_ID`. Trust `sub` includes GitHub owner/repo IDs. |
| AWS Secrets Manager | PLANNED / MAY OMIT | Only if a real secret is required | Intentionally avoided unless necessary |

## Infrastructure, CI, DevSecOps

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Terraform | LIVE VERIFIED | `plan`/`apply` of the intended stack | First bootstrap apply 2026-09-09: `34 added, 0 changed, 0 destroyed` in `eu-west-1`. Retry compute apply 2026-09-09: `6 added, 1 changed, 0 destroyed` (Lambda/API integration created; runtime init failure prevents workload LIVE VERIFICATION). Local gitignored state; GitHub apply stays disabled until remote state exists. |
| GitHub Actions CI | LIVE VERIFIED | Successful workflow runs on the repo | [ci run 34370805926](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805926) on `ubuntu-24.04`: go-quality, angular-quality, terraform-validate, trivy-and-sbom, cypress-local, linux-container all success. |
| GitHub Actions CD | LIVE VERIFIED | OIDC + immutable ECR push from main | [cd run 34370805611](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805611) assumed deploy role, pushed `sha-c22171c…`. GitHub `terraform apply` remains disabled (local state). |
| Trivy scanning | LIVE VERIFIED | Workflow step that can fail the build | CI fs + Terraform config + HTTP/Lambda image CRITICAL gate on run 34370805926. Documented IaC ignore: AWS-0011 (no WAF), AWS-0132 (SSE-S3 not CMK). |
| Syft SBOM | LIVE VERIFIED | Generated artifact attached to release/build | `anchore/sbom-action` uploaded `cloudops-sbom.spdx.json` (CI) and `cloudops-lambda-sbom.spdx.json` (CD). |
| Cosign keyless signing | IMPLEMENTED | Signed ECR image verified in CI | Prior verify failed (`ecr:GetDownloadUrlForLayer`). Permission now present; re-verify pending corrected image. Not LIVE VERIFIED. |
| Linux | IMPLEMENTED | Real Ubuntu CI + Linux OCI image | Runner `Ubuntu 24.04` x86_64. HTTP image non-root + SIGTERM. Lambda `provided.al2023` + `/var/task/bootstrap` in source. AWS Lambda runtime execution not LIVE VERIFIED until corrected digest is applied. |
| Docker / OCI containers | LIVE VERIFIED | Multi-stage image, non-root, digest identity | ECR historical `cloudops-prod-api@sha256:08ab5782…` (broken entrypoint). Corrected digest pending. |
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
| Event-driven analysis | TESTED | Ingest → EventBridge → SQS → worker, not in-request theatre | Local processor still TESTED. AWS bus/rule/queue/DLQ exist; worker and event source mapping do not. Not LIVE VERIFIED. |
| Idempotent ingestion | TESTED | Duplicate event_id does not double-apply | Processor + memory/DynamoDB conditional CreateEvent + HTTP 200 duplicate. Local only. |
| Public recruiter demo | IMPLEMENTED | Unauthenticated synthetic scenarios through real logic | Local console + `POST /demo/reset`. Not publicly deployed. |
| Self-dogfooding CI evidence | IMPLEMENTED | This repo's real SHA/run/deploy metadata visible in-product | Real metadata now exists: git SHA `c22171c3b6b8bd3946a2972f16283eac629f5ce3`, CI run `34370805926`, CD run `34370805611`, image digest `sha256:08ab57825cc1f43a1527ad474f20a146e8f55dacb2c49323f42601129efa1a53`. Not ingested because the public API does not exist. Keep LIVE PROJECT DATA separate from SYNTHETIC DEMO. |

## Testing and reliability

| Claim | Status | Evidence required | Current evidence |
| --- | --- | --- | --- |
| Go table-driven unit tests | TESTED | Engines covered | Domain/repo/API tests passing locally. Engine packages not started. |
| Integration tests | TESTED | Repository/event boundaries | DynamoDB-compatible fake + event processor tests. Not DynamoDB Local / LocalStack. |
| Cypress E2E | TESTED | Recruiter demo paths | `frontend/cypress/e2e/recruiter.cy.ts` against local API+Angular. Not public. |
| k6 performance test | PLANNED | Documented run + limitations | None |
| Dead letter handling | TESTED | SQS DLQ exists and is observable | In-process Processor DLQ TESTED. AWS queue `cloudops-prod-analysis-dlq` and alarm exist; poison path not exercised by a worker. |

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
