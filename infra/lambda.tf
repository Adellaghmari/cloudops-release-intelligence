resource "aws_lambda_function" "api" {
  count         = local.deploy_compute ? 1 : 0
  function_name = "${local.name_prefix}-api"
  role          = aws_iam_role.lambda_api.arn
  package_type  = "Image"
  image_uri     = var.api_image_uri
  timeout       = 15
  memory_size   = 512
  architectures = ["x86_64"]

  image_config {
    command = ["api"]
  }

  environment {
    variables = {
      APP_ENV              = "prod"
      APP_STORE            = "dynamodb"
      APP_SEED_LOCAL       = "true"
      APP_LOG_LEVEL        = "info"
      APP_SERVICE_NAME     = "cloudops-api"
      DDB_TABLE_NAME       = aws_dynamodb_table.main.name
      AWS_REGION           = var.aws_region
      EVENT_BUS_NAME       = aws_cloudwatch_event_bus.main.name
      EVENT_SOURCE         = "cloudops.release-intelligence"
      S3_RAW_EVENTS_BUCKET = aws_s3_bucket.raw.bucket
      APP_CORS_ORIGINS     = "https://${aws_cloudfront_distribution.web.domain_name}"
    }
  }

  tracing_config {
    mode = "Active"
  }
}

resource "aws_lambda_function" "worker" {
  count         = local.deploy_compute ? 1 : 0
  function_name = "${local.name_prefix}-worker"
  role          = aws_iam_role.lambda_worker.arn
  package_type  = "Image"
  image_uri     = local.worker_image_uri
  timeout       = 60
  memory_size   = 512
  architectures = ["x86_64"]

  image_config {
    command = ["worker"]
  }

  environment {
    variables = {
      APP_ENV        = "prod"
      APP_STORE      = "dynamodb"
      APP_SEED_LOCAL = "false"
      APP_LOG_LEVEL  = "info"
      DDB_TABLE_NAME = aws_dynamodb_table.main.name
      AWS_REGION     = var.aws_region
    }
  }

  tracing_config {
    mode = "Active"
  }
}

resource "aws_lambda_event_source_mapping" "worker" {
  count                   = local.deploy_compute ? 1 : 0
  event_source_arn        = aws_sqs_queue.analysis.arn
  function_name           = aws_lambda_function.worker[0].arn
  batch_size              = 5
  function_response_types = ["ReportBatchItemFailures"]
}

# One Lambda image contains both binaries. image_config.command selects
# `api` or `worker`. Do not publish a second repository for the worker.
