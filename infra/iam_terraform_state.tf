# GitHub OIDC roles' access to the remote Terraform state bucket.
# Owned by the MAIN stack so bootstrap never mutates main-stack IAM roles.
# Bucket itself is created by infra/bootstrap/ (separate state).
#
# Set terraform_state_bucket_arn after bootstrap apply (or use a PENDING
# placeholder for pre-migration informational plans only).

locals {
  terraform_state_enabled = var.terraform_state_bucket_arn != ""
  terraform_state_key     = var.terraform_state_key
  terraform_state_lock    = "${var.terraform_state_key}.tflock"
}

# Plan role: read state object; full lockfile acquire/release; no DeleteObject on state.
data "aws_iam_policy_document" "github_plan_terraform_state" {
  count = local.terraform_state_enabled ? 1 : 0

  statement {
    sid       = "ListStatePrefix"
    actions   = ["s3:ListBucket"]
    resources = [var.terraform_state_bucket_arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        local.terraform_state_key,
        "${local.terraform_state_key}/*",
        local.terraform_state_lock,
      ]
    }
  }

  statement {
    sid       = "StateObjectRead"
    actions   = ["s3:GetObject"]
    resources = ["${var.terraform_state_bucket_arn}/${local.terraform_state_key}"]
  }

  statement {
    sid = "LockObjectAcquireRelease"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = ["${var.terraform_state_bucket_arn}/${local.terraform_state_lock}"]
  }
}

resource "aws_iam_role_policy" "github_plan_terraform_state" {
  count  = local.terraform_state_enabled ? 1 : 0
  name   = "terraform-state"
  role   = aws_iam_role.github_plan.id
  policy = data.aws_iam_policy_document.github_plan_terraform_state[0].json
}

# Apply/deploy role: Get+Put state object (no DeleteObject on state); full lockfile.
data "aws_iam_policy_document" "github_deploy_terraform_state" {
  count = local.terraform_state_enabled ? 1 : 0

  statement {
    sid       = "ListStatePrefix"
    actions   = ["s3:ListBucket"]
    resources = [var.terraform_state_bucket_arn]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        local.terraform_state_key,
        "${local.terraform_state_key}/*",
        local.terraform_state_lock,
      ]
    }
  }

  statement {
    sid = "StateObjectReadWrite"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = ["${var.terraform_state_bucket_arn}/${local.terraform_state_key}"]
  }

  statement {
    sid = "LockObjectAcquireRelease"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
    ]
    resources = ["${var.terraform_state_bucket_arn}/${local.terraform_state_lock}"]
  }
}

resource "aws_iam_role_policy" "github_deploy_terraform_state" {
  count  = local.terraform_state_enabled ? 1 : 0
  name   = "terraform-state"
  role   = aws_iam_role.github_deploy.id
  policy = data.aws_iam_policy_document.github_deploy_terraform_state[0].json
}
