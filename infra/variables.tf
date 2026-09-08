variable "aws_region" {
  type        = string
  description = "Single-region portfolio environment."
  default     = "eu-west-1"
}

variable "project" {
  type    = string
  default = "cloudops"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "github_owner" {
  type        = string
  description = "GitHub org or user that owns the repository."
  default     = "Adellaghmari"
}

variable "github_repo" {
  type        = string
  description = "Repository name without owner."
  default     = "cloudops-release-intelligence"
}

variable "create_github_oidc_provider" {
  type        = bool
  description = "Set false if the account already has token.actions.githubusercontent.com."
  default     = true
}

variable "api_image_uri" {
  type        = string
  description = "ECR image URI including digest (sha256:...). Empty skips Lambda/API until the first image exists."
  default     = ""
}

variable "worker_image_uri" {
  type        = string
  description = "Worker image URI. Defaults to api_image_uri when empty."
  default     = ""
}

variable "budget_notification_email" {
  type        = string
  description = "Optional email for $5/$10 budget alerts. Leave empty to create the budget without subscribers."
  default     = ""
}

variable "cors_additional_origins" {
  type        = list(string)
  description = "Extra browser origins besides the CloudFront domain."
  default     = []
}
