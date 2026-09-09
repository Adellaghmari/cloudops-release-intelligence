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
| `cloudops-prod-github-plan` | `sub` = the ID-qualified repo `*` | Read-only plan from pull requests |

## Why this is safer than stored access keys

A static access key is a long-lived secret. Anyone who copies it can call AWS until it is rotated. OIDC issues a **short-lived** token for one workflow run, scoped to this repository and branch. There is no standing AWS key in GitHub Secrets for deploy.

## Live verification

OIDC is **LIVE VERIFIED**. CD run https://github.com/Adellaghmari/cloudops-release-intelligence/actions/runs/34370805611 called `aws sts get-caller-identity` and received `arn:aws:sts::912415493331:assumed-role/cloudops-prod-github-deploy/GitHubActions`.

GitHub `terraform apply` remains disabled while state is local (`ENABLE_TERRAFORM_APPLY` is unset).
