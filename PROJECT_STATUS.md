# Project status

**Project:** CloudOps Release Intelligence
**Phase:** 11–15 preparation (code complete, AWS not applied)
**Status:** IMPLEMENTATION IN REPOSITORY — NOTHING LIVE VERIFIED ON AWS
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

NOT LIVE VERIFIED:

- No public URL
- No Terraform apply
- No real DynamoDB / EventBridge / SQS / Lambda / CloudFront / ECR
- No GitHub Actions run on the real repository until the remote exists and CI is green
- No Linux container proof on this Windows workstation (Docker is not installed here)

## Implemented in this batch

- Angular ESLint (angular-eslint 21 + ESLint 9 flat config). Real `ng lint` run: 4 auto-fixable issues fixed, then clean.
- Multi-stage Linux Dockerfile (build → Lambda `provided.al2023` + distroless HTTP)
- Graceful shutdown runtime + unit test; CI scripts for SIGTERM / image inspect
- Terraform for the approved `eu-west-1` serverless stack (not applied)
- GitHub Actions Ubuntu CI/CD files (OIDC deploy gated on `AWS_DEPLOY_ROLE_ARN`)
- `POST /api/v1/demo/reset` and `GET /api/v1/status`
- Lambda HTTP adapter + SQS worker
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
| Terraform validate | Written; run in CI / local when Terraform CLI is available |
| k6 performance | NOT STARTED |

## Deployment status

| Environment | Status |
| --- | --- |
| Local | IMPLEMENTED (in-memory) |
| Staging | NOT USED (portfolio is local → prod) |
| Production | NOT STARTED (no apply) |
| Public frontend | NOT STARTED |
| Public API | NOT STARTED |

## Known limitations

- Persistence is process-local unless `APP_STORE=dynamodb` and a real table exist.
- DynamoDB adapter is TESTED against an in-process compatible fake. Docker/Java are unavailable on the Windows workstation, so DynamoDB Local was not run.
- EventBridge/SQS adapters are code + Terraform only. Not AWS verified.
- CloudFront serves the SPA only. The public API URL is API Gateway (avoids a Terraform cycle with Lambda CORS).
- First Terraform apply uses **local state**. See `docs/TERRAFORM_STATE.md` and `docs/COST_CHECKPOINT.md`.
- Cosign keyless signing is wired in CD after an image exists. Not proven until that job runs.
- Linux container behaviour is proven in GitHub Ubuntu CI, not on this laptop.
- Module path is normalized to the intended GitHub repository when the remote exists.

## Remaining work before LIVE VERIFIED

1. Create/push the real GitHub repository (this session, if `gh` succeeds)
2. User AWS login
3. Cost checkpoint approval
4. `terraform apply` (no image → no Lambda yet)
5. CI image push + second apply with digest
6. Public Cypress + six Northstar scenarios through the deployed API
7. Real EventBridge/SQS/DLQ/X-Ray/OIDC evidence

Phase 16 is not started.

## Phase tracker

| Phase | Name | Status |
| --- | --- | --- |
| 0 | Architecture and repository foundation | COMPLETE |
| 1 | Core domain models + Go API | COMPLETE |
| 2 | DynamoDB + event persistence | COMPLETE |
| 3 | Angular operations console | COMPLETE |
| 4 | Release Risk Engine | COMPLETE |
| 5 | Event architecture foundation (local ports) | COMPLETE (local; EventBridge/SQS not AWS verified) |
| 6 | Deployment Health Comparator + correlation | COMPLETE (local) |
| 7 | Change Impact Graph | COMPLETE (local) |
| 8 | OPA / Release Policy Gate | COMPLETE (local) |
| 9 | Rollback Readiness | COMPLETE (local) |
| 10 | Release Timeline + Release Replay | COMPLETE (local) |
| 11 | Real GitHub + self dogfooding | IMPLEMENTED in repo; LIVE VERIFIED pending remote + green Actions |
| 12 | AWS infrastructure with Terraform | IMPLEMENTED in repo; apply not run |
| 13 | CI/CD + OIDC + DevSecOps | IMPLEMENTED in repo; not executed on GitHub yet |
| 14 | CloudWatch / X-Ray | IMPLEMENTED in Terraform; no traces yet |
| 15 | Public synthetic recruiter demo | IMPLEMENTED locally; not publicly deployed |
| 16 | Security / reliability / performance hardening | NOT STARTED |
| 17 | Full live verification | NOT STARTED |
| 18 | GitHub repository finalization | IN PROGRESS |
