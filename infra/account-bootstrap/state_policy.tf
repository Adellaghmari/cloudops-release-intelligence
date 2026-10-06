data "aws_iam_policy_document" "terraform_apply_state" {
  statement {
    sid       = "ListApplicationState"
    actions   = ["s3:ListBucket"]
    resources = ["arn:aws:s3:::${local.state_bucket}"]

    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values = [
        local.state_key,
        local.state_lock_key,
      ]
    }
  }

  statement {
    sid = "ReadWriteApplicationState"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = ["arn:aws:s3:::${local.state_bucket}/${local.state_key}"]
  }

  statement {
    sid = "ManageApplicationStateLock"
    actions = [
      "s3:DeleteObject",
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = ["arn:aws:s3:::${local.state_bucket}/${local.state_lock_key}"]
  }
}

resource "aws_iam_role_policy" "terraform_apply_state" {
  name   = "application-state"
  role   = aws_iam_role.terraform_apply.id
  policy = data.aws_iam_policy_document.terraform_apply_state.json
}
