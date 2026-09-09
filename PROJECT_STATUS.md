# Project status

**Project:** CloudOps Release Intelligence
**Phase:** 11–15 (first AWS bootstrap applied; compute and public demo not created)
**Status:** NOT COMPLETE — AWS foundation inspected; product runtime not live
**Complete:** No

This file is the source of truth for what exists versus what is planned. It must not mark the project COMPLETE until the public product, AWS infrastructure, CI/CD, observability, and live verification criteria in the master build prompt are actually proven.

## Current truth

Locally TESTED (unchanged product engines plus new gates):

- Go, Gin, Angular 21, TypeScript, RxJS, SCSS
- DynamoDB adapter (in-process fake)
- Release risk, health comparison, correlation, graph, OPA/Rego, rollback, timeline, replay
- Event ingest / idempotency / local DLQ
- Cypress recruiter path (local)
- Angular ESLint (`npx ng lint` passes)
- Demo reset of synthetic Northstar only
- Operational evidence store (live-only)

LIVE VERIFIED (first Terraform bootstrap, 2026-09-09, account `912415493331`, region `eu-west-1`):

- Terraform apply of the intended first plan (`34 added, 0 changed, 0 destroyed`)
- S3 web bucket `cloudops-prod-web-7be25877` and raw bucket `cloudops-prod-raw-7be25877`: public access blocked, AES256 encryption, raw 90-day lifecycle
- CloudFront distribution `E2220NQVG6GU75` (`d34fwrlm14h6js.cloudfront.net`) with OAC to the web bucket (origin is empty; Angular is not uploaded; this is not a public recruiter demo)
- DynamoDB table `cloudops-prod-main` exists (`PAY_PER_REQUEST`, GSI1, GSI2, ACTIVE). No application items have been written.
- ECR repository `cloudops-prod-api` exists (`IMMUTABLE` tags, scan on push). No image has been pushed.
- EventBridge custom bus `cloudops-prod-release-events`, rule `cloudops-prod-analysis`, target = analysis SQS queue. No real events have been processed.
- SQS `cloudops-prod-analysis` + DLQ `cloudops-prod-analysis-dlq` with `maxReceiveCount=3`. No worker consumes the queue.
- CloudWatch log groups `/aws/lambda/cloudops-prod-api`, `/aws/lambda/cloudops-prod-worker`, `/aws/apigateway/cloudops-prod-http` with 14-day retention, plus DLQ alarm `cloudops-prod-analysis-dlq`. No application logs yet.
- AWS Budget `cloudops-prod-monthly` ($10 MONTHLY limit; $5 and $10 ACTUAL notifications). Subscriber email is not committed.
- GitHub OIDC provider `token.actions.githubusercontent.com` plus IAM roles `cloudops-prod-github-deploy` and `cloudops-prod-github-plan` (trust scoped to `Adellaghmari/cloudops-release-intelligence`). Actions have not assumed these roles.
- Lambda IAM roles `cloudops-prod-api` and `cloudops-prod-worker` exist. No Lambda functions exist.

NOT LIVE VERIFIED:

- Go workload on Lambda
- API Gateway / public API
- Public Angular recruiter demo (CloudFront origin is empty)
- Real EventBridge → SQS → worker processing
- Linux runtime on AWS
- X-Ray traces
- GitHub Actions deployment via OIDC
- DynamoDB items written by the application
- GitHub Actions `terraform apply` (must stay disabled while state is local)

## Implemented in this batch

- Angular ESLint (angular-eslint 21 + ESLint 9 flat config). Real `ng lint` run: 4 auto-fixable issues fixed, then clean.
- Multi-stage Linux Dockerfile (build → Lambda `provided.al2023` + distroless HTTP)
- Graceful shutdown runtime + unit test; CI scripts for SIGTERM / image inspect
- Terraform for the approved `eu-west-1` serverless stack — **first bootstrap applied**
- GitHub Actions Ubuntu CI/CD files (OIDC deploy gated on `AWS_DEPLOY_ROLE_ARN`)
- `POST /api/v1/demo/reset` and `GET /api/v1/status`
- Lambda HTTP adapter + SQS worker (code only; functions not created)
- EventBridge and S3 raw-evidence adapters
- Frontend System Status + LIVE PROJECT DATA / SYNTHETIC DEMO labels

