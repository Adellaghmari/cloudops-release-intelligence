# Terraform state

## Current strategy (first controlled environment)

State is **local** in `infra/terraform.tfstate`.

Why:

- There is no existing AWS account bootstrap in this repository.
- Creating an S3 backend in the same stack that first creates the account's buckets is a circular bootstrap.
- This environment is a single operator, single region, single apply path.

How secrets are handled:

- State may contain resource IDs and ARNs. It must never be committed.
- Secret `.tfvars` files are gitignored. `terraform.tfvars.example` has no secrets.
- AWS credentials are never stored in Terraform files. Apply uses the operator's local AWS login or GitHub OIDC.

Locking / concurrency:

- Local state has no remote lock. Do not run two applies at once.
- After a remote backend exists, use S3 + DynamoDB lock in a **separately documented** bootstrap, not this stack.

What must never be committed:

- `terraform.tfstate`
- `terraform.tfstate.backup`
- `*.tfvars` except `*.tfvars.example`
- crash logs and override files

`.terraform.lock.hcl` **is** committed so provider versions stay pinned.
