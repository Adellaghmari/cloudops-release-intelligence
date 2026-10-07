data "aws_iam_policy_document" "terraform_apply_iam" {
  statement {
    sid = "ReadCloudOpsRoles"
    actions = [
      "iam:GetRole",
      "iam:GetRolePolicy",
      "iam:ListAttachedRolePolicies",
      "iam:ListRolePolicies",
      "iam:ListRoleTags",
    ]
    resources = local.managed_role_arns
  }

  statement {
    sid = "ManageCloudOpsRoleConfiguration"
    actions = [
      "iam:DeleteRolePolicy",
      "iam:PutRolePolicy",
      "iam:TagRole",
      "iam:UntagRole",
      "iam:UpdateAssumeRolePolicy",
    ]
    resources = local.managed_role_arns
  }

  statement {
    sid     = "CreateBoundedEvidenceProducer"
    actions = ["iam:CreateRole"]
    resources = [
      "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-evidence-producer",
    ]

    condition {
      test     = "StringEquals"
      variable = "iam:PermissionsBoundary"
      values   = [local.evidence_boundary_arn]
    }
  }

  statement {
    sid     = "DeleteEvidenceProducer"
    actions = ["iam:DeleteRole"]
    resources = [
      "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-evidence-producer",
    ]
  }

  statement {
    sid     = "SetEvidenceProducerBoundary"
    actions = ["iam:PutRolePermissionsBoundary"]
    resources = [
      "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-evidence-producer",
    ]

    condition {
      test     = "StringEquals"
      variable = "iam:PermissionsBoundary"
      values   = [local.evidence_boundary_arn]
    }
  }

  statement {
    sid = "ManageApprovedReadOnlyAttachments"
    actions = [
      "iam:AttachRolePolicy",
      "iam:DetachRolePolicy",
    ]
    resources = [
      "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-deploy",
      "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-github-plan",
    ]

    condition {
      test     = "ArnEquals"
      variable = "iam:PolicyARN"
      values   = ["arn:aws:iam::aws:policy/ReadOnlyAccess"]
    }
  }

  statement {
    sid     = "PassLambdaRuntimeRolesOnly"
    actions = ["iam:PassRole"]
    resources = [
      "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-api",
      "arn:aws:iam::${local.account_id}:role/${local.name_prefix}-worker",
    ]

    condition {
      test     = "StringEquals"
      variable = "iam:PassedToService"
      values   = ["lambda.amazonaws.com"]
    }
  }

  statement {
    sid = "ReadGitHubOIDCProvider"
    actions = [
      "iam:GetOpenIDConnectProvider",
      "iam:ListOpenIDConnectProviders",
    ]
    resources = ["*"]
  }

  statement {
    sid = "MaintainGitHubOIDCProvider"
    actions = [
      "iam:AddClientIDToOpenIDConnectProvider",
      "iam:RemoveClientIDFromOpenIDConnectProvider",
      "iam:UpdateOpenIDConnectProviderThumbprint",
    ]
    resources = [data.aws_iam_openid_connect_provider.github.arn]
  }
}

resource "aws_iam_role_policy" "terraform_apply_iam" {
  name   = "cloudops-iam"
  role   = aws_iam_role.terraform_apply.id
  policy = data.aws_iam_policy_document.terraform_apply_iam.json

  depends_on = [aws_iam_policy.evidence_producer_boundary]
}
