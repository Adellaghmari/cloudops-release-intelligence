# Project status

**Project:** CloudOps Release Intelligence  
**Phase:** 16B Part 2 COMPLETE (IAM hardening applied; frontend waiter + GitHub Terraform PLAN proven)  
**Status:** Public demo LIVE; remote state + OIDC plan proven; GitHub apply still DISABLED  
**Complete:** No (controlled GitHub Terraform apply not authorized yet)

## LIVE VERIFIED baseline

- Frontend / health / ready HTTP 200
- Lambdas Active on `sha256:ef3778d5af9e80d155610d5ffb0a889e509a4ba3da3fee2ac6878c6c5287ade5`
- Operator `adel-admin` / `eu-west-1`

## Phase 16B Part 2 (this phase)

### Local IAM hardening apply
- Applied exact `infra/tfplan-hardening-16b-remote-1` via remote S3 backend: **2 add / 2 change / 0 destroy**
- Main remote state address count after apply: **55** (was 50; +2 role policies + related data sources)
- New remote state S3 version written; idle `.tflock` absent after apply

### IAM / OIDC verification
- Plan + deploy roles: state object Get/Put (no DeleteObject); lock object Get/Put/Delete; ListBucket prefix-restricted
- Plan trust immutable subjects `Adellaghmari@179922674` / `cloudops-release-intelligence@1361976389` for `main` + `pull_request` only (no `repo:…:*`, no `pull_request_target`)
- Deploy policy includes `cloudfront:GetInvalidation` on `E2220NQVG6GU75`
- Plan role retains AWS managed **ReadOnlyAccess** (broad refresh compatibility — **not** fully least privilege)
- No AdministratorAccess

### Frontend CD invalidation waiter
- Hard waiter (no soft-ignore) via OIDC deploy role
- Run: https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34389391377 (success)
- Assumed: `assumed-role/cloudops-prod-github-deploy/GitHubActions`
- Invalidation `IEJAC1M2X6OEE1YRXIV7RKLC8` → **Completed**

### GitHub Terraform PLAN
- `TERRAFORM_REMOTE_STATE_READY=true`
- `ENABLE_TERRAFORM_APPLY=false`
- Run: https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34389687373 (success)
- Assumed: `assumed-role/cloudops-prod-github-plan/GitHubActions`
- Remote S3 backend init + full refresh; plan result: **No changes (0/0/0)**
- Fork PRs blocked from AWS plan credentials; fmt/validate remain non-AWS

## Still DISABLED / not LIVE VERIFIED

- GitHub Terraform **apply**
- Custom domain / ACM
- Project C

## Next (authorized later)

1. Explicit approval for controlled GitHub no-op apply
2. Set `ENABLE_TERRAFORM_APPLY=true` temporarily
3. Replace apply-job hard-exit stub with real `terraform apply` under review
4. `workflow_dispatch` apply=true against the proven 0/0/0 config — then re-disable apply
