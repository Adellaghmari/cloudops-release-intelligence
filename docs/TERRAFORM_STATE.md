# Terraform state

## Current strategy (Phase 16A)

Main stack state remains **local** in `infra/terraform.tfstate` until Phase 16B migration.

GitHub Terraform **apply remains DISABLED** while state is local.

## Target architecture (Phase 16B)

```
infra/bootstrap/     → S3 state bucket + GitHub role state policies (own local state)
infra/               → product stack using backend "s3" { use_lockfile = true }
```

### Why a separate bootstrap stack

The state bucket cannot live in the same Terraform state it is meant to store.
`infra/bootstrap/` creates only:

- private versioned SSE-S3 state bucket
- Block Public Access + BucketOwnerEnforced
- deny insecure transport bucket policy
- least-privilege state/lock object access for existing GitHub plan + deploy roles

No Lambda, API, DynamoDB, CloudFront, or product queues in bootstrap.

### Locking

Use native S3 lockfile (`use_lockfile = true`).

Do **not** create a DynamoDB lock table (deprecated / unnecessary for this project).

### Credentials

No access keys in backend config, GitHub variables, or committed files.
Operators and GitHub Actions use OIDC / local AWS login.

Backend non-secrets (safe to document):

- bucket name
- state key
- region
- `use_lockfile = true`
- `encrypt = true`

### Migration procedure (DO NOT RUN in 16A)

1. Backup `infra/terraform.tfstate` (+ `.backup`) to an offline path; record `serial` / lineage.
2. `cd infra/bootstrap && terraform apply` (reviewed plan only).
3. Verify bucket: public access block, versioning, encryption, ownership, no public ACL/policy grants.
4. Rewrite `infra/backend.tf` to `backend "s3"` with `use_lockfile = true` (values from bootstrap outputs).
5. `cd infra && terraform init -migrate-state` (confirm prompts carefully).
6. `terraform state list` matches pre-migration inventory.
7. `terraform plan` → prefer `0/0/0` or only reviewed Phase 16 hardening.
8. Keep local backup until several successful remote plans/applies.
9. Only then enable controlled GitHub apply (`workflow_dispatch` / environment gate).

### What must never be committed

- `terraform.tfstate` / `*.tfstate*`
- `*.tfvars` (except examples)
- `tfplan*`
- `.aws-login-config`
- session tokens / access keys

`.terraform.lock.hcl` **is** committed.
