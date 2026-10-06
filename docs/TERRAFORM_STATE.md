# Terraform state

## Current strategy (Phase 16B Part 3 — PROJECT COMPLETE)

Main stack state is on **S3**:

- bucket: `cloudops-prod-tfstate-7d9fdd77`
- key: `cloudops-release-intelligence/prod/terraform.tfstate`
- region: `eu-west-1`
- `encrypt = true`
- `use_lockfile = true`

The original state-bucket foundation remains local under
`infra/bootstrap/terraform.tfstate` (intentional; backed up in gitignored
`infra/state-backups/`).

The production release control plane uses the same encrypted bucket with a
separate remote state key:

`cloudops-release-intelligence/prod/account-bootstrap/terraform.tfstate`

That stack owns only `cloudops-prod-terraform-apply`, its permission ceiling,
scoped policies, and the evidence producer boundary. Its one-time human bootstrap
is complete with zero drift.

GitHub Terraform **plan** is LIVE VERIFIED (`TERRAFORM_REMOTE_STATE_READY=true`).  
GitHub Terraform **controlled apply** is LIVE VERIFIED once (no-op), then the gate was re-disabled:

- `ENABLE_TERRAFORM_APPLY=false` (final posture)
- Re-enable only for an explicit maintenance window

Pre-migration main backup (gitignored):

- `infra/state-backups/main-pre-s3-migration-20260909.tfstate`
- SHA-256 recorded in Phase 16B Part 1 report (do not commit)

## Ownership split

| Stack | Owns |
| --- | --- |
| `infra/bootstrap/` | S3 state bucket + security only (BPA, versioning, SSE-S3, ownership, deny insecure transport) |
| `infra/account-bootstrap/` | Production Terraform apply role, permission boundaries, and scoped release-control policies |
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

## GitHub state IAM (main — applied in Part 2)

Gitignored vars / CI `TF_VAR_*` after bootstrap:

- `terraform_state_bucket_arn`
- `terraform_state_bucket_name`
- `terraform_state_key` (default matches backend key)

The GitHub plan role can list the exact state prefix, read the state object, and
acquire or release only its native lock object. It cannot write or delete the
state object.

The application deploy role retains read only visibility of the exact state and
lock objects for compatibility. It cannot write either object.

The separately bootstrapped `cloudops-prod-terraform-apply` role is the only
GitHub identity allowed to update the main state object. Its state permissions
are scoped to the exact main state key and lock object. The human bootstrap
stack retains separate ownership of its own state.

Plan and application deploy roles retain AWS managed `ReadOnlyAccess` for
refresh compatibility until a tested replacement exists. No role has
AdministratorAccess.

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
