resource "aws_cloudfront_distribution" "web" {
  enabled             = true
  comment             = "${local.name_prefix} recruiter demo"
  default_root_object = "index.html"
  price_class         = "PriceClass_100"
  wait_for_deployment = false

  origin {
    domain_name              = aws_s3_bucket.web.bucket_regional_domain_name
    origin_id                = "s3-web"
    origin_access_control_id = aws_cloudfront_origin_access_control.web.id
  }

  default_cache_behavior {
    target_origin_id       = "s3-web"
    viewer_protocol_policy = "redirect-to-https"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    compress               = true
    forwarded_values {
      query_string = false
      cookies { forward = "none" }
    }
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # Default *.cloudfront.net certificate: AWS reports MinimumProtocolVersion=TLSv1
  # regardless of TLSv1.2_2021. Do not fight that with perpetual drift.
  # Custom domain + ACM (us-east-1) is required for a modern viewer TLS policy.
  viewer_certificate {
    cloudfront_default_certificate = true
    minimum_protocol_version       = "TLSv1"
  }

  custom_error_response {
    error_code         = 403
    response_code      = 200
    response_page_path = "/index.html"
  }

  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
  }
}
