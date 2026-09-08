output "region" {
  value = var.aws_region
}

output "dynamodb_table_name" {
  value = aws_dynamodb_table.main.name
}

output "ecr_repository_url" {
  value = aws_ecr_repository.api.repository_url
}

output "raw_evidence_bucket" {
  value = aws_s3_bucket.raw.bucket
}

output "web_bucket" {
  value = aws_s3_bucket.web.bucket
}

output "event_bus_name" {
  value = aws_cloudwatch_event_bus.main.name
}

output "analysis_queue_url" {
  value = aws_sqs_queue.analysis.id
}

output "analysis_dlq_url" {
  value = aws_sqs_queue.dlq.id
}

output "cloudfront_domain" {
  value = aws_cloudfront_distribution.web.domain_name
}

output "cloudfront_url" {
  value = "https://${aws_cloudfront_distribution.web.domain_name}"
}

output "api_endpoint" {
  value = local.deploy_compute ? aws_apigatewayv2_api.http[0].api_endpoint : null
}

output "github_deploy_role_arn" {
  value = aws_iam_role.github_deploy.arn
}

output "github_plan_role_arn" {
  value = aws_iam_role.github_plan.arn
}

output "lambda_api_name" {
  value = local.deploy_compute ? aws_lambda_function.api[0].function_name : null
}

output "lambda_worker_name" {
  value = local.deploy_compute ? aws_lambda_function.worker[0].function_name : null
}
