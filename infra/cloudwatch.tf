resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/${local.name_prefix}-api"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "worker" {
  name              = "/aws/lambda/${local.name_prefix}-worker"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "apigw" {
  name              = "/aws/apigateway/${local.name_prefix}-http"
  retention_in_days = 14
}

resource "aws_cloudwatch_metric_alarm" "dlq" {
  alarm_name          = "${local.name_prefix}-analysis-dlq"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Maximum"
  threshold           = 1
  alarm_description   = "Analysis DLQ received a message. Investigate poison events."
  treat_missing_data  = "notBreaching"
  dimensions = {
    QueueName = aws_sqs_queue.dlq.name
  }
}
