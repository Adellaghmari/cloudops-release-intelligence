terraform {
  required_version = ">= 1.10.0, < 2.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.80"
    }
  }

  backend "s3" {
    bucket       = "cloudops-prod-tfstate-7d9fdd77"
    key          = "cloudops-release-intelligence/prod/account-bootstrap/terraform.tfstate"
    region       = "eu-west-1"
    encrypt      = true
    use_lockfile = true
  }
}

provider "aws" {
  region = "eu-west-1"

  default_tags {
    tags = {
      Project     = "cloudops"
      Environment = "prod"
      ManagedBy   = "terraform-account-bootstrap"
      Purpose     = "production-release-control"
    }
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"
}

locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = "eu-west-1"

  github_owner_id = "179922674"
  github_repo_id  = "1361976389"
  github_oidc_sub = "repo:Adellaghmari@${local.github_owner_id}/cloudops-release-intelligence@${local.github_repo_id}:environment:prod"

  name_prefix = "cloudops-prod"

  terraform_apply_boundary_arn = "arn:aws:iam::${local.account_id}:policy/${local.name_prefix}-terraform-apply-boundary"
  evidence_boundary_arn        = "arn:aws:iam::${local.account_id}:policy/${local.name_prefix}-evidence-producer-boundary"

  api_id                  = "8kci5uht3d"
  web_bucket              = "cloudops-prod-web-7be25877"
  evidence_bucket         = "cloudops-prod-raw-7be25877"
  state_bucket            = "cloudops-prod-tfstate-7d9fdd77"
  state_key               = "cloudops-release-intelligence/prod/terraform.tfstate"
  state_lock_key          = "${local.state_key}.tflock"
  cloudfront_distribution = "E2220NQVG6GU75"

  managed_role_arns = [
    "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-api",
    "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-worker",
    "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-deploy",
    "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-plan",
    "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-evidence-producer",
  ]
}

data "aws_iam_policy_document" "terraform_apply_assume" {
  statement {
    sid     = "GitHubProdEnvironment"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [data.aws_iam_openid_connect_provider.github.arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:sub"
      values   = [local.github_oidc_sub]
    }
  }
}

resource "aws_iam_role" "terraform_apply" {
  name                 = "${local.name_prefix}-terraform-apply"
  description          = "Applies reviewed CloudOps Terraform plans from the protected GitHub prod environment"
  assume_role_policy   = data.aws_iam_policy_document.terraform_apply_assume.json
  permissions_boundary = local.terraform_apply_boundary_arn
  max_session_duration = 3600

  depends_on = [aws_iam_policy.terraform_apply_boundary]
}

output "terraform_apply_role_arn" {
  value = aws_iam_role.terraform_apply.arn
}

output "terraform_apply_boundary_arn" {
  value = aws_iam_policy.terraform_apply_boundary.arn
}

output "evidence_producer_boundary_arn" {
  value = aws_iam_policy.evidence_producer_boundary.arn
}
