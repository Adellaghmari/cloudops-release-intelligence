resource "aws_sqs_queue" "dlq" {
  name                      = "${local.name_prefix}-analysis-dlq"
  message_retention_seconds = 1209600
}

resource "aws_sqs_queue" "analysis" {
  name                       = "${local.name_prefix}-analysis"
  visibility_timeout_seconds = 90
  message_retention_seconds  = 345600
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 3
  })
}
