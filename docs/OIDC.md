# GitHub OIDC → AWS

Normal deployment must not use long-lived `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` in GitHub Secrets.

## Identity provider

- Issuer: `https://token.actions.githubusercontent.com`
- Audience: `sts.amazonaws.com`
- Terraform: `aws_iam_openid_connect_provider.github` (or reuse an existing account provider)

## Roles

GitHub now embeds numeric owner/repo IDs in the `sub` claim (observed 2026-09-09):

`repo:Adellaghmari@179922674/cloudops-release-intelligence@1361976389:ref:refs/heads/main`

| Role | Trust | Use |
| --- | --- | --- |
| `cloudops-prod-github-deploy` | that `sub` for `main`, and `:environment:prod` | Push ECR, update Lambda, sync S3 |
| `cloudops-prod-github-plan` | ID-qualified `main` and same-repository pull-request subjects | Read-only Terraform plan |
| `cloudops-prod-github-evidence-producer` | exact ID-qualified `main` subject | Invoke only `POST /api/v1/events` on the CloudOps API |

## Why this is safer than stored access keys

A static access key is a long-lived secret. Anyone who copies it can call AWS until it is rotated. OIDC issues a **short-lived** token for one workflow run, scoped to this repository and branch. There is no standing AWS key in GitHub Secrets for deploy.

## Live verification

OIDC is **LIVE VERIFIED**. CD run https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805611 called `aws sts get-caller-identity` and received `arn:aws:sts::912415493331:assumed-role/cloudops-prod-github-deploy/GitHubActions`.

Remote encrypted S3 state and its native lockfile are LIVE VERIFIED. GitHub `terraform apply` remains disabled by design (`ENABLE_TERRAFORM_APPLY=false`).

The evidence-producer trust policy and signed workflow are **IMPLEMENTED LOCALLY, NOT YET DEPLOYED OR LIVE VERIFIED**. They use short-lived OIDC credentials and SigV4; no static AWS key or browser secret is introduced.
