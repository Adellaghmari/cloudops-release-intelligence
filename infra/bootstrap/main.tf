# Bootstrap stack: remote-state S3 foundation ONLY.
# Owns the bucket; does NOT attach policies to main-stack GitHub IAM roles.
# Bootstrap keeps its own small LOCAL terraform.tfstate (intentional).
#
# Phase 16A.1: plan only. Do NOT apply until Phase 16B.

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

locals {
  name_prefix = "${var.project}-${var.environment}"
  # Stable main-stack state key (also configured in the main stack variables).
  state_key = "cloudops-release-intelligence/prod/terraform.tfstate"
}

resource "random_id" "suffix" {
  byte_length = 4
}

resource "aws_s3_bucket" "tfstate" {
  bucket        = "${local.name_prefix}-tfstate-${random_id.suffix.hex}"
  force_destroy = false

  lifecycle {
    prevent_destroy = true
  }
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
    # After bootstrap apply (Phase 16B), set main-stack vars and replace infra/backend.tf with:
    terraform {
      backend "s3" {
        bucket       = "${aws_s3_bucket.tfstate.bucket}"
        key          = "${local.state_key}"
        region       = "${var.aws_region}"
        encrypt      = true
        use_lockfile = true
      }
    }
    # Then set in gitignored terraform.tfvars (or TF_VAR_*):
    #   terraform_state_bucket_arn  = "${aws_s3_bucket.tfstate.arn}"
    #   terraform_state_bucket_name = "${aws_s3_bucket.tfstate.bucket}"
    #   terraform_state_key         = "${local.state_key}"
  EOT
}
