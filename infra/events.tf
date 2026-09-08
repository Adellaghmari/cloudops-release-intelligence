resource "aws_cloudwatch_event_bus" "main" {
  name = "${local.name_prefix}-release-events"
}

resource "aws_cloudwatch_event_rule" "analysis" {
  name           = "${local.name_prefix}-analysis"
  event_bus_name = aws_cloudwatch_event_bus.main.name
  event_pattern = jsonencode({
    source = ["cloudops.release-intelligence"]
  })
}

resource "aws_cloudwatch_event_target" "analysis_queue" {
  rule           = aws_cloudwatch_event_rule.analysis.name
  event_bus_name = aws_cloudwatch_event_bus.main.name
  arn            = aws_sqs_queue.analysis.arn
}

resource "aws_sqs_queue_policy" "analysis_from_events" {
  queue_url = aws_sqs_queue.analysis.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "AllowEventBridge"
      Effect    = "Allow"
      Principal = { Service = "events.amazonaws.com" }
      Action    = "sqs:SendMessage"
      Resource  = aws_sqs_queue.analysis.arn
      Condition = {
        ArnEquals = { "aws:SourceArn" = aws_cloudwatch_event_rule.analysis.arn }
      }
    }]
  })
}
