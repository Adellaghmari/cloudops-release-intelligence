# Project status

**Project:** CloudOps Release Intelligence  
**Phase:** 16B Part 1 COMPLETE (remote state migrated; main hardening NOT applied)  
**Status:** Public demo LIVE; main state on S3 + lockfile proven  
**Complete:** No (main hardening apply + GitHub plan/apply gates remain)

## LIVE VERIFIED baseline

- Frontend / health / ready HTTP 200
- Lambdas Active on `sha256:ef3778d5af9e80d155610d5ffb0a889e509a4ba3da3fee2ac6878c6c5287ade5`
- Operator `adel-admin` / `eu-west-1`

## Phase 16B Part 1

### Bootstrap
- Applied `tfplan-bootstrap-16a1`: **7 add / 0 change / 0 destroy**
- Bucket `cloudops-prod-tfstate-7d9fdd77` — BPA, BucketOwnerEnforced, versioning, AES256, not public, no website
- Bootstrap state remains **local** (backed up under gitignored `infra/state-backups/`)

### Main state migration
- Pre-migration backup: `infra/state-backups/main-pre-s3-migration-20260909.tfstate`
- Address count before/after: **50 / 50** (match)
- Backend: S3 `cloudops-release-intelligence/prod/terraform.tfstate`, `use_lockfile=true`, `eu-west-1`
- Remote object exists (AES256 + versioning)
- Lock contention proven: concurrent plan failed with “Error acquiring the state lock”; lock released after cancel

### Fresh remote plan (DO NOT APPLY yet)
- `infra/tfplan-hardening-16b-remote-1` → **2 add / 2 change / 0 destroy**
- Adds GitHub plan/deploy terraform-state policies; updates plan OIDC trust + deploy IAM (`GetInvalidation`)
- Pre-migration plans discarded as stale for apply

### Still DISABLED
- GitHub Terraform apply (`ENABLE_TERRAFORM_APPLY`)
- Main hardening apply (Part 2)

## Next (Part 2)
1. Review/apply `tfplan-hardening-16b-remote-1` (or regenerate if config drifted)
2. Set `TERRAFORM_REMOTE_STATE_READY=true`
3. Prove GitHub OIDC terraform plan (same-repo only)
4. Later gated apply — not automatic on push
