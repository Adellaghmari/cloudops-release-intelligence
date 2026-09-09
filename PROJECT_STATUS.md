# Project status

**Project:** CloudOps Release Intelligence
**Phase:** 11–15 (first AWS bootstrap live; Linux image in ECR; second compute plan saved, not applied)
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
- ECR repository `cloudops-prod-api` exists (`IMMUTABLE` tags, scan on push). Production image pushed: tag `sha-c22171c3b6b8bd3946a2972f16283eac629f5ce3`, digest `sha256:08ab57825cc1f43a1527ad474f20a146e8f55dacb2c49323f42601129efa1a53` (linux/amd64). Not deployed to Lambda.
- EventBridge custom bus `cloudops-prod-release-events`, rule `cloudops-prod-analysis`, target = analysis SQS queue. No real events have been processed.
- SQS `cloudops-prod-analysis` + DLQ `cloudops-prod-analysis-dlq` with `maxReceiveCount=3`. No worker consumes the queue.
- CloudWatch log groups `/aws/lambda/cloudops-prod-api`, `/aws/lambda/cloudops-prod-worker`, `/aws/apigateway/cloudops-prod-http` with 14-day retention, plus DLQ alarm `cloudops-prod-analysis-dlq`. No application logs yet.
- AWS Budget `cloudops-prod-monthly` ($10 MONTHLY limit; $5 and $10 ACTUAL notifications). Subscriber email is not committed.
- GitHub OIDC: Actions assumed `cloudops-prod-github-deploy`. Proof: `aws sts get-caller-identity` returned `arn:aws:sts::912415493331:assumed-role/cloudops-prod-github-deploy/GitHubActions` in [cd run 34370805611](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805611). Trust uses GitHub owner/repo numeric IDs in `sub`.
- GitHub Actions CI green on Ubuntu 24.04: [ci run 34370805926](https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805926) (go-quality, angular-quality, terraform-validate, trivy-and-sbom, cypress-local, linux-container).
- Linux HTTP image inspected in CI: `os=linux arch=amd64 user=65532:65532`; SIGTERM accepted. Lambda image `user=1000:1000`.
- Trivy: filesystem/config gates passed on that CI run (documented IaC exceptions AWS-0011 / AWS-0132). Image CRITICAL gate passed. Syft SPDX SBOM uploaded from CI and CD.
- Lambda IAM roles `cloudops-prod-api` and `cloudops-prod-worker` exist. No Lambda functions exist.

NOT LIVE VERIFIED:

- Go workload on Lambda
- API Gateway / public API
- Public Angular recruiter demo (CloudFront origin is empty)
- Real EventBridge → SQS → worker processing
- Linux runtime on AWS Lambda
- X-Ray traces
- Cosign keyless verify (sign attempted; verify denied without `ecr:GetDownloadUrlForLayer`)
- DynamoDB items written by the application
- GitHub Actions `terraform apply` (must stay disabled while state is local)
- In-product LIVE PROJECT DATA rows (metadata exists in workflow artifacts; API cannot ingest yet)

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
| Production | FOUNDATION + ECR IMAGE (no Lambda / API Gateway) |
| Public frontend | NOT STARTED (CloudFront exists; SPA not uploaded) |
| Public API | NOT STARTED |

## Known limitations

- Persistence is process-local unless `APP_STORE=dynamodb` and the API writes to `cloudops-prod-main`. The table exists; the app has not written items.
- DynamoDB adapter is TESTED against an in-process compatible fake. Docker/Java are unavailable on the Windows workstation, so DynamoDB Local was not run.
- EventBridge/SQS resources exist. Delivery, worker consume, and DLQ poison-path behaviour are not AWS verified.
- CloudFront serves the SPA origin only. The public API URL will be API Gateway (avoids a Terraform cycle with Lambda CORS). SPA objects are not uploaded.
- Terraform state is **local** (`infra/terraform.tfstate`, gitignored). Do not enable GitHub terraform apply until a remote backend exists.
- Cosign keyless signing ran in CD; verify failed because the deploy role lacked `ecr:GetDownloadUrlForLayer`. That permission is in the saved second plan. Cosign is not LIVE VERIFIED.
- Linux container behaviour is proven on GitHub Ubuntu 24.04. The Lambda image is in ECR. This workstation still has no Docker.
- GitHub deploy IAM policy still uses `Resource: "*"` for several runtime APIs. That is documented debt, not least-privilege LIVE VERIFIED.

## Remaining work before a complete live product

1. Review and apply the saved second plan `infra/tfplan-compute` (`8 add, 5 change, 0 destroy`) using digest `sha256:08ab57825cc1f43a1527ad474f20a146e8f55dacb2c49323f42601129efa1a53` (not started)
2. Upload the Angular production build to the web bucket / CloudFront
3. Public Cypress + six Northstar scenarios through the deployed API
4. Real EventBridge → SQS → worker + X-Ray evidence
5. Remote Terraform state before any GitHub apply
6. Cosign verify after `ecr:GetDownloadUrlForLayer` is applied

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
| 12 | AWS infrastructure with Terraform | FIRST BOOTSTRAP LIVE VERIFIED; second compute plan saved, not applied |
| 13 | CI/CD + OIDC + DevSecOps | CI + OIDC assume-role + ECR push LIVE VERIFIED; GitHub terraform apply disabled |
| 14 | CloudWatch / X-Ray | Log groups + DLQ alarm LIVE VERIFIED; no app logs or X-Ray traces |
| 15 | Public synthetic recruiter demo | IMPLEMENTED locally; not publicly deployed |
| 16 | Security / reliability / performance hardening | NOT STARTED |
| 17 | Full live verification | NOT STARTED |
| 18 | GitHub repository finalization | IN PROGRESS |
