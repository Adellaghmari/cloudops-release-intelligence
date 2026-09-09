# Bootstrap stack: Terraform remote-state infrastructure ONLY.
# Local state for this stack is intentional (chicken-and-egg).
# Do NOT put product resources (Lambda, API, DynamoDB, CloudFront) here.
#
# Phase 16A: plan only. Do NOT apply until Phase 16B.

terraform {
  required_version = ">= 1.10.0, < 2.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.80"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform-bootstrap"
      Purpose     = "terraform-remote-state"
    }
  }
}

variable "aws_region" {
  type    = string
  default = "eu-west-1"
}

variable "project" {
  type    = string
  default = "cloudops"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "github_deploy_role_name" {
  type        = string
  description = "Existing main-stack GitHub deploy role name (state write/lock)."
  default     = "cloudops-prod-github-deploy"
}

variable "github_plan_role_name" {
  type        = string
  description = "Existing main-stack GitHub plan role name (state read/lock)."
  default     = "cloudops-prod-github-plan"
}

data "aws_caller_identity" "current" {}

data "aws_iam_role" "github_deploy" {
  name = var.github_deploy_role_name
}

data "aws_iam_role" "github_plan" {
  name = var.github_plan_role_name
}

locals {
  name_prefix = "${var.project}-${var.environment}"
  # Dedicated state key for the main product stack.
  state_key = "${local.name_prefix}/terraform.tfstate"
}

resource "random_id" "suffix" {
  byte_length = 4
}

resource "aws_s3_bucket" "tfstate" {
  bucket = "${local.name_prefix}-tfstate-${random_id.suffix.hex}"
}

resource "aws_s3_bucket_public_access_block" "tfstate" {
  bucket                  = aws_s3_bucket.tfstate.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_versioning" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_ownership_controls" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

# Deny non-TLS access to state objects.
data "aws_iam_policy_document" "tfstate_bucket" {
  statement {
    sid     = "DenyInsecureTransport"
    effect  = "Deny"
    actions = ["s3:*"]
    resources = [
      aws_s3_bucket.tfstate.arn,
      "${aws_s3_bucket.tfstate.arn}/*",
    ]
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }
}

resource "aws_s3_bucket_policy" "tfstate" {
  bucket = aws_s3_bucket.tfstate.id
  policy = data.aws_iam_policy_document.tfstate_bucket.json
}

# Plan role: read state + lockfile acquire/release (no product apply writes).
data "aws_iam_policy_document" "plan_state" {
  statement {
    sid       = "ListStateBucket"
    actions   = ["s3:ListBucket", "s3:GetBucketVersioning"]
    resources = [aws_s3_bucket.tfstate.arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = [local.state_key, "${local.state_key}.*", "${local.state_key}.tflock"]
    }
  }
  statement {
    sid = "StateObjectReadAndLock"
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = [
      "${aws_s3_bucket.tfstate.arn}/${local.state_key}",
      "${aws_s3_bucket.tfstate.arn}/${local.state_key}.tflock",
    ]
  }
}

resource "aws_iam_role_policy" "github_plan_state" {
  name   = "terraform-state"
  role   = data.aws_iam_role.github_plan.id
  policy = data.aws_iam_policy_document.plan_state.json
}

# Deploy/apply role: same lock + write updated state after apply.
data "aws_iam_policy_document" "deploy_state" {
  statement {
    sid       = "ListStateBucket"
    actions   = ["s3:ListBucket", "s3:GetBucketVersioning"]
    resources = [aws_s3_bucket.tfstate.arn]
  }
  statement {
    sid = "StateObjectReadWriteAndLock"
    actions = [
      "s3:GetObject",
      "s3:GetObjectVersion",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = [
      "${aws_s3_bucket.tfstate.arn}/${local.state_key}",
      "${aws_s3_bucket.tfstate.arn}/${local.state_key}.tflock",
    ]
  }
}

resource "aws_iam_role_policy" "github_deploy_state" {
  name   = "terraform-state"
  role   = data.aws_iam_role.github_deploy.id
  policy = data.aws_iam_policy_document.deploy_state.json
}

output "state_bucket" {
  value = aws_s3_bucket.tfstate.bucket
}

output "state_bucket_arn" {
  value = aws_s3_bucket.tfstate.arn
}

output "state_key" {
  value = local.state_key
}

output "backend_config_hint" {
  value = <<-EOT
    # After bootstrap apply (Phase 16B), replace infra/backend.tf with:
    terraform {
      backend "s3" {
        bucket       = "${aws_s3_bucket.tfstate.bucket}"
        key          = "${local.state_key}"
        region       = "${var.aws_region}"
        encrypt      = true
        use_lockfile = true
      }
    }
  EOT
}
