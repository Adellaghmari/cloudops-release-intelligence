# CloudOps account bootstrap

This Terraform root owns only the production release control plane:

- `cloudops-prod-terraform-apply`
- its fixed permission boundary
- its scoped inline policies
- the evidence producer permission boundary

It reuses the existing GitHub OIDC provider. It does not own application
runtime resources or the existing state bucket.

State is stored in the existing encrypted bucket under:

`cloudops-release-intelligence/prod/account-bootstrap/terraform.tfstate`

The first apply requires the independently authorized `adel-admin` session.
Normal application releases then use GitHub OIDC and the protected `prod`
environment. Never apply this stack from the production Terraform role because
that would allow the role to change its own permission ceiling.

Bootstrap updates that add Terraform provider read permissions must be applied
here before resuming a protected application release.
