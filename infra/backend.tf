# First controlled environment uses local state in this directory.
# terraform.tfstate is gitignored. See docs/TERRAFORM_STATE.md.
# A remote S3 backend is intentionally not created here to avoid
# bootstrapping a bucket that Terraform itself must manage.
terraform {
  backend "local" {
    path = "terraform.tfstate"
  }
}
