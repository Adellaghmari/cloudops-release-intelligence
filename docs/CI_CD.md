# CI/CD

GitHub Actions is the only pipeline. AWS access is **OIDC assumed roles**, not static `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`.

## Version identity

Every deployable is identified by:

- `git_sha` (full SHA)
- `version` = `0.1.0+${sha:0:7}` until tagged releases exist
- container `image_digest` for Lambda
- frontend object prefix `web/${sha}/`

`latest` may exist as a mutable convenience tag in ECR but **must never be the production identity**. Lambda is updated to a digest.

## Workflows (target)

### `pr.yml` — pull requests

- Go fmt (`gofmt` / `gofumpt` if adopted), `go vet`, `golangci-lint`
- `go test ./...`
- Angular lint + unit tests
- Terraform `fmt -check`, `init -backend=false`, `validate`
- Trivy filesystem + Terraform config scan
- Frontend production build (no deploy)

### `main.yml` — main branch

All PR checks, plus:

- Build Go binaries / container image tagged with SHA
- Trivy image scan (fail on CRITICAL)
- Syft SBOM attached to the workflow
- Cosign keyless sign of the image via GitHub OIDC
- Push to ECR
- Terraform plan
- Terraform apply to prod (portfolio single environment) **after** plan
- Smoke: `GET /api/v1/health` and `GET /api/v1/releases` against the public API
- Frontend build uploaded to the SHA prefix; CloudFront invalidation of `index.html` only

Promotion gate in a later hardening phase can require the policy engine PASS on the live `cloudops-api` release. v1 of the pipeline still deploys on green CI; the **product** policy gate is shown in-app, not as a circular blocker that prevents the demo from shipping.

### `e2e.yml`

Cypress against a deployed environment or a local stack in CI. Prefer local stack for PR, smoke against prod on main.

## OIDC

Trust: `token.actions.githubusercontent.com` → IAM role `github-cloudops-deploy`.

Conditions:

- `sub` limited to this repository and `ref:refs/heads/main` for apply
- PR jobs that only need to validate Terraform use a **plan-only** role with `ReadOnly` + `iam:PassRole` denied
- No `*:*` on deploy role. Scope to this project's Lambda, S3, DynamoDB, ECR, API Gateway, CloudFront, EventBridge, SQS, CloudWatch.

## Dogfood

A final job `publish-release-evidence` posts this pipeline's SHA, workflow run ID, test status, duration, image digest, and Trivy summary into `POST /api/v1/events` using the deploy role. That is the **live** strip in the UI.

Synthetic Northstar scenarios remain available when GitHub is down or the recruiter has no interest in this repo's plumbing.

## What must not happen

- Long-lived AWS keys in GitHub Secrets
- `--no-verify` culture
- Deploying an unsigned image if Cosign is wired (if Cosign is deferred, the CV matrix stays PLANNED)
- Mutable `latest` as the only recorded version
