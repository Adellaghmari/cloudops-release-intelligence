# CI/CD

GitHub Actions is the only pipeline. AWS access is **OIDC assumed roles**, not static `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`.

## Version identity

Every deployable is identified by:

- `git_sha` (full SHA)
- container `image_digest` for Lambda (immutable; never `latest` as production identity)
- frontend objects in the private web bucket behind CloudFront OAC

Current production Lambda digest:

`sha256:ef3778d5af9e80d155610d5ffb0a889e509a4ba3da3fee2ac6878c6c5287ade5`

## Workflows

### CI (`ci.yml`)

Go test/vet, Angular lint/unit/build, supply-chain checks as configured on main.

### CD (`cd.yml`)

`cd.yml` is the only production mutation path. It is manual dispatch from an
exact `main` SHA. A push to `main` runs CI but does not deploy.

The ordered release is:

1. Build, scan, generate an SBOM for, and publish one immutable Lambda candidate.
2. Resolve its ECR digest and use that digest as `api_image_uri`.
3. Save a Terraform plan and inspect its JSON actions.
4. Stop at the protected `prod` environment for human approval.
5. Verify plan metadata and apply the exact saved plan with
   `cloudops-prod-terraform-apply`.
6. Verify Terraform deployed the candidate digest, then deploy the Angular build
   to private S3 and wait for CloudFront invalidation.
7. Assume the bounded evidence producer role and SigV4-sign deployment evidence.
8. Verify public reads, anonymous write denial, the frontend asset, and zero
   Terraform drift.

The candidate job assumes the existing main-only deploy role with an inline
session policy limited to ECR. Terraform is the sole owner of Lambda image
updates. The later application job has read-only Lambda access and cannot update
the functions.

### Terraform (`terraform.yml`)

| Mode | When | Role | Gate |
| --- | --- | --- | --- |
| fmt/validate | push/PR/dispatch | none | always |
| preview | push/PR from the same repository | `cloudops-prod-github-plan` | `TERRAFORM_REMOTE_STATE_READY=true` |
| drift | manual dispatch from `main` | `cloudops-prod-github-plan` | requires zero change |
| release apply | protected `cd.yml` only | `cloudops-prod-terraform-apply` | saved plan, `ENABLE_TERRAFORM_APPLY=true`, `environment:prod` |

Release plans may contain intentional creates and updates. A tested plan guard
blocks every delete, replacement, and unknown action before approval. Drift mode
remains a separate zero-change check.

The short-lived release artifact contains the saved plan, summary, and metadata
binding the source SHA, candidate digest, Terraform version, workflow run,
workflow identity, Terraform configuration tree, backend bucket, state key, and
plan checksum. The apply job verifies every field and applies the saved plan
without recalculation.

Fork PRs never receive AWS credentials. No `pull_request_target`.

## First bootstrap versus normal release

**FIRST BOOTSTRAP** used the authenticated human `adel-admin` session once to
create the bounded production Terraform role and its remote state. That account
bootstrap is complete and is not part of normal releases.

**NORMAL RELEASE** uses only GitHub OIDC roles and temporary credentials. It
requires the protected `prod` environment reviewer after the plan is available.
The repository apply kill switch remains an additional explicit maintenance
control.

## OIDC

Immutable subjects use owner/repo IDs:

`repo:Adellaghmari@179922674/cloudops-release-intelligence@1361976389:…`

- Plan: `main` + `pull_request`
- Candidate and application deploy: `main`
- Terraform apply: `environment:prod`, with a main-only GitHub environment rule
- Evidence producer: exact `main` subject; only `execute-api:Invoke` on the canonical event-ingest route

## Final posture

- `TERRAFORM_REMOTE_STATE_READY=true`
- `ENABLE_TERRAFORM_APPLY=false` outside an explicitly approved release window
- Account bootstrap: **LIVE VERIFIED, ZERO DRIFT**
- Protected release workflow: **IMPLEMENTED, NOT YET EXECUTED**
- Secure producer workflow: **CONFIGURED LOCALLY, NOT YET DEPLOYED OR LIVE VERIFIED**
