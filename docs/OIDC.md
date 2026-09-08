# GitHub OIDC → AWS

Normal deployment must not use long-lived `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` in GitHub Secrets.

## Identity provider

- Issuer: `https://token.actions.githubusercontent.com`
- Audience: `sts.amazonaws.com`
- Terraform: `aws_iam_openid_connect_provider.github` (or reuse an existing account provider)

## Roles

| Role | Trust | Use |
| --- | --- | --- |
| `cloudops-prod-github-deploy` | `sub` = this repo `ref:refs/heads/main` (and `environment:prod`) | Push ECR, update Lambda, sync S3, apply image URI |
| `cloudops-prod-github-plan` | `sub` = this repo `*` | Read-only plan from pull requests |

## Why this is safer than stored access keys

A static access key is a long-lived secret. Anyone who copies it can call AWS until it is rotated. OIDC issues a **short-lived** token for one workflow run, scoped to this repository and branch. There is no standing AWS key in GitHub Secrets for deploy.

## Live verification

OIDC is **IMPLEMENTED** in Terraform and the `cd` workflow. It becomes **LIVE VERIFIED** only after:

1. The IAM role exists in the account
2. A GitHub Actions run on this repository successfully calls `aws sts get-caller-identity` using `configure-aws-credentials` with `role-to-assume`

Until that run exists, do not claim OIDC on a CV.
