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
  description = "Email for $5 warning and $10 stronger AWS Budget notifications. Provide via gitignored terraform.tfvars or TF_VAR_budget_notification_email. Not a secret; do not commit the real address."

  validation {
    condition     = can(regex("^[^@\\s]+@[^@\\s]+\\.[^@\\s]+$", var.budget_notification_email))
    error_message = "budget_notification_email must be a non-empty email address."
  }
}

variable "cors_additional_origins" {
  type        = list(string)
  description = "Extra browser origins besides the CloudFront domain."
  default     = []
}

variable "terraform_state_bucket_arn" {
  type        = string
  description = "ARN of the remote state bucket from infra/bootstrap. Empty skips GitHub state-access policies (set after bootstrap apply)."
  default     = ""
}

variable "terraform_state_bucket_name" {
  type        = string
  description = "Name of the remote state bucket (documentation / future backend wiring). Optional until bootstrap apply."
  default     = ""
}

variable "terraform_state_key" {
  type        = string
  description = "S3 object key for the main stack state (must match backend key)."
  default     = "cloudops-release-intelligence/prod/terraform.tfstate"
}
