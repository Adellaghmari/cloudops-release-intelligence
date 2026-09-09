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
| `infra/bootstrap/` | S3 state bucket + security only |
| `infra/` (main) | Product infrastructure + GitHub OIDC roles + GitHub remote-state IAM policies |

## GitHub state IAM (main — planned, not yet applied)

**Plan + apply roles** (exact objects):

- `s3:ListBucket` with prefix condition
- state object: `s3:GetObject` + `s3:PutObject` (**no** `DeleteObject`)
- lock object `<key>.tflock`: Get/Put/Delete

Fresh remote plan: `tfplan-hardening-16b-remote-1` (apply in Part 2).

## Migration completed locally with `adel-admin`

Do not re-run `init -migrate-state` unless intentionally remediating. Never apply pre-migration saved plans.

## Ownership split (corrected)

| Stack | Owns |
| --- | --- |
| `infra/bootstrap/` | S3 state bucket + security only (BPA, versioning, SSE-S3, ownership, deny insecure transport) |
| `infra/` (main) | Product infrastructure + GitHub OIDC roles + **GitHub remote-state IAM policies** |

Bootstrap must **not** attach policies to main-stack IAM roles (avoids cross-state lifecycle coupling).

## Target architecture (Phase 16B)

```
infra/bootstrap/  → private versioned SSE-S3 state bucket (local bootstrap state)
infra/            → product stack + GitHub state-access policies
                    backend "s3" { use_lockfile = true }
```

### Backend (main)

```hcl
terraform {
  backend "s3" {
    bucket       = "<from bootstrap output state_bucket>"
    key          = "cloudops-release-intelligence/prod/terraform.tfstate"
    region       = "eu-west-1"
    encrypt      = true
    use_lockfile = true
  }
}
```

No credentials in backend config. No DynamoDB lock table.

### GitHub state IAM (main stack)

After bootstrap apply, set gitignored vars:

- `terraform_state_bucket_arn`
- `terraform_state_bucket_name` (optional)
- `terraform_state_key` (default matches backend key)

**Plan role** (exact objects):

- `s3:ListBucket` on bucket with prefix condition
- `s3:GetObject` on state object only
- `s3:GetObject` / `PutObject` / `DeleteObject` on `<key>.tflock` only
- **No** `DeleteObject` or `PutObject` on the state object

**Apply/deploy role**:

- `s3:ListBucket` (prefix-conditioned)
- `s3:GetObject` / `PutObject` on state object (**no** `DeleteObject` on state)
- lock object Get/Put/Delete

### Migration sequence (DO NOT RUN in 16A.1)

1. Local operator `adel-admin` (not GitHub) applies **bootstrap** plan.
2. Backup main local state: `terraform state pull > ../state-backups/main-pre-migrate.json` (gitignored); record SHA-256 of the backup file (do not print state contents).
3. Configure `infra/backend.tf` for S3 + `use_lockfile`.
4. Set `terraform_state_bucket_arn` (and name) from bootstrap outputs into gitignored tfvars.
5. `cd infra && terraform init -migrate-state` as `adel-admin`.
6. `terraform state list` matches pre-migration inventory.
7. **Discard** any pre-migration saved plan (e.g. `tfplan-hardening-16a` / `16a1`). Regenerate a **fresh** plan against remote state.
8. Apply main hardening locally (includes GitHub state IAM + GetInvalidation scoping).
9. Test GitHub Terraform **plan** via OIDC.
10. Only later: controlled GitHub **apply** (`workflow_dispatch` + env + enable flag).

### Saved plan rule

Never apply a main plan generated against **local** state after the backend has migrated. Always replan on the remote backend.

### What must never be committed

- `terraform.tfstate` / `*.tfstate*`
- `state-backups/`
- `*.tfvars` (except examples)
- `tfplan*`
- `.aws-login-config`
- credentials / session tokens

`.terraform.lock.hcl` **is** committed.