## Test status

| Suite | Status |
| --- | --- |
| Go unit tests | Local quality gate required after this batch |
| Go integration tests | TESTED against DynamoDB-compatible fake (not DynamoDB Local) |
| Angular lint | PASSING locally |
| Angular unit tests | PASSING locally |
| Angular production build | PASSING locally |
| Cypress E2E | TESTED locally (recruiter path; not public) |
| Terraform validate | PASSING locally; first apply succeeded |
| k6 performance | NOT STARTED |

## Deployment status

| Environment | Status |
| --- | --- |
| Local | IMPLEMENTED (in-memory) |
| Staging | NOT USED (portfolio is local → prod) |
| Production | FIRST BOOTSTRAP ONLY (no Lambda / API Gateway) |
| Public frontend | NOT STARTED (CloudFront exists; SPA not uploaded) |
| Public API | NOT STARTED |

## Known limitations

- Persistence is process-local unless `APP_STORE=dynamodb` and the API writes to `cloudops-prod-main`. The table exists; the app has not written items.
- DynamoDB adapter is TESTED against an in-process compatible fake. Docker/Java are unavailable on the Windows workstation, so DynamoDB Local was not run.
- EventBridge/SQS resources exist. Delivery, worker consume, and DLQ poison-path behaviour are not AWS verified.
- CloudFront serves the SPA origin only. The public API URL will be API Gateway (avoids a Terraform cycle with Lambda CORS). SPA objects are not uploaded.
- Terraform state is **local** (`infra/terraform.tfstate`, gitignored). Do not enable GitHub terraform apply until a remote backend exists.
- Cosign keyless signing is wired in CD after an image exists. Not proven until that job runs.
- Linux container behaviour is proven in GitHub Ubuntu CI, not on this laptop. No image is in ECR.
- GitHub deploy IAM policy still uses `Resource: "*"` for several runtime APIs. That is documented debt, not least-privilege LIVE VERIFIED.

## Remaining work before a complete live product

1. Build and push the Linux Lambda image to ECR by digest (not started)
2. Second Terraform apply with `api_image_uri=<ecr>@sha256:…` to create Lambda + HTTP API (not started)
3. Upload the Angular production build to the web bucket / CloudFront
4. Public Cypress + six Northstar scenarios through the deployed API
5. Real EventBridge/SQS/DLQ/X-Ray/OIDC assume-role evidence
6. Remote Terraform state before any GitHub apply

Phase 16 is not started.

## Phase tracker

| Phase | Name | Status |
| --- | --- | --- |
| 0 | Architecture and repository foundation | COMPLETE |
| 1 | Core domain models + Go API | COMPLETE |
| 2 | DynamoDB + event persistence | COMPLETE (local adapter; AWS table exists, empty) |
| 3 | Angular operations console | COMPLETE (local; not on CloudFront) |
| 4 | Release Risk Engine | COMPLETE |
| 5 | Event architecture foundation (local ports) | COMPLETE (local; AWS bus/queue exist, no processing) |
| 6 | Deployment Health Comparator + correlation | COMPLETE (local) |
| 7 | Change Impact Graph | COMPLETE (local) |
| 8 | OPA / Release Policy Gate | COMPLETE (local) |
| 9 | Rollback Readiness | COMPLETE (local) |
| 10 | Release Timeline + Release Replay | COMPLETE (local) |
| 11 | Real GitHub + self dogfooding | IMPLEMENTED in repo; LIVE VERIFIED pending green Actions + live evidence rows |
| 12 | AWS infrastructure with Terraform | FIRST BOOTSTRAP LIVE VERIFIED; Lambda/API Gateway not created |
| 13 | CI/CD + OIDC + DevSecOps | IMPLEMENTED in repo; OIDC roles exist; Actions have not assumed them |
| 14 | CloudWatch / X-Ray | Log groups + DLQ alarm LIVE VERIFIED; no app logs or X-Ray traces |
| 15 | Public synthetic recruiter demo | IMPLEMENTED locally; not publicly deployed |
| 16 | Security / reliability / performance hardening | NOT STARTED |
| 17 | Full live verification | NOT STARTED |
| 18 | GitHub repository finalization | IN PROGRESS |
