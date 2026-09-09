# Remote state backend (Phase 16B). Bucket created by infra/bootstrap (separate local state).
# Non-secret identifiers only — no access keys or session tokens.
terraform {
  backend "s3" {
    bucket       = "cloudops-prod-tfstate-7d9fdd77"
    key          = "cloudops-release-intelligence/prod/terraform.tfstate"
    region       = "eu-west-1"
    encrypt      = true
    use_lockfile = true
  }
}
