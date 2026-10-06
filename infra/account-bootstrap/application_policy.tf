data "aws_iam_policy_document" "terraform_apply_compute" {
  statement {
    sid = "ManageCloudOpsApiGateway"
    actions = [
      "apigateway:DELETE",
      "apigateway:GET",
      "apigateway:PATCH",
      "apigateway:POST",
      "apigateway:PUT",
    ]
    resources = [
      "arn:aws:apigateway:${local.region}::/apis/${local.api_id}",
      "arn:aws:apigateway:${local.region}::/apis/${local.api_id}/*",
    ]
  }

  statement {
    sid = "ManageCloudOpsFunctions"
    actions = [
      "lambda:AddPermission",
      "lambda:CreateFunction",
      "lambda:DeleteFunction",
      "lambda:GetFunction",
      "lambda:GetFunctionConfiguration",
      "lambda:GetPolicy",
      "lambda:ListTags",
      "lambda:RemovePermission",
      "lambda:TagResource",
      "lambda:UntagResource",
      "lambda:UpdateFunctionCode",
      "lambda:UpdateFunctionConfiguration",
    ]
    resources = [
      "arn:aws:lambda:${local.region}:${local.account_id}:function:${local.name_prefix}-api",
      "arn:aws:lambda:${local.region}:${local.account_id}:function:${local.name_prefix}-worker",
    ]
  }

  statement {
    sid = "ManageCloudOpsWorkerMapping"
    actions = [
      "lambda:CreateEventSourceMapping",
      "lambda:DeleteEventSourceMapping",
      "lambda:GetEventSourceMapping",
      "lambda:UpdateEventSourceMapping",
    ]
    resources = ["*"]

    condition {
      test     = "ArnEquals"
      variable = "lambda:FunctionArn"
      values   = ["arn:aws:lambda:${local.region}:${local.account_id}:function:${local.name_prefix}-worker"]
    }
  }

  statement {
    sid = "ManageCloudOpsQueues"
    actions = [
      "sqs:CreateQueue",
      "sqs:DeleteQueue",
      "sqs:GetQueueAttributes",
      "sqs:GetQueueUrl",
      "sqs:ListQueueTags",
      "sqs:SetQueueAttributes",
      "sqs:TagQueue",
      "sqs:UntagQueue",
    ]
    resources = [
      "arn:aws:sqs:${local.region}:${local.account_id}:${local.name_prefix}-analysis",
      "arn:aws:sqs:${local.region}:${local.account_id}:${local.name_prefix}-analysis-dlq",
    ]
  }

  statement {
    sid = "ManageCloudOpsEventBridge"
    actions = [
      "events:CreateEventBus",
      "events:DeleteEventBus",
      "events:DeleteRule",
      "events:DescribeEventBus",
      "events:DescribeRule",
      "events:ListTagsForResource",
      "events:ListTargetsByRule",
      "events:PutRule",
      "events:PutTargets",
      "events:RemoveTargets",
      "events:TagResource",
      "events:UntagResource",
    ]
    resources = [
      "arn:aws:events:${local.region}:${local.account_id}:event-bus/${local.name_prefix}-release-events",
      "arn:aws:events:${local.region}:${local.account_id}:rule/${local.name_prefix}-release-events/${local.name_prefix}-analysis",
    ]
  }

  statement {
    sid = "ManageCloudOpsDynamoDB"
    actions = [
      "dynamodb:CreateTable",
      "dynamodb:DeleteTable",
      "dynamodb:DescribeContinuousBackups",
      "dynamodb:DescribeTable",
      "dynamodb:DescribeTimeToLive",
      "dynamodb:ListTagsOfResource",
      "dynamodb:TagResource",
      "dynamodb:UntagResource",
      "dynamodb:UpdateContinuousBackups",
      "dynamodb:UpdateTable",
      "dynamodb:UpdateTimeToLive",
    ]
    resources = [
      "arn:aws:dynamodb:${local.region}:${local.account_id}:table/${local.name_prefix}-main",
      "arn:aws:dynamodb:${local.region}:${local.account_id}:table/${local.name_prefix}-main/index/*",
    ]
  }

  statement {
    sid = "ManageCloudOpsLogs"
    actions = [
      "logs:CreateLogGroup",
      "logs:DeleteLogGroup",
      "logs:ListTagsForResource",
      "logs:PutRetentionPolicy",
      "logs:TagResource",
      "logs:UntagResource",
    ]
    resources = [
      "arn:aws:logs:${local.region}:${local.account_id}:log-group:/aws/apigateway/${local.name_prefix}-http",
      "arn:aws:logs:${local.region}:${local.account_id}:log-group:/aws/lambda/${local.name_prefix}-api",
      "arn:aws:logs:${local.region}:${local.account_id}:log-group:/aws/lambda/${local.name_prefix}-worker",
    ]
  }

  statement {
    sid       = "DescribeCloudOpsLogs"
    actions   = ["logs:DescribeLogGroups"]
    resources = ["*"]
  }

  statement {
    sid = "ManageCloudOpsAlarm"
    actions = [
      "cloudwatch:DeleteAlarms",
      "cloudwatch:DescribeAlarms",
      "cloudwatch:PutMetricAlarm",
    ]
    resources = ["arn:aws:cloudwatch:${local.region}:${local.account_id}:alarm:${local.name_prefix}-analysis-dlq"]
  }
}

