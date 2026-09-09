locals {
  name_prefix = "${var.project}-${var.environment}"
  common_tags = {
    Project     = var.project
    Environment = var.environment
    ManagedBy   = "terraform"
  }

  worker_image_uri = var.worker_image_uri != "" ? var.worker_image_uri : var.api_image_uri
  deploy_compute   = var.api_image_uri != ""

  github_repo_full      = "${var.github_owner}/${var.github_repo}"
  github_owner_id       = "179922674"
  github_repo_id        = "1361976389"
  github_repo_full_oidc = "${var.github_owner}@${local.github_owner_id}/${var.github_repo}@${local.github_repo_id}"
}
