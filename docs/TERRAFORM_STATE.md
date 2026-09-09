# Terraform state

## Current strategy (Phase 16B Part 1 — LIVE)

Main stack state is on **S3**:

- bucket: `cloudops-prod-tfstate-7d9fdd77`
- key: `cloudops-release-intelligence/prod/terraform.tfstate`
- region: `eu-west-1`
- `encrypt = true`
- `use_lockfile = true`

Bootstrap stack state remains **local** under `infra/bootstrap/terraform.tfstate` (intentional; backed up in gitignored `infra/state-backups/`).

GitHub Terraform **apply remains DISABLED**.

Pre-migration main backup (gitignored):

- `infra/state-backups/main-pre-s3-migration-20260909.tfstate`
- SHA-256 recorded in Phase 16B Part 1 report (do not commit)

## Ownership split

| Stack | Owns |
| --- | --- |
| `infra/bootstrap/` | S3 state bucket + security only (BPA, versioning, SSE-S3, ownership, deny insecure transport) |
| `infra/` (main) | Product infrastructure + GitHub OIDC roles + GitHub remote-state IAM policies |

Bootstrap must **not** attach policies to main-stack IAM roles.

## Backend (main)

```hcl
terraform {
  backend "s3" {
    bucket       = "cloudops-prod-tfstate-7d9fdd77"
    key          = "cloudops-release-intelligence/prod/terraform.tfstate"
    region       = "eu-west-1"
    encrypt      = true
    use_lockfile = true
  }
}
```

No credentials in backend config. No DynamoDB lock table.

## GitHub state IAM (main — planned in Part 2 apply)

Gitignored vars after bootstrap:

- `terraform_state_bucket_arn`
- `terraform_state_bucket_name`
- `terraform_state_key` (default matches backend key)

**Plan + apply roles** (exact objects):

- `s3:ListBucket` with prefix condition
- state object: `s3:GetObject` + `s3:PutObject` (**no** `DeleteObject`)
- lock object `<key>.tflock`: Get/Put/Delete

Fresh remote plan: `tfplan-hardening-16b-remote-1` (**do not apply pre-migration plans**).

Plan role also retains AWS managed `ReadOnlyAccess` for refresh compatibility until a tested replacement exists. No AdministratorAccess.

## Locking

Native S3 lockfile proven in Part 1 (normal plan + contention test). Never use `-lock=false`.

## What must never be committed

- `terraform.tfstate` / `*.tfstate*`
- `state-backups/`
- `*.tfvars` (except examples)
- `tfplan*`
- `.aws-login-config`
- credentials / session tokens

`.terraform.lock.hcl` **is** committed.
