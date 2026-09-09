# Project status

**Project:** CloudOps Release Intelligence
**Phase:** 11–15 (compute applied; Lambda entrypoint fix in flight; runtime not live)
**Status:** NOT COMPLETE — foundation and CI/OIDC/ECR proven; product runtime not live
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
- ECR repository `cloudops-prod-api` exists (`IMMUTABLE` tags, scan on push). Previously deployed image: tag `sha-c22171c3b6b8bd3946a2972f16283eac629f5ce3`, digest `sha256:08ab57825cc1f43a1527ad474f20a146e8f55dacb2c49323f42601129efa1a53` (linux/amd64). That image lacked `/var/task/bootstrap`, so Lambda init failed with `Runtime.InvalidEntrypoint`. A corrected image (bootstrap dispatcher) must be pushed and applied via Terraform before runtime can be LIVE VERIFIED.
- EventBridge custom bus `cloudops-prod-release-events`, rule `cloudops-prod-analysis`, target = analysis SQS queue. No successful worker processing yet.
- SQS `cloudops-prod-analysis` + DLQ `cloudops-prod-analysis-dlq` with `maxReceiveCount=3`, visibility 360s. Event source mapping exists and is Enabled; worker has not successfully processed messages.
- CloudWatch log groups `/aws/lambda/cloudops-prod-api`, `/aws/lambda/cloudops-prod-worker`, `/aws/apigateway/cloudops-prod-http` with 14-day retention, plus DLQ alarm `cloudops-prod-analysis-dlq`. Application init error logs for `cloudops-prod-api` show `Runtime.InvalidEntrypoint` on the old image.
- AWS Budget `cloudops-prod-monthly` ($10 MONTHLY limit; $5 and $10 ACTUAL notifications). Subscriber email is not committed.
- GitHub OIDC: Actions assumed `cloudops-prod-github-deploy`. Proof: `aws sts get-caller-identity` returned `arn:aws:sts::912415493331:assumed-role/cloudops-prod-github-deploy/GitHubActions` in [cd run 34370805611](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805611). Trust uses GitHub owner/repo numeric IDs in `sub`.
- GitHub Actions CI green on Ubuntu 24.04 for the pre-fix image: [ci run 34370805926](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805926). Post-fix CI must re-prove bootstrap + RIE.
- Linux HTTP image inspected in CI: `os=linux arch=amd64 user=65532:65532`; SIGTERM accepted. Lambda image non-root is the AWS sandbox (no Dockerfile `USER`); HTTP distroless remains the non-root proof.
- Trivy: filesystem/config gates passed on that CI run (documented IaC exceptions AWS-0011 / AWS-0132). Image CRITICAL gate passed. Syft SPDX SBOM uploaded from CI and CD.
- Lambda IAM roles `cloudops-prod-api` and `cloudops-prod-worker` exist. Lambda functions exist and are `Active` on the old digest; init fails until the corrected image is applied.

NOT LIVE VERIFIED:

- Go workload on Lambda (successful Runtime API execution)
- API Gateway / public API
- Public Angular recruiter demo (CloudFront origin is empty)
- Real EventBridge → SQS → worker processing
- Linux runtime on AWS Lambda
- Successful X-Ray application traces
- Cosign keyless verify (permission `ecr:GetDownloadUrlForLayer` now present; re-verify pending corrected image)
- DynamoDB items written by the application
- GitHub Actions `terraform apply` (must stay disabled while state is local)
- In-product LIVE PROJECT DATA rows (metadata exists in workflow artifacts; API cannot ingest yet)

## Implemented in this batch

- Angular ESLint (angular-eslint 21 + ESLint 9 flat config). Real `ng lint` run: 4 auto-fixable issues fixed, then clean.
- Multi-stage Linux Dockerfile (build → Lambda `provided.al2023` with `/var/task/bootstrap` dispatcher + distroless HTTP)
- Graceful shutdown runtime + unit test; CI scripts for SIGTERM / image inspect / Lambda RIE proof
- Terraform for the approved `eu-west-1` serverless stack — **first bootstrap applied**; compute applied; `AWS_REGION` not set as a user Lambda env var
- GitHub Actions Ubuntu CI/CD files (OIDC deploy gated on `AWS_DEPLOY_ROLE_ARN`)
- `POST /api/v1/demo/reset` and `GET /api/v1/status`
- Lambda HTTP adapter (`aws-lambda-go` + `httpadapter`) + SQS worker (`ReportBatchItemFailures`)
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
| Production | FOUNDATION + compute applied (Lambdas Active on old image; InvalidEntrypoint until corrected digest applied) |
| Public frontend | NOT STARTED (CloudFront exists; SPA not uploaded) |
| Public API | NOT STARTED |

## Known limitations

- Persistence is process-local unless `APP_STORE=dynamodb` and the API writes to `cloudops-prod-main`. The table exists; the app has not written items.
- DynamoDB adapter is TESTED against an in-process compatible fake. Docker/Java are unavailable on the Windows workstation, so DynamoDB Local was not run.
- EventBridge/SQS resources exist. Delivery, worker consume, and DLQ poison-path behaviour are not AWS verified.
- CloudFront serves the SPA origin only. The public API URL will be API Gateway (avoids a Terraform cycle with Lambda CORS). SPA objects are not uploaded.
- Terraform state is **local** (`infra/terraform.tfstate`, gitignored). Do not enable GitHub terraform apply until a remote backend exists.
- Cosign keyless signing ran in CD; verify previously failed without `ecr:GetDownloadUrlForLayer` (permission now in deploy role). Cosign verify is not LIVE VERIFIED until the corrected image run succeeds.
- Linux container behaviour is proven on GitHub Ubuntu 24.04. The broken Lambda image is in ECR; corrected image pending push. This workstation still has no Docker.
- GitHub deploy IAM policy still uses `Resource: "*"` for several runtime APIs. That is documented debt, not least-privilege LIVE VERIFIED.

## Remaining work before a complete live product

1. Push corrected Lambda image (with `/var/task/bootstrap`), review fresh Terraform plan for the new digest, then apply (do not use `aws lambda update-function-code` outside Terraform).
2. Re-verify real `/api/v1/*` responses (health/ready/services/releases) and that `APP_SEED_LOCAL=true` safely initializes SYNTHETIC Northstar data in `cloudops-prod-main`.
3. Prove the real AWS EventBridge → SQS → worker processing end-to-end (including duplicate `event_id` idempotency) and validate DLQ bounded retries.
4. Confirm Cosign sign+verify on the corrected image (keyless GitHub OIDC).
5. Upload the Angular production build to the web bucket / CloudFront (Phase 15 public frontend).
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
| 11 | Real GitHub + self dogfooding | CI LIVE VERIFIED; live evidence rows not ingested (no public API yet) |
| 12 | AWS infrastructure with Terraform | FIRST BOOTSTRAP LIVE VERIFIED; compute layer applied (Lambda/API integration created); Lambda init failing |
| 13 | CI/CD + OIDC + DevSecOps | CI + OIDC assume-role + ECR push LIVE VERIFIED; GitHub terraform apply disabled |
| 14 | CloudWatch / X-Ray | Log groups + DLQ alarm LIVE VERIFIED; Runtime.InvalidEntrypoint logs present; X-Ray trace IDs visible for failed init |
| 15 | Public synthetic recruiter demo | IMPLEMENTED locally; not publicly deployed |
| 16 | Security / reliability / performance hardening | NOT STARTED |
| 17 | Full live verification | NOT STARTED |
| 18 | GitHub repository finalization | IN PROGRESS |