resource "aws_iam_role_policy" "terraform_apply_compute" {
  name   = "cloudops-compute"
  role   = aws_iam_role.terraform_apply.id
  policy = data.aws_iam_policy_document.terraform_apply_compute.json
}

data "aws_iam_policy_document" "terraform_apply_storage_edge" {
  statement {
    sid = "ManageCloudOpsBuckets"
    actions = [
      "s3:CreateBucket",
      "s3:DeleteBucket",
      "s3:DeleteBucketPolicy",
      "s3:GetBucketLocation",
      "s3:GetBucketPolicy",
      "s3:GetBucketPublicAccessBlock",
      "s3:GetBucketTagging",
      "s3:GetBucketVersioning",
      "s3:GetEncryptionConfiguration",
      "s3:GetLifecycleConfiguration",
      "s3:ListBucket",
      "s3:PutBucketPolicy",
      "s3:PutBucketPublicAccessBlock",
      "s3:PutBucketTagging",
      "s3:PutBucketVersioning",
      "s3:PutEncryptionConfiguration",
      "s3:PutLifecycleConfiguration",
    ]
    resources = [
      "arn:aws:s3:::${local.web_bucket}",
      "arn:aws:s3:::${local.evidence_bucket}",
    ]
  }

  statement {
    sid = "ReadCloudOpsBucketObjectsForStateRefresh"
    actions = [
      "s3:GetObject",
    ]
    resources = [
      "arn:aws:s3:::${local.web_bucket}/*",
      "arn:aws:s3:::${local.evidence_bucket}/*",
    ]
  }

  statement {
    sid = "ManageCloudOpsDistribution"
    actions = [
      "cloudfront:DeleteDistribution",
      "cloudfront:GetDistribution",
      "cloudfront:GetDistributionConfig",
      "cloudfront:ListTagsForResource",
      "cloudfront:TagResource",
      "cloudfront:UntagResource",
      "cloudfront:UpdateDistribution",
    ]
    resources = [
      "arn:aws:cloudfront::${local.account_id}:distribution/${local.cloudfront_distribution}",
    ]
  }

  statement {
    sid = "ManageCloudOpsOriginAccessControl"
    actions = [
      "cloudfront:CreateOriginAccessControl",
      "cloudfront:DeleteOriginAccessControl",
      "cloudfront:GetOriginAccessControl",
      "cloudfront:GetOriginAccessControlConfig",
      "cloudfront:UpdateOriginAccessControl",
    ]
    resources = ["*"]
  }

  statement {
    sid = "ManageCloudOpsRepository"
    actions = [
      "ecr:CreateRepository",
      "ecr:DeleteLifecyclePolicy",
      "ecr:DeleteRepository",
      "ecr:DescribeRepositories",
      "ecr:GetLifecyclePolicy",
      "ecr:ListTagsForResource",
      "ecr:PutLifecyclePolicy",
      "ecr:TagResource",
      "ecr:UntagResource",
    ]
    resources = ["arn:aws:ecr:${local.region}:${local.account_id}:repository/${local.name_prefix}-api"]
  }

  statement {
    sid = "ManageCloudOpsBudget"
    actions = [
      "budgets:CreateBudget",
      "budgets:DeleteBudget",
      "budgets:DescribeBudget",
      "budgets:ModifyBudget",
    ]
    resources = ["arn:aws:budgets::${local.account_id}:budget/${local.name_prefix}-monthly"]
  }
}

resource "aws_iam_role_policy" "terraform_apply_storage_edge" {
  name   = "cloudops-storage-edge"
  role   = aws_iam_role.terraform_apply.id
  policy = data.aws_iam_policy_document.terraform_apply_storage_edge.json
}
