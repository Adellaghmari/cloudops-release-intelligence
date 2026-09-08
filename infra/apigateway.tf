resource "aws_apigatewayv2_api" "http" {
  count         = local.deploy_compute ? 1 : 0
  name          = "${local.name_prefix}-http"
  protocol_type = "HTTP"

  cors_configuration {
    allow_origins = concat(["https://${aws_cloudfront_distribution.web.domain_name}"], var.cors_additional_origins)
    allow_methods = ["GET", "POST", "OPTIONS"]
    allow_headers = ["content-type", "x-request-id", "x-correlation-id"]
    max_age       = 3600
  }
}

resource "aws_apigatewayv2_integration" "api" {
  count                  = local.deploy_compute ? 1 : 0
  api_id                 = aws_apigatewayv2_api.http[0].id
  integration_type       = "AWS_PROXY"
  integration_uri        = aws_lambda_function.api[0].invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "proxy" {
  count     = local.deploy_compute ? 1 : 0
  api_id    = aws_apigatewayv2_api.http[0].id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.api[0].id}"
}

resource "aws_apigatewayv2_stage" "prod" {
  count       = local.deploy_compute ? 1 : 0
  api_id      = aws_apigatewayv2_api.http[0].id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = 40
    throttling_rate_limit  = 20
  }

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.apigw.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      status         = "$context.status"
      integration    = "$context.integrationErrorMessage"
      responseLength = "$context.responseLength"
    })
  }
}

resource "aws_lambda_permission" "apigw" {
  count         = local.deploy_compute ? 1 : 0
  statement_id  = "AllowAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api[0].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http[0].execution_arn}/*/*"
}
