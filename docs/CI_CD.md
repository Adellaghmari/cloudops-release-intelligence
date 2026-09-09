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

Path-filtered:

- Backend paths → build/push Lambda image (OIDC deploy role), update functions
- Frontend / `cd.yml` → Angular build, S3 sync, CloudFront invalidation **with GetInvalidation waiter**
- Docs-only changes must not mint a new backend image

### Terraform (`terraform.yml`)

| Mode | When | Role | Gate |
| --- | --- | --- | --- |
| fmt/validate | push/PR/dispatch | none | always |
| plan | push/PR (same-repo)/dispatch | `cloudops-prod-github-plan` | `TERRAFORM_REMOTE_STATE_READY=true` |
| apply | **workflow_dispatch only** | `cloudops-prod-github-deploy` + `environment:prod` | `apply=true` **and** `ENABLE_TERRAFORM_APPLY=true` |

Apply generates a same-job plan (`-out=tfplan`). If the plan is not a no-op (`-detailed-exitcode` ≠ 0), apply is refused. Successful controlled no-op apply was proven in Phase 16B Part 3; the apply gate is then left **false**.

Fork PRs never receive AWS credentials. No `pull_request_target`.

## OIDC

Immutable subjects use owner/repo IDs:

`repo:Adellaghmari@179922674/cloudops-release-intelligence@1361976389:…`

- Plan: `main` + `pull_request`
- Deploy/apply: `main` + `environment:prod`

## Final posture

- `TERRAFORM_REMOTE_STATE_READY=true`
- `ENABLE_TERRAFORM_APPLY=false`
