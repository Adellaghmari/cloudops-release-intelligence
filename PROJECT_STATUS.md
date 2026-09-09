# Project status

**Project:** CloudOps Release Intelligence  
**Phase:** 16B Part 3 COMPLETE  
**Status:** **PROJECT B COMPLETE**  
**Complete:** Yes (custom domain optional; Project C not started)

## LIVE VERIFIED baseline

- Frontend `https://d34fwrlm14h6js.cloudfront.net` HTTP 200
- Backend `/api/v1/health` + `/api/v1/ready` HTTP 200
- Lambdas Active on `sha256:ef3778d5af9e80d155610d5ffb0a889e509a4ba3da3fee2ac6878c6c5287ade5`
- Operator `adel-admin` / `eu-west-1`

## Phase 16B Part 3 — controlled GitHub Terraform apply

### Workflow safety
- Apply only on `workflow_dispatch` + `apply=true` + `ENABLE_TERRAFORM_APPLY=true` + `environment: prod`
- Never on push/PR
- Same-job plan must be no-op (`-detailed-exitcode`) or apply is refused
- Apply uses `cloudops-prod-github-deploy` via OIDC (not plan role / static keys)
- Deploy role also has AWS managed **ReadOnlyAccess** for Terraform refresh during apply (broad read — not fully least privilege)

### Controlled no-op apply proof
- Gate enabled only for the run, then immediately set `ENABLE_TERRAFORM_APPLY=false`
- Run: https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34399074308 (**success**)
- SHA: `8232ee0`
- Assumed: `assumed-role/cloudops-prod-github-deploy/GitHubActions`
- Terraform **v1.10.5**; S3 backend `cloudops-prod-tfstate-7d9fdd77` / `cloudops-release-intelligence/prod/terraform.tfstate` / `use_lockfile`
- Same-job plan: **No changes**
- Apply: **0 added / 0 changed / 0 destroyed**
- Lock acquired/released; idle `.tflock` absent afterward

### Negative apply-gate proof
- After disable: dispatch with `apply=true` → apply job **skipped**
- Run: https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34399312899
- `ENABLE_TERRAFORM_APPLY=false` (final operational posture)

### Final posture
| Gate | Value |
| --- | --- |
| `TERRAFORM_REMOTE_STATE_READY` | `true` |
| `ENABLE_TERRAFORM_APPLY` | `false` |

## Prior LIVE VERIFIED (16B Part 1–2)

- Remote S3 state + native lockfile + lock contention
- IAM hardening apply (state policies, GetInvalidation, immutable plan OIDC)
- Frontend CloudFront invalidation waiter
- GitHub Terraform PLAN via `cloudops-prod-github-plan`

## Explicit non-goals (not required for COMPLETE)

- Custom domain / ACM / TLSv1.2_2021 viewer policy on default `*.cloudfront.net`
- Fully least-privilege plan/apply IAM while ReadOnlyAccess remains
- Project C (Secure Integration Hub)
- Backend image rebuild

## Optional polish (separate future work)

1. Purchased domain + ACM (us-east-1) + CloudFront alias + modern TLS policy
2. Narrower replacement for ReadOnlyAccess on plan/apply roles
3. Temporary re-enable of apply for intentional infra changes only
